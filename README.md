# Nutanix VDisk gRPC Client

## Summary
A Go client for Nutanix VDisk gRPC APIs, providing streaming read/write operations for virtual disks with support for compression, checksums, various disk identifier types, and batch/concurrent operations.

## Features
- **Streaming Operations**: Bidirectional gRPC streaming for efficient data transfer
- **Batch Operations**: Concurrent execution of multiple VDisk operations for testing performance and concurrency
- **Throughput Testing**: Continuous load testing with configurable duration, concurrency limits, and real-time metrics
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
  
  # Throughput testing flags:
  -throughput_mode
        Enable throughput testing mode (default: false)
  -test_duration duration
        Duration to run throughput test (default: 10m)
  -max_concurrent int
        Maximum concurrent requests (default: 10)
  -report_interval duration
        Interval for intermediate throughput reports (default: 30s)
  
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

### Throughput Testing Examples

The client supports comprehensive throughput testing for performance analysis and load testing. Throughput mode runs continuous operations for a specified duration with configurable concurrency limits.

#### Basic Throughput Testing
```bash
# Read throughput test for 10 minutes with max 10 concurrent requests
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -throughput_mode=true -test_duration=10m -max_concurrent=10 -report_interval=30s -read_length=1048576 -vdisk_auth_token="your_auth_token"

# Write throughput test for 5 minutes with max 8 concurrent requests
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -throughput_mode=true -test_duration=5m -max_concurrent=8 -report_interval=15s -write_length=1024 -write_data="ThroughputTest" -vdisk_auth_token="your_auth_token"
```

#### High Load Testing
```bash
# High concurrency read test for 2 minutes with frequent reporting
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=read -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -throughput_mode=true -test_duration=2m -max_concurrent=20 -report_interval=10s -read_length=4096 -vdisk_auth_token="your_auth_token"

# Stress test with maximum concurrency for 30 seconds
./vdisk-client -vdisk_server="localhost:9090" -vdisk_operation=write -vm_disk_uuid="12345678-1234-5678-9012-123456789012" -throughput_mode=true -test_duration=30s -max_concurrent=25 -report_interval=5s -write_length=2048 -write_data="StressTest" -vdisk_auth_token="your_auth_token"
```

#### Throughput Testing Features
- **Continuous Operation**: Runs requests in a loop for the specified duration
- **Concurrency Control**: Limits maximum concurrent requests (max 10 as requested)
- **Real-time Metrics**: Provides intermediate reports at configurable intervals
- **Comprehensive Results**: Detailed throughput metrics including:
  - Requests per second
  - Bytes per second (MB/s, GB/s)
  - Success/failure rates
  - Latency statistics (min, max, average)
  - Total data transfer amounts
- **Load Distribution**: Varies request offsets to distribute load across the disk
- **Duration Control**: Configurable test duration (default 10 minutes)
- **Interval Reporting**: Customizable reporting intervals for monitoring progress

#### Sample Throughput Output
```
=== Intermediate Throughput Report (Elapsed: 30s) ===
Total Requests: 1247
Successful: 1245, Failed: 2
Requests/sec: 41.57
Bytes/sec: 43516928.00 (41.50 MB/s)
Total Data: 1305507840 bytes (1245.00 MB)
Avg Latency: 240ms
Min Latency: 105ms, Max Latency: 892ms
========================================

FINAL THROUGHPUT TEST RESULTS
============================================================
Test Duration: 10m0s
Operation Type: read
Max Concurrent: 10

Request Statistics:
  Total Requests: 24936
  Successful: 24930 (99.98%)
  Failed: 6 (0.02%)

Throughput Metrics:
  Requests/sec: 41.55
  Bytes/sec: 43350272.00
  MB/sec: 41.36
  GB/sec: 0.0404

Data Transfer:
  Total Bytes: 26010280960
  Total MB: 24809.00
  Total GB: 24.23

Latency Statistics:
  Average: 241ms
  Minimum: 98ms
  Maximum: 1.2s

Efficiency Metrics:
  Avg bytes per request: 1043456.00
  Requests per minute: 2493.00
  MB per minute: 2481.60
============================================================
```

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