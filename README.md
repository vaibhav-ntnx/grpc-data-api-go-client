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
## Run
Options available
```
./cbt-grpc-go-client -h
Usage of ./cbt-grpc-go-client:
  -disk_recovery_point_ext_id string
        Disk Recovery Point Ext ID
  -function string
        API to call (vm or volume_group)
  -pc_ip string
        Prism Central IP
  -pe_socket string
        PE socket in ip:port format
  -recovery_point_ext_id string
        Recovery Point Ext ID
  -reference_disk_recovery_point_ext_id string
        Reference Disk Recovery Point Ext ID
  -reference_recovery_point_ext_id string
        Reference Recovery Point Ext ID
  -reference_vm_recovery_point_ext_id string
        Reference VM Recovery Point Ext ID
  -vm_recovery_point_ext_id string
        VM Recovery Point Ext ID
```
Example command 
```
./cbt-grpc-go-client -function vm -pc_ip 10.61.4.92 -pe_socket 10.61.42.78:50051 -recovery_point_ext_id c9812280-bcf6-47e9-b050-457209c24ad2 -vm_recovery_point_ext_id cbfb97d6-8464-41f9-8a45-e0e6796af9dc -disk_recovery_point_ext_id c5819a63-1037-4be5-af97-cc16244cc0e5 
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
