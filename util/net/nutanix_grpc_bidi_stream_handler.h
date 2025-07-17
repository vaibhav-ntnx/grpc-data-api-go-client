#ifndef _UTIL_NET_NUTANIX_GRPC_BIDI_STREAM_HANDLER_H_
#define _UTIL_NET_NUTANIX_GRPC_BIDI_STREAM_HANDLER_H_

#include <grpcpp/impl/codegen/async_stream.h>
#include <type_traits>
#include <queue>
#include <chrono>

#include "util/base/basictypes.h"
#include "util/net/ip_util.h"
#include "util/net/nutanix_grpc_event_handler.h"

DECLARE_string(grpc_bidi_stream_experimental_client_ip_in_metadata);
DECLARE_int32(grpc_write_retry_max_attempts);
DECLARE_int32(grpc_write_retry_initial_delay_ms);

namespace nutanix { namespace net {

const string source_ip_str = "source-ip";

// Structure to hold queued write operations
template<typename WriteType>
struct QueuedWrite {
  shared_ptr<const WriteType> write_val;
  int retry_count;
  std::chrono::steady_clock::time_point next_retry;

  QueuedWrite(shared_ptr<const WriteType>&& val) 
    : write_val(std::move(val)), 
      retry_count(0),
      next_retry(std::chrono::steady_clock::now()) {}
};

// ... existing code ...

template<typename RpcArg, typename RpcRet, typename StreamType>
class NutanixGrpcBidiStreamHandler :
  public std::enable_shared_from_this<NutanixGrpcBidiStreamHandler<
    RpcArg, RpcRet, StreamType>> {
 public:
  // ... existing typedefs and enums ...

  enum class WriteStatus {
    kNoError = 0,
    kConcurrentWriteInProgress,
    kAborted,
    kWriteEndClosed,
    kQueuedForRetry  // New status for queued writes
  };

 protected:
  // ... existing protected members ...

  // Queue for pending writes
  std::queue<QueuedWrite<WriteType>> write_queue_;

  // Maximum number of write retry attempts
  const int max_write_retry_attempts_;

  // Initial delay between write retries in milliseconds
  const int write_retry_initial_delay_ms_;

  // Timer for write retries
  std::chrono::steady_clock::time_point next_write_retry_;

  // Process the next write in queue
  void ProcessNextWrite() {
    thread::ScopedSpinRWLocker sl(&spin_rw_lock_);
    
    if (write_in_progress_ || write_queue_.empty() || 
        writes_done_ || stream_cancelled_ || stream_finished_) {
      return;
    }

    auto now = std::chrono::steady_clock::now();
    auto& next_write = write_queue_.front();

    if (now < next_write.next_retry) {
      // Not time to retry yet
      return;
    }

    write_in_progress_ = true;
    write_val_ = next_write.write_val;
    
    TAGGED_VLOG(2) << "Processing queued write, retry count: " 
                   << next_write.retry_count;
    
    stream_->Write(*write_val_.get(),
                   reinterpret_cast<void *>(write_event_.get()));
    
    write_queue_.pop();
  }

  // Schedule next write retry with exponential backoff
  void ScheduleWriteRetry(shared_ptr<const WriteType>&& write_val, 
                         int retry_count) {
    if (retry_count >= max_write_retry_attempts_) {
      TAGGED_LOG(ERROR) << "Max write retries exceeded, failing write";
      stream_write_done_cb_(WriteStatus::kAborted, std::move(write_val));
      return;
    }

    QueuedWrite<WriteType> queued_write(std::move(write_val));
    queued_write.retry_count = retry_count;
    
    // Exponential backoff
    int delay_ms = write_retry_initial_delay_ms_ * (1 << retry_count);
    queued_write.next_retry = std::chrono::steady_clock::now() + 
                             std::chrono::milliseconds(delay_ms);

    write_queue_.push(std::move(queued_write));
    
    TAGGED_VLOG(2) << "Scheduled write retry #" << retry_count 
                   << " with delay " << delay_ms << "ms";
  }

  // Modified WriteDoneEvent to handle retries
  void WriteDoneEvent(const bool ok) {
    DCHECK(write_val_);
    DCHECK(stream_write_done_cb_);
    CHECK(write_in_progress_);
    TAGGED_VLOG(4) << "Write done with ok-status: " << std::boolalpha << ok;

    if (!ok) {
      TAGGED_LOG(ERROR) << "Write failed, will retry if attempts remain";
      
      thread::ScopedSpinRWLocker sl(&spin_rw_lock_);
      write_in_progress_ = false;

      // Get retry count for current write
      int retry_count = 0;
      if (!write_queue_.empty()) {
        retry_count = write_queue_.front().retry_count + 1;
      }

      // Schedule retry
      ScheduleWriteRetry(std::move(write_val_), retry_count);
      
      // Process next write if any
      ProcessNextWrite();
      return;
    }

    shared_ptr<const WriteType> tmp_write_val = std::move(write_val_);

    thread::ScopedSpinRWLocker sl(&spin_rw_lock_);
    write_in_progress_ = false;

    // Process next write in queue if any
    ProcessNextWrite();

    stream_write_done_cb_(WriteStatus::kNoError, std::move(tmp_write_val));
  }

  // Modified WriteToStream to use write queue
  WriteStatus WriteToStream(shared_ptr<const WriteType>&& write_val) {
    DCHECK(write_val);
    DCHECK(stream_write_done_cb_);

    thread::ScopedSpinRWLocker sl(&spin_rw_lock_);
    
    if (writes_done_ || stream_cancelled_ || stream_finished_) {
      TAGGED_LOG(ERROR) << "Cannot Write to a stream that is already marked "
                        << "with writes done " << OUTVARS(writes_done_,
                                                        stream_cancelled_,
                                                        stream_finished_);
      return WriteStatus::kWriteEndClosed;
    }

    if (write_in_progress_) {
      // Queue the write instead of failing
      QueuedWrite<WriteType> queued_write(std::move(write_val));
      write_queue_.push(std::move(queued_write));
      
      TAGGED_VLOG(2) << "Write queued, current queue size: " 
                     << write_queue_.size();
      
      return WriteStatus::kQueuedForRetry;
    }

    DCHECK(!write_val_);
    DCHECK(write_event_);

    write_in_progress_ = true;
    write_val_ = std::move(write_val);
    stream_->Write(*write_val_.get(),
                   reinterpret_cast<void *>(write_event_.get()));
    
    return WriteStatus::kNoError;
  }

  // Constructor updated to initialize retry parameters
  NutanixGrpcBidiStreamHandler(
    NutanixGrpcEventHandler *const grpc_event_handler,
    StreamStartFunc&& stream_start_func,
    HandleStreamConnectionFunc&& stream_connection_cb,
    HandleStreamReadFunc&& stream_read_cb,
    HandleStreamClosedFunc&& stream_closed_cb,
    StreamWriteDoneFunc&& stream_write_done_cb,
    const string& stream_identifier) :
    grpc_event_handler_(grpc_event_handler),
    read_val_(nullptr),
    write_val_(nullptr),
    stream_(nullptr),
    finish_status_(grpc::Status::OK),
    stream_start_func_(move(stream_start_func)),
    stream_connection_cb_(move(stream_connection_cb)),
    stream_read_cb_(move(stream_read_cb)),
    stream_closed_cb_(move(stream_closed_cb)),
    stream_write_done_cb_(move(stream_write_done_cb)),
    stream_identifier_(stream_identifier),
    reads_done_(false),
    writes_done_(false),
    write_in_progress_(false),
    stream_finished_(false),
    stream_cancelled_(false),
    reads_paused_(false),
    read_scheduled_(false),
    max_write_retry_attempts_(FLAGS_grpc_write_retry_max_attempts),
    write_retry_initial_delay_ms_(FLAGS_grpc_write_retry_initial_delay_ms) { }

  // ... rest of the existing code ...
};

// ... rest of the file ... 