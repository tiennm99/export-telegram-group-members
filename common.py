"""Export persistence in Redis.

Each group's members are stored per export as one self-contained JSON key:

    <prefix>:run:<phone>:<yyyymmddhhmmss>:<group_id>
        ->  {group_id, title, time, members:[{id,username,first_name,last_name}]}

All groups in a single run share one timestamp. No key references another, so
deleting any key can never corrupt another's state. Writing a group export is a
single atomic SET; listing scans the run keys.
"""

import json
from datetime import datetime

from config import key, phone, redis_client


def new_run_time():
    """Timestamp shared by every group in one run: yyyymmddhhmmss."""
    return datetime.now().strftime('%Y%m%d%H%M%S')


def member_dict(member):
    return {
        'id': member.id,
        'username': member.username,
        'first_name': member.first_name,
        'last_name': member.last_name,
    }


def save_group_export(group_id, title, members, run_time):
    """Atomically persist one group's members for a run as a single JSON value."""
    record = {
        'group_id': group_id,
        'title': title,
        'time': run_time,
        'members': [member_dict(m) for m in members],
    }
    redis_client.set(
        key('run', phone, run_time, str(group_id)),
        json.dumps(record, ensure_ascii=False),
    )


def list_exports():
    """Return all group-export records, sorted by (time, group_id).

    Tolerates keys deleted mid-scan and corrupt/non-JSON values (skips them).
    """
    records = []
    for export_key in redis_client.scan_iter(match=key('run', phone, '*')):
        raw = redis_client.get(export_key)
        if raw is None:  # deleted between scan and get
            continue
        try:
            records.append(json.loads(raw))
        except (ValueError, TypeError):  # corrupt / non-JSON value: skip, don't abort
            continue
    records.sort(key=lambda r: (r.get('time', ''), r.get('group_id', 0)))
    return records
