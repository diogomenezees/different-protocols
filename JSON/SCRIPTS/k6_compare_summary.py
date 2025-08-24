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

def read_summary_export(path):
    with open(path, 'r', encoding='utf-8') as f:
        data = json.load(f)
    metrics = data.get("metrics", {})
    http_reqs = metrics.get("http_reqs", {})
    http_req_duration = metrics.get("http_req_duration", {})
    avg_rps = http_reqs.get("rate", 0.0)
    p50 = http_req_duration.get("p(50)", http_req_duration.get("percentiles", {}).get("p(50)", 0.0))
    p90 = http_req_duration.get("p(90)", http_req_duration.get("percentiles", {}).get("p(90)", 0.0))
    p95 = http_req_duration.get("p(95)", http_req_duration.get("percentiles", {}).get("p(95)", 0.0))
    p99 = http_req_duration.get("p(99)", http_req_duration.get("percentiles", {}).get("p(99)", 0.0))
    return avg_rps, p50, p90, p95, p99

def read_detailed(path):
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
                req_counts[sec] += int(val)
            elif metric == 'http_req_duration':
                v = float(val)
                durs_by_sec[sec].append(v)
                all_durs.append(v)
    all_seconds = sorted(set(req_counts.keys()) | set(durs_by_sec.keys()))
    rows = []
    for s in all_seconds:
        rps = req_counts.get(s, 0)
        rows.append({'second': s, 'rps': rps})
    df = pd.DataFrame(rows)
    total_reqs = df['rps'].sum() if not df.empty else 0
    total_secs = len(df)
    avg_rps = total_reqs / total_secs if total_secs > 0 else 0.0
    if all_durs:
        p50 = float(np.percentile(all_durs, 50))
        p90 = float(np.percentile(all_durs, 90))
        p95 = float(np.percentile(all_durs, 95))
        p99 = float(np.percentile(all_durs, 99))
    else:
        p50 = p90 = p95 = p99 = 0.0
    return avg_rps, p50, p90, p95, p99

def detect_and_read(path):
    try:
        with open(path, 'r', encoding='utf-8') as f:
            first = f.read(2000)
            f.seek(0)
            if '"metrics"' in first:
                return read_summary_export(path)
            else:
                return read_detailed(path)
    except Exception as e:
        print(f"Erro lendo {path}: {e}")
        return 0,0,0,0,0

def wide_fig():
    return plt.figure(figsize=(13.5,4.2), dpi=160)

def main():
    parser = argparse.ArgumentParser(description="Comparar 3 tentativas da mesma abordagem k6")
    parser.add_argument('--labels', nargs=3, required=True, help='Labels para as 3 tentativas')
    parser.add_argument('files', nargs=3, help='Arquivos JSON do k6 (summary-export ou out json)')
    args = parser.parse_args()

    labels = args.labels
    files = args.files

    results = []
    for lab, path in zip(labels, files):
        avg_rps, p50, p90, p95, p99 = detect_and_read(path)
        results.append((lab, avg_rps, p50, p90, p95, p99))

    df = pd.DataFrame(results, columns=['label','avg_rps','p50','p90','p95','p99'])
    df.to_csv('compare_attempts.csv', index=False)

    fig = wide_fig()
    plt.title("Average RPS")
    x = np.arange(len(labels))
    plt.bar(x, df['avg_rps'])
    plt.xticks(x, labels)
    plt.ylabel("req/s")
    plt.tight_layout()
    fig.savefig("attempts_avg_rps.png")

    for col in ['p50','p90','p95','p99']:
        fig = wide_fig()
        plt.title(f"Latency {col.upper()} (ms)")
        plt.bar(x, df[col])
        plt.xticks(x, labels)
        plt.ylabel("ms")
        plt.tight_layout()
        fig.savefig(f"attempts_{col}.png")

    print("Arquivos gerados: compare_attempts.csv, attempts_avg_rps.png, attempts_p50.png, attempts_p90.png, attempts_p95.png, attempts_p99.png")

if __name__ == '__main__':
    main()
