# -*- coding: utf-8 -*-
import sys
import json
from datetime import datetime
from collections import defaultdict

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

def parse_time(ts):
    # Example input: "2025-08-17T21:19:39.56094788Z" (8 fractional digits)
    # Normalize to microseconds (6 digits) to satisfy strptime.
    s = ts.rstrip('Z')
    if 'T' not in s:
        # Fallback: try plain seconds
        try:
            return datetime.strptime(s, '%Y-%m-%d %H:%M:%S').replace(microsecond=0)
        except Exception:
            return datetime.utcnow().replace(microsecond=0)
    if '.' in s:
        base, frac = s.split('.', 1)
        # Keep only digits in fraction
        frac_digits = ''.join(c for c in frac if c.isdigit())
        # Truncate or pad to 6 digits (microseconds)
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
    return dt.replace(microsecond=0)  # group by whole second

def load_points(path):
    req_counts = defaultdict(int)
    durations_by_sec = defaultdict(list)

    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
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
                    durations_by_sec[sec].append(float(val))
                except Exception:
                    pass

    return req_counts, durations_by_sec

def build_dataframe(req_counts, durations_by_sec):
    all_seconds = sorted(set(req_counts.keys()) | set(durations_by_sec.keys()))
    rows = []
    for sec in all_seconds:
        rps = req_counts.get(sec, 0)
        durs = durations_by_sec.get(sec, [])
        if durs:
            p50 = float(np.percentile(durs, 50))
            p90 = float(np.percentile(durs, 90))
            p95 = float(np.percentile(durs, 95))
            p99 = float(np.percentile(durs, 99))
        else:
            p50 = p90 = p95 = p99 = float('nan')
        rows.append({
            'second': sec,
            'rps': rps,
            'p50_ms': p50,
            'p90_ms': p90,
            'p95_ms': p95,
            'p99_ms': p99
        })
    df = pd.DataFrame(rows).sort_values('second').reset_index(drop=True)
    return df

def plot_series(df):
    plt.figure()
    plt.title('K6 - RPS over time')
    plt.plot(df['second'], df['rps'])
    plt.xlabel('Time (s)')
    plt.ylabel('Req/s')
    plt.tight_layout()
    plt.savefig('k6_rps_timeseries.png', dpi=160)

    plt.figure()
    plt.title('K6 - P95 latency (ms) over time')
    plt.plot(df['second'], df['p95_ms'])
    plt.xlabel('Time (s)')
    plt.ylabel('ms')
    plt.tight_layout()
    plt.savefig('k6_p95_timeseries.png', dpi=160)

def main():
    if len(sys.argv) < 2:
        print('Usage: python k6_timeseries_plots_win.py <k6_out_json_file>')
        sys.exit(1)

    path = sys.argv[1]
    req_counts, durations_by_sec = load_points(path)
    if not req_counts and not durations_by_sec:
        print('Warning: no points parsed. Ensure file was generated with: k6 run --out json=...')
    df = build_dataframe(req_counts, durations_by_sec)

    df.to_csv('k6_timeseries.csv', index=False)
    plot_series(df)

    print('Done. Files generated:')
    print(' - k6_timeseries.csv')
    print(' - k6_rps_timeseries.png')
    print(' - k6_p95_timeseries.png')

if __name__ == '__main__':
    main()
