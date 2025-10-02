// Package main implements a stub server that can handle both HTTP and gRPC requests based on predefined stubs.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kogxi/stub-server/internal/api"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

func allowH2c(next http.Handler) http.Handler {
	h2server := &http2.Server{IdleTimeout: time.Second * 60}
	return h2c.NewHandler(next, h2server)
}

var (
	address      = flag.String("address", ":50051", "Port to listen on")
	protoDir     = flag.String("proto", "", "Path to proto files")
	protoStubDir = flag.String("stubs", "", "Path to gRPC stubs")
	httpStubDir  = flag.String("http", "", "Path to HTTP stubs")
)

func newHandler(httpStubDir string, protoDir string, protoStubDir string) (http.Handler, error) {
	handler := api.NewHandler()
	var err error
	if httpStubDir != "" {
		handler, err = handler.WithHTTP(httpStubDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP handler: %w", err)
		}
	}

	if protoDir != "" && protoStubDir != "" {
		handler, err = handler.WithProto(protoDir, protoStubDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC handler: %w", err)
		}
	}

	return allowH2c(handler), nil
}

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	handler, err := newHandler(*httpStubDir, *protoDir, *protoStubDir)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := &http.Server{Addr: *address, Handler: handler}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		slog.Info("listening", slog.String("address", srv.Addr))
		return srv.ListenAndServe()
	})

	eg.Go(func() error {
		<-ctx.Done()
		slog.Info("closing server")
		return srv.Close()
	})

	err = eg.Wait()
	if err != nil {
		slog.ErrorContext(ctx, "server stopped", slog.String("error", err.Error()))
	}
}
