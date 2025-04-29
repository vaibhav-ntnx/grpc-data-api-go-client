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

	"github.com/alfaz-ahmed/cbt-grpc-go-client/com/nutanix/dataprotection/v4/content"
	vmservice "github.com/alfaz-ahmed/cbt-grpc-go-client/com/nutanix/dataprotection/v4/content"
)

type DiscoverClusterResponse struct {
	Data struct {
		JwtToken string `json:"jwtToken"`
	} `json:"data"`
}

var (
	peSocket                        = flag.String("pe_socket", "", "PE socket in ip:port format")
	function                        = flag.String("function", "", "API to call (vm or volume_group)")
	pcIP                            = flag.String("pc_ip", "", "Prism Central IP")
	recoveryPointExtID              = flag.String("recovery_point_ext_id", "", "Recovery Point Ext ID")
	vmRecoveryPointExtID            = flag.String("vm_recovery_point_ext_id", "", "VM Recovery Point Ext ID")
	diskRecoveryPointExtID          = flag.String("disk_recovery_point_ext_id", "", "Disk Recovery Point Ext ID")
	referenceRecoveryPointExtID     = flag.String("reference_recovery_point_ext_id", "", "Reference Recovery Point Ext ID")
	referenceDiskRecoveryPointExtID = flag.String("reference_disk_recovery_point_ext_id", "", "Reference Disk Recovery Point Ext ID")
	referenceVmRecoveryPointExtID   = flag.String("reference_vm_recovery_point_ext_id", "", "Reference VM Recovery Point Ext ID")
)

func fetchJwtToken() (string, error) {
	url := fmt.Sprintf("https://%s:9440/api/dataprotection/v4.0/config/recovery-points/%s/$actions/discover-cluster", *pcIP, *recoveryPointExtID)
	requestBody := map[string]interface{}{
		"operation": "COMPUTE_CHANGED_REGIONS",
		"spec": map[string]interface{}{
			"$objectType": "dataprotection.v4.content.ComputeChangedRegionsClusterDiscoverSpec",
			"diskRecoveryPoint": map[string]interface{}{
				"$objectType":            "dataprotection.v4.content.VmDiskRecoveryPointReference",
				"recoveryPointExtId":     *recoveryPointExtID,
				"vmRecoveryPointExtId":   *vmRecoveryPointExtID,
				"diskRecoveryPointExtId": *diskRecoveryPointExtID,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %v", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic YWRtaW46TnV0YW5peC4xMjM=")

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

func createGrpcChannel(serverAddress string) (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

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

func computeVmChangedRegions(serverAddress, jwtToken string) error {
	conn, err := createGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := vmservice.NewVMRecoveryPointComputeChangedRegionsServiceClient(conn)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", fmt.Sprintf("Bearer %s", jwtToken))

	request := &vmservice.VmRecoveryPointComputeChangedRegionsArg{
		RecoveryPointExtId:   *recoveryPointExtID,
		VmRecoveryPointExtId: *vmRecoveryPointExtID,
		ExtId:                *diskRecoveryPointExtID,
	}

	if *referenceRecoveryPointExtID != "" {
		request.Body = &content.VmRecoveryPointChangedRegionsComputeSpec{
			Base: &content.VmDiskRecoveryPointClusterDiscoverSpec{
				Base: &content.BaseRecoveryPointSpec{
					ReferenceRecoveryPointExtId:     *referenceRecoveryPointExtID,
					ReferenceDiskRecoveryPointExtId: *referenceDiskRecoveryPointExtID,
				},
				ReferenceVmRecoveryPointExtId: *referenceVmRecoveryPointExtID,
			},
			Offset:    0,
			Length:    1024 * 1024,
			BlockSize: 65536,
		}
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
		if response.Content.GetChangedRegionArrayData() != nil {
			for _, region := range response.Content.GetChangedRegionArrayData().Value {
				fmt.Printf("Offset: %d, Length: %d, Type: %s\n", region.Offset, region.Length, region.RegionType)
			}
		} else if response.Content.GetErrorResponseData() != nil {
			fmt.Printf("Error: %s\n", response.Content.GetErrorResponseData().Value)
			break
		}
	}
	fmt.Println("\nStreaming complete.")
	return nil
}

func main() {
	flag.Parse()

	if *peSocket == "" || *pcIP == "" || *recoveryPointExtID == "" || *vmRecoveryPointExtID == "" || *diskRecoveryPointExtID == "" {
		log.Fatal("Missing required arguments. Use -h for help.")
	}
	if *function != "vm" && *function != "volume_group" {
		log.Fatal("Function must be 'vm' or 'volume_group'.")
	}

	if *function == "vm" {
		jwtToken, err := fetchJwtToken()
		if err != nil {
			log.Fatalf("Failed to fetch JWT token: %v", err)
		}
		if err := computeVmChangedRegions(*peSocket, jwtToken); err != nil {
			log.Fatalf("Error computing VM changed regions: %v", err)
		}
	}
}
