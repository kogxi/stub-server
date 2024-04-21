package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	helloworldpb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// go:generate

func TestHTTPServer(t *testing.T) {
	handler, err := newHandler("../examples/httpstubs", "", "")
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	defer server.Close()

	url, err := url.JoinPath(server.URL, "helloworld")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"message": "Hello from http stub"}`, string(body))
}

func TestGrpcServer(t *testing.T) {
	err := os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "ignore")
	require.NoError(t, err)

	handler, err := newHandler("", "../examples/protos", "../examples/protostubs")
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	defer server.Close()
	url, _ := strings.CutPrefix(server.URL, "http://")
	c, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := helloworldpb.NewGreeterClient(c)
	reply, err := client.SayHello(context.TODO(), &helloworldpb.HelloRequest{
		Name: "Jane",
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello from proto stub", reply.Message)
}
