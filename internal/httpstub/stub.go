package httpstub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// Stub represents a predefined HTTP stub.
type Stub struct {
	Path     string   `json:"path"`
	Method   string   `json:"method"`
	Response Response `json:"response"`
}

// Response represents an HTTP response defined in a stub.
type Response struct {
	Header http.Header    `json:"header"`
	Body   map[string]any `json:"body"`
	Status int            `json:"status"`
}

func (s *Stub) validate() error {
	if s.Path == "" {
		return errors.New(`"path" field is required`)
	}

	if s.Response.Status == 0 {
		return errors.New(`"status" field is required`)
	}

	return nil
}

func loadStubs(dir string, storage *Storage) error {
	if err := filepath.WalkDir(dir, walk(storage)); err != nil {
		return fmt.Errorf("read stubs from dir %v: %w", dir, err)
	}
	return nil
}

func walk(storage *Storage) fs.WalkDirFunc {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if filepath.Ext(path) != ".json" {
				return nil
			}

			stub, err := loadFile(path)
			if err != nil {
				return fmt.Errorf("load stub from %v: %w", path, err)
			}

			storage.Add(stub)
		}
		return nil
	}
}

func loadFile(path string) (s Stub, err error) {
	f, err := os.Open(path)
	if err != nil {
		return Stub{}, fmt.Errorf("open file: %v: %w", path, err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close file: %w", closeErr))
		}
	}()

	var stub Stub
	if err := json.NewDecoder(f).Decode(&stub); err != nil {
		return Stub{}, fmt.Errorf("unmarshal stub %v: %w", path, err)
	}

	if err = stub.validate(); err != nil {
		return Stub{}, fmt.Errorf("stub validation %v: %w", path, err)
	}
	return stub, nil
}
