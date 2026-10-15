import io
import os
import unittest
from contextlib import redirect_stdout
from unittest.mock import MagicMock, patch

os.environ.setdefault('REDIS_URL', 'redis://localhost')

import common
import compare_groups
import crawl


class FakeChat:
    def __init__(self, group_id):
        self.id = group_id
        self.title = f'Group {group_id}'


class FakeMember:
    def __init__(self, member_id, username=None):
        self.id = member_id
        self.username = username
        self.first_name = None
        self.last_name = None


def group(group_id, member_ids):
    return {
        'group_id': group_id,
        'title': f'Group {group_id}',
        'members': [{'id': i, 'username': None, 'first_name': None, 'last_name': None}
                    for i in member_ids],
    }


class SplitMembersTest(unittest.TestCase):
    def test_splits_by_member_id(self):
        only_a, only_b, both = compare_groups.split_members(
            group(1, [3, 1, 2]), group(2, [4, 2, 3]),
        )

        self.assertEqual([m['id'] for m in only_a], [1])
        self.assertEqual([m['id'] for m in only_b], [4])
        self.assertEqual([m['id'] for m in both], [2, 3])


class CompareGroupsMainTest(unittest.TestCase):
    def test_requires_two_group_ids(self):
        with self.assertRaises(SystemExit):
            compare_groups.parse_args(['-100123'])

    @patch.object(common, 'save_group_export')
    @patch.object(crawl, 'save_session')
    @patch.object(crawl, 'load_session', return_value='stored-session')
    @patch.object(crawl, 'TelegramClient')
    def test_fetches_live_and_does_not_save_exports(
        self, telegram_client, load_session, save_session, save_group_export,
    ):
        client = MagicMock()
        telegram_client.return_value = client
        client.is_user_authorized.return_value = True
        client.get_entity.side_effect = [FakeChat(101), FakeChat(202)]
        client.get_participants.side_effect = [
            [FakeMember(2, 'bob'), FakeMember(1, 'amy')],
            [FakeMember(3, 'cat'), FakeMember(2, 'bob')],
        ]
        app_config = {'api_id': 1, 'api_hash': 'h', 'phone': 'p', 'group_ids': [9]}

        out = io.StringIO()
        with (
            patch('config.load_app_config', return_value=app_config),
            patch.object(crawl, 'Chat', FakeChat),
            patch.object(crawl, 'Channel', FakeChat),
            patch.object(crawl, 'StringSession', return_value='session'),
            redirect_stdout(out),
        ):
            result = compare_groups.main(['101', '202'])

        self.assertEqual(result, 0)
        save_group_export.assert_not_called()
        lines = out.getvalue().splitlines()
        self.assertIn('  2   @bob', lines)
        self.assertIn('- 1   @amy', lines)
        self.assertIn('+ 3   @cat', lines)

    @patch.object(crawl, 'save_session')
    @patch.object(crawl, 'load_session', return_value='stored-session')
    @patch.object(crawl, 'TelegramClient')
    def test_non_group_returns_error(self, telegram_client, load_session, save_session):
        client = MagicMock()
        telegram_client.return_value = client
        client.is_user_authorized.return_value = True
        client.get_entity.return_value = object()
        app_config = {'api_id': 1, 'api_hash': 'h', 'phone': 'p', 'group_ids': [9]}

        with (
            patch('config.load_app_config', return_value=app_config),
            patch.object(crawl, 'StringSession', return_value='session'),
            redirect_stdout(io.StringIO()),
            patch('sys.stderr', io.StringIO()),
        ):
            result = compare_groups.main(['101', '202'])

        self.assertEqual(result, 1)
        client.get_participants.assert_not_called()


if __name__ == '__main__':
    unittest.main()
