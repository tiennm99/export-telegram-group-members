# Quickstart: Telegram Group Members Export

## Prerequisites

- Go 1.21 or higher
- Telegram API credentials (API ID and API Hash) - obtain from https://my.telegram.org
- Phone number registered with Telegram

## Setup

1. **Get Telegram API credentials:**
   - Visit https://my.telegram.org
   - Login with your phone number
   - Go to "API development tools"
   - Create new application to get API ID and API Hash

2. **Clone and setup the project:**
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

3. **Create environment file:**
   ```bash
   cp .env.example .env
   ```

4. **Configure environment variables:**
   ```env
   PHONE=+1234567890          # Your phone number in international format
   API_ID=12345678             # Your Telegram API ID
   API_HASH=your_api_hash      # Your Telegram API Hash
   GROUP_ID=-1001234567890     # Target group ID (negative for supergroups)

   # Optional proxy settings (if needed)
   PROXY_HOST=proxy.example.com
   PROXY_PORT=443
   PROXY_SECRET=your_secret
   ```

## Building the Application

```bash
go mod init telegram-exporter
go get github.com/celestix/gotgproto
go get github.com/xelaj/mtproto
go build -o telegram-exporter main.go
```

## Running the Application

```bash
# Run the application (authentication code will be requested)
./telegram-exporter

# Or run directly with Go
go run main.go
```

## Expected Output

The application will:
1. Prompt for authentication code sent to your Telegram
2. Connect to the specified group
3. Export member information to `out/{groupname}_members_YYYYMMDD.json`
4. Display progress during the export process

## Example Output File

```json
{
  "groupId": -1001234567890,
  "groupName": "Example Group",
  "exportTime": "2025-12-06T17:11:00Z",
  "memberCount": 150,
  "members": [
    {
      "id": 123456789,
      "username": "example_user",
      "firstName": "Example",
      "lastName": "User"
    },
    ...
  ]
}
```

## Troubleshooting

**Authentication Issues:**
- Ensure phone number is in international format (+1234567890)
- Verify API credentials are correct
- Check if phone number is properly registered with Telegram

**Group Access Issues:**
- Verify the group ID is correct
- Ensure you have access to the group
- Confirm you're using the correct format (negative ID for supergroups)

**Network Issues:**
- Check internet connection
- If behind firewall, try using proxy settings
- Verify API endpoints are not blocked in your region