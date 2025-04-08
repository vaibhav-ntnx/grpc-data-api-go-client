package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	// import the generated package for your protobuf definitions
	"github.com/alfaz-ahmed/cbt-grpc-go-client/com/nutanix/dataprotection/v4/content"
	vmservice "github.com/alfaz-ahmed/cbt-grpc-go-client/com/nutanix/dataprotection/v4/content"
)

// These would be from your generated Go protobuf files
// import your_proto_package "path/to/your/proto/package"

// Constants from the Python code
const (
	PCIP                                 = "10.33.81.74"
	RECOVERY_POINT_EXT_ID                = "4f30f7ab-9236-4dbb-9122-aa76194a358d"
	VM_RECOVERY_POINT_EXT_ID             = "52b90e12-5abd-4a3d-8b1e-652637e9b319"
	DISK_RECOVERY_POINT_EXT_ID           = "b17a08ea-3344-4362-ba20-659292c836eb"
	REFERENCE_RECOVERY_POINT_EXT_ID      = "9d5bba3c-67d8-4a15-8b55-0c6584ae7891"
	REFERENCE_DISK_RECOVERY_POINT_EXT_ID = "0e675252-d5b4-4ebb-8856-951ca5cf6b40"
	REFERENCE_VM_RECOVERY_POINT_EXT_ID   = "bfef1eef-a42a-4300-9ef3-a5db041b7e10"
)

// JWT token response structure
type DiscoverClusterResponse struct {
	Data struct {
		JwtToken string `json:"jwtToken"`
	} `json:"data"`
}

// fetchJwtToken fetches a JWT token for authentication
func fetchJwtToken() (string, error) {
	url := fmt.Sprintf("https://%s:9440/api/dataprotection/v4.0/config/recovery-points/%s/$actions/discover-cluster", PCIP, RECOVERY_POINT_EXT_ID)

	requestBody := map[string]interface{}{
		"operation": "COMPUTE_CHANGED_REGIONS",
		"spec": map[string]interface{}{
			"$objectType": "dataprotection.v4.content.ComputeChangedRegionsClusterDiscoverSpec",
			"diskRecoveryPoint": map[string]interface{}{
				"$objectType":            "dataprotection.v4.content.VmDiskRecoveryPointReference",
				"recoveryPointExtId":     RECOVERY_POINT_EXT_ID,
				"vmRecoveryPointExtId":   VM_RECOVERY_POINT_EXT_ID,
				"diskRecoveryPointExtId": DISK_RECOVERY_POINT_EXT_ID,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %v", err)
	}

	// Create HTTP client with TLS configuration that skips certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Create request
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic YWRtaW46TnV0YW5peC4xMjM=") // Replace with actual base64 if needed

	fmt.Printf("Requesting JWT token from: %s\n", url)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get JWT token. HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response DiscoverClusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if response.Data.JwtToken == "" {
		return "", fmt.Errorf("JWT token not found in response")
	}

	fmt.Println("JWT Token fetched successfully.")
	return response.Data.JwtToken, nil
}

// createGrpcChannel creates a secure gRPC channel
func createGrpcChannel(serverAddress string) (*grpc.ClientConn, error) {
	// Create TLS credentials with InsecureSkipVerify set to true
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	// Create channel with options
	conn, err := grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return conn, nil
}

// computeVmChangedRegions makes the gRPC call to compute VM changed regions
func computeVmChangedRegions(serverAddress, jwtToken string) error {
	conn, err := createGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// This is a placeholder - you need to generate the actual protobuf Go code
	client := vmservice.NewVMRecoveryPointComputeChangedRegionsServiceClient(conn)

	// Create context with metadata for authentication
	ctx := metadata.AppendToOutgoingContext(
		context.Background(),
		"authorization", fmt.Sprintf("Bearer %s", jwtToken),
	)

	request := &vmservice.VmRecoveryPointComputeChangedRegionsArg{
		RecoveryPointExtId:   RECOVERY_POINT_EXT_ID,
		VmRecoveryPointExtId: VM_RECOVERY_POINT_EXT_ID,
		ExtId:                DISK_RECOVERY_POINT_EXT_ID,
		Body: &content.VmRecoveryPointChangedRegionsComputeSpec{
			Base: &content.VmDiskRecoveryPointClusterDiscoverSpec{
				Base: &content.BaseRecoveryPointSpec{
					ReferenceRecoveryPointExtId:     REFERENCE_RECOVERY_POINT_EXT_ID,
					ReferenceDiskRecoveryPointExtId: REFERENCE_DISK_RECOVERY_POINT_EXT_ID,
				},
				ReferenceVmRecoveryPointExtId: REFERENCE_VM_RECOVERY_POINT_EXT_ID,
			},
			Offset:    0,
			Length:    1024 * 1024,
			BlockSize: 65536,
		},
	}

	stream, err := client.VmRecoveryPointComputeChangedRegions(ctx, request)
	if err != nil {
		return fmt.Errorf("RPC failed: %v", err)
	}

	fmt.Println("Initiating VM Changed Regions Streaming RPC...")
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %v", err)
		}

		// Process response based on what field is set
		if response.Content.GetChangedRegionArrayData().ProtoReflect() != nil {
			fmt.Println("\nReceived Changed Region Batch:")
			for _, region := range response.Content.GetChangedRegionArrayData().Value {
				fmt.Printf("Offset: %d, Length: %d, Type: %s\n",
					region.Offset, region.Length, region.RegionType)
			}
		} else if response.Content.GetErrorResponseData() != nil {
			fmt.Printf("\nError: %s\n", response.Content.GetErrorResponseData().Value)
			break
		}
	}
	fmt.Println("\nStreaming complete.")

	// Placeholder code instead of actual implementation
	fmt.Println("VM Changed Regions: Implementation would use the protobuf Go code")
	return nil
}

// computeVolumeGroupChangedRegions makes the gRPC call to compute volume group changed regions
func computeVolumeGroupChangedRegions(serverAddress string) error {
	conn, err := createGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// This is a placeholder - you need to generate the actual protobuf Go code
	client := vmservice.NewVolumeGroupRecoveryPointComputeChangedRegionsServiceClient(conn)

	// In the actual implementation, you would create the request like this:

	request := &vmservice.VolumeGroupRecoveryPointComputeChangedRegionsArg{
		RecoveryPointExtId:            "dummy-recovery-point-ext-id",
		VolumeGroupRecoveryPointExtId: "dummy-volume-group-recovery-point-ext-id",
		ExtId:                         "dummy-disk-recovery-point-ext-id",
		Body: &content.VolumeGroupRecoveryPointChangedRegionsComputeSpec{
			Base: &content.VolumeGroupDiskRecoveryPointClusterDiscoverSpec{
				Base: &content.BaseRecoveryPointSpec{
					ReferenceRecoveryPointExtId:     "ref-recovery-point-ext-id",
					ReferenceDiskRecoveryPointExtId: "ref-disk-recovery-point-ext-id",
				},
				ReferenceVolumeGroupRecoveryPointExtId: "ref-volume-group-recovery-point-ext-id",
			},
			Offset:    0,
			Length:    1024 * 1024,
			BlockSize: 65536,
		},
	}

	response, err := client.VolumeGroupRecoveryPointComputeChangedRegions(context.Background(), request)
	if err != nil {
		return fmt.Errorf("RPC failed: %v", err)
	}

	if response.Content.GetChangedRegionArrayData() != nil {
		fmt.Println("Changed Regions:")
		for _, region := range response.Content.GetChangedRegionArrayData().Value {
			fmt.Printf("Offset: %d, Length: %d, Type: %s\n",
				region.Offset, region.Length, region.RegionType)
		}
	} else if response.Content.GetErrorResponseData() != nil {
		fmt.Printf("Error: %s\n", response.Content.GetErrorResponseData().Value)
	}

	// Placeholder code instead of actual implementation
	fmt.Println("Volume Group Changed Regions: Implementation would use the protobuf Go code")
	return nil
}

func main() {

	// Define command line flags
	serverFlag := flag.String("server", "", "Server in ip:port format")
	functionFlag := flag.String("function", "", "Which API to call (vm or volume_group)")
	flag.Parse()

	if *serverFlag == "" {
		log.Fatal("Server address is required. Use -server flag.")
	}
	if *functionFlag != "vm" && *functionFlag != "volume_group" {
		log.Fatal("Function must be 'vm' or 'volume_group'. Use -function flag.")
	}

	if *functionFlag == "vm" {
		jwtToken, err := fetchJwtToken()
		if err != nil {
			log.Fatalf("Failed to fetch JWT token: %v", err)
		}
		fmt.Println("JWT Token:", jwtToken)

		if err := computeVmChangedRegions(*serverFlag, jwtToken); err != nil {
			log.Fatalf("Error computing VM changed regions: %v", err)
		}
	} else if *functionFlag == "volume_group" {
		if err := computeVolumeGroupChangedRegions(*serverFlag); err != nil {
			log.Fatalf("Error computing volume group changed regions: %v", err)
		}
	}
}
