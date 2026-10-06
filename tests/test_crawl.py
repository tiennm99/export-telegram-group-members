import os
import unittest
from unittest.mock import MagicMock, call, patch

os.environ.setdefault('REDIS_URL', 'redis://localhost')

import crawl


class FakeChat:
    def __init__(self, group_id):
        self.id = group_id
        self.title = f'Group {group_id}'


class FakeMember:
    def __init__(self, member_id):
        self.id = member_id


class CrawlArgumentsTest(unittest.TestCase):
    def test_accepts_multiple_group_ids(self):
        args = crawl.parse_args(['-100123', '-100456'])

        self.assertEqual(args.group_ids, [-100123, -100456])

    def test_allows_no_group_ids(self):
        args = crawl.parse_args([])

        self.assertEqual(args.group_ids, [])

    def test_rejects_non_integer_group_id(self):
        with (
            patch.object(crawl, 'load_app_config') as load_app_config,
            patch.object(crawl, 'TelegramClient') as telegram_client,
            self.assertRaises(SystemExit),
        ):
            crawl.main(['not-a-group'])

        load_app_config.assert_not_called()
        telegram_client.assert_not_called()


class CrawlMainTest(unittest.TestCase):
    @patch.object(crawl, 'save_group_export')
    @patch.object(crawl, 'save_session')
    @patch.object(crawl, 'load_session', return_value='stored-session')
    @patch.object(crawl, 'load_app_config')
    @patch.object(crawl, 'TelegramClient')
    def test_crawls_configured_group_ids(
        self,
        telegram_client,
        load_app_config,
        load_session,
        save_session,
        save_group_export,
    ):
        load_app_config.return_value = {
            'api_id': 123,
            'api_hash': 'hash',
            'phone': '+15551234567',
            'group_ids': [101, 202],
        }
        client = MagicMock()
        telegram_client.return_value = client
        client.is_user_authorized.return_value = True
        client.session.save.return_value = 'updated-session'
        client.get_entity.side_effect = [FakeChat(101), FakeChat(202)]
        client.get_participants.side_effect = [
            [FakeMember(2), FakeMember(1)],
            [FakeMember(4), FakeMember(3)],
        ]

        with (
            patch.object(crawl, 'Chat', FakeChat),
            patch.object(crawl, 'Channel', FakeChat),
            patch.object(crawl, 'StringSession', return_value='session'),
            patch.object(crawl, 'new_run_time', return_value='20260724120000'),
        ):
            result = crawl.main([])

        self.assertEqual(result, 0)
        self.assertEqual(client.get_entity.call_args_list, [call(101), call(202)])
        save_session.assert_called_once_with('updated-session')
        self.assertEqual(save_group_export.call_count, 2)
        first_members = save_group_export.call_args_list[0].args[2]
        self.assertEqual([member.id for member in first_members], [1, 2])

    @patch.object(crawl, 'save_group_export')
    @patch.object(crawl, 'save_session')
    @patch.object(crawl, 'load_session', return_value='stored-session')
    @patch.object(crawl, 'load_app_config')
    @patch.object(crawl, 'TelegramClient')
    def test_supplied_ids_override_configured_ids(
        self,
        telegram_client,
        load_app_config,
        load_session,
        save_session,
        save_group_export,
    ):
        load_app_config.return_value = {
            'api_id': 123,
            'api_hash': 'hash',
            'phone': '+15551234567',
            'group_ids': [999],
        }
        client = MagicMock()
        telegram_client.return_value = client
        client.is_user_authorized.return_value = True
        client.session.save.return_value = 'updated-session'
        client.get_entity.side_effect = [FakeChat(101), FakeChat(202)]
        client.get_participants.side_effect = [
            [FakeMember(2), FakeMember(1)],
            [FakeMember(4), FakeMember(3)],
        ]

        with (
            patch.object(crawl, 'Chat', FakeChat),
            patch.object(crawl, 'Channel', FakeChat),
            patch.object(crawl, 'StringSession', return_value='session'),
            patch.object(crawl, 'new_run_time', return_value='20260724120000'),
        ):
            result = crawl.main(['101', '202'])

        self.assertEqual(result, 0)
        self.assertEqual(client.get_entity.call_args_list, [call(101), call(202)])
        self.assertNotIn(call(999), client.get_entity.call_args_list)


if __name__ == '__main__':
    unittest.main()
