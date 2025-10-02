# Stub Server
The Stub Server is a simple stubs server for HTTP(s) and gRPC.
No need to invoke the proto compiler - the proto files are loaded dynamically.
It supports streaming and unary RPC calls for gRPC.
The HTTP and gRPC server run on the same port.

# Usage

## Installation
`go install github.com/kogxi/stub-server/cmd`

## Parameters
| Name | Usage | Required | Default |
|-|-|-|-|
| address | Address to listen on | `false`| `:58001` |
| cert | Path to the `cert` file | `false`| - |
| key | Path to the `key` file | `false`| - |
| proto | Directory containing the `.proto` files| `false`| - |
| stubs | Directory containing the `.json` gRPC stub files| `true`| - |
| http | Directory containing the `.json` HTTP stub files| `true`| - |

## HTTP stub server

The HTTP(s) stub requires only the `path` fields. By default the stub server returns the HTTP status code 200 (OK).

### Minimal example
```JSON
{
    "path": "/helloworld"
}
```

### Fully specified example
```JSON
{
    "path": "/helloworld",
    "method": "GET",
    "response": {
        "header": {
          "Content-Type":  ["application/json"]
        },
        "body": {"message": "Hello from http stub"},
        "status": 201
    }
}
```

To start the HTTP stub server one needs to specify the path to the HTTP stub dir.
`./stub-server --http ./examples/httpstubs`

## gRPC stub server

The gRPC stub requires the `service`, `method` and `outputs` fields.

### Unary success example
```JSON
{
    "service": "helloworld.Greeter",
    "method": "SayHello",
    "output": {
        "data": {
            "message": "Hello from proto stub"
        }
    }
}
```

### Unary Error example 
```JSON
{
    "service": "helloworld.Greeter",
    "method": "SayHello",
    "output": {
        "error": {
            "code": 3,
            "message": "Invalid request"
        }
    }
}
```

To start the gRPC stub server one needs to specify the path to the gRPC stub directory and the path to the proto files. E.g., `./stub-server --proto ./examples/protos --stubs ./examples/protostubs`

To start HTTP and gRPC server you can combine the two commands:
`./stub-server --proto ./examples/protos" --stubs "./examples/protostubs --http ./examples/httpstubs`
