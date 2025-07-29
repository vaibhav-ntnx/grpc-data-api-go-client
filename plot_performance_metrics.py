#!/usr/bin/env python3

import pandas as pd
import matplotlib.pyplot as plt
import argparse
import sys
import os

def create_plots(csv_file, output_dir='plots'):
    """
    Generate line plots and histograms for performance metrics
    """
    try:
        # Read CSV data
        df = pd.read_csv(csv_file)
        print(f"Loaded data with {len(df)} rows and {len(df.columns)} columns")
        
        # Create output directory if it doesn't exist
        os.makedirs(output_dir, exist_ok=True)
        
        # Get process names by parsing column headers
        processes = []
        for col in df.columns:
            if 'Avg. RSS (MB)' in col:
                process_name = col.replace(' Avg. RSS (MB)', '').lower()
                processes.append(process_name)
        
        print(f"Found processes: {processes}")
        
        # Set up matplotlib style
        plt.style.use('default')
        fig_size = (12, 8)
        
        for process in processes:
            rss_col = f"{process.title()} Avg. RSS (MB)"
            cpu_col = f"{process.title()} Avg. %CPU"
            
            if rss_col not in df.columns or cpu_col not in df.columns:
                print(f"Warning: Missing columns for {process}")
                continue
                
            # Line plots
            create_line_plots(df, process, rss_col, cpu_col, output_dir, fig_size)
            
            # Histograms  
            create_histograms(df, process, rss_col, cpu_col, output_dir, fig_size)
        
        print(f"All plots saved to {output_dir}/ directory")
        
    except FileNotFoundError:
        print(f"Error: CSV file '{csv_file}' not found")
        sys.exit(1)
    except Exception as e:
        print(f"Error processing data: {e}")
        sys.exit(1)

def create_line_plots(df, process, rss_col, cpu_col, output_dir, fig_size):
    """Create line plots for RSS and CPU metrics"""
    
    # RSS Line Plot
    plt.figure(figsize=fig_size)
    plt.plot(df['Iteration'], df[rss_col], marker='o', linewidth=2, markersize=4)
    plt.title(f'{process.title()} Process - Average RSS Memory Usage Over Time', fontsize=14, fontweight='bold')
    plt.xlabel('Iteration (Sequence Number)', fontsize=12)
    plt.ylabel('Average RSS Memory (MB)', fontsize=12)
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(f'{output_dir}/{process}_rss_line_plot.png', dpi=300, bbox_inches='tight')
    plt.close()
    
    # CPU Line Plot
    plt.figure(figsize=fig_size)
    plt.plot(df['Iteration'], df[cpu_col], marker='s', linewidth=2, markersize=4, color='orange')
    plt.title(f'{process.title()} Process - Average CPU Usage Over Time', fontsize=14, fontweight='bold')
    plt.xlabel('Iteration (Sequence Number)', fontsize=12)
    plt.ylabel('Average CPU Usage (%)', fontsize=12)
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(f'{output_dir}/{process}_cpu_line_plot.png', dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Created line plots for {process}")

def create_histograms(df, process, rss_col, cpu_col, output_dir, fig_size):
    """Create histograms for RSS and CPU metrics"""
    
    # RSS Histogram
    plt.figure(figsize=fig_size)
    plt.hist(df[rss_col], bins=20, alpha=0.7, color='skyblue', edgecolor='black')
    plt.title(f'{process.title()} Process - Distribution of Average RSS Memory Usage', fontsize=14, fontweight='bold')
    plt.xlabel('Average RSS Memory (MB)', fontsize=12)
    plt.ylabel('Frequency', fontsize=12)
    plt.grid(True, alpha=0.3, axis='y')
    
    # Add statistics
    mean_rss = df[rss_col].mean()
    std_rss = df[rss_col].std()
    plt.axvline(mean_rss, color='red', linestyle='--', linewidth=2, label=f'Mean: {mean_rss:.2f} MB')
    plt.legend()
    
    plt.tight_layout()
    plt.savefig(f'{output_dir}/{process}_rss_histogram.png', dpi=300, bbox_inches='tight')
    plt.close()
    
    # CPU Histogram
    plt.figure(figsize=fig_size)
    plt.hist(df[cpu_col], bins=20, alpha=0.7, color='lightcoral', edgecolor='black')
    plt.title(f'{process.title()} Process - Distribution of Average CPU Usage', fontsize=14, fontweight='bold')
    plt.xlabel('Average CPU Usage (%)', fontsize=12)
    plt.ylabel('Frequency', fontsize=12)
    plt.grid(True, alpha=0.3, axis='y')
    
    # Add statistics
    mean_cpu = df[cpu_col].mean()
    std_cpu = df[cpu_col].std()
    plt.axvline(mean_cpu, color='red', linestyle='--', linewidth=2, label=f'Mean: {mean_cpu:.2f}%')
    plt.legend()
    
    plt.tight_layout()
    plt.savefig(f'{output_dir}/{process}_cpu_histogram.png', dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Created histograms for {process}")

def create_combined_plots(csv_file, output_dir='plots'):
    """Create combined plots showing all processes together"""
    
    df = pd.read_csv(csv_file)
    
    # Get process names
    processes = []
    for col in df.columns:
        if 'Avg. RSS (MB)' in col:
            process_name = col.replace(' Avg. RSS (MB)', '').lower()
            processes.append(process_name)
    
    if len(processes) > 1:
        # Combined RSS plot
        plt.figure(figsize=(14, 8))
        for process in processes:
            rss_col = f"{process.title()} Avg. RSS (MB)"
            if rss_col in df.columns:
                plt.plot(df['Iteration'], df[rss_col], marker='o', linewidth=2, label=f'{process.title()}', markersize=3)
        
        plt.title('All Processes - Average RSS Memory Usage Comparison', fontsize=14, fontweight='bold')
        plt.xlabel('Iteration (Sequence Number)', fontsize=12)
        plt.ylabel('Average RSS Memory (MB)', fontsize=12)
        plt.legend()
        plt.grid(True, alpha=0.3)
        plt.tight_layout()
        plt.savefig(f'{output_dir}/combined_rss_comparison.png', dpi=300, bbox_inches='tight')
        plt.close()
        
        # Combined CPU plot
        plt.figure(figsize=(14, 8))
        for process in processes:
            cpu_col = f"{process.title()} Avg. %CPU"
            if cpu_col in df.columns:
                plt.plot(df['Iteration'], df[cpu_col], marker='s', linewidth=2, label=f'{process.title()}', markersize=3)
        
        plt.title('All Processes - Average CPU Usage Comparison', fontsize=14, fontweight='bold')
        plt.xlabel('Iteration (Sequence Number)', fontsize=12)
        plt.ylabel('Average CPU Usage (%)', fontsize=12)
        plt.legend()
        plt.grid(True, alpha=0.3)
        plt.tight_layout()
        plt.savefig(f'{output_dir}/combined_cpu_comparison.png', dpi=300, bbox_inches='tight')
        plt.close()
        
        print("Created combined comparison plots")

def main():
    parser = argparse.ArgumentParser(description='Generate performance metric plots from CSV data')
    parser.add_argument('csv_file', help='Input CSV file path (output from envoy_perf_monitor.py)')
    parser.add_argument('--output-dir', '-o', default='plots', help='Output directory for plots (default: plots)')
    parser.add_argument('--combined', '-c', action='store_true', help='Also create combined plots for multiple processes')
    
    args = parser.parse_args()
    
    if not os.path.exists(args.csv_file):
        print(f"Error: CSV file '{args.csv_file}' does not exist")
        sys.exit(1)
    
    # Generate individual plots
    create_plots(args.csv_file, args.output_dir)
    
    # Generate combined plots if requested
    if args.combined:
        create_combined_plots(args.csv_file, args.output_dir)

if __name__ == "__main__":
    main() 