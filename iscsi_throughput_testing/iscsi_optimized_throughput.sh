#!/bin/bash

# Optimized iSCSI throughput testing - targets network bandwidth levels
# Addresses: queue depth, multiple sessions, block sizes, network tuning

set -e

# ========== Configuration ==========
TARGET_IP="10.37.130.67"
IQN="iqn.2010-06.com.nutanix:testvg2-d708fe79-aec3-404f-a863-8b4987f88ad0"
MOUNT_DIR="/mnt/iscsi_test"
TOTAL_SIZE_GB=10
TOTAL_SIZE_MB=$((TOTAL_SIZE_GB * 1024))

# Performance parameters
ISCSI_QUEUE_DEPTH=128    # Increase from default 32
NUM_SESSIONS=4           # Multiple sessions for parallelism
BLOCK_SIZE_MB=64         # Larger blocks for better throughput
IO_ENGINE="libaio"       # Use async I/O

# Network tuning parameters
TCP_WINDOW_SIZE=16777216  # 16MB TCP window
TCP_RMEM="4096 65536 16777216"
TCP_WMEM="4096 65536 16777216"

echo "[*] Starting optimized iSCSI throughput test..."
echo "Target: ~940 MB/s (network bandwidth limit)"

# ========== System Optimization ==========
echo "[*] Applying system optimizations..."

# Backup original network settings
echo "[*] Backing up original network settings..."
ORIG_TCP_RMEM=$(cat /proc/sys/net/ipv4/tcp_rmem)
ORIG_TCP_WMEM=$(cat /proc/sys/net/ipv4/tcp_wmem)
ORIG_TCP_WINDOW_SCALING=$(cat /proc/sys/net/ipv4/tcp_window_scaling)

# Apply network optimizations
echo "$TCP_RMEM" | sudo tee /proc/sys/net/ipv4/tcp_rmem > /dev/null
echo "$TCP_WMEM" | sudo tee /proc/sys/net/ipv4/tcp_wmem > /dev/null
echo "1" | sudo tee /proc/sys/net/ipv4/tcp_window_scaling > /dev/null
echo "1" | sudo tee /proc/sys/net/ipv4/tcp_timestamps > /dev/null

# ========== iSCSI Optimization ==========
echo "[*] Configuring iSCSI parameters..."

# Ensure iSCSI target is discovered
sudo iscsiadm -m discovery -t sendtargets -p "$TARGET_IP" || true

# Set iSCSI session parameters BEFORE login
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" -o update -n node.session.queue_depth -v $ISCSI_QUEUE_DEPTH
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" -o update -n node.conn[0].iscsi.MaxRecvDataSegmentLength -v 262144
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" -o update -n node.conn[0].iscsi.MaxXmitDataSegmentLength -v 262144
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" -o update -n node.session.nr_sessions -v $NUM_SESSIONS

# Login with optimized parameters
echo "[*] Logging into iSCSI target with $NUM_SESSIONS sessions..."
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --login || true
sleep 10

# ========== Device Detection ==========
echo "[*] Detecting iSCSI devices..."
DEVICES=()
for i in $(seq 1 10); do
    DEVICE=$(lsblk -ndo NAME,TYPE | grep disk | awk '{print $1}' | tail -n1)
    if [ -n "$DEVICE" ]; then
        DEVICES+=("/dev/$DEVICE")
        break
    fi
    sleep 1
done

if [ ${#DEVICES[@]} -eq 0 ]; then
    echo "[!] No iSCSI devices found. Exiting."
    exit 1
fi

MAIN_DEVICE="${DEVICES[0]}"
echo "[*] Using device: $MAIN_DEVICE"

# ========== Raw Device Test (No Filesystem) ==========
echo "[*] Running RAW device performance test..."
echo "This bypasses filesystem overhead for maximum throughput"

# Raw write test
echo "[*] Raw write test (${TOTAL_SIZE_GB}GB)..."
RAW_WRITE_START=$(date +%s.%N)
sudo dd if=/dev/zero of="$MAIN_DEVICE" bs=${BLOCK_SIZE_MB}M count=$((TOTAL_SIZE_MB / BLOCK_SIZE_MB)) oflag=direct status=progress
RAW_WRITE_END=$(date +%s.%N)
RAW_WRITE_TIME=$(echo "$RAW_WRITE_END - $RAW_WRITE_START" | bc)
RAW_WRITE_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $RAW_WRITE_TIME" | bc)

echo "[✓] Raw write: ${RAW_WRITE_THROUGHPUT} MB/s"

# Raw read test  
echo "[*] Raw read test (${TOTAL_SIZE_GB}GB)..."
RAW_READ_START=$(date +%s.%N)
sudo dd if="$MAIN_DEVICE" of=/dev/null bs=${BLOCK_SIZE_MB}M count=$((TOTAL_SIZE_MB / BLOCK_SIZE_MB)) iflag=direct status=progress
RAW_READ_END=$(date +%s.%N)
RAW_READ_TIME=$(echo "$RAW_READ_END - $RAW_READ_START" | bc)
RAW_READ_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $RAW_READ_TIME" | bc)

echo "[✓] Raw read: ${RAW_READ_THROUGHPUT} MB/s"

# ========== Filesystem Test ==========
echo "[*] Setting up optimized filesystem test..."

# Create partition and filesystem with performance options
echo -e "o\nn\np\n1\n\n\nw" | sudo fdisk "$MAIN_DEVICE" >/dev/null 2>&1
sleep 3

# Format with performance optimizations
sudo mkfs.ext4 -F -E stride=32,stripe-width=128 -O ^has_journal "${MAIN_DEVICE}1" >/dev/null
sleep 3

sudo mkdir -p "$MOUNT_DIR"
sudo mount -o defaults,noatime,data=writeback,barrier=0 "${MAIN_DEVICE}1" "$MOUNT_DIR"

# Filesystem write test
echo "[*] Filesystem write test (${TOTAL_SIZE_GB}GB)..."
FS_WRITE_START=$(date +%s.%N)
sudo dd if=/dev/zero of="$MOUNT_DIR/testfile" bs=${BLOCK_SIZE_MB}M count=$((TOTAL_SIZE_MB / BLOCK_SIZE_MB)) oflag=direct status=progress
sync
FS_WRITE_END=$(date +%s.%N)
FS_WRITE_TIME=$(echo "$FS_WRITE_END - $FS_WRITE_START" | bc)
FS_WRITE_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $FS_WRITE_TIME" | bc)

echo "[✓] Filesystem write: ${FS_WRITE_THROUGHPUT} MB/s"

# Filesystem read test
echo "[*] Filesystem read test (${TOTAL_SIZE_GB}GB)..."
FS_READ_START=$(date +%s.%N)
sudo dd if="$MOUNT_DIR/testfile" of=/dev/null bs=${BLOCK_SIZE_MB}M iflag=direct status=progress
FS_READ_END=$(date +%s.%N)
FS_READ_TIME=$(echo "$FS_READ_END - $FS_READ_START" | bc)
FS_READ_THROUGHPUT=$(echo "scale=2; $TOTAL_SIZE_MB / $FS_READ_TIME" | bc)

echo "[✓] Filesystem read: ${FS_READ_THROUGHPUT} MB/s"

# ========== Performance Summary ==========
echo ""
echo "PERFORMANCE COMPARISON"
echo "====================="
echo "Network baseline (iperf3): ~940 MB/s"
echo ""
echo "Raw Device (no filesystem):"
echo "  Write: ${RAW_WRITE_THROUGHPUT} MB/s"
echo "  Read:  ${RAW_READ_THROUGHPUT} MB/s"
echo ""
echo "Filesystem (ext4 optimized):"
echo "  Write: ${FS_WRITE_THROUGHPUT} MB/s"
echo "  Read:  ${FS_READ_THROUGHPUT} MB/s"
echo ""

# Calculate efficiency
RAW_EFFICIENCY=$(echo "scale=1; $RAW_READ_THROUGHPUT / 940 * 100" | bc)
FS_EFFICIENCY=$(echo "scale=1; $FS_READ_THROUGHPUT / 940 * 100" | bc)

echo "Network efficiency:"
echo "  Raw device: ${RAW_EFFICIENCY}%"
echo "  Filesystem: ${FS_EFFICIENCY}%"

# ========== Cleanup ==========
echo ""
echo "[*] Cleaning up..."
sudo umount "$MOUNT_DIR" 2>/dev/null || true
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --logout || true
sudo rm -rf "$MOUNT_DIR"

# Restore network settings
echo "[*] Restoring original network settings..."
echo "$ORIG_TCP_RMEM" | sudo tee /proc/sys/net/ipv4/tcp_rmem > /dev/null
echo "$ORIG_TCP_WMEM" | sudo tee /proc/sys/net/ipv4/tcp_wmem > /dev/null

echo "[✓] Optimization test completed."

# ========== Recommendations ==========
echo ""
echo "OPTIMIZATION RECOMMENDATIONS"
echo "============================"
if (( $(echo "$RAW_READ_THROUGHPUT < 800" | bc -l) )); then
    echo "• Check iSCSI target configuration (queue depth, multiple sessions)"
    echo "• Verify network interface settings (MTU, interrupt coalescence)"
    echo "• Consider CPU affinity for iSCSI and network interrupts"
    echo "• Check for network packet loss: ping -f $TARGET_IP"
fi

if (( $(echo "$FS_READ_THROUGHPUT < 600" | bc -l) )); then
    echo "• Filesystem overhead is significant"
    echo "• Consider XFS instead of ext4 for better performance"
    echo "• Use larger block sizes (current: ${BLOCK_SIZE_MB}MB)"
fi

echo "• Monitor with: iostat -x 1 during test"
echo "• Check network: ss -i | grep $TARGET_IP" 