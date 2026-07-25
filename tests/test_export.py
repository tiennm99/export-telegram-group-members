import csv
import io
import types
import unittest
from contextlib import redirect_stderr, redirect_stdout
from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import MagicMock, patch

import export as export_command


def export_record(group_id=-100123, run_time='20260724120000'):
    return {
        'group_id': group_id,
        'title': 'Test Group',
        'time': run_time,
        'members': [
            {
                'id': 1,
                'username': 'alice',
                'first_name': 'Alice',
                'last_name': 'Example',
            },
            {
                'id': 2,
                'username': None,
                'first_name': 'Bob',
                'last_name': None,
            },
        ],
    }


class ExportCommandTest(unittest.TestCase):
    def test_exports_exact_group_and_time(self):
        record = export_record()
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            stdout = io.StringIO()
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stdout(stdout),
            ):
                result = export_command.main(['-100123', '20260724120000'])

            output_path = output_dir / '-100123-20260724120000.csv'
            self.assertEqual(result, 0)
            common.get_group_export.assert_called_once_with(
                -100123,
                '20260724120000',
            )
            common.latest_group_export_time.assert_not_called()
            self.assertTrue(output_path.is_file())
            self.assertIn(str(output_path), stdout.getvalue())

            with output_path.open(encoding='utf-8', newline='') as csv_file:
                rows = list(csv.DictReader(csv_file))

            self.assertEqual(
                csv.DictReader(io.StringIO(output_path.read_text())).fieldnames,
                export_command.CSV_FIELDS,
            )
            self.assertEqual(rows[0], {
                'id': '1',
                'username': 'alice',
                'first_name': 'Alice',
                'last_name': 'Example',
            })
            self.assertEqual(rows[1]['username'], '')
            self.assertEqual(rows[1]['last_name'], '')

    def test_uses_latest_export_when_time_is_omitted(self):
        latest = export_record(run_time='20260724120000')
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=latest),
            latest_group_export_time=MagicMock(
                return_value='20260724120000',
            ),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stdout(io.StringIO()),
            ):
                result = export_command.main(['-100123'])

            self.assertEqual(result, 0)
            common.latest_group_export_time.assert_called_once_with(-100123)
            common.get_group_export.assert_called_once_with(
                -100123,
                '20260724120000',
            )
            self.assertTrue(
                (output_dir / '-100123-20260724120000.csv').is_file(),
            )

    def test_uses_first_configured_group_when_both_args_are_omitted(self):
        record = export_record(group_id=-100999)
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(
                return_value='20260724120000',
            ),
        )
        config = types.SimpleNamespace(
            load_app_config=MagicMock(return_value={'group_ids': [-100999]}),
        )

        with TemporaryDirectory() as temp_dir:
            with (
                patch.dict('sys.modules', {'common': common, 'config': config}),
                patch.object(
                    export_command,
                    'OUTPUT_DIR',
                    Path(temp_dir) / 'output',
                ),
                redirect_stdout(io.StringIO()),
            ):
                result = export_command.main([])

        self.assertEqual(result, 0)
        config.load_app_config.assert_called_once_with()
        common.latest_group_export_time.assert_called_once_with(-100999)
        common.get_group_export.assert_called_once_with(
            -100999,
            '20260724120000',
        )

    def test_default_group_config_failure_returns_clean_error(self):
        common = types.SimpleNamespace(
            get_group_export=MagicMock(),
            latest_group_export_time=MagicMock(),
        )
        config = types.SimpleNamespace(
            load_app_config=MagicMock(
                side_effect=RuntimeError('Redis unavailable'),
            ),
        )
        stderr = io.StringIO()

        with (
            patch.dict('sys.modules', {'common': common, 'config': config}),
            redirect_stderr(stderr),
        ):
            result = export_command.main([])

        self.assertEqual(result, 1)
        self.assertIn('could not read export configuration', stderr.getvalue())
        common.latest_group_export_time.assert_not_called()
        common.get_group_export.assert_not_called()

    def test_common_import_failure_returns_clean_error(self):
        stderr = io.StringIO()
        real_import = __import__

        def fail_common_import(name, *args, **kwargs):
            if name == 'common':
                raise ValueError('Redis URL is invalid')
            return real_import(name, *args, **kwargs)

        with (
            patch('builtins.__import__', side_effect=fail_common_import),
            redirect_stderr(stderr),
        ):
            result = export_command.main(['-100123', '20260724120000'])

        self.assertEqual(result, 1)
        self.assertIn('could not read export configuration', stderr.getvalue())

    def test_missing_exact_export_returns_error_without_creating_output(self):
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=None),
            latest_group_export_time=MagicMock(),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            stderr = io.StringIO()
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stderr(stderr),
            ):
                result = export_command.main(['-100123', '20260724120000'])

            self.assertEqual(result, 1)
            self.assertFalse(output_dir.exists())
            self.assertIn('export not found', stderr.getvalue())

    def test_missing_group_history_returns_error_without_creating_output(self):
        common = types.SimpleNamespace(
            get_group_export=MagicMock(),
            latest_group_export_time=MagicMock(return_value=None),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            stderr = io.StringIO()
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stderr(stderr),
            ):
                result = export_command.main(['-100123'])

            self.assertEqual(result, 1)
            self.assertFalse(output_dir.exists())
            self.assertIn('no exports found', stderr.getvalue())

    def test_rejects_invalid_time_before_loading_redis(self):
        for invalid_time in (
            '../unsafe',
            '20261301120000',
            '20260230010101',
            '20260724246000',
        ):
            with self.subTest(invalid_time=invalid_time):
                with (
                    patch.dict('sys.modules', {'common': None}),
                    self.assertRaises(SystemExit),
                ):
                    export_command.main(['-100123', invalid_time])

    def test_rejects_mismatched_record_without_creating_output(self):
        record = export_record(group_id='../escaped')
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            stderr = io.StringIO()
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stderr(stderr),
            ):
                result = export_command.main(['-100123', '20260724120000'])

            self.assertEqual(result, 1)
            self.assertFalse(output_dir.exists())
            self.assertIn('group ID does not match', stderr.getvalue())

    def test_rejects_impossible_time_from_redis_key(self):
        record = export_record(run_time='99999999999999')
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(
                return_value='99999999999999',
            ),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            stderr = io.StringIO()
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stderr(stderr),
            ):
                result = export_command.main(['-100123'])

            self.assertEqual(result, 1)
            self.assertFalse(output_dir.exists())
            self.assertIn('invalid crawl time', stderr.getvalue())

    def test_rejects_malformed_members_without_replacing_existing_csv(self):
        record = export_record()
        record['members'] = None
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            output_dir.mkdir()
            output_path = output_dir / '-100123-20260724120000.csv'
            output_path.write_text('existing export', encoding='utf-8')
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stderr(io.StringIO()),
            ):
                result = export_command.main(['-100123', '20260724120000'])

            self.assertEqual(result, 1)
            self.assertEqual(
                output_path.read_text(encoding='utf-8'),
                'existing export',
            )

    def test_escapes_spreadsheet_formulas_and_uses_private_permissions(self):
        record = export_record()
        record['members'][0]['first_name'] = ' =HYPERLINK("https://example.com")'
        record['members'][0]['last_name'] = '\n+SUM(1,1)'
        common = types.SimpleNamespace(
            get_group_export=MagicMock(return_value=record),
            latest_group_export_time=MagicMock(),
        )

        with TemporaryDirectory() as temp_dir:
            output_dir = Path(temp_dir) / 'output'
            with (
                patch.dict('sys.modules', {'common': common}),
                patch.object(export_command, 'OUTPUT_DIR', output_dir),
                redirect_stdout(io.StringIO()),
            ):
                result = export_command.main(['-100123', '20260724120000'])

            output_path = output_dir / '-100123-20260724120000.csv'
            with output_path.open(encoding='utf-8', newline='') as csv_file:
                rows = list(csv.DictReader(csv_file))

            self.assertEqual(result, 0)
            self.assertEqual(
                rows[0]['first_name'],
                '\' =HYPERLINK("https://example.com")',
            )
            self.assertEqual(rows[0]['last_name'], "'\n+SUM(1,1)")
            self.assertEqual(output_dir.stat().st_mode & 0o777, 0o700)
            self.assertEqual(output_path.stat().st_mode & 0o777, 0o600)


if __name__ == '__main__':
    unittest.main()
