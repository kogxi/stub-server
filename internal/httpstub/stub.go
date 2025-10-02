package httpstub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Response represents an HTTP response defined in a stub.
type Response struct {
	Header http.Header    `json:"header"`
	Body   map[string]any `json:"body"`
	Status int            `json:"status"`
}

// HTTPStub represents a predefined HTTP stub.
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

func loadFile(path string) (s HTTPStub, err error) {
	f, err := os.Open(path)
	if err != nil {
		return HTTPStub{}, fmt.Errorf("failed to open file: %v: %w", path, err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close file: %w", closeErr))
		}
	}()

	stubJSON, err := io.ReadAll(f)
	if err != nil {
		return HTTPStub{}, fmt.Errorf("failed to read file %v: %w", path, err)
	}
	var stub HTTPStub
	err = json.Unmarshal(stubJSON, &stub)
	if err != nil {
		return HTTPStub{}, fmt.Errorf("failed to unmarshal stub %v: %w", path, err)
	}

	err = stub.validate()
	if err != nil {
		return HTTPStub{}, fmt.Errorf("stub validation failed: %w", err)
	}
	return stub, nil
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

			stub, err := loadFile(path)
			if err != nil {
				return fmt.Errorf("failed to load stub from %v: %w", path, err)
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
