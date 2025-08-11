//go:build !scaletest
// +build !scaletest

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `VDisk gRPC Client - Supports single, batch, and throughput testing operations

Usage:
  %s [flags]

Examples:
  # Single read operation
  %s -vdisk_server=localhost:9090 -vdisk_operation=read -vm_disk_uuid=12345 -read_offset=0 -read_length=1024

  # Single write operation  
  %s -vdisk_server=localhost:9090 -vdisk_operation=write -vm_disk_uuid=12345 -write_offset=0 -write_length=10 -write_data="Hello"

  # Batch read operations (5 concurrent reads)
  %s -vdisk_server=localhost:9090 -vdisk_operation=read -vm_disk_uuid=12345 -batch_mode=true -batch_size=5 -read_offset=0 -read_length=1024

  # Batch write operations (3 concurrent writes with 100ms delay between starts)
  %s -vdisk_server=localhost:9090 -vdisk_operation=write -vm_disk_uuid=12345 -batch_mode=true -batch_size=3 -batch_delay=100ms -write_offset=0 -write_length=10 -write_data="Hello"

  # Throughput test - read operations for 10 minutes with max 10 concurrent requests
  %s -vdisk_server=localhost:9090 -vdisk_operation=read -vm_disk_uuid=12345 -throughput_mode=true -test_duration=10m -max_concurrent=10 -report_interval=30s -read_length=1024

  # Throughput test - write operations for 5 minutes with max 8 concurrent requests
  %s -vdisk_server=localhost:9090 -vdisk_operation=write -vm_disk_uuid=12345 -throughput_mode=true -test_duration=5m -max_concurrent=8 -report_interval=15s -write_length=1024 -write_data="ThroughputTest"

  # High throughput test - read operations for 2 minutes with max 20 concurrent requests, reporting every 10 seconds
  %s -vdisk_server=localhost:9090 -vdisk_operation=read -vm_disk_uuid=12345 -throughput_mode=true -test_duration=2m -max_concurrent=20 -report_interval=10s -read_length=4096

  # High throughput test with connection pool - read operations with 5 gRPC connections and max 50 concurrent requests
  %s -vdisk_server=localhost:9090 -vdisk_operation=read -vm_disk_uuid=12345 -throughput_mode=true -test_duration=5m -max_concurrent=50 -connection_pool_size=5 -report_interval=15s -read_length=8192

Flags:
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	// Check if help was requested
	if flag.NFlag() == 0 {
		printUsage()
		return
	}

	err := runVDiskOperation()
	if err != nil {
		log.Fatalf("VDisk operation failed: %v", err)
	}
}
