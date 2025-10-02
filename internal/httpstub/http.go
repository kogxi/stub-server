// Package httpstub provides an HTTP handler that serves predefined HTTP responses.
package httpstub

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// StubHandler is an HTTP handler that serves predefined HTTP stubs.
type StubHandler struct {
	stubs map[string]HTTPStub
}

var _ http.Handler = &StubHandler{}

// NewHandler creates a new StubHandler by loading HTTP stubs from the specified directory.
func NewHandler(stubDir string) (*StubHandler, error) {
	stubs, err := loadStubs(stubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load HTTP stubs from %v: %w ", stubDir, err)
	}

	sm := make(map[string]HTTPStub, len(stubs))

	for _, s := range stubs {
		sm[s.Path] = s
	}

	return &StubHandler{
		stubs: sm,
	}, nil
}

// ServeHTTP serves HTTP requests based on the loaded stubs.
func (s *StubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stub, ok := s.stubs[r.URL.Path]
	if !ok {
		slog.Error("unknown stub", slog.String("path", r.URL.Path))
		http.Error(w, "unknown stub", http.StatusNotFound)
		return
	}

	if stub.Method != "" && r.Method != stub.Method {
		slog.ErrorContext(r.Context(), "method not allowed", slog.String("expected", stub.Method), slog.String("got", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Body != nil {
		// Read the full body to a noop Writer
		_, err := io.Copy(newWriter(), r.Body)
		if err != nil {
			slog.ErrorContext(r.Context(), "error reading body", slog.String("error", err.Error()))
			http.Error(w, "error reading request body", http.StatusInternalServerError)
			return
		}
	}

	for k, val := range stub.Response.Header {
		for _, v := range val {
			w.Header().Set(k, v)
		}
	}

	w.WriteHeader(stub.Response.Status)

	if stub.Response.Body == nil {
		return
	}

	err := json.NewEncoder(w).Encode(stub.Response.Body)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
