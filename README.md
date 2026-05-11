# export-telegram-group-members

Export Telegram group members to CSV using Telethon — admin auth required for full member visibility.

## How to use

1. Clone this repository:

```bash
git clone https://github.com/tiennm99/export-telegram-group-members.git
```

2. Install requirements:

```bash
pip install -r requirements.txt
```

3. Create a new Telegram app at [https://my.telegram.org](https://my.telegram.org) and get the `api_id` and `api_hash`.
4. Copy `.env.example` to `.env` and fill in the `api_id`, `api_hash`, and `phone`.
5. Run the script:

```bash
python main.py
```

## Output format

Results are saved to a timestamped folder (e.g. `2025-01-15 10-30-00/`). Each group produces one CSV file named after the group title:

```
id,username,first_name,last_name
123456789,johndoe,John,Doe
987654321,,Jane,Smith
```

Columns: `id`, `username`, `first_name`, `last_name`. Members without a username have an empty `username` field.

## Rate limits and visibility notes

- Telegram limits how fast you can fetch participants; large groups may take longer.
- For **supergroups**, only admins can retrieve the full member list — regular members see a partial list or get an error.
- For private groups where you are not a member, access will be denied.

## License

Apache-2.0 — see [LICENSE](LICENSE).
