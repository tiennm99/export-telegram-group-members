import argparse
import csv
import os
import sys
import tempfile
from datetime import datetime
from pathlib import Path

OUTPUT_DIR = Path('output')
CSV_FIELDS = ['id', 'username', 'first_name', 'last_name']
FORMULA_PREFIXES = ('=', '+', '-', '@', '\t', '\r')


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description='Export a Redis-backed Telegram group crawl to CSV.',
    )
    parser.add_argument(
        'group_id',
        type=int,
        nargs='?',
        help='Telegram group ID (defaults to the first configured group)',
    )
    parser.add_argument(
        'timecrawl',
        nargs='?',
        type=parse_run_time,
        help='Crawl time in yyyymmddhhmmss format (defaults to latest)',
    )
    return parser.parse_args(argv)


def parse_run_time(value):
    if not is_valid_run_time(value):
        if len(value) != 14 or not value.isdigit():
            raise argparse.ArgumentTypeError(
                'timecrawl must use yyyymmddhhmmss format',
            )
        raise argparse.ArgumentTypeError(
            'timecrawl must be a valid date and time',
        )
    return value


def is_valid_run_time(value):
    if not isinstance(value, str) or len(value) != 14 or not value.isdigit():
        return False
    try:
        datetime.strptime(value, '%Y%m%d%H%M%S')
    except ValueError:
        return False
    return True


def main(argv=None):
    args = parse_args(argv)

    group_id = args.group_id
    try:
        # Lazy imports keep `python export.py --help` working without REDIS_URL.
        from common import get_group_export, latest_group_export_time

        if group_id is None:
            from config import load_app_config

            group_id = load_app_config()['group_ids'][0]

        run_time = args.timecrawl
        if run_time is None:
            run_time = latest_group_export_time(group_id)
            if run_time is None:
                print(f'no exports found for group {group_id}', file=sys.stderr)
                return 1
        record = get_group_export(group_id, run_time)
    except Exception as exc:
        print(f'could not read export configuration or history: {exc}', file=sys.stderr)
        return 1

    if record is None:
        print(
            f'export not found or invalid for group {group_id} at {run_time}',
            file=sys.stderr,
        )
        return 1

    try:
        members = validate_record(record, group_id, run_time)
        output_path = write_csv(group_id, run_time, members)
    except (OSError, ValueError, csv.Error) as exc:
        print(f'could not export group {group_id} at {run_time}: {exc}', file=sys.stderr)
        return 1

    print(f'exported {len(members)} members to {output_path}')
    return 0


def validate_record(record, expected_group_id, expected_time):
    if not is_valid_run_time(expected_time):
        raise ValueError('Redis export key has an invalid crawl time')
    if not isinstance(record, dict):
        raise ValueError('Redis export is not an object')
    if record.get('group_id') != expected_group_id:
        raise ValueError('Redis export group ID does not match its key')
    if record.get('time') != expected_time:
        raise ValueError('Redis export time does not match its key')

    members = record.get('members')
    if not isinstance(members, list):
        raise ValueError('Redis export members must be a list')

    validated = []
    for index, member in enumerate(members):
        if not isinstance(member, dict):
            raise ValueError(f'member {index} is not an object')
        member_id = member.get('id')
        if not isinstance(member_id, int) or isinstance(member_id, bool):
            raise ValueError(f'member {index} has an invalid ID')

        row = {'id': member_id}
        for field in CSV_FIELDS[1:]:
            value = member.get(field)
            if value is not None and not isinstance(value, str):
                raise ValueError(f'member {index} has an invalid {field}')
            row[field] = escape_formula(value)
        validated.append(row)
    return validated


def escape_formula(value):
    if value and (
        value[0] in ('\t', '\r', '\n')
        or value.lstrip().startswith(FORMULA_PREFIXES)
    ):
        return f"'{value}"
    return value


def write_csv(group_id, run_time, members, output_dir=None):
    output_dir = output_dir or OUTPUT_DIR
    output_path = output_dir / f'{group_id}-{run_time}.csv'
    output_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
    output_dir.chmod(0o700)

    temp_path = None
    try:
        with tempfile.NamedTemporaryFile(
            mode='w',
            encoding='utf-8',
            newline='',
            dir=output_dir,
            prefix=f'.{output_path.name}.',
            delete=False,
        ) as csv_file:
            temp_path = Path(csv_file.name)
            writer = csv.DictWriter(csv_file, fieldnames=CSV_FIELDS)
            writer.writeheader()
            writer.writerows(members)
        os.chmod(temp_path, 0o600)
        os.replace(temp_path, output_path)
    except Exception:
        if temp_path is not None:
            temp_path.unlink(missing_ok=True)
        raise

    return output_path


if __name__ == '__main__':
    raise SystemExit(main())
