package main

import (
	"context"
	"flag"
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

func allowh2c(next http.Handler) http.Handler {
	h2server := &http2.Server{IdleTimeout: time.Second * 60}
	return h2c.NewHandler(next, h2server)
}

var (
	address      = flag.String("address", ":50051", "Port to listen on")
	protoDir     = flag.String("proto", "/Users/mathias/Documents/GitHub/stub-server/examples/protos", "Path to proto files")
	protoStubDir = flag.String("stubs", "/Users/mathias/Documents/Github/stub-server/examples/stubs", "Path to gRPC stubs")
	httpStubDir  = flag.String("http", "/Users/mathias/Documents/Github/mock/httpstubs", "Path to HTTP stubs")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	handler := api.NewHandler()

	handler, err := handler.WithHTTP(*httpStubDir)
	if err != nil {
		slog.Error("failed to create HTTP handler", slog.String("error", err.Error()))
	}

	handler, err = handler.WithProto(*protoDir, *protoStubDir)
	if err != nil {
		slog.Error("failed to create gRPC handler", slog.String("error", err.Error()))
	}

	srv := &http.Server{Addr: *address, Handler: allowh2c(handler)}

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
		slog.Info("server stopped")
	}
}
