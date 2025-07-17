# Nutanix VDisk gRPC Client

## Summary
A Go client for Nutanix VDisk gRPC APIs, providing streaming read/write operations for virtual disks with support for compression, checksums, various disk identifier types, and batch/concurrent operations.

## Features
- **Streaming Operations**: Bidirectional gRPC streaming for efficient data transfer
- **Batch Operations**: Concurrent execution of multiple VDisk operations for testing performance and concurrency
- **Multiple Disk Identifiers**: Recovery point UUID, VM disk UUID, Volume Group disk UUID
- **Data Compression**: LZ4, Snappy, Zlib compression support
- **Data Integrity**: CRC32, SHA1, SHA256 checksum verification
- **Authentication**: Bearer token authentication
- **TLS Support**: Configurable TLS/non-TLS connections
- **Error Handling**: Comprehensive error handling for network, server, and data errors

## Dependencies

This project uses Go modules for dependency management. Key dependencies:

- google.golang.org/grpc - For gRPC communication
- google.golang.org/protobuf - For Protocol Buffers support

## Building

```bash
# Clone the repository
git clone https://github.com/vaibhav-ntnx/grpc-data-api-go-client.git
cd grpc-data-api-go-client

# Install dependencies
go mod download

# Build the project
go build -o vdisk-client .

# Build the tester for parallel scale test
go build -tags=scaletest -o vdisk-scale-tester
```

## Usage

### VDisk Operations

Options available:
```
./vdisk-client -h
VDisk flags:
  -vdisk_server string
        VDisk server address in ip:port format
  -vdisk_operation string
        VDisk operation (read or write)
  -vdisk_auth_token string
        Authentication token for VDisk service
  -vdisk_use_tls
        Use TLS for gRPC connection (default: false)
  -vdisk_skip_tls_verify
        Skip TLS certificate verification (default: true)
  
  # Batch operation flags:
  -batch_mode
        Enable batch mode for concurrent operations (default: false)
  -batch_size int
        Number of concurrent operations to run (default: 1)
  -batch_delay duration
        Delay between starting batch operations (default: 0s)
  
  # Disk identifier flags (choose one):
  -disk_recovery_point_uuid string
        Disk recovery point UUID
  -vm_disk_uuid string
        VM disk UUID
  -vg_disk_uuid string
        Volume group disk UUID
  
  # Read operation flags:
  -read_offset int
        Read offset in bytes (default 0)
  -read_length int
        Read length in bytes (0 for entire disk)
  -max_response_size int
        Maximum response size in bytes (default 1048576)
  
  # Write operation flags:
  -write_offset int
        Write offset in bytes (default 0)
  -write_length int
        Write length in bytes (default 0)
  -write_data string
        Data to write (hex string)
  -compression_type string
        Compression type (none, lz4, snappy, zlib) (default "none")
  -checksum_type string
        Checksum type (none, crc32, sha1, sha256) (default "none")
  -sequence_number int
        Sequence number for write ordering (default 0)
```

### VDisk Read Operation Examples

#### Non-TLS Connection (Default)
```bash
# Read from VM disk UUID
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -read_offset=0 -read_length=1048576 -max_response_size=4194304 -vdisk_auth_token="your_auth_token"

# Read from recovery point UUID
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -disk_recovery_point_uuid="12345678-1234-5678-9012-123456789012" -read_offset=0 -read_length=0 -max_response_size=4194304 -vdisk_auth_token="your_auth_token"

# Read from Volume Group disk UUID
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vg_disk_uuid="12345678-1234-5678-9012-123456789012" -read_offset=0 -read_length=1048576 -vdisk_auth_token="your_auth_token"
```

#### TLS Connection
```bash
# Read from VM disk UUID with TLS
./vdisk-client -vdisk_server="localhost:9443" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -read_offset=0 -read_length=1048576 -max_response_size=4194304 -vdisk_use_tls=true -vdisk_skip_tls_verify=true -vdisk_auth_token="your_auth_token"
```

### VDisk Write Operation Examples

#### Non-TLS Connection (Default)
```bash
# Write to VM disk UUID with compression
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -write_offset=0 -write_length=1024 -write_data="Hello, VDisk!" -compression_type=lz4 -checksum_type=crc32 -sequence_number=1 -vdisk_auth_token="your_auth_token"

# Write to Volume Group disk UUID with checksum
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=write -vg_disk_uuid="12345678-1234-5678-9012-123456789012" -write_offset=0 -write_length=1024 -write_data="Hello, VDisk!" -compression_type=snappy -checksum_type=sha256 -sequence_number=1 -vdisk_auth_token="your_auth_token"
```

#### TLS Connection
```bash
# Write to VM disk UUID with TLS
./vdisk-client -vdisk_server="localhost:9443" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -write_offset=0 -write_length=1024 -write_data="Hello, VDisk!" -compression_type=lz4 -checksum_type=crc32 -sequence_number=1 -vdisk_use_tls=true -vdisk_skip_tls_verify=true -vdisk_auth_token="your_auth_token"
```

### Batch Operation Examples

The client supports batch/concurrent operations for testing performance and concurrency. In batch mode, multiple operations run concurrently using goroutines.

#### Batch Read Operations
```bash
# 5 concurrent read operations
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -batch_mode=true -batch_size=5 -read_offset=0 -read_length=1048576 -vdisk_auth_token="your_auth_token"

# 10 concurrent reads with staggered start (50ms delay between starts)
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -batch_mode=true -batch_size=10 -batch_delay=50ms -read_offset=0 -read_length=524288 -vdisk_auth_token="your_auth_token"
```

#### Batch Write Operations
```bash
# 3 concurrent write operations
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -batch_mode=true -batch_size=3 -write_offset=0 -write_length=1024 -write_data="BatchData" -compression_type=lz4 -checksum_type=crc32 -sequence_number=100 -vdisk_auth_token="your_auth_token"

# 20 concurrent writes with TLS and staggered start
./vdisk-client -vdisk_server="localhost:9443" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -batch_mode=true -batch_size=20 -batch_delay=25ms -write_offset=0 -write_length=2048 -write_data="ConcurrentTest" -compression_type=snappy -checksum_type=sha256 -vdisk_use_tls=true -vdisk_auth_token="your_auth_token"
```

#### Batch Operation Features
- **Concurrent Execution**: Operations run in parallel using goroutines
- **Unique Data Per Operation**: Each operation uses slightly different offsets and data to avoid conflicts
- **Comprehensive Results**: Detailed summary including success rate, timing, and data processed
- **Staggered Start**: Optional delay between starting operations to simulate real-world scenarios
- **Thread-Safe**: Uses WaitGroup for proper synchronization

## Examples

Run the example script to see various usage patterns:
```bash
chmod +x vdisk-examples.sh
./vdisk-examples.sh
```

## TLS Configuration

The client supports both TLS and non-TLS connections:

- **Default**: Non-TLS connections (suitable for local development and non-TLS gRPC servers)
- **TLS**: Use `-vdisk_use_tls=true` to enable TLS
- **TLS Verification**: Use `-vdisk_skip_tls_verify=false` to enable certificate verification (default is to skip verification)

## Code generation from proto

```bash
# Generate Go code from VDisk proto (run from project root)
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       protos/stargate_vdisk_rpc_svc.proto
```

## Project Structure

```
grpc-data-api-go-client/
├── main.go                               # Main entry point
├── vdisk-utils.go                        # VDisk gRPC client implementation
├── vdisk-examples.sh                     # Example usage scripts
├── go.mod                                # Go module dependencies
├── go.sum                                # Go module checksums
├── protos/                               # Protocol buffer definitions
│   ├── stargate_vdisk_rpc_svc.proto      # VDisk service proto definition
│   ├── stargate_vdisk_rpc_svc.pb.go      # Generated VDisk protobuf code
│   └── stargate_vdisk_rpc_svc_grpc.pb.go # Generated VDisk gRPC code
└── README.md                             # This file
```

## API Reference

### VDisk Service Methods

- **VDiskStreamRead**: Streaming read operation from virtual disk
- **VDiskStreamWrite**: Streaming write operation to virtual disk

### Disk Identifier Types

- **Recovery Point UUID**: Identifies disk by recovery point
- **VM Disk UUID**: Identifies disk by VM disk UUID
- **Volume Group Disk UUID**: Identifies disk by volume group disk UUID

### Compression Types

- **None**: No compression
- **LZ4**: LZ4 compression
- **Snappy**: Snappy compression
- **Zlib**: Zlib compression

### Checksum Types

- **None**: No checksum
- **CRC32**: CRC32 checksum
- **SHA1**: SHA1 checksum
- **SHA256**: SHA256 checksum