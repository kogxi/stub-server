package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	helloworldpb "google.golang.org/grpc/examples/helloworld/helloworld"
	routeguide "google.golang.org/grpc/examples/route_guide/routeguide"
)

// go:generate

func TestHTTPServer(t *testing.T) {
	// handler, err := newHandler("../examples/httpstubs", "", "")
	// require.NoError(t, err)

	// server := httptest.NewServer(handler)
	// defer server.Close()

	// url, err := url.JoinPath(server.URL, "helloworld")
	// require.NoError(t, err)
	url := "http://localhost:50051/helloworld"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
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

	// handler, err := newHandler("", "../examples/protos", "../examples/protostubs")
	// require.NoError(t, err)

	// server := httptest.NewServer(handler)
	// defer server.Close()
	// url, _ := strings.CutPrefix(server.URL, "http://")
	url := "localhost:50051"
	c, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	{
		client := helloworldpb.NewGreeterClient(c)
		reply, err := client.SayHello(context.TODO(), &helloworldpb.HelloRequest{
			Name: "Jane",
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello from proto stub", reply.Message)
	}

	{
		routeguide.NewRouteGuideClient(c)
		client := routeguide.NewRouteGuideClient(c)
		stream, err := client.ListFeatures(context.TODO(), &routeguide.Rectangle{
			Lo: &routeguide.Point{Latitude: 400000000, Longitude: -750000000},
			Hi: &routeguide.Point{Latitude: 420000000, Longitude: -730000000},
		})
		require.NoError(t, err)

		results := make([]*routeguide.Feature, 0, 3)
		for {
			feature, err := stream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			results = append(results, feature)
		}

		require.Len(t, results, 3)

		assert.Equal(t, "#1", results[0].Name)
		assert.Equal(t, int32(409146138), results[0].Location.Latitude)
		assert.Equal(t, int32(-746188906), results[0].Location.Longitude)

		assert.Equal(t, "#2", results[1].Name)
		assert.Equal(t, int32(413628156), results[1].Location.Latitude)
		assert.Equal(t, int32(-749015468), results[1].Location.Longitude)

		assert.Equal(t, "#3", results[2].Name)
		assert.Equal(t, int32(419999544), results[2].Location.Latitude)
		assert.Equal(t, int32(733555590), results[2].Location.Longitude)
	}

	{
		routeguide.NewRouteGuideClient(c)
		client := routeguide.NewRouteGuideClient(c)
		stream, err := client.RecordRoute(context.TODO())
		require.NoError(t, err)

		err = stream.Send(&routeguide.Point{Latitude: 20, Longitude: -40})
		require.NoError(t, err)
		err = stream.Send(&routeguide.Point{Latitude: 10, Longitude: -500})
		require.NoError(t, err)
		err = stream.Send(&routeguide.Point{Latitude: 124234, Longitude: -12142352})
		require.NoError(t, err)

		summary, err := stream.CloseAndRecv()
		require.NoError(t, err)
		assert.Equal(t, int32(10), summary.PointCount)
		assert.Equal(t, int32(5), summary.FeatureCount)
		assert.Equal(t, int32(1000), summary.Distance)
		assert.Equal(t, int32(120), summary.ElapsedTime)
	}
}
