#!/bin/bash

# Usage: ./iscsi_read_verify_throughput.sh
# Optimized for 10GB write/read throughput testing

set -e

# ========== Configuration ==========
TARGET_IP="10.33.8.46"
IQN="iqn.2010-06.com.nutanix:vg1-3740cfaa-3460-4d65-9dfb-f66022ef95e4-tgt0"
MOUNT_DIR="/mnt/iscsi_test"
BLOCK_SIZE=1M
TOTAL_SIZE_GB=10           # 10GB of data
TOTAL_SIZE_MB=$((TOTAL_SIZE_GB * 1024))  # 10240 MB
CHUNK_SIZE_MB=512          # Read in 512MB chunks for progress monitoring

# ========== iSCSI Login ==========
echo "[*] Logging into iSCSI target..."
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --login || true
sleep 5

# ========== Detect iSCSI Device ==========
DEVICE=$(lsblk -ndo NAME,TYPE | grep disk | awk '{print $1}' | tail -n1)
DEVICE="/dev/$DEVICE"

if [ ! -b "$DEVICE" ]; then
  echo "[!] iSCSI device not found. Exiting."
  exit 1
fi

echo "[*] Using device: $DEVICE"

# ========== Partition, Format, Mount ==========
echo "[*] Setting up device $DEVICE..."
echo -e "o\nn\np\n1\n\n\nw" | sudo fdisk "$DEVICE" >/dev/null 2>&1
sleep 3
sudo mkfs.ext4 -F "${DEVICE}1" >/dev/null
sudo mkdir -p "$MOUNT_DIR"
sudo mount "${DEVICE}1" "$MOUNT_DIR"

# ========== Write 10GB data ==========
echo "[*] Writing ${TOTAL_SIZE_GB}GB test file..."
WRITE_START=$(date +%s.%N)
sudo dd if=/dev/urandom of="$MOUNT_DIR/testfile" bs=$BLOCK_SIZE count=$TOTAL_SIZE_MB oflag=direct status=progress
sync
WRITE_END=$(date +%s.%N)
WRITE_TIME=$(echo "$WRITE_END - $WRITE_START" | bc)
WRITE_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $WRITE_TIME" | bc)

echo "[✓] Write completed:"
echo "    Size: ${TOTAL_SIZE_GB}GB (${TOTAL_SIZE_MB}MB)"
echo "    Time: ${WRITE_TIME}s"
echo "    Throughput: ${WRITE_THROUGHPUT} MB/s"
echo ""

# ========== Read 10GB data with throughput monitoring ==========
echo "[*] Starting optimized ${TOTAL_SIZE_GB}GB read test..."
echo "timestamp,chunk_mb,elapsed_sec,instant_mbps,cumulative_mbps"

READ_START=$(date +%s.%N)
TOTAL_READ_MB=0
CHUNK_COUNT=$((TOTAL_SIZE_MB / CHUNK_SIZE_MB))

# Create a function to read in chunks and monitor progress
read_with_monitoring() {
    local chunk_start chunk_end chunk_time instant_mbps cumulative_mbps elapsed_time
    
    for ((i=1; i<=CHUNK_COUNT; i++)); do
        chunk_start=$(date +%s.%N)
        
        # Read chunk using dd with skip and count
        sudo dd if="$MOUNT_DIR/testfile" of=/dev/null bs=1M skip=$((CHUNK_SIZE_MB * (i-1))) count=$CHUNK_SIZE_MB iflag=direct status=none 2>/dev/null
        
        chunk_end=$(date +%s.%N)
        chunk_time=$(echo "$chunk_end - $chunk_start" | bc)
        elapsed_time=$(echo "$chunk_end - $READ_START" | bc)
        
        TOTAL_READ_MB=$((TOTAL_READ_MB + CHUNK_SIZE_MB))
        
        # Calculate instant and cumulative throughput
        instant_mbps=$(echo "scale=2; $CHUNK_SIZE_MB / $chunk_time" | bc)
        cumulative_mbps=$(echo "scale=2; $TOTAL_READ_MB / $elapsed_time" | bc)
        
        echo "$(date +%s),$CHUNK_SIZE_MB,$elapsed_time,$instant_mbps,$cumulative_mbps"
    done
}

# Execute the monitored read
read_with_monitoring

READ_END=$(date +%s.%N)
TOTAL_READ_TIME=$(echo "$READ_END - $READ_START" | bc)
TOTAL_READ_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $TOTAL_READ_TIME" | bc)

# ========== Final Read Stats ==========
echo ""
echo "[✓] Read test completed."
echo "=============================="
echo "Total data read    : ${TOTAL_SIZE_GB}GB (${TOTAL_SIZE_MB}MB)"
echo "Total read time    : ${TOTAL_READ_TIME}s"
echo "Average throughput : ${TOTAL_READ_THROUGHPUT} MB/s"
echo "=============================="

# ========== Performance Summary ==========
echo ""
echo "PERFORMANCE SUMMARY"
echo "==================="
echo "Write: ${WRITE_THROUGHPUT} MB/s"
echo "Read:  ${TOTAL_READ_THROUGHPUT} MB/s"

# ========== Cleanup ==========
echo ""
echo "[*] Cleaning up..."
sudo umount "$MOUNT_DIR"
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --logout || true
sudo rm -rf "$MOUNT_DIR"

echo "[✓] Done."

