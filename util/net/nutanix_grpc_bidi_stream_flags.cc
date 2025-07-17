#include "gflags/gflags.h"

DEFINE_string(grpc_bidi_stream_experimental_client_ip_in_metadata, "",
              "Experimental flag to set client IP in metadata");

DEFINE_int32(grpc_write_retry_max_attempts, 3,
             "Maximum number of retry attempts for failed writes");

DEFINE_int32(grpc_write_retry_initial_delay_ms, 100,
             "Initial delay in milliseconds between write retries"); 