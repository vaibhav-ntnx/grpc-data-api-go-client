# Nutanix CBT gRPC Client

## Summary
A simple go client to call Nutanix Changed Block Tracking (CBT) APIs over gRPC.
The client first calls discover_cluster REST API to fetch auth token which is set as NTNX_IGW_SESSION
into the CBT gRPC call, used by the backend for authn.

## Dependencies

This project uses Go modules for dependency management. Key dependencies:

- google.golang.org/grpc - For gRPC communication
- github.com/joho/godotenv - For environment variable management
- google.golang.org/protobuf - For Protocol Buffers support

## Building

```bash
# Clone the repository
git clone https://github.com/alfaz-ahmed/cbt-grpc-go-client.git
cd cbt-grpc-go-client

# Install dependencies
go mod download

# Build the project
go build -o cbt-grpc-go-client .
```

## Code generation from protos
```bash
 protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative com/nutanix/dataprotection/v4/error/error.proto
 protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative com/nutanix/dataprotection/v4/content/*.proto
 protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative com/nutanix/dataprotection/v4/*.proto
 protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative com/nutanix/common/v1/config/config.proto
 protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative com/nutanix/common/v1/response/response.proto
```
