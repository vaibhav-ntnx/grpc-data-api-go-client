import psutil
import time
import csv
import argparse
import sys

def collect_metrics(pids, output_file, interval, iterations):
    try:
        # Filter out None PIDs and create process objects
        processes = {}
        for name, pid in pids.items():
            if pid is not None:
                try:
                    processes[name] = psutil.Process(pid)
                    print(f"Monitoring {name} process (PID: {pid})")
                except psutil.NoSuchProcess:
                    print(f"Warning: No process found with PID {pid} for {name}")
                    
        if not processes:
            print("No valid processes to monitor. Exiting.")
            return

        # Build dynamic CSV header based on monitored processes
        header = ["Iteration", "Epoch_Time"]
        for name in processes.keys():
            header.extend([
                f"{name.title()} Avg. RSS (MB)",
                f"{name.title()} Max RSS (MB)", 
                f"{name.title()} Avg. %CPU",
                f"{name.title()} Max %CPU"
            ])

        # Initialize metrics tracking
        metrics = {}
        for name in processes.keys():
            metrics[name] = {
                'total_rss': 0.0,
                'total_cpu': 0.0,
                'max_rss': 0.0,
                'max_cpu': 0.0
            }

        # Open the output CSV file
        with open(output_file, mode='w', newline='') as file:
            csv_writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
            csv_writer.writerow(header)
            
            counter = 1
            total_duration = interval * iterations
            print(f"Starting monitoring for {len(processes)} processes...")
            print(f"Configuration: {iterations} iterations, {interval}s interval (~{total_duration/60:.1f} minutes total)")
            print("Press Ctrl+C to stop monitoring")

            # Collect metrics for specified iterations
            while counter <= iterations:
                timestamp = int(time.time())  # Epoch time in seconds only
                row_data = [counter, timestamp]

                # Collect metrics for each monitored process
                for name, process in processes.items():
                    try:
                        # Memory metrics
                        memory_info = process.memory_info()
                        rss_mb = memory_info.rss / 1024 / 1024
                        metrics[name]['max_rss'] = max(rss_mb, metrics[name]['max_rss'])
                        metrics[name]['total_rss'] += rss_mb

                        # CPU metrics (use min of 1 second or interval for accuracy)
                        cpu_interval = min(1.0, interval)
                        cpu_percent = process.cpu_percent(interval=cpu_interval)
                        metrics[name]['max_cpu'] = max(cpu_percent, metrics[name]['max_cpu'])
                        metrics[name]['total_cpu'] += cpu_percent

                        # Add to row data
                        avg_rss = metrics[name]['total_rss'] / counter
                        avg_cpu = metrics[name]['total_cpu'] / counter
                        row_data.extend([avg_rss, metrics[name]['max_rss'], avg_cpu, metrics[name]['max_cpu']])

                    except psutil.NoSuchProcess:
                        print(f"Warning: {name} process (PID: {pids[name]}) no longer exists")
                        # Add zeros for missing process
                        row_data.extend([0, 0, 0, 0])
                    except Exception as e:
                        print(f"Error collecting metrics for {name}: {e}")
                        row_data.extend([0, 0, 0, 0])

                csv_writer.writerow(row_data)
                file.flush()

                # Show human-readable time for user feedback
                readable_time = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(timestamp))
                print(f"Iteration {counter}/{iterations} completed at {readable_time} (epoch: {timestamp})")
                counter += 1

                # Sleep for remaining interval time (accounting for CPU measurement time)
                if counter <= iterations:
                    remaining_sleep = interval - cpu_interval if counter > 1 else interval
                    if remaining_sleep > 0:
                        time.sleep(remaining_sleep)

        print(f"Monitoring completed. Results saved to {output_file}")
    
    except KeyboardInterrupt:
        print("\nMonitoring stopped by user")
    except Exception as e:
        print(f"An error occurred: {str(e)}")

def main():
    parser = argparse.ArgumentParser(description='Monitor process performance metrics')
    parser.add_argument('--envoy-pid', type=int, help='PID of Envoy process')
    parser.add_argument('--mercury-pid', type=int, help='PID of Mercury process')
    parser.add_argument('--idf-pid', type=int, help='PID of IDF process')
    parser.add_argument('--mock-pid', type=int, help='PID of Mock process')
    parser.add_argument('--output', '-o', default='metrics.csv', help='Output CSV file (default: metrics.csv)')
    parser.add_argument('--interval', '-i', type=int, default=2, help='Monitoring interval in seconds (default: 2)')
    parser.add_argument('--iterations', '-n', type=int, default=60, help='Number of monitoring iterations (default: 60)')
    
    args = parser.parse_args()
    
    # Validate arguments
    if args.interval < 1:
        print("Error: Interval must be at least 1 second")
        sys.exit(1)
    
    if args.iterations < 1:
        print("Error: Iterations must be at least 1")
        sys.exit(1)
    
    # Build PID dictionary
    pids = {
        'envoy': args.envoy_pid,
        'mercury': args.mercury_pid,
        'idf': args.idf_pid,
        'mock': args.mock_pid
    }
    
    # Check if at least one PID is provided
    if not any(pid is not None for pid in pids.values()):
        print("Error: Please provide at least one PID to monitor")
        print("Usage examples:")
        print("  python envoy_perf_monitor.py --envoy-pid 1234")
        print("  python envoy_perf_monitor.py --envoy-pid 1234 --interval 5 --iterations 120")
        print("  python envoy_perf_monitor.py --envoy-pid 1234 --mercury-pid 5678 -i 1 -n 30")
        sys.exit(1)
    
    collect_metrics(pids, args.output, args.interval, args.iterations)

if __name__ == "__main__":
    main()