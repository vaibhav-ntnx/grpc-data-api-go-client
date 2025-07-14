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

echo -e "\n=== VDisk Examples completed ===" 