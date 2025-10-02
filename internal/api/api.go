// Package api provides an HTTP handler that can route requests to either an
// HTTP stub server or a gRPC stub server based on the request properties.
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kogxi/stub-server/internal/grpcstub"
	"github.com/kogxi/stub-server/internal/httpstub"
	"google.golang.org/grpc"
)

// Server represents a server that can handle both HTTP and gRPC requests.
type Server struct {
	grpcServer  *grpc.Server
	httpHandler http.Handler
}

var _ http.Handler = &Server{}

// WithProto configures the server to handle gRPC requests using the provided
// proto and stub directories.
func (s *Server) WithProto(protoDir string, stubDir string) (*Server, error) {
	server, err := grpcstub.NewServer(protoDir, stubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	s.grpcServer = server

	return s, nil
}

// WithHTTP configures the server to handle HTTP requests using the provided
// HTTP stubs directory.
func (s *Server) WithHTTP(httpStubs string) (*Server, error) {
	handler, err := httpstub.NewHandler(httpStubs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP handler: %w", err)
	}

	s.httpHandler = handler

	return s, nil
}

// NewHandler creates a new Server instance with no handlers configured.
func NewHandler() *Server {
	mux := http.NewServeMux()

	s := &Server{}

	mux.Handle("/", s)

	return s
}

// ServeHTTP routes incoming HTTP requests to either the gRPC server or the HTTP
// handler based on the request properties. If the request is a gRPC
// request (HTTP/2 with "application/grpc" content type), it is forwarded to the
// gRPC server. Otherwise, it is handled by the HTTP handler.
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
