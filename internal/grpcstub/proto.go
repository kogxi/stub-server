package grpcstub

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
)

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

func Load(dir string) ([]ProtoStub, error) {
	stubs := make([]ProtoStub, 0)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Ext(path) != ".json" {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %v: %w", path, err)
			}
			defer file.Close()

			stubJSON, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read file %v: %w", d.Name(), err)
			}
			var stub ProtoStub
			err = json.Unmarshal(stubJSON, &stub)
			if err != nil {
				return fmt.Errorf("failed to unmarshal stub %v: %w", d.Name(), err)
			}

			err = stub.validate()
			if err != nil {
				return fmt.Errorf("stub validation failed: %w", err)
			}
			stubs = append(stubs, stub)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(`error reading dir "%v": %w`, dir, err)
	}
	return stubs, nil
}
