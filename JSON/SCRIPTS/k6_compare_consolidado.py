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

def read_k6_detailed_json(path):
    req_counts = defaultdict(int)
    durs_by_sec = defaultdict(list)
    all_durs = []

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
                    v = float(val)
                    durs_by_sec[sec].append(v)
                    all_durs.append(v)
                except Exception:
                    pass

    all_seconds = sorted(set(req_counts.keys()) | set(durs_by_sec.keys()))
    rows = []
    for s in all_seconds:
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
    df = pd.DataFrame(rows)
    if not df.empty and 'second' in df.columns:
        df = df.sort_values('second').reset_index(drop=True)
    else:
        # cria DataFrame vazio com colunas padrão
        df = pd.DataFrame(columns=['second','rps','p50_ms','p90_ms','p95_ms','p99_ms'])

    return df, all_durs

def wide_fig():
    return plt.figure(figsize=(13.5, 4.2), dpi=160)

def main():
    parser = argparse.ArgumentParser(description='Compare 3 k6 detailed JSON files and output wide charts.')
    parser.add_argument('--labels', nargs=3, required=True, help='Three labels, in order of the 3 input files')
    parser.add_argument('files', nargs=3, help='Three k6 detailed JSON files from --out json=...')
    args = parser.parse_args()

    labels = args.labels
    files = args.files

    series = []
    percentiles_flat = []
    avg_rps_values = []

    for lab, path in zip(labels, files):
        df, all_durs = read_k6_detailed_json(path)
        series.append((lab, df))
        total_reqs = df['rps'].sum()
        total_secs = len(df)
        avg_rps = total_reqs / total_secs if total_secs > 0 else 0.0
        avg_rps_values.append(avg_rps)
        if all_durs:
            p50 = float(np.percentile(all_durs, 50))
            p90 = float(np.percentile(all_durs, 90))
            p95 = float(np.percentile(all_durs, 95))
            p99 = float(np.percentile(all_durs, 99))
        else:
            p50 = p90 = p95 = p99 = float('nan')
        percentiles_flat.append((lab, p50, p90, p95, p99))

    fig = wide_fig()
    plt.title('Average RPS (per second mean)')
    x = np.arange(len(labels))
    plt.bar(x, avg_rps_values)
    plt.xticks(x, labels, rotation=0, ha='center')
    plt.ylabel('Req/s')
    plt.tight_layout()
    fig.savefig('compare_avg_rps.png')

    labs = [t[0] for t in percentiles_flat]
    p50s = [t[1] for t in percentiles_flat]
    p90s = [t[2] for t in percentiles_flat]
    p95s = [t[3] for t in percentiles_flat]
    p99s = [t[4] for t in percentiles_flat]

    for name, vals in [('p50', p50s), ('p90', p90s), ('p95', p95s), ('p99', p99s)]:
        fig = wide_fig()
        plt.title(f'Latency {name.upper()} (ms)')
        x = np.arange(len(labs))
        plt.bar(x, vals)
        plt.xticks(x, labs, rotation=0, ha='center')
        plt.ylabel('ms')
        plt.tight_layout()
        fig.savefig(f'compare_{name}.png')

    fig = wide_fig()
    plt.title('RPS over time')
    for lab, df in series:
        if not df.empty:
            plt.plot(df['second'], df['rps'], label=lab)
    plt.xlabel('Time (s)')
    plt.ylabel('Req/s')
    plt.legend()
    plt.tight_layout()
    fig.savefig('timeseries_rps.png')

    fig = wide_fig()
    plt.title('Latency P95 (ms) over time')
    for lab, df in series:
        if not df.empty:
            plt.plot(df['second'], df['p95_ms'], label=lab)
    plt.xlabel('Time (s)')
    plt.ylabel('ms')
    plt.legend()
    plt.tight_layout()
    fig.savefig('timeseries_p95.png')

    for lab, df in series:
        safe_lab = lab.lower().replace(' ', '_')
        df.to_csv(f'k6_{safe_lab}_timeseries.csv', index=False)

    if series:
        merged = None
        for lab, df in series:
            sdf = df[['second', 'rps', 'p50_ms', 'p90_ms', 'p95_ms', 'p99_ms']].copy()
            sdf.columns = ['second'] + [f'{lab}__rps', f'{lab}__p50', f'{lab}__p90', f'{lab}__p95', f'{lab}__p99']
            if merged is None:
                merged = sdf
            else:
                merged = pd.merge(merged, sdf, on='second', how='outer')
        if merged is not None:
            merged = merged.sort_values('second')
            merged.to_csv('k6_timeseries_merged.csv', index=False)

if __name__ == '__main__':
    main()
