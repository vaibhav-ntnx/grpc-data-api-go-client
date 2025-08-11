#!/usr/bin/env python3

import pandas as pd
import argparse
import sys

def analyze_metrics(csv_file):
    """
    Analyze performance metrics to understand why avg and max might be similar
    """
    try:
        df = pd.read_csv(csv_file)
        print(f"Loaded {len(df)} rows of data\n")
        
        # Find process columns
        processes = []
        for col in df.columns:
            if 'Avg. RSS (MB)' in col:
                process_name = col.replace(' Avg. RSS (MB)', '').lower()
                processes.append(process_name)
        
        for process in processes:
            print(f"=== {process.upper()} ANALYSIS ===")
            
            avg_rss_col = f"{process.title()} Avg. RSS (MB)"
            max_rss_col = f"{process.title()} Max RSS (MB)"
            
            if avg_rss_col not in df.columns or max_rss_col not in df.columns:
                print(f"Missing columns for {process}")
                continue
            
            avg_values = df[avg_rss_col]
            max_values = df[max_rss_col]
            
            # Basic statistics
            print(f"Average RSS column stats:")
            print(f"  Min: {avg_values.min():.2f} MB")
            print(f"  Max: {avg_values.max():.2f} MB") 
            print(f"  Mean: {avg_values.mean():.2f} MB")
            print(f"  Std Dev: {avg_values.std():.4f} MB")
            print(f"  Range: {avg_values.max() - avg_values.min():.2f} MB")
            
            print(f"\nMax RSS column stats:")
            print(f"  Min: {max_values.min():.2f} MB")
            print(f"  Max: {max_values.max():.2f} MB")
            print(f"  Mean: {max_values.mean():.2f} MB")
            print(f"  Std Dev: {max_values.std():.4f} MB")
            
            # Check if they're identical
            identical_count = (avg_values == max_values).sum()
            print(f"\nRows where Avg RSS = Max RSS: {identical_count}/{len(df)} ({100*identical_count/len(df):.1f}%)")
            
            # Show first 10 values
            print(f"\nFirst 10 iterations:")
            print("Iter | Avg RSS | Max RSS | Difference")
            print("-" * 40)
            for i in range(min(10, len(df))):
                diff = max_values.iloc[i] - avg_values.iloc[i]
                print(f"{i+1:4d} | {avg_values.iloc[i]:7.2f} | {max_values.iloc[i]:7.2f} | {diff:10.4f}")
            
            # Show where they differ most
            differences = abs(max_values - avg_values)
            max_diff_idx = differences.idxmax()
            max_diff_value = differences.iloc[max_diff_idx]
            
            print(f"\nLargest difference at iteration {max_diff_idx + 1}:")
            print(f"  Avg: {avg_values.iloc[max_diff_idx]:.2f} MB")
            print(f"  Max: {max_values.iloc[max_diff_idx]:.2f} MB")
            print(f"  Difference: {max_diff_value:.4f} MB")
            
            # Check for very stable values
            if avg_values.std() < 1.0:  # Less than 1 MB standard deviation
                print(f"\n⚠️  {process.upper()} has very stable memory usage (std dev < 1 MB)")
                print("This is why average and max appear similar.")
            
            print("\n" + "="*50 + "\n")
            
    except Exception as e:
        print(f"Error analyzing data: {e}")
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description='Analyze performance metrics CSV')
    parser.add_argument('csv_file', help='CSV file to analyze')
    
    args = parser.parse_args()
    
    analyze_metrics(args.csv_file)

if __name__ == "__main__":
    main() 