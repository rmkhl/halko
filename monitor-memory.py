#!/usr/bin/env python3
"""
Process Memory Monitor

Monitors memory usage of running processes to detect memory leaks.
Provides real-time console output and optional CSV logging.

Works with any Linux processes - monitors by name using /proc filesystem.

Usage:
    python3 monitor-memory.py                    # Monitor with console output only
    python3 monitor-memory.py -i 5               # Custom interval (seconds)
    python3 monitor-memory.py -o memlog.csv      # Save to CSV file
    python3 monitor-memory.py -t 3600            # Run for 1 hour then exit
    python3 monitor-memory.py -p controlunit powerunit  # Monitor specific processes

Requirements:
    None - uses only Python standard library
"""

import argparse
import csv
import datetime
import os
import sys
import time
from collections import defaultdict
from typing import Dict, List, Optional, Tuple


class MemoryMonitor:
    """Monitor process memory usage and detect leaks"""

    def __init__(self, process_names: List[str], output_file: Optional[str],
                 interval: int, duration: Optional[int] = None):
        self.process_names = process_names
        self.output_file = output_file
        self.interval = interval
        self.duration = duration
        self.running = True
        self.start_time = time.time()

        # Memory tracking for leak detection
        self.memory_history: Dict[str, List[float]] = defaultdict(list)
        self.baseline_memory: Dict[str, float] = {}

        # CSV setup (optional)
        self.csv_file = None
        self.csv_writer = None
        self.csv_file = None
        self.csv_writer = None

    def setup_csv(self):
        """Initialize CSV file with headers (only if output file specified)"""
        if not self.output_file:
            return

        self.csv_file = open(self.output_file, 'w', newline='')
        self.csv_writer = csv.writer(self.csv_file)

        # CSV header
        header = ['timestamp', 'elapsed_seconds']
        for process in self.process_names:
            header.extend([
                f'{process}_mem_mb',
                f'{process}_pid'
            ])
        self.csv_writer.writerow(header)
        self.csv_file.flush()

    def cleanup(self):
        """Close CSV file"""
        if self.csv_file:
            self.csv_file.close()

    def find_process_by_name(self, name: str) -> Optional[int]:
        """Find process ID by name, returns the first match"""
        try:
            # Search through /proc for matching process names
            for pid in os.listdir('/proc'):
                if not pid.isdigit():
                    continue

                try:
                    # Read command line to get process name
                    with open(f'/proc/{pid}/cmdline', 'r') as f:
                        cmdline = f.read()
                        # cmdline uses null bytes as separators
                        cmd_parts = cmdline.split('\x00')
                        if cmd_parts and cmd_parts[0]:
                            # Get the actual executable name
                            exe_name = os.path.basename(cmd_parts[0])
                            if name in exe_name or exe_name in name:
                                return int(pid)
                except (FileNotFoundError, PermissionError, ProcessLookupError):
                    # Process may have died or we don't have permission
                    continue
        except Exception:
            pass
        return None

    def get_process_memory(self, pid: int) -> Optional[float]:
        """Get memory usage for a process in MB (RSS - Resident Set Size)"""
        try:
            with open(f'/proc/{pid}/status', 'r') as f:
                for line in f:
                    if line.startswith('VmRSS:'):
                        # VmRSS is in kB, convert to MB
                        kb = int(line.split()[1])
                        return kb / 1024.0
        except (FileNotFoundError, PermissionError, ProcessLookupError):
            return None
        return None

    def get_process_stats(self, process_name: str) -> Optional[Dict]:
        """Get memory statistics for a specific process"""
        pid = self.find_process_by_name(process_name)
        if not pid:
            return None

        memory_mb = self.get_process_memory(pid)
        if memory_mb is None:
            return None

        return {
            'usage_mb': memory_mb,
            'pid': pid
        }

    def check_for_leaks(self, process: str, current_mb: float) -> Optional[Dict]:
        """Check if a process shows signs of memory leak using linear regression"""
        history = self.memory_history[process]
        history.append(current_mb)

        # Keep last 60 samples only
        if len(history) > 60:
            history.pop(0)

        # Set baseline after first few samples
        if process not in self.baseline_memory and len(history) >= 5:
            self.baseline_memory[process] = sum(history[:5]) / 5

        # Need at least 20 samples for leak detection
        if len(history) < 20:
            return None

        # Calculate trend using simple linear regression
        n = len(history)
        x_mean = (n - 1) / 2
        y_mean = sum(history) / n

        numerator = sum((i - x_mean) * (y - y_mean) for i, y in enumerate(history))
        denominator = sum((i - x_mean) ** 2 for i in range(n))

        if denominator == 0:
            return None

        slope = numerator / denominator  # MB per sample

        # Convert to MB per hour
        mb_per_hour = slope * (3600 / self.interval)

        # Flag as potential leak if growing >5MB/hour consistently
        if mb_per_hour > 5.0:
            growth_from_baseline = current_mb - self.baseline_memory[process]
            return {
                'rate_mb_per_hour': mb_per_hour,
                'growth_from_baseline_mb': growth_from_baseline
            }

        return None

    def format_duration(self, seconds: float) -> str:
        """Format duration in human-readable format"""
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)

        if hours > 0:
            return f"{hours}h {minutes}m {secs}s"
        elif minutes > 0:
            return f"{minutes}m {secs}s"
        else:
            return f"{secs}s"

    def monitor_loop(self):
        """Main monitoring loop"""
        self.setup_csv()

        print(f"Monitoring processes: {', '.join(self.process_names)}")
        if self.output_file:
            print(f"Logging to: {self.output_file}")
        print(f"Interval: {self.interval} seconds")
        if self.duration:
            print(f"Duration: {self.format_duration(self.duration)}")
        print("-" * 80)

        # Check if any processes exist at start
        found_any = any(self.find_process_by_name(p) for p in self.process_names)
        if not found_any:
            print(f"⚠️  Warning: None of the specified processes are currently running")
            print(f"    Will continue monitoring and log data if processes start...")
        print()

        iteration = 0

        try:
            while self.running:
                timestamp = datetime.datetime.now().isoformat()
                elapsed = time.time() - self.start_time

                # Check duration limit
                if self.duration and elapsed > self.duration:
                    print(f"\n✓ Monitoring completed after {self.format_duration(elapsed)}")
                    break

                # Collect stats for all processes
                row = [timestamp, f"{elapsed:.1f}"]
                stats_data = {}

                for process in self.process_names:
                    stats = self.get_process_stats(process)
                    if stats:
                        row.extend([
                            f"{stats['usage_mb']:.2f}",
                            str(stats['pid'])
                        ])
                        stats_data[process] = stats
                    else:
                        row.extend(['N/A', 'N/A'])

                # Write to CSV (if enabled)
                if self.csv_writer:
                    self.csv_writer.writerow(row)
                    self.csv_file.flush()

                # Update leak detection history every iteration
                leak_results = {}
                for process, stats in stats_data.items():
                    leak_info = self.check_for_leaks(process, stats['usage_mb'])
                    leak_results[process] = leak_info

                # Print summary every 10 iterations
                if iteration % 10 == 0:
                    print(f"[{self.format_duration(elapsed)}] Memory usage:")
                    if not stats_data:
                        print("  (no processes found)")
                    for process, stats in stats_data.items():
                        leak_info = leak_results.get(process)
                        leak_indicator = ""

                        if leak_info:
                            leak_indicator = f" ⚠️  LEAK: +{leak_info['rate_mb_per_hour']:.1f} MB/hour"

                        print(f"  {process:15s}: {stats['usage_mb']:7.2f} MB "
                              f"(PID: {stats['pid']}){leak_indicator}")
                iteration += 1
                time.sleep(self.interval)

        except KeyboardInterrupt:
            print("\n\nMonitoring stopped by user")
        finally:
            self.cleanup()
            self.print_summary()

    def print_summary(self):
        """Print final summary statistics"""
        print("\n" + "=" * 80)
        print("MONITORING SUMMARY")
        print("=" * 80)

        elapsed = time.time() - self.start_time
        print(f"Total monitoring time: {self.format_duration(elapsed)}")
        print(f"Samples collected: {max(len(h) for h in self.memory_history.values()) if self.memory_history else 0}")
        print()

        print("Memory statistics per process:")
        for process, history in self.memory_history.items():
            if not history:
                continue

            min_mb = min(history)
            max_mb = max(history)
            avg_mb = sum(history) / len(history)
            growth = max_mb - min_mb

            print(f"\n  {process}:")
            print(f"    Min:    {min_mb:7.2f} MB")
            print(f"    Max:    {max_mb:7.2f} MB")
            print(f"    Avg:    {avg_mb:7.2f} MB")
            print(f"    Growth: {growth:7.2f} MB ({(growth/min_mb*100):.1f}%)")

            # Final leak check
            if len(history) >= 20:
                leak_info = self.check_for_leaks(process, history[-1])
                if leak_info:
                    print(f"    ⚠️  POTENTIAL LEAK DETECTED:")
                    print(f"        Rate: +{leak_info['rate_mb_per_hour']:.1f} MB/hour")
                    print(f"        Total growth: +{leak_info['growth_from_baseline_mb']:.1f} MB from baseline")
                else:
                    print(f"    ✓ No memory leak detected")

        if self.output_file:
            print(f"\nDetailed log saved to: {self.output_file}")


def main():
    parser = argparse.ArgumentParser(
        description='Monitor process memory usage for Halko services',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                                    # Monitor default halko processes
  %(prog)s -i 5 -t 3600                       # Monitor for 1 hour with 5s interval
  %(prog)s -p controlunit simulator           # Monitor specific processes
  %(prog)s -o custom_log.csv                  # Custom output file
        """
    )

    parser.add_argument('-p', '--processes', nargs='+',
                        default=['controlunit', 'powerunit', 'simulator', 'sensorunit'],
                        help='List of process names to monitor (default: all halko services)')
    parser.add_argument('-i', '--interval', type=int, default=10,
                        help='Sampling interval in seconds (default: 10)')
    parser.add_argument('-o', '--output', default=None,
                        help='Output CSV file (optional, console output only if not specified)')
    parser.add_argument('-t', '--duration', type=int, default=None,
                        help='Total monitoring duration in seconds (default: run indefinitely)')
    parser.add_argument('--version', action='version', version='%(prog)s 1.0')

    args = parser.parse_args()

    # Validate arguments
    if args.interval < 1:
        print("Error: Interval must be at least 1 second", file=sys.stderr)
        sys.exit(1)

    if args.duration and args.duration < args.interval:
        print("Error: Duration must be greater than interval", file=sys.stderr)
        sys.exit(1)

    # Create monitor
    monitor = MemoryMonitor(
        process_names=args.processes,
        output_file=args.output,
        interval=args.interval,
        duration=args.duration
    )

    # Start monitoring (KeyboardInterrupt is handled in monitor_loop)
    monitor.monitor_loop()


if __name__ == '__main__':
    main()
