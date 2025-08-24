# -*- coding: utf-8 -*-
import sys
import json
import argparse
from datetime import datetime
from collections import defaultdict

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

def parse_time(ts):
    s = ts.rstrip('Z')
    if 'T' not in s:
        return datetime.strptime(s, '%Y-%m-%d %H:%M:%S').replace(microsecond=0)
    if '.' in s:
        base, frac = s.split('.', 1)
        frac_digits = ''.join(c for c in frac if c.isdigit())
        if len(frac_digits) > 6:
            frac_digits = frac_digits[:6]
        else:
            frac_digits = frac_digits.ljust(6, '0')
        s = f'{base}.{frac_digits}'
        fmt = '%Y-%m-%dT%H:%M:%S.%f'
    else:
        s = s + '.000000'
        fmt = '%Y-%m-%dT%H:%M:%S.%f'
    dt = datetime.strptime(s, fmt)
    return dt.replace(microsecond=0)

def read_detailed_timeseries(path):
    req_counts = defaultdict(int)
    durs_by_sec = defaultdict(list)

    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except Exception:
                continue
            if obj.get('type') != 'Point':
                continue

            metric = obj.get('metric')
            data = obj.get('data', {})
            ts = data.get('time')
            val = data.get('value')
            if ts is None or val is None:
                continue
            sec = parse_time(ts)

            if metric == 'http_reqs':
                try:
                    req_counts[sec] += int(val)
                except Exception:
                    req_counts[sec] += int(round(float(val)))
            elif metric == 'http_req_duration':
                try:
                    durs_by_sec[sec].append(float(val))
                except Exception:
                    pass

    seconds = sorted(set(req_counts.keys()) | set(durs_by_sec.keys()))
    rows = []
    for s in seconds:
        rps = req_counts.get(s, 0)
        durs = durs_by_sec.get(s, [])
        if durs:
            p50 = float(np.percentile(durs, 50))
            p90 = float(np.percentile(durs, 90))
            p95 = float(np.percentile(durs, 95))
            p99 = float(np.percentile(durs, 99))
        else:
            p50 = p90 = p95 = p99 = float('nan')
        rows.append({'second': s, 'rps': rps, 'p50_ms': p50, 'p90_ms': p90, 'p95_ms': p95, 'p99_ms': p99})

    df = pd.DataFrame(rows).sort_values('second').reset_index(drop=True)
    return df

def wide_fig():
    return plt.figure(figsize=(13.5, 4.2), dpi=160)

def main():
    parser = argparse.ArgumentParser(description="Plot time-series (RPS and P95) for 3 attempts from k6 detailed JSON.")
    parser.add_argument('--labels', nargs=3, required=True, help='Labels for the three attempts')
    parser.add_argument('files', nargs=3, help='Three k6 detailed JSON files from --out json=...')
    args = parser.parse_args()

    labels = args.labels
    files = args.files

    series = []
    for lab, path in zip(labels, files):
        df = read_detailed_timeseries(path)
        df.to_csv(f'k6_attempt_{lab.lower().replace(" ", "_")}.csv', index=False)
        series.append((lab, df))

    # Merge by second to align time for CSV export (outer join)
    merged = None
    for lab, df in series:
        sdf = df[['second', 'rps', 'p95_ms']].copy()
        sdf.columns = ['second', f'{lab}__rps', f'{lab}__p95']
        if merged is None:
            merged = sdf
        else:
            merged = pd.merge(merged, sdf, on='second', how='outer')
    if merged is not None:
        merged = merged.sort_values('second')
        merged.to_csv('k6_attempts_timeseries_merged.csv', index=False)

    # Plot RPS (all attempts on same axes)
    fig = wide_fig()
    plt.title('RPS over time (3 attempts)')
    for lab, df in series:
        if not df.empty:
            plt.plot(df['second'], df['rps'], label=lab)
    plt.xlabel('Time (s)')
    plt.ylabel('Req/s')
    plt.legend()
    plt.tight_layout()
    fig.savefig('attempts_timeseries_rps.png')

    # Plot P95 (all attempts on same axes)
    fig = wide_fig()
    plt.title('Latency P95 (ms) over time (3 attempts)')
    for lab, df in series:
        if not df.empty:
            plt.plot(df['second'], df['p95_ms'], label=lab)
    plt.xlabel('Time (s)')
    plt.ylabel('ms')
    plt.legend()
    plt.tight_layout()
    fig.savefig('attempts_timeseries_p95.png')

    print('Generated:')
    print(' - attempts_timeseries_rps.png')
    print(' - attempts_timeseries_p95.png')
    print(' - k6_attempts_timeseries_merged.csv')
    print(' - k6_attempt_<label>.csv for each attempt')

if __name__ == '__main__':
    main()
