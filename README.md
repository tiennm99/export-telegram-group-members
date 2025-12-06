# Telegram Group Members Exporter

A Go application that exports Telegram group members to JSON format.

## Features

- Export members from Telegram groups using group ID
- Support for both regular groups and supergroups/channels
- Configurable member information fields
- Rate limiting and retry handling
- Progress tracking during export
- Support for proxy connections

## Prerequisites

- Go 1.21 or higher
- Telegram API credentials (API ID and API Hash) from https://my.telegram.org
- Phone number registered with Telegram

## Setup

1. **Get Telegram API credentials:**
   - Visit https://my.telegram.org
   - Login with your phone number
   - Go to "API development tools"
   - Create new application to get API ID and API Hash

2. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

3. **Create environment file:**
   ```bash
   cp .env.example .env
   ```

4. **Configure environment variables in `.env`:**
   ```env
   PHONE=+1234567890
   API_ID=12345678
   API_HASH=your_api_hash_here
   GROUP_ID=-1001234567890
   OUTPUT_DIR=out
   VERBOSE=false

   # Export options (default: true for ID, username, first_name, last_name; false for others for privacy)
   INCLUDE_ID=true
   INCLUDE_USERNAME=true
   INCLUDE_FIRST_NAME=true
   INCLUDE_LAST_NAME=true
   INCLUDE_IS_BOT=false
   INCLUDE_IS_SCAM=false
   INCLUDE_IS_FAKE=false
   INCLUDE_PHONE_NUMBER=false

   # Optional proxy settings
   PROXY_HOST=proxy.example.com
   PROXY_PORT=443
   PROXY_SECRET=your_proxy_secret
   ```

## Usage

### Basic Usage

1. Run the application:
   ```bash
   go run main.go
   ```

2. If `GROUP_ID` is not specified in the environment, the application will show a list of accessible groups for you to select from.

3. Enter the authentication code sent to your Telegram when prompted.

4. The application will export the member list to a JSON file in the `out/` directory.

### Configuration Options

- `GROUP_ID`: Specify a group ID directly to export from that group
- `OUTPUT_DIR`: Directory to save the exported JSON file (default: `out`)
- `VERBOSE`: Enable verbose logging (default: `false`)
- `INCLUDE_*`: Control which member fields to include in the export

## Export Format

The exported JSON file contains:

```json
{
  "members": [
    {
      "id": 123456789,
      "username": "username",
      "first_name": "First",
      "last_name": "Last",
      "is_bot": false,
      "is_scam": false,
      "is_fake": false,
      "phone_number": "1234567890"
    }
  ],
  "memberCount": 150,
  "exportTime": "2025-12-06T17:11:00Z"
}
```

## Building

To build the application:

```bash
go build -o telegram-exporter main.go
```

Then run:

```bash
./telegram-exporter
```

## Error Handling

The application includes robust error handling for:
- Authentication failures
- Invalid group IDs or inaccessible groups
- Rate limiting by Telegram API
- Network timeouts
- Transient network failures (with retry logic)

## Performance

- Supports exporting large groups with thousands of members
- Implements pagination to handle large datasets
- Includes performance timing to monitor export duration

## Security Considerations

- Store your Telegram API credentials securely
- Don't commit `.env` file to version control
- Be mindful of privacy settings when exporting member information
- Only export member information you have permission to access

## Troubleshooting

### Authentication Issues
- Ensure phone number is in international format (+1234567890)
- Verify API credentials are correct
- Check that phone number is properly registered with Telegram

### Group Access Issues
- Verify the group ID is correct
- Ensure you have access to the group
- Confirm you're using the correct format (negative ID for supergroups)

### Network Issues
- Check internet connection
- If behind firewall, try using proxy settings
- Verify API endpoints are not blocked in your region

## License

This project is licensed under the terms specified in the LICENSE file.