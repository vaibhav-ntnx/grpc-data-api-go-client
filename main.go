//go:build !scaletest
// +build !scaletest

package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()

	err := runVDiskOperation()
	if err != nil {
		log.Fatalf("VDisk operation failed: %v", err)
	}
}
