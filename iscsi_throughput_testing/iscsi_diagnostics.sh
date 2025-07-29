#!/bin/bash

# iSCSI Performance Diagnostics
# Run this in a separate terminal while running throughput tests

TARGET_IP="10.33.8.46"
IQN="iqn.2010-06.com.nutanix:vg1-3740cfaa-3460-4d65-9dfb-f66022ef95e4-tgt0"
INTERVAL=2

echo "iSCSI Performance Diagnostics"
echo "=============================="
echo "Monitor this while running throughput tests"
echo "Target: $TARGET_IP"
echo ""

while true; do
    clear
    echo "$(date) - iSCSI Performance Monitor"
    echo "======================================"
    
    # Network interface stats
    echo "1. Network Interface Performance:"
    for iface in $(ip route get $TARGET_IP | grep -oP 'dev \K\w+' | head -1); do
        if [ -n "$iface" ]; then
            RX_BYTES=$(cat /sys/class/net/$iface/statistics/rx_bytes)
            TX_BYTES=$(cat /sys/class/net/$iface/statistics/tx_bytes)
            RX_PACKETS=$(cat /sys/class/net/$iface/statistics/rx_packets)
            TX_PACKETS=$(cat /sys/class/net/$iface/statistics/tx_packets)
            RX_ERRORS=$(cat /sys/class/net/$iface/statistics/rx_errors)
            TX_ERRORS=$(cat /sys/class/net/$iface/statistics/tx_errors)
            RX_DROPPED=$(cat /sys/class/net/$iface/statistics/rx_dropped)
            TX_DROPPED=$(cat /sys/class/net/$iface/statistics/tx_dropped)
            
            echo "   Interface: $iface"
            echo "   RX: $(($RX_BYTES / 1024 / 1024)) MB, Packets: $RX_PACKETS, Errors: $RX_ERRORS, Dropped: $RX_DROPPED"
            echo "   TX: $(($TX_BYTES / 1024 / 1024)) MB, Packets: $TX_PACKETS, Errors: $TX_ERRORS, Dropped: $TX_DROPPED"
        fi
    done
    echo ""
    
    # TCP connection stats
    echo "2. TCP Connection to iSCSI Target:"
    ss -i dst $TARGET_IP | grep -A 5 -B 1 "3260\|iscsi" || echo "   No active iSCSI connections found"
    echo ""
    
    # iSCSI session information
    echo "3. iSCSI Session Status:"
    if sudo iscsiadm -m session -P 1 2>/dev/null | grep -q "$TARGET_IP"; then
        echo "   Active sessions:"
        sudo iscsiadm -m session -P 1 | grep -A 5 -B 2 "$TARGET_IP" | head -10
        
        # Queue depth info
        echo "   Current queue depths:"
        for dev in $(lsscsi | grep -i nutanix | awk '{print $6}' 2>/dev/null); do
            if [ -n "$dev" ]; then
                QUEUE_DEPTH=$(cat /sys/block/$(basename $dev)/queue/nr_requests 2>/dev/null || echo "N/A")
                echo "   $dev: $QUEUE_DEPTH requests"
            fi
        done
    else
        echo "   No active iSCSI sessions"
    fi
    echo ""
    
    # Disk I/O stats
    echo "4. Disk I/O Statistics:"
    iostat -x 1 1 | grep -E "(Device|sd[a-z]+)" | tail -5
    echo ""
    
    # Memory and CPU
    echo "5. System Resources:"
    echo "   Memory: $(free -h | grep Mem | awk '{print $3 "/" $2}')"
    echo "   Load: $(uptime | awk -F'load average:' '{print $2}')"
    echo ""
    
    # Network buffer information
    echo "6. Network Buffer Settings:"
    echo "   TCP rmem: $(cat /proc/sys/net/ipv4/tcp_rmem)"
    echo "   TCP wmem: $(cat /proc/sys/net/ipv4/tcp_wmem)"
    echo "   TCP window scaling: $(cat /proc/sys/net/ipv4/tcp_window_scaling)"
    echo ""
    
    # IRQ information
    echo "7. Interrupt Information:"
    echo "   Network IRQs per second:"
    grep -E "$(ip route get $TARGET_IP | grep -oP 'dev \K\w+' | head -1)" /proc/interrupts | head -3
    echo ""
    
    echo "Press Ctrl+C to stop monitoring"
    echo "Refreshing every ${INTERVAL} seconds..."
    
    sleep $INTERVAL
done 