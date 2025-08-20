#!/usr/bin/env python3

import argparse
import os
import sys
import pandas as pd
import matplotlib.pyplot as plt


def load_envoy_metrics(metrics_csv: str):
    df = pd.read_csv(metrics_csv)
    # Validate required columns
    required_cols = ["Epoch_Time", "Envoy Avg. RSS (MB)"]
    for col in required_cols:
        if col not in df.columns:
            raise ValueError(f"Missing column '{col}' in metrics CSV: {metrics_csv}")

    # Use numeric epoch seconds on X
    x = pd.to_numeric(df["Epoch_Time"], errors="coerce")
    y_rss = pd.to_numeric(df["Envoy Avg. RSS (MB)"], errors="coerce")

    # Drop rows with NaNs in either series
    mask = x.notna() & y_rss.notna()
    return x[mask], y_rss[mask]


def load_vdisk_throughput(throughput_csv: str, prefer_column: str = "mb_per_sec"):
    df = pd.read_csv(throughput_csv)
    # Validate columns
    if "epoch_second" not in df.columns:
        raise ValueError(f"Missing column 'epoch_second' in throughput CSV: {throughput_csv}")

    throughput_col = None
    if prefer_column in df.columns:
        throughput_col = prefer_column
    elif "bytes_per_sec" in df.columns:
        throughput_col = "bytes_per_sec"
    else:
        # Fallback to requests_per_sec if no bytes-based rate present, though units differ
        candidate = "requests_per_sec"
        if candidate in df.columns:
            throughput_col = candidate
        else:
            raise ValueError(
                "Throughput CSV must contain one of: 'mb_per_sec', 'bytes_per_sec', 'requests_per_sec'"
            )

    x = pd.to_numeric(df["epoch_second"], errors="coerce")
    y = pd.to_numeric(df[throughput_col], errors="coerce")

    # Normalize to MB/s if possible
    if throughput_col == "bytes_per_sec":
        y = y / (1024 * 1024)
    # If requests_per_sec was selected, keep it as-is but label will indicate RPS

    mask = x.notna() & y.notna()
    return x[mask], y[mask], throughput_col


def plot_combined(envoy_x, envoy_rss_mb, vdisk_x, vdisk_mb_per_sec, vdisk_unit_label: str, output_path: str):
    os.makedirs(os.path.dirname(output_path), exist_ok=True) if os.path.dirname(output_path) else None

    fig, ax_left = plt.subplots(figsize=(14, 8))

    # Left axis: Envoy RSS (MB)
    line1, = ax_left.plot(envoy_x, envoy_rss_mb, color="tab:blue", linewidth=2, label="Envoy RSS (MB)")
    ax_left.set_xlabel("Epoch Time (s)", fontsize=12)
    ax_left.set_ylabel("Envoy RSS (MB)", color="tab:blue", fontsize=12)
    ax_left.tick_params(axis='y', labelcolor="tab:blue")
    ax_left.grid(True, which="both", axis="both", alpha=0.3)

    # Right axis: VDisk throughput (MB/s or RPS)
    ax_right = ax_left.twinx()
    line2, = ax_right.plot(vdisk_x, vdisk_mb_per_sec, color="tab:red", linewidth=2, label=f"VDisk Throughput ({vdisk_unit_label})")
    ax_right.set_ylabel(f"VDisk Throughput ({vdisk_unit_label})", color="tab:red", fontsize=12)
    ax_right.tick_params(axis='y', labelcolor="tab:red")

    # Title and legend
    plt.title("Envoy RSS vs VDisk Throughput over Epoch Time", fontsize=14, fontweight="bold")

    # Build combined legend
    lines = [line1, line2]
    labels = [line.get_label() for line in lines]
    ax_left.legend(lines, labels, loc="upper left")

    plt.tight_layout()
    plt.savefig(output_path, dpi=300, bbox_inches='tight')
    plt.close(fig)


def main():
    parser = argparse.ArgumentParser(
        description="Plot Envoy RSS (MB) and VDisk Throughput on the same Epoch-time axis"
    )
    parser.add_argument("--metrics-csv", default="metrics.csv", help="Path to Envoy metrics CSV (default: metrics.csv)")
    parser.add_argument(
        "--throughput-csv",
        default="throughput_metrics.csv",
        help="Path to VDisk throughput CSV (default: throughput_metrics.csv)",
    )
    parser.add_argument(
        "--output",
        "-o",
        default="plots/combined_envoy_rss_vdisk_throughput.png",
        help="Output plot image path",
    )
    parser.add_argument(
        "--throughput-col",
        choices=["mb_per_sec", "bytes_per_sec", "requests_per_sec"],
        default="mb_per_sec",
        help="Which throughput column to use from throughput CSV (default: mb_per_sec)",
    )

    args = parser.parse_args()

    if not os.path.exists(args.metrics_csv):
        print(f"Error: metrics CSV not found: {args.metrics_csv}")
        sys.exit(1)
    if not os.path.exists(args.throughput_csv):
        print(f"Error: throughput CSV not found: {args.throughput_csv}")
        sys.exit(1)

    try:
        envoy_x, envoy_rss_mb = load_envoy_metrics(args.metrics_csv)
        vdisk_x, vdisk_series, used_col = load_vdisk_throughput(args.throughput_csv, prefer_column=args.throughput_col)

        # Determine label for the throughput axis
        if used_col == "mb_per_sec":
            unit_label = "MB/s"
        elif used_col == "bytes_per_sec":
            unit_label = "MB/s"  # We convert bytes/sec to MB/s above
        else:
            unit_label = "RPS"

        plot_combined(envoy_x, envoy_rss_mb, vdisk_x, vdisk_series, unit_label, args.output)
        print(f"Saved combined plot to {args.output}")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()


