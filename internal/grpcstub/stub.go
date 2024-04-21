package grpcstub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Repository interface {
	Add(stub ProtoStub)
	Get(service string, method string, in json.RawMessage) (Output, bool)
}

type GRPCService struct {
	stubs      Repository
	sdMap      map[string]protoreflect.ServiceDescriptor
	grpcServer *grpc.Server
}

func New(srv *grpc.Server, r Repository) GRPCService {
	s := GRPCService{
		stubs:      r,
		sdMap:      map[string]protoreflect.ServiceDescriptor{},
		grpcServer: srv,
	}
	return s
}

func (s *GRPCService) loadStubs(dir string) error {
	stubs, err := Load(dir)
	if err != nil {
		return fmt.Errorf("failed to load stubs: %w", err)
	}

	for _, stub := range stubs {
		if s.sdMap[stub.Service] == nil {
			return fmt.Errorf(`no service "%v" registered`, stub.Service)
		}
		s.stubs.Add(stub)
	}

	return nil
}

func (s *GRPCService) LoadSpecs(protoDir string) error {
	err := filepath.WalkDir(protoDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Ext(path) != ".proto" {
				return nil
			}
			n, err := filepath.Rel(protoDir, path)
			if err != nil {
				return err
			}

			return s.registerProto(protoDir, n)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load specs: %w", err)
	}

	reflection.Register(s.grpcServer)
	return nil
}

func (s *GRPCService) Handler(srv any, ctx context.Context, deccode func(any) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	stream := grpc.ServerTransportStreamFromContext(ctx)
	arr := strings.Split(stream.Method(), "/")
	serviceName := arr[1]
	methodName := arr[2]
	service, ok := s.sdMap[serviceName]
	if !ok {
		slog.ErrorContext(ctx, "No stub found", slog.String("service", serviceName))
		return nil, status.Error(codes.NotFound, "service "+serviceName+" not found")
	}

	method := service.Methods().ByName(protoreflect.Name(methodName))
	if method == nil {
		return nil, status.Error(codes.NotFound, "method "+methodName+" not found")
	}
	input := dynamicpb.NewMessage(method.Input())

	err := deccode(input)
	if err != nil {
		slog.ErrorContext(ctx, "failed to decode input message", slog.String("error", err.Error()))
	}

	jsonInput, err := protojson.Marshal(input)
	if err != nil {
		slog.Error("failed to marshall input", slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, "failed to marshall input")
	}

	resp, ok := s.stubs.Get(serviceName, methodName, jsonInput)
	if !ok {
		return nil, status.Error(codes.NotFound, "no stub found")
	}

	if resp.Data != nil {
		output := dynamicpb.NewMessage(method.Output())

		err = protojson.Unmarshal(resp.Data, output)
		if err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal response", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to unmarshal response")
		}

		return output, nil
	}

	if resp.Code != nil {
		return nil, status.Error(*resp.Code, resp.Error)
	}

	return nil, status.Error(codes.Unimplemented, resp.Error)
}

func (s *GRPCService) StreamHandler(srv any, stream grpc.ServerStream) error {
	slog.Info("received streaming call")

	return status.Error(codes.Unimplemented, "")
}

func (s *GRPCService) registerProto(protoDir string, protoFileName string) error {
	// Skip the file if it is already registered
	if _, err := protoregistry.GlobalFiles.FindFileByPath(protoFileName); err == nil {
		return nil
	}

	fh, err := os.Open(path.Join(protoDir, protoFileName))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer fh.Close()

	handler := reporter.NewHandler(nil)
	node, err := parser.Parse(protoFileName, fh, handler)
	if err != nil {
		return fmt.Errorf("parse proto: %w", err)
	}

	res, err := parser.ResultFromAST(node, true, handler)
	if err != nil {
		return fmt.Errorf("convert from AST: %w", err)
	}

	// recursively register dependencies
	for _, d := range res.FileDescriptorProto().Dependency {
		err = s.registerProto(protoDir, d)
		if err != nil {
			return err
		}
	}

	fd, err := protodesc.NewFile(res.FileDescriptorProto(), protoregistry.GlobalFiles)
	if err != nil {
		return fmt.Errorf("convert to FileDescriptor: %w", err)
	}

	_, err = protoregistry.GlobalTypes.FindMessageByName(fd.FullName())
	if err == protoregistry.NotFound {
		if err := protoregistry.GlobalFiles.RegisterFile(fd); err != nil {
			return fmt.Errorf("register file: %w", err)
		}
	}

	for i := 0; i < fd.Messages().Len(); i++ {
		msg := fd.Messages().Get(i)
		if err := protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(msg)); err != nil {
			return fmt.Errorf("register message %q: %w", msg.FullName(), err)
		}
	}
	for i := 0; i < fd.Extensions().Len(); i++ {
		ext := fd.Extensions().Get(i)
		if err := protoregistry.GlobalTypes.RegisterExtension(dynamicpb.NewExtensionType(ext)); err != nil {
			return fmt.Errorf("register extension %q: %w", ext.FullName(), err)
		}
	}

	for svcNum := 0; svcNum < fd.Services().Len(); svcNum++ {
		svc := fd.Services().Get(svcNum)
		serviceName := string(svc.FullName())
		s.sdMap[serviceName] = svc
		gsd := grpc.ServiceDesc{ServiceName: serviceName, HandlerType: (*interface{})(nil)}
		for methodNum := 0; methodNum < svc.Methods().Len(); methodNum++ {
			m := svc.Methods().Get(methodNum)
			gsd.Methods = append(gsd.Methods, grpc.MethodDesc{MethodName: string(m.Name()), Handler: s.Handler})
		}
		s.grpcServer.RegisterService(&gsd, s)
	}

	return nil
}
