#!/bin/bash

# Usage: ./iscsi_read_throughput.sh <interval_in_seconds>
# Example: ./iscsi_read_throughput.sh 10

set -e

TARGET_IP="10.33.8.46"
IQN="iqn.2010-06.com.nutanix:vg1-3740cfaa-3460-4d65-9dfb-f66022ef95e4-tgt0"
MOUNT_DIR="/mnt/iscsi_test"
FILE_SIZE_MB=1024
BLOCK_SIZE=1M
DEVICE=""
INTERVAL="$1"

# Validate input
if [[ -z "$INTERVAL" || "$INTERVAL" -le 0 ]]; then
  echo "Usage: $0 <interval_in_seconds>"
  exit 1
fi

echo "[*] Logging into iSCSI target..."
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --login || true
sleep 5

# Detect device
DEVICE=$(lsblk -ndo NAME,TYPE | grep disk | awk '{print $1}' | tail -n1)
DEVICE="/dev/$DEVICE"

if [ ! -b "$DEVICE" ]; then
  echo "[!] iSCSI device not found. Exiting."
  exit 1
fi

# Partition, format, mount
echo "[*] Preparing device $DEVICE..."
echo -e "o\nn\np\n1\n\n\nw" | sudo fdisk "$DEVICE" >/dev/null 2>&1
sleep 3
sudo mkfs.ext4 -F "${DEVICE}1" >/dev/null
sudo mkdir -p "$MOUNT_DIR"
sudo mount "${DEVICE}1" "$MOUNT_DIR"

# Create test file
echo "[*] Creating test file of ${FILE_SIZE_MB} MB..."
sudo dd if=/dev/zero of="$MOUNT_DIR/testfile" bs=1M count=$FILE_SIZE_MB status=none oflag=direct

# Start read throughput test
echo "[*] Starting read throughput test for ${INTERVAL} seconds..."
echo "timestamp,MB_read"

START_TIME=$(date +%s)
END_TIME=$((START_TIME + INTERVAL))
TOTAL_MB=0

while [ "$(date +%s)" -lt "$END_TIME" ]; do
  NOW=$(date +%s)
  
  OUT=$(sudo dd if="$MOUNT_DIR/testfile" of=/dev/null bs=$BLOCK_SIZE iflag=direct count=100 2>&1)
  BYTES=$(echo "$OUT" | grep -oP '^\d+(?= bytes)' || echo 0)
  MB=$(echo "$BYTES / 1024 / 1024" | bc)

  echo "$NOW,$MB"
  TOTAL_MB=$((TOTAL_MB + MB))

  sleep 1
done

ACTUAL_DURATION=$(( $(date +%s) - START_TIME ))

# Show final stats
echo
echo "[✓] Test completed."
echo "-----------------------------"
echo "Total MB read     : $TOTAL_MB"
echo "Duration (seconds): $ACTUAL_DURATION"
if [ "$ACTUAL_DURATION" -gt 0 ]; then
  AVG_MBPS=$(echo "scale=2; $TOTAL_MB / $ACTUAL_DURATION" | bc)
else
  AVG_MBPS="0"
fi
echo "Avg Throughput    : $AVG_MBPS MB/s"
echo "-----------------------------"

# Cleanup
echo "[*] Cleaning up..."
sudo umount "$MOUNT_DIR"
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --logout || true
sudo rm -rf "$MOUNT_DIR"

echo "[✓] Done."

