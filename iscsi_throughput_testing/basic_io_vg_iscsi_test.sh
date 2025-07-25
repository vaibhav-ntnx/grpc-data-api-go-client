#!/bin/bash

# Configurations
TARGET_IP="10.33.8.46"
IQN="iqn.2010-06.com.nutanix:vg1-3740cfaa-3460-4d65-9dfb-f66022ef95e4-tgt0"
MOUNT_DIR="/mnt/iscsi_test"
DEVICE=""

echo "[*] Logging into iSCSI target..."
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --login

echo "[*] Waiting for device to appear..."
sleep 5

# Detect the new disk (assumes it's the last added /dev/sdX)
DEVICE=$(lsblk -ndo NAME,TYPE | grep disk | awk '{print $1}' | tail -n1)
DEVICE="/dev/$DEVICE"

if [ ! -b "$DEVICE" ]; then
  echo "[!] iSCSI device not found. Exiting."
  exit 1
fi

echo "[*] Creating partition on $DEVICE..."
echo -e "o\nn\np\n1\n\n\nw" | sudo fdisk "$DEVICE"
sleep 3

echo "[*] Formatting partition with ext4..."
sudo mkfs.ext4 "${DEVICE}1"

echo "[*] Creating mount directory..."
sudo mkdir -p "$MOUNT_DIR"

echo "[*] Mounting partition..."
sudo mount "${DEVICE}1" "$MOUNT_DIR"

echo "[*] Performing write I/O test..."
sudo dd if=/dev/zero of="$MOUNT_DIR/testfile" bs=1M count=100 oflag=direct

echo "[*] Performing read I/O test..."
sudo dd if="$MOUNT_DIR/testfile" of=/dev/null bs=1M iflag=direct

echo "[*] Cleaning up..."
sudo rm -f "$MOUNT_DIR/testfile"
sudo umount "$MOUNT_DIR"
sudo iscsiadm -m node -T "$IQN" -p "$TARGET_IP" --logout

echo "[✓] Done."

