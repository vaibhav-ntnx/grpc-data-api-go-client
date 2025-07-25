#!/bin/bash

# VDisk gRPC Client Examples
# This script demonstrates how to use the VDisk gRPC client for read and write operations

echo "=== VDisk gRPC Client Examples ==="

# Build the client
echo "Building the VDisk gRPC client..."
go build -o vdisk-client .

# Example 1: VDisk Read Operation with VM disk UUID (non-TLS)
echo -e "\n1. VDisk Read Operation (VM Disk UUID - non-TLS)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -read_offset=0 \
    -read_length=1048576 \
    -max_response_size=4194304 \
    -vdisk_auth_token="your_auth_token_here"

# Example 2: VDisk Read Operation with Recovery Point UUID (non-TLS)
echo -e "\n2. VDisk Read Operation (Recovery Point UUID - non-TLS)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -disk_recovery_point_uuid="12345678-1234-5678-9012-123456789012" \
    -read_offset=0 \
    -read_length=0 \
    -max_response_size=4194304 \
    -vdisk_auth_token="your_auth_token_here"

# Example 3: VDisk Write Operation with VG disk UUID (non-TLS)
echo -e "\n3. VDisk Write Operation (VG Disk UUID - non-TLS)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=write \
    -vg_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -write_offset=0 \
    -write_length=1024 \
    -write_data="Hello, VDisk!" \
    -compression_type=lz4 \
    -checksum_type=crc32 \
    -sequence_number=1 \
    -vdisk_auth_token="your_auth_token_here"

# Example 4: VDisk Write Operation with VM disk UUID and TLS enabled
echo -e "\n4. VDisk Write Operation (VM Disk UUID with TLS)"
./vdisk-client \
    -vdisk_server="localhost:9443" \
    -vdisk_operation=write \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -write_offset=4096 \
    -write_length=2048 \
    -write_data="VDisk Data with TLS" \
    -compression_type=snappy \
    -checksum_type=sha256 \
    -sequence_number=2 \
    -vdisk_use_tls=true \
    -vdisk_skip_tls_verify=true \
    -vdisk_auth_token="your_auth_token_here"

# Example 5: VDisk Read Operation with Volume Group disk UUID (non-TLS)
echo -e "\n5. VDisk Read Operation (Volume Group Disk UUID - non-TLS)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -vg_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -read_offset=0 \
    -read_length=2097152 \
    -max_response_size=8388608 \
    -vdisk_auth_token="your_auth_token_here"

echo -e "\n=== Batch Operations Examples ==="

# Example 6: Batch VDisk Read Operations (5 concurrent reads)
echo -e "\n6. Batch VDisk Read Operations (5 concurrent reads)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -batch_mode=true \
    -batch_size=5 \
    -read_offset=0 \
    -read_length=1048576 \
    -max_response_size=4194304 \
    -vdisk_auth_token="your_auth_token_here"

# Example 7: Batch VDisk Write Operations (3 concurrent writes)
echo -e "\n7. Batch VDisk Write Operations (3 concurrent writes)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=write \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -batch_mode=true \
    -batch_size=3 \
    -write_offset=0 \
    -write_length=1024 \
    -write_data="BatchData" \
    -compression_type=lz4 \
    -checksum_type=crc32 \
    -sequence_number=100 \
    -vdisk_auth_token="your_auth_token_here"

# Example 8: Batch VDisk Read Operations with delay between starts (10 concurrent reads, 50ms delay)
echo -e "\n8. Batch VDisk Read Operations with staggered start (10 concurrent reads, 50ms delay)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -batch_mode=true \
    -batch_size=10 \
    -batch_delay=50ms \
    -read_offset=0 \
    -read_length=524288 \
    -max_response_size=2097152 \
    -vdisk_auth_token="your_auth_token_here"

# Example 9: Large Batch VDisk Write Operations with TLS (20 concurrent writes)
echo -e "\n9. Large Batch VDisk Write Operations with TLS (20 concurrent writes)"
./vdisk-client \
    -vdisk_server="localhost:9443" \
    -vdisk_operation=write \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -batch_mode=true \
    -batch_size=20 \
    -batch_delay=25ms \
    -write_offset=0 \
    -write_length=2048 \
    -write_data="ConcurrentWriteTest" \
    -compression_type=snappy \
    -checksum_type=sha256 \
    -sequence_number=200 \
    -vdisk_use_tls=true \
    -vdisk_skip_tls_verify=true \
    -vdisk_auth_token="your_auth_token_here"

echo -e "\n=== Throughput Testing Examples ==="

# Example 10: Throughput test - read operations for 2 minutes with max 10 concurrent requests
echo -e "\n10. Throughput Test - Read Operations (2 minutes, max 10 concurrent, 30s reports)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=read \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -throughput_mode=true \
    -test_duration=2m \
    -max_concurrent=10 \
    -report_interval=30s \
    -read_offset=0 \
    -read_length=1048576 \
    -max_response_size=4194304 \
    -vdisk_auth_token="your_auth_token_here"

# Example 11: Throughput test - write operations for 1 minute with max 8 concurrent requests
echo -e "\n11. Throughput Test - Write Operations (1 minute, max 8 concurrent, 15s reports)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=write \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -throughput_mode=true \
    -test_duration=1m \
    -max_concurrent=8 \
    -report_interval=15s \
    -write_offset=0 \
    -write_length=1024 \
    -write_data="ThroughputTestData" \
    -compression_type=lz4 \
    -checksum_type=crc32 \
    -sequence_number=1000 \
    -vdisk_auth_token="your_auth_token_here"

# Example 12: High throughput test - read operations with TLS for 10 minutes
echo -e "\n12. High Throughput Test - Read with TLS (10 minutes, max 20 concurrent, 1m reports)"
./vdisk-client \
    -vdisk_server="localhost:9443" \
    -vdisk_operation=read \
    -vm_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -throughput_mode=true \
    -test_duration=10m \
    -max_concurrent=20 \
    -report_interval=1m \
    -read_offset=0 \
    -read_length=2097152 \
    -max_response_size=8388608 \
    -vdisk_use_tls=true \
    -vdisk_skip_tls_verify=true \
    -vdisk_auth_token="your_auth_token_here"

# Example 13: Stress test - write operations for 30 seconds with maximum concurrency
echo -e "\n13. Stress Test - Write Operations (30 seconds, max 25 concurrent, 10s reports)"
./vdisk-client \
    -vdisk_server="localhost:9090" \
    -vdisk_operation=write \
    -vg_disk_uuid="12345678-1234-5678-9012-123456789012" \
    -throughput_mode=true \
    -test_duration=30s \
    -max_concurrent=25 \
    -report_interval=10s \
    -write_offset=0 \
    -write_length=4096 \
    -write_data="StressTestPayload" \
    -compression_type=snappy \
    -checksum_type=sha256 \
    -sequence_number=2000 \
    -vdisk_auth_token="your_auth_token_here"

echo -e "\n=== VDisk Examples completed ===" 