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
                f"{name.title()} Current RSS (MB)",
                f"{name.title()} Avg. RSS (MB)",
                f"{name.title()} Min RSS (MB)",
                f"{name.title()} Max RSS (MB)", 
                f"{name.title()} Current %CPU",
                f"{name.title()} Avg. %CPU",
                f"{name.title()} Min %CPU",
                f"{name.title()} Max %CPU"
            ])

        # Initialize metrics tracking
        metrics = {}
        for name in processes.keys():
            metrics[name] = {
                'total_rss': 0.0,
                'total_cpu': 0.0,
                'min_rss': float('inf'),
                'max_rss': 0.0,
                'min_cpu': float('inf'),
                'max_cpu': 0.0,
                'rss_values': [],  # Store all values for debugging
                'cpu_values': []
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
                timestamp = int(time.time())
                row_data = [counter, timestamp]

                # Collect metrics for each monitored process
                for name, process in processes.items():
                    try:
                        # Memory metrics
                        memory_info = process.memory_info()
                        rss_mb = round(memory_info.rss / 1024 / 1024, 2)  # Round to 2 decimal places
                        
                        # Update metrics
                        metrics[name]['total_rss'] += rss_mb
                        metrics[name]['min_rss'] = min(rss_mb, metrics[name]['min_rss'])
                        metrics[name]['max_rss'] = max(rss_mb, metrics[name]['max_rss'])
                        metrics[name]['rss_values'].append(rss_mb)

                        # CPU metrics
                        cpu_interval = min(1.0, interval)
                        cpu_percent = round(process.cpu_percent(interval=cpu_interval), 2)
                        
                        metrics[name]['total_cpu'] += cpu_percent
                        metrics[name]['min_cpu'] = min(cpu_percent, metrics[name]['min_cpu'])
                        metrics[name]['max_cpu'] = max(cpu_percent, metrics[name]['max_cpu'])
                        metrics[name]['cpu_values'].append(cpu_percent)

                        # Calculate averages
                        avg_rss = round(metrics[name]['total_rss'] / counter, 2)
                        avg_cpu = round(metrics[name]['total_cpu'] / counter, 2)

                        # Add to row data: Current, Avg, Min, Max for both RSS and CPU
                        row_data.extend([
                            rss_mb, avg_rss, metrics[name]['min_rss'], metrics[name]['max_rss'],
                            cpu_percent, avg_cpu, metrics[name]['min_cpu'], metrics[name]['max_cpu']
                        ])

                        # Debug output for first few iterations
                        if counter <= 5:
                            print(f"  {name} - Current RSS: {rss_mb} MB, Avg: {avg_rss} MB, Min: {metrics[name]['min_rss']} MB, Max: {metrics[name]['max_rss']} MB")

                    except psutil.NoSuchProcess:
                        print(f"Warning: {name} process (PID: {pids[name]}) no longer exists")
                        # Add zeros for missing process
                        row_data.extend([0, 0, 0, 0, 0, 0, 0, 0])
                    except Exception as e:
                        print(f"Error collecting metrics for {name}: {e}")
                        row_data.extend([0, 0, 0, 0, 0, 0, 0, 0])

                csv_writer.writerow(row_data)
                file.flush()

                # Show human-readable time for user feedback
                readable_time = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(timestamp))
                print(f"Iteration {counter}/{iterations} completed at {readable_time}")
                counter += 1

                # Sleep for remaining interval time
                if counter <= iterations:
                    remaining_sleep = interval - cpu_interval if counter > 1 else interval
                    if remaining_sleep > 0:
                        time.sleep(remaining_sleep)

        # Print final statistics
        print(f"\nMonitoring completed. Results saved to {output_file}")
        print("\nFinal Statistics:")
        for name, data in metrics.items():
            if data['rss_values']:
                rss_range = data['max_rss'] - data['min_rss']
                cpu_range = data['max_cpu'] - data['min_cpu']
                print(f"{name.title()}:")
                print(f"  RSS: Min={data['min_rss']:.2f} MB, Max={data['max_rss']:.2f} MB, Range={rss_range:.2f} MB")
                print(f"  CPU: Min={data['min_cpu']:.2f}%, Max={data['max_cpu']:.2f}%, Range={cpu_range:.2f}%")
                print(f"  Total samples: {len(data['rss_values'])}")
    
    except KeyboardInterrupt:
        print("\nMonitoring stopped by user")
    except Exception as e:
        print(f"An error occurred: {str(e)}")

def main():
    parser = argparse.ArgumentParser(description='Monitor process performance metrics with debugging')
    parser.add_argument('--envoy-pid', type=int, help='PID of Envoy process')
    parser.add_argument('--mercury-pid', type=int, help='PID of Mercury process')
    parser.add_argument('--idf-pid', type=int, help='PID of IDF process')
    parser.add_argument('--mock-pid', type=int, help='PID of Mock process')
    parser.add_argument('--output', '-o', default='metrics_debug.csv', help='Output CSV file (default: metrics_debug.csv)')
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
        print("  python envoy_perf_monitor_debug.py --envoy-pid 1234")
        print("  python envoy_perf_monitor_debug.py --envoy-pid 1234 --interval 5 --iterations 120")
        sys.exit(1)
    
    collect_metrics(pids, args.output, args.interval, args.iterations)

if __name__ == "__main__":
    main() 