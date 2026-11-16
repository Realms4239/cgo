#!/usr/bin/env python3
"""Consolidation of the 2026-11-16/17 exhaustive live-bench layers.

Read-only: never writes guest data, never touches the frozen corpus.
Usage: python3 consolidate.py <run-dir> [run-dir ...]
Each run-dir is a local copy of a guest bench-data/<run-id> directory.
Prints a compact Markdown table with per-cell counts, medians of the valid
cells, and an explicit list of columns that are NOT usable from a given run.
"""
import csv
import glob
import os
import statistics
import sys

# Columns invalidated by the tc -j root-counter bug (fixed in 5e00cd9).
# Runs launched with the old /home/bench/cgo binary must not use these
# columns as evidence of zero loss or as a cost model.
LEGACY_ZERO_COLUMNS = ("drops", "retransmissions", "wasted_bytes", "cost_ar_per_h")
NUMERIC = ("rtt_p50_ms", "rtt_p95_ms", "qdi_ms", "voip_r", "jfi_pct",
           "small_p95_ms", "deadline_ok_pct", "bulk_goodput_mbps", "cpu_pct",
           "rtt_base_p50_ms", "rtt_base_p95_ms")


def load(path):
    with open(os.path.join(path, "aqm_eval.csv"), encoding="utf-8") as f:
        return list(csv.DictReader(f))


def med(rows, col):
    vals = []
    for r in rows:
        try:
            v = float(r[col])
            if v > 0 or col == "qdi_ms":
                vals.append(v)
        except (ValueError, TypeError):
            pass
    return f"{statistics.median(vals):.1f}" if vals else "—"


def main(paths):
    print("| run | cells | valid | invalid | degraded | rtt_p95 med | small_p95 med | goodput med |")
    print("|---|---:|---:|---:|---:|---:|---:|---:|")
    for p in paths:
        rows = load(p)
        valid = [r for r in rows if r["gate_status"] == "valid"]
        counts = {k: sum(1 for r in rows if r["gate_status"] == k)
                  for k in ("valid", "invalid", "degraded")}
        print(f"| {os.path.basename(p.rstrip('/'))} | {len(rows)} | "
              f"{counts['valid']} | {counts['invalid']} | {counts['degraded']} | "
              f"{med(valid, 'rtt_p95_ms')} | {med(valid, 'small_p95_ms')} | "
              f"{med(valid, 'bulk_goodput_mbps')} |")
        if any(all(float(r[c] or 0) == 0 for r in rows) for c in LEGACY_ZERO_COLUMNS):
            print(f"\n> {os.path.basename(p.rstrip('/'))}: "
                  f"{', '.join(LEGACY_ZERO_COLUMNS)} are structurally zero "
                  f"(old binary, pre-`5e00cd9`). Not loss-free proof; do not cost-model.")


if __name__ == "__main__":
    args = sys.argv[1:]
    if not args:
        args = sorted(glob.glob("docs/verification/*/run-*"))
    main(args)
