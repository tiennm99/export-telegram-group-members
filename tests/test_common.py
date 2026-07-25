import os
import unittest
from unittest.mock import MagicMock, patch

os.environ.setdefault('REDIS_URL', 'redis://localhost')

import common


class LatestGroupExportTimeTest(unittest.TestCase):
    def test_uses_latest_timestamp_from_redis_keys_without_fetching_records(self):
        redis_client = MagicMock()
        redis_client.scan_iter.return_value = [
            'telegram-export:group:-100123:20260724120000',
            'telegram-export:group:-100123:20260723120000',
        ]

        with patch.object(common, 'redis_client', redis_client):
            run_time = common.latest_group_export_time(-100123)

        self.assertEqual(run_time, '20260724120000')
        redis_client.scan_iter.assert_called_once_with(
            match='telegram-export:group:-100123:*',
        )
        redis_client.get.assert_not_called()

    def test_ignores_malformed_keys(self):
        redis_client = MagicMock()
        redis_client.scan_iter.return_value = [
            'telegram-export:group:-100123:not-a-time',
        ]

        with patch.object(common, 'redis_client', redis_client):
            run_time = common.latest_group_export_time(-100123)

        self.assertIsNone(run_time)


if __name__ == '__main__':
    unittest.main()
