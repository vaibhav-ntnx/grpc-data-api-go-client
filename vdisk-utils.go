package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	
	"github.com/vaibhav-ntnx/grpc-data-api-go-client/protos"
)

var (
	vdiskServerAddress = flag.String("vdisk_server", "", "VDisk server address in ip:port format")
	vdiskOperation     = flag.String("vdisk_operation", "", "VDisk operation (read or write)")
	vdiskAuthToken     = flag.String("vdisk_auth_token", "", "Authentication token for VDisk service")
	vdiskUseTLS        = flag.Bool("vdisk_use_tls", false, "Use TLS for gRPC connection (default: false)")
	vdiskSkipTLSVerify = flag.Bool("vdisk_skip_tls_verify", true, "Skip TLS certificate verification (default: true)")
	
	// Disk identifier flags
	diskRecoveryPointUuid = flag.String("disk_recovery_point_uuid", "", "Disk recovery point UUID")
	vmDiskUuid           = flag.String("vm_disk_uuid", "", "VM disk UUID")
	vgDiskUuid           = flag.String("vg_disk_uuid", "", "Volume group disk UUID")
	
	// Read operation flags
	readOffset         = flag.Int64("read_offset", 0, "Read offset in bytes")
	readLength         = flag.Int64("read_length", 0, "Read length in bytes (0 for entire disk)")
	maxResponseSize    = flag.Int64("max_response_size", 1024*1024, "Maximum response size in bytes")
	
	// Write operation flags
	writeOffset        = flag.Int64("write_offset", 0, "Write offset in bytes")
	writeLength        = flag.Int64("write_length", 0, "Write length in bytes")
	writeData          = flag.String("write_data", "", "Data to write (hex string)")
	compressionType    = flag.String("compression_type", "none", "Compression type (none, lz4, snappy, zlib)")
	checksumType       = flag.String("checksum_type", "none", "Checksum type (none, crc32, sha1, sha256)")
	sequenceNumber     = flag.Int64("sequence_number", 0, "Sequence number for write ordering")
)

func createVDiskGrpcChannel(serverAddress string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	
	if *vdiskUseTLS {
		// Use TLS credentials
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: *vdiskSkipTLSVerify,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
		fmt.Println("Connecting with TLS...")
	} else {
		// Use insecure connection (no TLS)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		fmt.Println("Connecting without TLS...")
	}
	
	// Add call options
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(100*1024*1024),
		grpc.MaxCallSendMsgSize(100*1024*1024),
	))

	conn, err := grpc.Dial(serverAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to VDisk server: %v", err)
	}

	return conn, nil
}

func createDiskIdentifier() *protos.DiskIdentifier {
	diskId := &protos.DiskIdentifier{}
	
	if *diskRecoveryPointUuid != "" {
		diskId.Identifier = &protos.DiskIdentifier_DiskRecoveryPoint{
			DiskRecoveryPoint: &protos.DiskRecoveryPoint{
				RecoveryPointUuid: diskRecoveryPointUuid,
			},
		}
	} else if *vmDiskUuid != "" {
		diskId.Identifier = &protos.DiskIdentifier_VmDiskUuid{
			VmDiskUuid: *vmDiskUuid,
		}
	} else if *vgDiskUuid != "" {
		diskId.Identifier = &protos.DiskIdentifier_VgDiskUuid{
			VgDiskUuid: *vgDiskUuid,
		}
	}
	
	return diskId
}

func getCompressionType(compressionStr string) protos.CompressionType {
	switch compressionStr {
	case "lz4":
		return protos.CompressionType_kLZ4Compression
	case "snappy":
		return protos.CompressionType_kSnappyCompression
	case "zlib":
		return protos.CompressionType_kZlibCompression
	default:
		return protos.CompressionType_kNoCompression
	}
}

func getChecksumType(checksumStr string) protos.ChecksumType {
	switch checksumStr {
	case "crc32":
		return protos.ChecksumType_kCRC32
	case "sha1":
		return protos.ChecksumType_kSHA1
	case "sha256":
		return protos.ChecksumType_kSHA256
	default:
		return protos.ChecksumType_kNoChecksum
	}
}

func vdiskStreamRead(serverAddress, authToken string) error {
	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)
	
	ctx := context.Background()
	if authToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	stream, err := client.VDiskStreamRead(ctx)
	if err != nil {
		return fmt.Errorf("failed to create read stream: %v", err)
	}

	// Send read request
	readReq := &protos.VDiskReadArg{
		DiskId:          createDiskIdentifier(),
		Offset:          readOffset,
		Length:          readLength,
		MaxResponseSize: maxResponseSize,
	}

	fmt.Printf("Sending read request: offset=%d, length=%d\n", *readOffset, *readLength)
	err = stream.Send(readReq)
	if err != nil {
		return fmt.Errorf("failed to send read request: %v", err)
	}

	// Close the send side of the stream
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("failed to close send stream: %v", err)
	}

	// Receive responses
	fmt.Println("Receiving read responses...")
	totalBytesRead := 0
	responseCount := 0
	
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream completed normally")
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %v", err)
		}

		responseCount++
		fmt.Printf("Received response #%d\n", responseCount)

		// Check for error message - but don't treat success messages as errors
		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			// Only treat as error if it's not a success message
			if errorMsg != "" && errorMsg != "Read operation successful" && 
			   errorMsg != "Operation completed successfully" {
				return fmt.Errorf("server error: %s", errorMsg)
			}
			// Log success messages but don't treat as errors
			if errorMsg == "Read operation successful" || errorMsg == "Operation completed successfully" {
				fmt.Printf("Server message: %s\n", errorMsg)
			}
		}

		fmt.Printf("Response details: ranges=%d, data_size=%d bytes\n", 
			len(response.RangeVec), len(response.Data))
		
		// Display range information
		for i, dataRange := range response.RangeVec {
			fmt.Printf("  Range %d: offset=%d, length=%d, zero_data=%t\n", 
				i+1, *dataRange.Offset, *dataRange.Length, *dataRange.ZeroData)
		}

		// Display total disk size if available
		if response.TotalDiskSize != nil {
			fmt.Printf("  Total disk size: %d bytes\n", *response.TotalDiskSize)
		}

		totalBytesRead += len(response.Data)
		
		// Check if there's more data
		if response.HasMoreData != nil {
			fmt.Printf("  Has more data: %t\n", *response.HasMoreData)
			if !*response.HasMoreData {
				fmt.Println("No more data available")
				break
			}
		}
	}

	fmt.Printf("Read operation completed successfully. Total bytes read: %d, responses received: %d\n", 
		totalBytesRead, responseCount)
	return nil
}

func vdiskStreamWrite(serverAddress, authToken string) error {
	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)
	
	ctx := context.Background()
	if authToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	stream, err := client.VDiskStreamWrite(ctx)
	if err != nil {
		return fmt.Errorf("failed to create write stream: %v", err)
	}

	// Create write request
	writeReq := &protos.VDiskWriteArg{
		DiskId: createDiskIdentifier(),
		RangeVec: []*protos.DiskDataRange{
			{
				Offset:   writeOffset,
				Length:   writeLength,
				ZeroData: func() *bool { b := false; return &b }(),
			},
		},
		CompressionType: func() *protos.CompressionType { ct := getCompressionType(*compressionType); return &ct }(),
		ChecksumType:    func() *protos.ChecksumType { ct := getChecksumType(*checksumType); return &ct }(),
		Data:            []byte(*writeData), // In real scenario, would decode hex string
		SequenceNumber:  sequenceNumber,
	}

	fmt.Printf("Sending write request: offset=%d, length=%d, data_size=%d bytes\n", 
		*writeOffset, *writeLength, len(writeReq.Data))
	
	err = stream.Send(writeReq)
	if err != nil {
		return fmt.Errorf("failed to send write request: %v", err)
	}

	// Close the send side of the stream
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("failed to close send stream: %v", err)
	}

	// Receive responses
	fmt.Println("Receiving write responses...")
	responseCount := 0
	
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream completed normally")
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %v", err)
		}

		responseCount++
		fmt.Printf("Received response #%d\n", responseCount)

		// Check for error message - but don't treat success messages as errors
		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			// Only treat as error if it's not a success message
			if errorMsg != "" && errorMsg != "Write operation successful" && 
			   errorMsg != "Operation completed successfully" {
				return fmt.Errorf("server error: %s", errorMsg)
			}
			// Log success messages but don't treat as errors
			if errorMsg == "Write operation successful" || errorMsg == "Operation completed successfully" {
				fmt.Printf("Server message: %s\n", errorMsg)
			}
		}

		// Display response details
		success := false
		if response.Success != nil {
			success = *response.Success
		}
		
		offset := int64(0)
		if response.Offset != nil {
			offset = *response.Offset
		}
		
		length := int64(0)
		if response.Length != nil {
			length = *response.Length
		}
		
		bytesWritten := int64(0)
		if response.BytesWritten != nil {
			bytesWritten = *response.BytesWritten
		}
		
		sequenceNumber := int64(0)
		if response.SequenceNumber != nil {
			sequenceNumber = *response.SequenceNumber
		}
		
		fmt.Printf("Write response: success=%t, offset=%d, length=%d, bytes_written=%d, sequence=%d\n",
			success, offset, length, bytesWritten, sequenceNumber)
	}

	fmt.Printf("Write operation completed successfully. Responses received: %d\n", responseCount)
	return nil
}

func runVDiskOperation() error {
	if *vdiskServerAddress == "" {
		return fmt.Errorf("vdisk_server address is required")
	}

	if *vdiskOperation == "" {
		return fmt.Errorf("vdisk_operation is required (read or write)")
	}

	// Validate disk identifier
	if *diskRecoveryPointUuid == "" && *vmDiskUuid == "" && *vgDiskUuid == "" {
		return fmt.Errorf("one of disk_recovery_point_uuid, vm_disk_uuid, or vg_disk_uuid is required")
	}

	start := time.Now()
	var err error

	switch *vdiskOperation {
	case "read":
		err = vdiskStreamRead(*vdiskServerAddress, *vdiskAuthToken)
	case "write":
		if *writeData == "" {
			return fmt.Errorf("write_data is required for write operation")
		}
		err = vdiskStreamWrite(*vdiskServerAddress, *vdiskAuthToken)
	default:
		return fmt.Errorf("invalid vdisk_operation: %s (must be 'read' or 'write')", *vdiskOperation)
	}

	if err != nil {
		return fmt.Errorf("VDisk operation failed: %v", err)
	}

	fmt.Printf("VDisk %s operation completed in %v\n", *vdiskOperation, time.Since(start))
	return nil
}