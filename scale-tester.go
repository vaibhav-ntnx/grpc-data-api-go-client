//go:build scaletest
// +build scaletest

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var jwtToken string

func runRequest(id int) {
	start := time.Now()

	var err error
	switch *function {
	case "vm":
		err = computeVmChangedRegions(*peSocket, jwtToken)
	case "volume_group":
		err = computeVolumeGroupChangedRegions(*peSocket, jwtToken)
	default:
		err = fmt.Errorf("unknown function: %s", *function)
	}
	duration := time.Since(start)
	if err != nil {
		fmt.Printf("[#%d] Request failed: %v (Duration: %v)\n", id, err, duration)
	} else {
		fmt.Printf("[#%d] Request succeeded (Duration: %v)\n", id, duration)
	}
}

func main() {
	parallel := flag.Int("parallel", 1, "Number of parallel requests")
	flag.Parse()

	if *peSocket == "" || *pcIP == "" || *recoveryPointExtID == "" || *diskRecoveryPointExtID == "" {
		fmt.Println("Missing required flags. Use -h for help.")
		os.Exit(1)
	}

	switch *function {
	case "vm":
		if *vmRecoveryPointExtID == "" {
			fmt.Println("Missing vm_recovery_point_ext_id for vm function")
			os.Exit(1)
		}
	case "volume_group":
		if *volumeGroupRecoveryPointExtID == "" {
			fmt.Println("Missing volume_group_recovery_point_ext_id for volume_group function")
			os.Exit(1)
		}
	default:
		fmt.Println("Function must be 'vm' or 'volume_group'")
		os.Exit(1)
	}

	var err error
	jwtToken, err = fetchJwtToken()
	if err != nil {
		fmt.Printf("Failed to fetch JWT token: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	id := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nInterrupt received. Stopping...")
			return
		default:
			var wg sync.WaitGroup
			for i := 0; i < *parallel; i++ {
				wg.Add(1)
				go func(localID int) {
					defer wg.Done()
					runRequest(localID)
				}(id)
				id++
			}
			wg.Wait()
		}
	}
}
