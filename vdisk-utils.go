package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/vaibhav-ntnx/grpc-data-api-go-client/protos"
)

var (
	vdiskServerAddress = flag.String("vdisk_server", "", "VDisk server address in ip:port format")
	vdiskOperation     = flag.String("vdisk_operation", "", "VDisk operation (read or write)")
	vdiskAuthToken     = flag.String("vdisk_auth_token", "", "Authentication token for VDisk service")
	vdiskUseTLS        = flag.Bool("vdisk_use_tls", true, "Use TLS for gRPC connection (default: false)")
	vdiskSkipTLSVerify = flag.Bool("vdisk_skip_tls_verify", true, "Skip TLS certificate verification (default: true)")

	// Authentication flags
	authType       = flag.String("auth_type", "cookie", "Authentication type (cookie, bearer, basic)")
	cookieValue    = flag.String("cookie_value", "NTNX_IGW_SESSION=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1c2VyX3Byb2ZpbGUiOiJ7XCJfcGVybWFuZW50XCI6IHRydWUsIFwiYXV0aGVudGljYXRlZFwiOiB0cnVlLCBcImFwcF9kYXRhXCI6IHt9LCBcInVzZXJuYW1lXCI6IFwiYWRtaW5cIiwgXCJkb21haW5cIjogbnVsbCwgXCJ1c2VydHlwZVwiOiBcImxvY2FsXCIsIFwibGVnYWN5X2FkbWluX2F1dGhvcml0aWVzXCI6IFtcIlJPTEVfQ0xVU1RFUl9BRE1JTlwiLCBcIlJPTEVfQ0xVU1RFUl9WSUVXRVJcIiwgXCJST0xFX1VTRVJfQURNSU5cIl0sIFwicm9sZXNcIjogW1wiUHJpc20gQWRtaW5cIiwgXCJQcmlzbSBWaWV3ZXJcIiwgXCJTdXBlciBBZG1pblwiXSwgXCJhdXRoX2luZm9cIjoge1widGVuYW50X3V1aWRcIjogXCIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDBcIiwgXCJzZXJ2aWNlX25hbWVcIjogbnVsbCwgXCJ1c2VyX3V1aWRcIjogXCIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDBcIiwgXCJ0b2tlbl9hdWRpZW5jZVwiOiBudWxsLCBcInRva2VuX2lzc3VlclwiOiBudWxsLCBcInVzZXJuYW1lXCI6IFwiYWRtaW5cIiwgXCJyZW1vdGVfYXV0aG9yaXphdGlvblwiOiBudWxsLCBcInJlbW90ZV9hdXRoX2pzb25cIjogbnVsbCwgXCJ3b3JrZmxvd190eXBlXCI6IG51bGwsIFwidXNlcl9ncm91cF91dWlkc1wiOiBudWxsfSwgXCJsb2dfdXVpZFwiOiBcImUxYjBhYzAyLWFiMjYtNGI4ZS1iMGU4LWQ1ZmM4YWQ0NTViNFwiLCBcIm9yaWdfdG9rZW5faXNzX3RpbWVcIjogMTc1MzY5NDM1OH0iLCJqdGkiOiI3ZTVmMmY0OC04NmI3LTM1ZDItOWI4My1jODI2Mzg4OGU1ZWUiLCJpc3MiOiJBdGhlbmEiLCJpYXQiOjE3NTM2OTQzNTgsImV4cCI6MTc1MzY5NTI1OH0.IKoCJit58Wa2VwDSGGE7WXQPtS9SoP4Kl2efen4nau-3Wkp9Tk1b_-7rirQI6X-Zm3fsXWFfXrcDHay1JlFm1rYl2OL1gf75-N3RfJOo1odetuxo4cQYpXpvKI7LyFFCoMxncRSMXfRylZ_Yb7i_rmqRo6CdxI3UJhNplw0nl6QlfPuNbb9TKgKRIsJCQcQoZAIl_C-XvLFuBC_-RaQijVEuDYta-A3u9g9kcmuFDnFmAsTN1DWNy9CXrusHBZp3VGPtdk3lmJ0g_HQiNbZO9-xAZrkcuSFqMBMhiyRsJ6w4ZgIK8appMoToQFeX1J3RnFRaErYrClxs6hmAnxUEew;NTNX_MERCURY_IGW_SESSION=CgVhZG1pbhCaiJ3EBhoHTWVyY3VyeSAA|26y8+JOdk/pM1CWEYjQvwlQTWjZ1WcSYIrjnN8+jKkY=", "Cookie value for cookie authentication")
	basicAuthValue = flag.String("basic_auth_value", "YWRtaW46TnV0YW5peC4xMjM=", "Base64 encoded credentials for basic auth")

	// Batch operation flags
	batchMode  = flag.Bool("batch_mode", false, "Enable batch mode for concurrent operations")
	batchSize  = flag.Int("batch_size", 1, "Number of concurrent operations to run")
	batchDelay = flag.Duration("batch_delay", 0, "Delay between starting batch operations")

	// Throughput testing flags
	throughputMode = flag.Bool("throughput_mode", false, "Enable throughput testing mode")
	testDuration   = flag.Duration("test_duration", 10*time.Minute, "Duration to run throughput test")
	maxConcurrent  = flag.Int("max_concurrent", 10, "Maximum concurrent requests")
	reportInterval = flag.Duration("report_interval", 30*time.Second, "Interval for intermediate throughput reports")

	// Disk identifier flags
	diskRecoveryPointUuid = flag.String("disk_recovery_point_uuid", "", "Disk recovery point UUID")
	vmDiskUuid            = flag.String("vm_disk_uuid", "", "VM disk UUID")
	vgDiskUuid            = flag.String("vg_disk_uuid", "", "Volume group disk UUID")

	// Read operation flags
	readOffset      = flag.Int64("read_offset", 0, "Read offset in bytes")
	readLength      = flag.Int64("read_length", 0, "Read length in bytes (0 for entire disk)")
	maxResponseSize = flag.Int64("max_response_size", 1024*1024, "Maximum response size in bytes")

	// Write operation flags
	writeOffset     = flag.Int64("write_offset", 0, "Write offset in bytes")
	writeLength     = flag.Int64("write_length", 0, "Write length in bytes")
	writeData       = flag.String("write_data", "", "Data to write (hex string)")
	compressionType = flag.String("compression_type", "none", "Compression type (none, lz4, snappy, zlib)")
	checksumType    = flag.String("checksum_type", "none", "Checksum type (none, crc32, sha1, sha256)")
	sequenceNumber  = flag.Int64("sequence_number", 0, "Sequence number for write ordering")
)

// BatchOperationResult holds the result of a single operation in a batch
type BatchOperationResult struct {
	OperationID   int
	Success       bool
	Error         error
	Duration      time.Duration
	BytesRead     int
	BytesWritten  int64
	ResponseCount int
}

// ThroughputMetrics tracks throughput testing metrics
type ThroughputMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalBytes         int64
	TotalDuration      time.Duration
	StartTime          time.Time
	EndTime            time.Time
	MinLatency         time.Duration
	MaxLatency         time.Duration
	TotalLatency       time.Duration
	RequestsPerSecond  float64
	BytesPerSecond     float64
}

// ThroughputResult represents the result of a single throughput test operation
type ThroughputResult struct {
	Success      bool
	Error        error
	Duration     time.Duration
	BytesRead    int64
	BytesWritten int64
	Timestamp    time.Time
}

// Connection pool for throughput testing
var (
	connectionPool      sync.Map // map[string]*grpc.ClientConn
	connectionPoolMutex sync.RWMutex
	cachedAuthContext   context.Context
	authContextOnce     sync.Once
)

// getCachedAuthContext returns a cached authentication context for throughput testing
func getCachedAuthContext() context.Context {
	authContextOnce.Do(func() {
		cachedAuthContext = createAuthContext(context.Background())
	})
	return cachedAuthContext
}

// getPooledConnection gets or creates a pooled connection
func getPooledConnection(serverAddress string) (*grpc.ClientConn, error) {
	if conn, ok := connectionPool.Load(serverAddress); ok {
		if grpcConn, ok := conn.(*grpc.ClientConn); ok {
			// Check if connection is still valid
			if grpcConn.GetState().String() != "SHUTDOWN" {
				return grpcConn, nil
			}
			// Remove invalid connection
			connectionPool.Delete(serverAddress)
		}
	}

	// Create new connection
	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		return nil, err
	}

	// Store in pool
	connectionPool.Store(serverAddress, conn)
	return conn, nil
}

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

	// Add keepalive settings for better connection management
	kacp := keepalive.ClientParameters{
		Time:                30 * time.Second, // Send pings every 30 seconds
		Timeout:             5 * time.Second,  // Wait 5 seconds for ping ack
		PermitWithoutStream: true,             // Send pings even without active streams
	}
	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	fmt.Printf("Using authentication type: %s\n", *authType)

	conn, err := grpc.Dial(serverAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to VDisk server: %v", err)
	}

	return conn, nil
}

// createAuthContext creates a context with appropriate authentication headers
func createAuthContext(baseCtx context.Context) context.Context {
	ctx := baseCtx

	switch *authType {
	case "bearer":
		if *vdiskAuthToken != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", *vdiskAuthToken))
			fmt.Println("Using Bearer token authentication")
		} else {
			fmt.Println("Warning: Bearer auth selected but no token provided")
		}
	case "basic":
		if *basicAuthValue != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Basic %s", *basicAuthValue))
			fmt.Println("Using Basic authentication")
		} else {
			fmt.Println("Warning: Basic auth selected but no credentials provided")
		}
	case "cookie":
		if *cookieValue != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "Cookie", *cookieValue)
			fmt.Println("Using Cookie authentication")
		} else {
			fmt.Println("Warning: Cookie auth selected but no cookie provided")
		}
	default:
		fmt.Printf("Warning: Unknown authentication type '%s', proceeding without authentication\n", *authType)
	}

	return ctx
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

func vdiskStreamRead(serverAddress string) error {
	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)

	ctx := context.Background()
	ctx = createAuthContext(ctx)

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

	// Close the send side of the stream
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("failed to close send stream: %v", err)
	}
	return nil
}

func vdiskStreamWrite(serverAddress string) error {
	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)

	ctx := context.Background()
	ctx = createAuthContext(ctx)

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

// vdiskStreamReadSingle performs a single read operation and returns the result
func vdiskStreamReadSingle(serverAddress string, operationID int) BatchOperationResult {
	start := time.Now()
	result := BatchOperationResult{
		OperationID: operationID,
		Success:     false,
	}

	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)

	ctx := context.Background()
	ctx = createAuthContext(ctx)

	stream, err := client.VDiskStreamRead(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to create read stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Send read request with offset adjusted by operation ID for testing
	readReq := &protos.VDiskReadArg{
		DiskId:          createDiskIdentifier(),
		Offset:          func() *int64 { offset := *readOffset + int64(operationID)*1024; return &offset }(),
		Length:          readLength,
		MaxResponseSize: maxResponseSize,
	}

	fmt.Printf("[Op %d] Sending read request: offset=%d, length=%d\n", operationID, *readReq.Offset, *readLength)
	err = stream.Send(readReq)
	if err != nil {
		result.Error = fmt.Errorf("failed to send read request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Receive responses
	totalBytesRead := 0
	responseCount := 0

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Error = fmt.Errorf("stream error: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		responseCount++

		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			if errorMsg != "" && errorMsg != "Read operation successful" &&
				errorMsg != "Operation completed successfully" {
				result.Error = fmt.Errorf("server error: %s", errorMsg)
				result.Duration = time.Since(start)
				return result
			}
		}

		totalBytesRead += len(response.Data)

		if response.HasMoreData != nil && !*response.HasMoreData {
			break
		}
	}

	// Close the send side of the stream
	err = stream.CloseSend()
	if err != nil {
		result.Error = fmt.Errorf("failed to close send stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = true
	result.BytesRead = totalBytesRead
	result.ResponseCount = responseCount
	result.Duration = time.Since(start)

	fmt.Printf("[Op %d] Read completed: %d bytes, %d responses, %v\n",
		operationID, totalBytesRead, responseCount, result.Duration)

	return result
}

// vdiskStreamWriteSingle performs a single write operation and returns the result
func vdiskStreamWriteSingle(serverAddress string, operationID int) BatchOperationResult {
	start := time.Now()
	result := BatchOperationResult{
		OperationID: operationID,
		Success:     false,
	}

	conn, err := createVDiskGrpcChannel(serverAddress)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	defer conn.Close()

	client := protos.NewStargateVDiskRpcSvcClient(conn)

	ctx := context.Background()
	ctx = createAuthContext(ctx)

	stream, err := client.VDiskStreamWrite(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to create write stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create write request with offset adjusted by operation ID for testing
	writeReq := &protos.VDiskWriteArg{
		DiskId: createDiskIdentifier(),
		RangeVec: []*protos.DiskDataRange{
			{
				Offset:   func() *int64 { offset := *writeOffset + int64(operationID)*1024; return &offset }(),
				Length:   writeLength,
				ZeroData: func() *bool { b := false; return &b }(),
			},
		},
		CompressionType: func() *protos.CompressionType { ct := getCompressionType(*compressionType); return &ct }(),
		ChecksumType:    func() *protos.ChecksumType { ct := getChecksumType(*checksumType); return &ct }(),
		Data:            []byte(fmt.Sprintf("%s_op%d", *writeData, operationID)), // Unique data per operation
		SequenceNumber:  func() *int64 { seq := *sequenceNumber + int64(operationID); return &seq }(),
	}

	fmt.Printf("[Op %d] Sending write request: offset=%d, length=%d, sequence=%d\n",
		operationID, *writeReq.RangeVec[0].Offset, *writeLength, *writeReq.SequenceNumber)

	err = stream.Send(writeReq)
	if err != nil {
		result.Error = fmt.Errorf("failed to send write request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Close the send side of the stream
	err = stream.CloseSend()
	if err != nil {
		result.Error = fmt.Errorf("failed to close send stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Receive responses
	responseCount := 0
	var totalBytesWritten int64 = 0

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Error = fmt.Errorf("stream error: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		responseCount++

		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			if errorMsg != "" && errorMsg != "Write operation successful" &&
				errorMsg != "Operation completed successfully" {
				result.Error = fmt.Errorf("server error: %s", errorMsg)
				result.Duration = time.Since(start)
				return result
			}
		}

		if response.BytesWritten != nil {
			totalBytesWritten += *response.BytesWritten
		}
	}

	result.Success = true
	result.BytesWritten = totalBytesWritten
	result.ResponseCount = responseCount
	result.Duration = time.Since(start)

	fmt.Printf("[Op %d] Write completed: %d bytes written, %d responses, %v\n",
		operationID, totalBytesWritten, responseCount, result.Duration)

	return result
}

// runBatchVDiskOperations runs multiple VDisk operations concurrently
func runBatchVDiskOperations() error {
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

	if *vdiskOperation == "write" && *writeData == "" {
		return fmt.Errorf("write_data is required for write operation")
	}

	fmt.Printf("Starting batch %s operations: %d concurrent operations\n", *vdiskOperation, *batchSize)

	start := time.Now()
	results := make([]BatchOperationResult, *batchSize)

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Run operations concurrently using a for loop as requested
	for i := 0; i < *batchSize; i++ {
		wg.Add(1)

		go func(operationID int) {
			defer wg.Done()

			// Add delay between starting operations if specified
			if *batchDelay > 0 && operationID > 0 {
				time.Sleep(time.Duration(operationID) * *batchDelay)
			}

			var result BatchOperationResult
			switch *vdiskOperation {
			case "read":
				result = vdiskStreamReadSingle(*vdiskServerAddress, operationID)
			case "write":
				result = vdiskStreamWriteSingle(*vdiskServerAddress, operationID)
			}

			results[operationID] = result
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	totalDuration := time.Since(start)

	// Print summary results
	fmt.Printf("\n=== Batch Operation Summary ===\n")
	fmt.Printf("Total operations: %d\n", *batchSize)
	fmt.Printf("Total time: %v\n", totalDuration)

	successCount := 0
	totalBytes := int64(0)
	totalResponses := 0

	for _, result := range results {
		if result.Success {
			successCount++
			if *vdiskOperation == "read" {
				totalBytes += int64(result.BytesRead)
			} else {
				totalBytes += result.BytesWritten
			}
			totalResponses += result.ResponseCount
		} else {
			fmt.Printf("Operation %d failed: %v\n", result.OperationID, result.Error)
		}
	}

	fmt.Printf("Successful operations: %d/%d\n", successCount, *batchSize)
	fmt.Printf("Total bytes processed: %d\n", totalBytes)
	fmt.Printf("Total responses: %d\n", totalResponses)
	fmt.Printf("Average operation time: %v\n", totalDuration/time.Duration(*batchSize))

	if successCount != *batchSize {
		return fmt.Errorf("batch operation partially failed: %d/%d operations succeeded", successCount, *batchSize)
	}

	fmt.Printf("All batch operations completed successfully!\n")
	return nil
}

// runThroughputSingleOperation performs a single operation for throughput testing
func runThroughputSingleOperation(serverAddress string, operationID int64, semaphore chan struct{}) ThroughputResult {
	// Acquire semaphore slot
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	start := time.Now()
	result := ThroughputResult{
		Success:   false,
		Timestamp: start,
	}

	conn, err := getPooledConnection(serverAddress)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}
	// Don't close - reuse pooled connection

	client := protos.NewStargateVDiskRpcSvcClient(conn)

	ctx := getCachedAuthContext()

	switch *vdiskOperation {
	case "read":
		result = performThroughputRead(client, ctx, operationID, start)
	case "write":
		result = performThroughputWrite(client, ctx, operationID, start)
	default:
		result.Error = fmt.Errorf("invalid operation: %s", *vdiskOperation)
		result.Duration = time.Since(start)
	}

	return result
}

// performThroughputRead performs a read operation for throughput testing
func performThroughputRead(client protos.StargateVDiskRpcSvcClient, ctx context.Context, operationID int64, start time.Time) ThroughputResult {
	result := ThroughputResult{Timestamp: start}

	stream, err := client.VDiskStreamRead(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to create read stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Send read request with varying offset for load distribution
	readReq := &protos.VDiskReadArg{
		DiskId:          createDiskIdentifier(),
		Offset:          func() *int64 { offset := *readOffset + (operationID%1000)*1024; return &offset }(),
		Length:          readLength,
		MaxResponseSize: maxResponseSize,
	}

	err = stream.Send(readReq)
	if err != nil {
		result.Error = fmt.Errorf("failed to send read request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	var totalBytesRead int64 = 0
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Error = fmt.Errorf("stream error: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			if errorMsg != "" && errorMsg != "Read operation successful" &&
				errorMsg != "Operation completed successfully" {
				result.Error = fmt.Errorf("server error: %s", errorMsg)
				result.Duration = time.Since(start)
				return result
			}
		}

		totalBytesRead += int64(len(response.Data))

		if response.HasMoreData != nil && !*response.HasMoreData {
			break
		}
	}

	err = stream.CloseSend()
	if err != nil {
		result.Error = fmt.Errorf("failed to close send stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = true
	result.BytesRead = totalBytesRead
	result.Duration = time.Since(start)
	return result
}

// performThroughputWrite performs a write operation for throughput testing
func performThroughputWrite(client protos.StargateVDiskRpcSvcClient, ctx context.Context, operationID int64, start time.Time) ThroughputResult {
	result := ThroughputResult{Timestamp: start}

	stream, err := client.VDiskStreamWrite(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to create write stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create write request with varying offset and sequence for load distribution
	writeReq := &protos.VDiskWriteArg{
		DiskId: createDiskIdentifier(),
		RangeVec: []*protos.DiskDataRange{
			{
				Offset:   func() *int64 { offset := *writeOffset + (operationID%1000)*1024; return &offset }(),
				Length:   writeLength,
				ZeroData: func() *bool { b := false; return &b }(),
			},
		},
		CompressionType: func() *protos.CompressionType { ct := getCompressionType(*compressionType); return &ct }(),
		ChecksumType:    func() *protos.ChecksumType { ct := getChecksumType(*checksumType); return &ct }(),
		Data:            []byte(fmt.Sprintf("%s_throughput_%d", *writeData, operationID)),
		SequenceNumber:  func() *int64 { seq := *sequenceNumber + operationID; return &seq }(),
	}

	err = stream.Send(writeReq)
	if err != nil {
		result.Error = fmt.Errorf("failed to send write request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	err = stream.CloseSend()
	if err != nil {
		result.Error = fmt.Errorf("failed to close send stream: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	var totalBytesWritten int64 = 0
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Error = fmt.Errorf("stream error: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		if response.ErrorMessage != nil {
			errorMsg := *response.ErrorMessage
			if errorMsg != "" && errorMsg != "Write operation successful" &&
				errorMsg != "Operation completed successfully" {
				result.Error = fmt.Errorf("server error: %s", errorMsg)
				result.Duration = time.Since(start)
				return result
			}
		}

		if response.BytesWritten != nil {
			totalBytesWritten += *response.BytesWritten
		}
	}

	result.Success = true
	result.BytesWritten = totalBytesWritten
	result.Duration = time.Since(start)
	return result
}

// runThroughputTest runs continuous throughput testing for specified duration
func runThroughputTest() error {
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

	if *vdiskOperation == "write" && *writeData == "" {
		return fmt.Errorf("write_data is required for write operation")
	}

	fmt.Printf("Starting throughput test: %s operations for %v\n", *vdiskOperation, *testDuration)
	fmt.Printf("Max concurrent requests: %d\n", *maxConcurrent)
	fmt.Printf("Report interval: %v\n", *reportInterval)

	// Metrics tracking
	var metrics ThroughputMetrics
	metrics.StartTime = time.Now()
	metrics.MinLatency = time.Hour // Initialize to high value

	// Semaphore to limit concurrent requests
	semaphore := make(chan struct{}, *maxConcurrent)

	// Channel for results
	resultChan := make(chan ThroughputResult, *maxConcurrent*2)

	// WaitGroup for goroutines
	var wg sync.WaitGroup

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *testDuration)
	defer cancel()

	// Operation counter
	var operationID int64 = 0

	// Start result processor goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		processThroughputResults(resultChan, &metrics)
	}()

	// Start periodic reporting goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		reportThroughputPeriodically(ctx, &metrics)
	}()

	// Main request generation loop
	requestTicker := time.NewTicker(1 * time.Millisecond) // Start new requests frequently
	defer requestTicker.Stop()

	fmt.Println("Throughput test started...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Test duration completed, stopping new requests...")
			goto cleanup
		case <-requestTicker.C:
			// Launch new request if we have capacity
			select {
			case semaphore <- struct{}{}:
				opID := atomic.AddInt64(&operationID, 1)
				go func(id int64) {
					result := runThroughputSingleOperation(*vdiskServerAddress, id, semaphore)
					resultChan <- result
				}(opID)
			default:
				// Semaphore full, skip this tick
			}
		}
	}

cleanup:
	// Wait for remaining operations to complete (with timeout)
	done := make(chan struct{})
	go func() {
		// Wait for semaphore to be empty (all operations done)
		for i := 0; i < *maxConcurrent; i++ {
			semaphore <- struct{}{}
		}
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All operations completed")
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout waiting for operations to complete")
	}

	// Signal completion and wait for processors
	close(resultChan)
	cancel()
	wg.Wait()

	// Final metrics calculation
	metrics.EndTime = time.Now()
	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)

	if metrics.TotalRequests > 0 {
		metrics.RequestsPerSecond = float64(metrics.TotalRequests) / metrics.TotalDuration.Seconds()
		metrics.BytesPerSecond = float64(metrics.TotalBytes) / metrics.TotalDuration.Seconds()
	}

	// Print final results
	printFinalThroughputResults(&metrics)

	// Cleanup connection pool
	fmt.Println("Cleaning up connection pool...")
	cleanupConnectionPool()

	return nil
}

// processThroughputResults processes incoming results and updates metrics
func processThroughputResults(resultChan <-chan ThroughputResult, metrics *ThroughputMetrics) {
	for result := range resultChan {
		atomic.AddInt64(&metrics.TotalRequests, 1)

		if result.Success {
			atomic.AddInt64(&metrics.SuccessfulRequests, 1)
			atomic.AddInt64(&metrics.TotalBytes, result.BytesRead+result.BytesWritten)

			// Update latency metrics (not thread-safe, but close enough for reporting)
			if metrics.MinLatency == 0 || result.Duration < metrics.MinLatency {
				metrics.MinLatency = result.Duration
			}
			if result.Duration > metrics.MaxLatency {
				metrics.MaxLatency = result.Duration
			}
			metrics.TotalLatency += result.Duration
		} else {
			atomic.AddInt64(&metrics.FailedRequests, 1)
			fmt.Printf("Operation failed: %v\n", result.Error)
		}
	}
}

// reportThroughputPeriodically prints intermediate throughput reports
func reportThroughputPeriodically(ctx context.Context, metrics *ThroughputMetrics) {
	ticker := time.NewTicker(*reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			printIntermediateThroughputReport(metrics)
		}
	}
}

// printIntermediateThroughputReport prints current throughput metrics
func printIntermediateThroughputReport(metrics *ThroughputMetrics) {
	totalReqs := atomic.LoadInt64(&metrics.TotalRequests)
	successReqs := atomic.LoadInt64(&metrics.SuccessfulRequests)
	failedReqs := atomic.LoadInt64(&metrics.FailedRequests)
	totalBytes := atomic.LoadInt64(&metrics.TotalBytes)

	elapsed := time.Since(metrics.StartTime)

	var rps, bps float64
	if elapsed.Seconds() > 0 {
		rps = float64(totalReqs) / elapsed.Seconds()
		bps = float64(totalBytes) / elapsed.Seconds()
	}

	fmt.Printf("\n=== Intermediate Throughput Report (Elapsed: %v) ===\n", elapsed.Truncate(time.Second))
	fmt.Printf("Total Requests: %d\n", totalReqs)
	fmt.Printf("Successful: %d, Failed: %d\n", successReqs, failedReqs)
	fmt.Printf("Requests/sec: %.2f\n", rps)
	fmt.Printf("Bytes/sec: %.2f (%.2f MB/s)\n", bps, bps/(1024*1024))
	fmt.Printf("Total Data: %d bytes (%.2f MB)\n", totalBytes, float64(totalBytes)/(1024*1024))

	if successReqs > 0 {
		avgLatency := metrics.TotalLatency / time.Duration(successReqs)
		fmt.Printf("Avg Latency: %v\n", avgLatency)
		fmt.Printf("Min Latency: %v, Max Latency: %v\n", metrics.MinLatency, metrics.MaxLatency)
	}
	fmt.Println("========================================")
}

// printFinalThroughputResults prints comprehensive final throughput results
func printFinalThroughputResults(metrics *ThroughputMetrics) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("FINAL THROUGHPUT TEST RESULTS\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	fmt.Printf("Test Duration: %v\n", metrics.TotalDuration.Truncate(time.Second))
	fmt.Printf("Operation Type: %s\n", *vdiskOperation)
	fmt.Printf("Max Concurrent: %d\n", *maxConcurrent)

	fmt.Printf("\nRequest Statistics:\n")
	fmt.Printf("  Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successful: %d (%.2f%%)\n", metrics.SuccessfulRequests,
		float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	fmt.Printf("  Failed: %d (%.2f%%)\n", metrics.FailedRequests,
		float64(metrics.FailedRequests)/float64(metrics.TotalRequests)*100)

	fmt.Printf("\nThroughput Metrics:\n")
	fmt.Printf("  Requests/sec: %.2f\n", metrics.RequestsPerSecond)
	fmt.Printf("  Bytes/sec: %.2f\n", metrics.BytesPerSecond)
	fmt.Printf("  MB/sec: %.2f\n", metrics.BytesPerSecond/(1024*1024))
	fmt.Printf("  GB/sec: %.4f\n", metrics.BytesPerSecond/(1024*1024*1024))

	fmt.Printf("\nData Transfer:\n")
	fmt.Printf("  Total Bytes: %d\n", metrics.TotalBytes)
	fmt.Printf("  Total MB: %.2f\n", float64(metrics.TotalBytes)/(1024*1024))
	fmt.Printf("  Total GB: %.4f\n", float64(metrics.TotalBytes)/(1024*1024*1024))

	if metrics.SuccessfulRequests > 0 {
		avgLatency := metrics.TotalLatency / time.Duration(metrics.SuccessfulRequests)
		fmt.Printf("\nLatency Statistics:\n")
		fmt.Printf("  Average: %v\n", avgLatency)
		fmt.Printf("  Minimum: %v\n", metrics.MinLatency)
		fmt.Printf("  Maximum: %v\n", metrics.MaxLatency)

		fmt.Printf("\nEfficiency Metrics:\n")
		fmt.Printf("  Avg bytes per request: %.2f\n", float64(metrics.TotalBytes)/float64(metrics.SuccessfulRequests))
		fmt.Printf("  Requests per minute: %.2f\n", metrics.RequestsPerSecond*60)
		fmt.Printf("  MB per minute: %.2f\n", metrics.BytesPerSecond*60/(1024*1024))
	}

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
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

	// Check if throughput mode is enabled
	if *throughputMode {
		return runThroughputTest()
	}

	// Check if batch mode is enabled
	if *batchMode {
		return runBatchVDiskOperations()
	}

	start := time.Now()
	var err error

	switch *vdiskOperation {
	case "read":
		err = vdiskStreamRead(*vdiskServerAddress)
	case "write":
		if *writeData == "" {
			return fmt.Errorf("write_data is required for write operation")
		}
		err = vdiskStreamWrite(*vdiskServerAddress)
	default:
		return fmt.Errorf("invalid vdisk_operation: %s (must be 'read' or 'write')", *vdiskOperation)
	}

	if err != nil {
		return fmt.Errorf("VDisk operation failed: %v", err)
	}

	fmt.Printf("VDisk %s operation completed in %v\n", *vdiskOperation, time.Since(start))
	return nil
}

// cleanupConnectionPool closes all pooled connections
func cleanupConnectionPool() {
	connectionPool.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*grpc.ClientConn); ok {
			conn.Close()
		}
		connectionPool.Delete(key)
		return true
	})
}
