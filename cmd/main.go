// Package main implements a stub server that can handle both HTTP and gRPC requests based on predefined stubs.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/kogxi/stub-server/internal/handler"
	"golang.org/x/sync/errgroup"
)

var (
	address      = flag.String("address", ":50051", "Port to listen on")
	protoDir     = flag.String("proto", "", "Path to proto files")
	protoStubDir = flag.String("stubs", "", "Path to gRPC stubs")
	httpStubDir  = flag.String("http", "", "Path to HTTP stubs")
	tlsCert      = flag.String("cert", "", "Path to TLS certificate")
	tlsCertKey   = flag.String("key", "", "Path to TLS certificate key")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	handler, err := handler.New(*httpStubDir, *protoDir, *protoStubDir)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	tls, err := loadTLS(*tlsCert, *tlsCertKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load TLS config", slog.String("error", err.Error()))
	}

	srv := &http.Server{
		Addr:      *address,
		Handler:   handler,
		TLSConfig: tls,
	}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		slog.Info("Listening", slog.String("address", srv.Addr))
		return srv.ListenAndServe()
	})

	eg.Go(func() error {
		<-ctx.Done()
		slog.Info("Closing server")
		return srv.Close()
	})

	err = eg.Wait()
	if err != nil {
		slog.ErrorContext(ctx, "Server stopped", slog.String("error", err.Error()))
	}
}

func loadTLS(certFile string, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load cert: %w", err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	return cfg, nil
}
