# Data Model: Telegram Group Members Export

## Entity Definitions

### GroupMember
**Description**: Represents a Telegram user within a group

**Fields**:
- `ID` (int64): Unique user identifier from Telegram
- `Username` (string): User's Telegram username (may be empty)
- `FirstName` (string): User's first name
- `LastName` (string): User's last name (may be empty)
- `PhoneNumber` (string): User's phone number (restricted by privacy settings)
- `IsBot` (bool): Whether the user is a bot account
- `IsScam` (bool): Whether the user is marked as scam
- `IsFake` (bool): Whether the user is marked as fake

**Validation Rules**:
- ID must be a positive integer
- FirstName cannot be empty (fallback to LastName if needed)

### ExportConfiguration
**Description**: Settings that define the export parameters

**Fields**:
- `APIID` (int): Telegram API ID from environment
- `APIHash` (string): Telegram API hash from environment
- `PhoneNumber` (string): Phone number for authentication
- `GroupID` (int64): Target group ID for member export
- `OutputDir` (string): Directory for output files (default: "out")
- `OutputFormat` (string): Output format (default: "json")
- `ProxyHost` (string): Optional proxy host
- `ProxyPort` (int): Optional proxy port
- `ProxySecret` (string): Optional proxy secret

**Validation Rules**:
- APIID must be a positive integer
- APIHash must be a valid hex string
- PhoneNumber must be in international format
- GroupID must be a valid Telegram group ID

### ExportResult
**Description**: Represents the result of an export operation

**Fields**:
- `GroupID` (int64): The group ID that was exported
- `GroupName` (string): Name of the group (retrieved from Telegram)
- `MemberCount` (int): Number of members exported
- `OutputFilePath` (string): Path to the generated file
- `ExportTime` (time.Time): Timestamp of export completion
- `Success` (bool): Whether the export was successful
- `ErrorMessage` (string): Error message if export failed

## State Transitions

### Export Process States
1. **Initialization**: Configuration loaded and validated
2. **Authentication**: User authentication in progress
3. **Connecting**: Connection to Telegram established
4. **Retrieving**: Group members being retrieved
5. **Processing**: Members data being processed
6. **Exporting**: Data being written to file
7. **Completed**: Export finished successfully
8. **Failed**: Export process failed

## Data Flow

### Member Data Flow
```
Telegram API → GroupMember entities → JSON serialization → Output file
```

### Configuration Data Flow
```
Environment variables → ExportConfiguration → Telegram client → Export process
```