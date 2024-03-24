package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kogxi/stub-server/internal/grpcstub"
	"github.com/kogxi/stub-server/internal/httpstub"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer  *grpc.Server
	httpHandler http.Handler
}

var _ http.Handler = &Server{}

func (s *Server) WithProto(protoDir string, stubDir string) (*Server, error) {
	server, err := grpcstub.NewServer(protoDir, stubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	s.grpcServer = server

	return s, nil
}

func (s *Server) WithHTTP(httpStubs string) (*Server, error) {
	handler, err := httpstub.NewHandler(httpStubs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP handler: %w", err)
	}

	s.httpHandler = handler

	return s, nil
}

func NewHandler() *Server {
	mux := http.NewServeMux()

	s := &Server{}

	mux.Handle("/", s)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.HasPrefix(
		r.Header.Get("Content-Type"), "application/grpc") {
		if s.grpcServer == nil {
			http.Error(w, "No gRPC stub server configured", http.StatusNotImplemented)
			return
		}
		s.grpcServer.ServeHTTP(w, r)
		return
	}

	if s.httpHandler == nil {
		http.Error(w, "No HTTP stub server configured", http.StatusNotImplemented)
		return
	}

	s.httpHandler.ServeHTTP(w, r)
}
