//go:build !scaletest
// +build !scaletest

package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	flag.Parse()

	if *peSocket == "" || *pcIP == "" || *recoveryPointExtID == "" || *diskRecoveryPointExtID == "" {
		log.Fatal("Missing required arguments. Use -h for help.")
	}

	switch *function {
	case "vm":
		if *vmRecoveryPointExtID == "" {
			log.Fatal("Missing vm_recovery_point_ext_id for vm function")
		}
	case "volume_group":
		if *volumeGroupRecoveryPointExtID == "" {
			log.Fatal("Missing volume_group_recovery_point_ext_id for volume_group function")
		}
	default:
		log.Fatal("Function must be 'vm' or 'volume_group'.")
	}

	jwtToken, err := fetchJwtToken()
	if err != nil {
		log.Fatalf("Failed to fetch JWT token: %v", err)
	}

	start := time.Now()
	if *function == "vm" {
		err = computeVmChangedRegions(*peSocket, jwtToken)
	} else {
		err = computeVolumeGroupChangedRegions(*peSocket, jwtToken)
	}
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Request completed in %v\n", time.Since(start))
}
