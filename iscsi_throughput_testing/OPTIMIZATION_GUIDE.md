# iSCSI Performance Optimization Guide

## Problem Summary
Your network shows ~940 MB/s bandwidth (7.38 Gbps) with iperf3, but iSCSI was only achieving ~130 MB/s (14% efficiency).

## Root Causes Identified

1. **Low iSCSI queue depth** (default: 32)
2. **Single iSCSI session** (no parallelism)
3. **Small block sizes** (1MB)
4. **Filesystem overhead** (ext4 with default settings)
5. **Suboptimal network parameters**

## Optimizations Applied

### 1. iSCSI Parameters
- **Queue depth**: 32 → 128 (4x increase)
- **Multiple sessions**: 1 → 4 (parallel connections)
- **Data segment size**: Increased to 256KB
- **Block size**: 1MB → 64MB (better throughput)

### 2. Network Tuning
- **TCP window scaling**: Enabled
- **TCP buffers**: Increased to 16MB
- **TCP receive/send buffers**: Optimized for high bandwidth

### 3. Filesystem Optimizations
- **Journal**: Disabled for testing
- **Mount options**: noatime, data=writeback, barrier=0
- **Stride/stripe**: Optimized for SSD/RAID

## Usage Instructions

### Step 1: Run the Optimized Test
```bash
./iscsi_throughput_testing/iscsi_optimized_throughput.sh
```

This script will:
- Apply all optimizations automatically
- Test both raw device and filesystem performance
- Compare against your 940 MB/s network baseline
- Show efficiency percentages

### Step 2: Monitor Performance (Optional)
In a separate terminal, run the diagnostic monitor:
```bash
./iscsi_throughput_testing/iscsi_diagnostics.sh
```

This shows real-time:
- Network interface statistics
- TCP connection details
- iSCSI session info
- I/O statistics
- System resources

## Expected Results

With optimizations, you should see:
- **Raw device**: 700-850 MB/s (75-90% efficiency)
- **Filesystem**: 500-700 MB/s (55-75% efficiency)

## Troubleshooting

### If still under 700 MB/s:

1. **Check iSCSI target configuration**:
   ```bash
   # Verify target supports multiple sessions
   sudo iscsiadm -m session -P 3 | grep -i session
   ```

2. **Network interface settings**:
   ```bash
   # Check MTU and ring buffers
   ethtool eth0  # replace with your interface
   ```

3. **CPU/IRQ affinity**:
   ```bash
   # Check if interrupts are balanced
   cat /proc/interrupts | grep eth0
   ```

4. **Packet loss check**:
   ```bash
   ping -f 10.33.8.46 -c 1000
   ```

### If filesystem performance is poor:

1. **Try XFS instead of ext4**:
   ```bash
   sudo mkfs.xfs -f /dev/sdX1
   sudo mount -o noatime,largeio,inode64 /dev/sdX1 /mnt/test
   ```

2. **Increase block size further**:
   - Modify `BLOCK_SIZE_MB=64` to `BLOCK_SIZE_MB=128`

3. **Use raw device for maximum performance**:
   - Skip filesystem entirely for pure throughput

## Performance Comparison

| Configuration | Expected Throughput | Efficiency |
|---------------|-------------------|------------|
| Original (1MB, queue=32) | ~130 MB/s | 14% |
| Optimized (64MB, queue=128, 4 sessions) | ~750 MB/s | 80% |
| Raw device optimal | ~850 MB/s | 90% |

## Additional Notes

- The script automatically restores original network settings after testing
- Raw device testing bypasses filesystem overhead entirely
- Multiple iSCSI sessions require target support (most modern targets support this)
- Results may vary based on target storage performance and network latency 