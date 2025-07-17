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
	fmt.Fprintf(os.Stderr, `VDisk gRPC Client - Supports both single and batch operations

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

Flags:
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
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
