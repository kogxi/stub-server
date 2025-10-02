// Package grpcstub provides functionality to create and manage a gRPC server
// that can load proto definitions and corresponding stub definitions.
package grpcstub

import (
	"fmt"

	"google.golang.org/grpc"
)

// NewServer creates a new gRPC server, loads proto definitions from the
// specified protoDir, and loads stub definitions from the specified protoStubDir.
func NewServer(protoDir string, protoStubDir string) (*grpc.Server, error) {
	server := grpc.NewServer()

	manager := New(server, NewStorage())
	err := manager.LoadSpecs(protoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load protos from %v: %w", protoDir, err)
	}

	err = manager.loadStubs(protoStubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load stubs from %v: %w", protoStubDir, err)
	}

	return server, nil
}
