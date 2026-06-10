"""Export persistence in Redis.

Each group's members are stored per export as one self-contained JSON key:

    <prefix>:group:<group_id>:<yyyymmddhhmmss>
        ->  {group_id, title, time, members:[{id,username,first_name,last_name}]}

All groups in a single run share one timestamp. No key references another, so
deleting any key can never corrupt another's state. Writing a group export is a
single atomic SET; listing scans the group export keys.
"""

import json
from datetime import datetime

from config import key, redis_client


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
        key('group', str(group_id), run_time),
        json.dumps(record, ensure_ascii=False),
    )


def list_exports():
    """Return all group-export records, sorted by (time, group_id).

    Tolerates keys deleted mid-scan and corrupt/non-JSON values (skips them).
    """
    records = []
    for export_key in redis_client.scan_iter(match=key('group', '*', '*')):
        if not _is_group_export_key(export_key):
            continue
        raw = redis_client.get(export_key)
        if raw is None:  # deleted between scan and get
            continue
        try:
            records.append(json.loads(raw))
        except (ValueError, TypeError):  # corrupt / non-JSON value: skip, don't abort
            continue
    records.sort(key=lambda r: (r.get('time', ''), r.get('group_id', 0)))
    return records


def list_group_exports(group_id):
    """Return export records for one group, sorted by run time."""
    group_id = str(group_id)
    return [
        record for record in list_exports()
        if str(record.get('group_id')) == group_id
    ]


def get_group_export(group_id, run_time):
    """Return a group export at one run time, or None when missing."""
    raw = redis_client.get(key('group', str(group_id), run_time))
    if raw is None:
        return None
    try:
        return json.loads(raw)
    except (ValueError, TypeError):
        return None


def diff_group_members(before_record, after_record):
    """Compare two group export records by Telegram member id."""
    before_members = _member_index(before_record)
    after_members = _member_index(after_record)

    before_ids = set(before_members)
    after_ids = set(after_members)

    added = [after_members[member_id] for member_id in sorted(after_ids - before_ids)]
    removed = [before_members[member_id] for member_id in sorted(before_ids - after_ids)]
    return added, removed


def _member_index(record):
    return {
        int(member['id']): member
        for member in record.get('members', [])
        if member.get('id') is not None
    }


def _is_run_time(value):
    return len(value) == 14 and value.isdigit()


def _is_group_export_key(export_key):
    parts = export_key.split(':')
    return len(parts) == 4 and _is_run_time(parts[3])
