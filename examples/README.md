# Examples

The proto examples messages are copied from the [grpc-go](https://github.com/grpc/grpc-go) repository.

## HTTP stub server

To start the HTTP stub server one needs to specify the path to the HTTP stub dir.
`./stub-server --http "./examples/httpstubs"`

## gRPC stub server

To start the gRPC stub server one needs to specify the path to the gRPC stub directory and the path to the proto files. E.g., `./stub-server --proto "./examples/protos" --stubs "./examples/protostubs"`