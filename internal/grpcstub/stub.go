package grpcstub

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
)

// Stream represents a stream of gRPC responses.
type Stream struct {
	Data  []json.RawMessage `json:"data"`
	Error string            `json:"error"`
	Code  *codes.Code       `json:"code,omitempty"`
	Delay int               `json:"delay,omitempty"`
}

func (s *Stream) validate() error {
	if s.Code == nil && len(s.Data) == 0 && s.Error == "" {
		return fmt.Errorf(`stream can't be empty`)
	}
	return nil
}

// Output represents the output of a gRPC method, which can be a single response or a stream.
type Output struct {
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
	Code   *codes.Code     `json:"code,omitempty"`
	Stream *Stream         `json:"stream"`
}

func (o *Output) validate() error {
	if o.Code == nil && o.Data == nil && o.Error == "" && o.Stream == nil {
		return fmt.Errorf(`output can't be empty`)
	}

	if o.Stream != nil {
		return o.Stream.validate()
	}
	return nil
}

// ProtoStub represents a gRPC stub definition.
type ProtoStub struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Matcher string `json:"matcher"`
	Output  Output `json:"output"`
}

func (s *ProtoStub) validate() error {
	if s.Service == "" {
		return fmt.Errorf(`"service" field is required`)
	}
	if s.Method == "" {
		return fmt.Errorf(`"method" field is required`)
	}

	return s.Output.validate()
}

func (s *GRPCService) loadStubs(dir string) error {
	stubs, err := load(dir)
	if err != nil {
		return fmt.Errorf("load stubs: %w", err)
	}

	for _, stub := range stubs {
		if s.sdMap[stub.Service] == nil {
			return fmt.Errorf(`no service "%v" registered`, stub.Service)
		}
		s.stubs.Add(stub)
	}

	return nil
}

func load(dir string) ([]ProtoStub, error) {
	stubs := make([]ProtoStub, 0)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Ext(path) != ".json" {
				return nil
			}

			stub, err := loadFile(path)
			if err != nil {
				return fmt.Errorf("load stub from file %v: %w", path, err)
			}

			stubs = append(stubs, stub)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(`read dir "%v": %w`, dir, err)
	}
	return stubs, nil
}

func loadFile(path string) (s ProtoStub, err error) {
	f, err := os.Open(path)
	if err != nil {
		return ProtoStub{}, fmt.Errorf("open file: %v: %w", path, err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close file: %w", closeErr))
		}
	}()

	var stub ProtoStub
	if err := json.NewDecoder(f).Decode(&stub); err != nil {
		return ProtoStub{}, fmt.Errorf("unmarshal stub %v: %w", path, err)
	}

	if err := stub.validate(); err != nil {
		return ProtoStub{}, fmt.Errorf("stub validation %v: %w", path, err)
	}
	return stub, nil
}
