"""One-off migration: import old local CSV export folders into Redis.

Old layout : '<YYYY-MM-DD HH-MM-SS>/<sanitized group title>.csv'
New key    : <prefix>:run:<phone>:<yyyymmddhhmmss>:<group_id>
             -> {group_id, title, time, members:[{id,username,first_name,last_name}]}

Idempotent: re-running overwrites the same keys. CSV filenames carry sanitized
titles (special chars stripped), so we map by exact filename stem to the real
title + group_id. Empty CSV fields are normalized to None to match live exports.
"""

import csv
import glob
import json
import os
import sys
from datetime import datetime

# Vietnamese group titles need UTF-8 stdout (Windows console defaults to cp1252).
sys.stdout.reconfigure(encoding='utf-8')

from config import key, phone, redis_client

# Sanitized CSV filename stem -> (group_id, real title)
GROUP_MAP = {
    'ZingPlay Game Studios': (-1001480682135, 'ZingPlay Game Studios'),
    'ZPS HCM': (-230962353, 'ZPS HCM'),
    'ZPS HCM - Xin nghỉ (phépđi trễ)': (-1001660824205, 'ZPS HCM - Xin nghỉ (phép/đi trễ)'),
}

FOLDER_GLOB = '20[0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]-[0-9][0-9]-[0-9][0-9]'


def _norm(value):
    """CSV stores missing fields as ''; live exports use None. Normalize to None."""
    return value if value else None


def parse_members(csv_path):
    members = []
    with open(csv_path, encoding='UTF-8', newline='') as f:
        reader = csv.DictReader(f)
        for row in reader:
            members.append({
                'id': int(row['id']),
                'username': _norm(row['username']),
                'first_name': _norm(row['first_name']),
                'last_name': _norm(row['last_name']),
            })
    return members


def main():
    folders = sorted(glob.glob(FOLDER_GLOB))
    keys_written = 0
    members_total = 0
    skipped = []

    for folder in folders:
        run_time = datetime.strptime(folder, '%Y-%m-%d %H-%M-%S').strftime('%Y%m%d%H%M%S')
        for csv_path in sorted(glob.glob(os.path.join(folder, '*.csv'))):
            stem = os.path.splitext(os.path.basename(csv_path))[0]
            if stem not in GROUP_MAP:
                skipped.append(csv_path)
                continue
            group_id, title = GROUP_MAP[stem]
            members = parse_members(csv_path)
            record = {'group_id': group_id, 'title': title, 'time': run_time, 'members': members}
            redis_client.set(
                key('run', phone, run_time, str(group_id)),
                json.dumps(record, ensure_ascii=False),
            )
            keys_written += 1
            members_total += len(members)
            print(f'{run_time} {group_id} {title}: {len(members)} members')

    print(f'\nmigrated {keys_written} key(s) from {len(folders)} folder(s); '
          f'{members_total} member rows total.')
    if skipped:
        print(f'skipped {len(skipped)} unmapped CSV(s):')
        for path in skipped:
            print(f'  {path}')


if __name__ == '__main__':
    main()
