package httpstub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	Header http.Header    `json:"header"`
	Body   map[string]any `json:"body"`
	Status int            `json:"status"`
}

type HTTPStub struct {
	Path     string   `json:"path"`
	Method   string   `json:"method"`
	Response Response `json:"response"`
}

func (s *HTTPStub) validate() error {
	if s.Path == "" {
		return fmt.Errorf(`"path" field is required`)
	}

	return nil
}

func loadStubs(dir string) ([]HTTPStub, error) {
	stubs := make([]HTTPStub, 0)
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
			var stub HTTPStub
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
		return nil, fmt.Errorf("error reading dir %v: %w", dir, err)
	}
	return stubs, nil
}
