package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/ext"
	"github.com/xelaj/mtproto"
	"github.com/xelaj/mtproto/telegram"
	"github.com/joho/godotenv"
)

// GroupMember represents a Telegram user within a group
type GroupMember struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsBot     bool   `json:"is_bot,omitempty"`
	IsScam    bool   `json:"is_scam,omitempty"`
	IsFake    bool   `json:"is_fake,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// ExportConfiguration settings that define the export parameters
type ExportConfiguration struct {
	APIID       int    `env:"API_ID"`
	APIHash     string `env:"API_HASH"`
	PhoneNumber string `env:"PHONE"`
	GroupID     int64  `env:"GROUP_ID"`
	OutputDir   string `env:"OUTPUT_DIR"`
	ProxyHost   string `env:"PROXY_HOST"`
	ProxyPort   int    `env:"PROXY_PORT"`
	ProxySecret string `env:"PROXY_SECRET"`
	Verbose     bool   `env:"VERBOSE"`
}

// ExportOptions specifies which fields to include in export
type ExportOptions struct {
	IncludeID        bool
	IncludeUsername  bool
	IncludeFirstName bool
	IncludeLastName  bool
	IncludeIsBot     bool
	IncludeIsScam    bool
	IncludeIsFake    bool
	IncludePhoneNumber bool
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	GroupID        int64     `json:"group_id"`
	GroupName      string    `json:"group_name"`
	MemberCount    int       `json:"member_count"`
	OutputFilePath string    `json:"output_file_path"`
	ExportTime     time.Time `json:"export_time"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// LoadExportOptions reads environment variables for export options
func LoadExportOptions() ExportOptions {
	return ExportOptions{
		IncludeID:        getBoolEnv("INCLUDE_ID", true),
		IncludeUsername:  getBoolEnv("INCLUDE_USERNAME", true),
		IncludeFirstName: getBoolEnv("INCLUDE_FIRST_NAME", true),
		IncludeLastName:  getBoolEnv("INCLUDE_LAST_NAME", true),
		IncludeIsBot:     getBoolEnv("INCLUDE_IS_BOT", false), // Default to false for privacy
		IncludeIsScam:    getBoolEnv("INCLUDE_IS_SCAM", false), // Default to false
		IncludeIsFake:    getBoolEnv("INCLUDE_IS_FAKE", false), // Default to false
		IncludePhoneNumber: getBoolEnv("INCLUDE_PHONE_NUMBER", false), // Default to false for privacy
	}
}

// getBoolEnv gets boolean environment variable or returns default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "1", "t", "T", "true", "TRUE", "True", "yes", "YES", "Yes", "on", "ON", "On":
			return true
		case "0", "f", "F", "false", "FALSE", "False", "no", "NO", "No", "off", "OFF", "Off":
			return false
		default:
			return defaultValue
		}
	}
	return defaultValue
}

// LoadConfig reads environment variables into ExportConfiguration
func LoadConfig() (*ExportConfiguration, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	config := &ExportConfiguration{
		APIID:       getIntEnv("API_ID", 0),
		APIHash:     getEnv("API_HASH", ""),
		PhoneNumber: getEnv("PHONE", ""),
		GroupID:     getInt64Env("GROUP_ID", 0),
		OutputDir:   getEnv("OUTPUT_DIR", "out"),
		ProxyHost:   getEnv("PROXY_HOST", ""),
		ProxyPort:   getIntEnv("PROXY_PORT", 0),
		ProxySecret: getEnv("PROXY_SECRET", ""),
		Verbose:     getBoolEnv("VERBOSE", false),
	}

	return config, nil
}

// ValidateConfig validates required configuration fields
func ValidateConfig(config *ExportConfiguration) error {
	if config.APIID == 0 {
		return fmt.Errorf("API_ID is required")
	}
	if config.APIHash == "" {
		return fmt.Errorf("API_HASH is required")
	}
	if config.PhoneNumber == "" {
		return fmt.Errorf("PHONE is required")
	}
	if config.GroupID == 0 {
		return fmt.Errorf("GROUP_ID is required")
	}

	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets integer environment variable or returns default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

// getInt64Env gets int64 environment variable or returns default value
func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

// InitializeTelegramClient initializes Telegram client using gotgproto with configuration values
func InitializeTelegramClient(config *ExportConfiguration) (*gotgproto.Client, error) {
	// Create device config with timeout settings
	deviceConfig := &mtproto.DeviceConfig{
		DeviceModel:    "Go Telegram Exporter",
		SystemVersion:  "1.0.0",
		AppVersion:     "1.0.0",
		SystemLangCode: "en",
		LangCode:       "en",
	}

	// Create MTProto config with timeout settings
	mtprotoConfig := &mtproto.Config{
		AppID:   config.APIID,
		AppHash: config.APIHash,
		Device:  deviceConfig,
		Timeout: 30 * time.Second, // 30 second timeout for network requests
	}

	client, err := gotgproto.NewClient(
		mtprotoConfig,
		gotgproto.ClientTypeUser(config.PhoneNumber),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram client: %v", err)
	}

	return client, nil
}

// AuthenticateTelegram handles the authentication flow with code request and sign-in
func AuthenticateTelegram(client *gotgproto.Client, config *ExportConfiguration) error {
	// Send code request
	sendCodeErr := client.SendCode(config.PhoneNumber)
	if sendCodeErr != nil {
		return fmt.Errorf("failed to send authentication code to %s. Please check your phone number and API credentials: %v", config.PhoneNumber, sendCodeErr)
	}

	// Ask user for the code they received
	fmt.Print("Enter the code you received: ")
	var code string
	fmt.Scanln(&code)

	// Sign in with the code
	authResult, signInErr := client.SignIn(config.PhoneNumber, code)
	if signInErr != nil {
		// If 2FA is enabled, we need to handle password
		if authResult.RequiredAuth == gotgproto.AuthTypePassword {
			fmt.Print("2FA Password required. Enter password: ")
			var password string
			fmt.Scanln(&password)

			_, signInErr = client.AuthWithPassword(password)
			if signInErr != nil {
				return fmt.Errorf("failed to authenticate with 2FA password. Please check your password and try again: %v", signInErr)
			}
		} else {
			return fmt.Errorf("failed to sign in with the provided code. Please check the code and try again: %v", signInErr)
		}
	}

	fmt.Println("Authentication successful!")
	return nil
}

// GetGroupMembers retrieves all members from specified group ID using Telegram client
func GetGroupMembers(client *gotgproto.Client, groupID int64) ([]GroupMember, error) {
	// Convert group ID to input peer
	inputPeer := &telegram.InputPeerChat{
		ChatId: int32(groupID),
	}

	// Check if it's a supergroup (negative ID)
	if groupID < 0 {
		inputPeer = &telegram.InputPeerChannel{
			ChannelId:    int32(-groupID), // Channels use positive IDs
			AccessHash:   0,               // Will be retrieved if needed
		}
	}

	// Get full channel info to retrieve access hash and other details
	// We'll try to get it first to see if it's a channel
	result, err := callWithRetry(func() (interface{}, error) {
		return client.API().ChannelsGetChannels(&telegram.ChannelsGetChannels{
			ID: []telegram.InputChannel{
				&telegram.InputChannel{
					ChannelId:  int32(-groupID),
					AccessHash: 0,
				},
			},
		})
	}, 3)

	if err == nil && result != nil {
		channel := result.(*telegram.MessagesChats)
		if len(channel.Chats) > 0 {
			chat := channel.Chats[0]
			if chatChannel, ok := chat.(*telegram.Channel); ok {
				// It's a supergroup/channel
				inputPeer = &telegram.InputPeerChannel{
					ChannelId:  int32(chatChannel.Id),
					AccessHash: chatChannel.AccessHash,
				}
			}
		}
	}

	// Get participants from the group/channel
	var allMembers []GroupMember
	var offset int32 = 0
	const limit int32 = 200 // Telegram's limit for one request

	// Track progress
	totalRetrieved := 0

	for {
		// Get participants in batches
		var participants *telegram.ChannelsChannelParticipants
		var err error

		switch peer := inputPeer.(type) {
		case *telegram.InputPeerChannel:
			req := &telegram.ChannelsGetParticipants{
				Channel: &telegram.InputChannel{
					ChannelId:  peer.ChannelId,
					AccessHash: peer.AccessHash,
				},
				Filter: &telegram.ChannelParticipantsRecent{},
				Offset: offset,
				Limit:  limit,
				Hash:   0,
			}

			// Make the API call with retry logic
			result, err := callWithRetry(func() (interface{}, error) {
				return client.API().ChannelsGetParticipants(req)
			}, 3)
			if err != nil {
				return nil, fmt.Errorf("failed to get channel participants after retries: %v", err)
			}
			participants = result.(*telegram.ChannelsChannelParticipants)
		default:
			return nil, fmt.Errorf("unsupported peer type for getting members")
		}

		// Convert telegram users to GroupMember
		for _, participant := range participants.Participants {
			var user *telegram.User

			// Find the user in the users list
			for _, u := range participants.Users {
				if u.GetId() == participant.GetUserId() {
					user = u
					break
				}
			}

			if user != nil {
				groupMember := GroupMember{
					ID:        int64(user.GetId()),
					Username:  user.GetUsername(),
					FirstName: user.GetFirstName(),
					LastName:  user.GetLastName(),
					IsBot:     user.GetBot(),
					IsScam:    user.GetScam(),
					IsFake:    user.GetFake(),
					PhoneNumber: user.GetPhone(),
				}
				allMembers = append(allMembers, groupMember)
			}
		}

		// Update progress
		totalRetrieved += len(participants.Participants)
		totalParticipants := totalRetrieved // Approximate total since Telegram doesn't give exact count upfront

		// For a more accurate progress, we'll just show current progress based on what we've retrieved
		fmt.Printf("Retrieved %d members so far...\n", totalRetrieved)

		// Check if we've retrieved all members
		if int32(len(participants.Participants)) < limit {
			break
		}

		offset += limit
	}

	fmt.Printf("Total members retrieved: %d\n", len(allMembers))

	return allMembers, nil
}

// callWithRetry executes an API call with retry logic for transient failures
func callWithRetry(apiCall func() (interface{}, error), maxRetries int) (interface{}, error) {
	var result interface{}
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		result, err = apiCall()
		if err == nil {
			return result, nil // Success
		}

		// Check if it's a FloodWait error (rate limiting)
		if isFloodWaitError(err.Error()) {
			waitSeconds := 1 << retry // Exponential backoff: 1, 2, 4 seconds
			if waitSeconds > 30 { // Cap at 30 seconds
				waitSeconds = 30
			}
			fmt.Printf("Rate limited by Telegram. Waiting for %d seconds...\n", waitSeconds)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		// Check if it's a network/timeout error that might be transient
		if isTransientError(err.Error()) {
			waitSeconds := 1 << retry // Exponential backoff: 1, 2, 4 seconds
			if waitSeconds > 10 { // Cap at 10 seconds for transient errors
				waitSeconds = 10
			}
			fmt.Printf("Transient network error occurred. Retrying in %d seconds... (attempt %d/%d)\n",
				waitSeconds, retry+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		// Not a retryable error, return immediately
		return nil, err
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

// isFloodWaitError checks if the error is related to rate limiting/flood wait
func isFloodWaitError(errorMsg string) bool {
	// Telegram flood wait errors typically contain these keywords
	return containsAny(errorMsg, []string{"FLOOD_WAIT", "FloodWait", "too many requests", "rate limit", "slow down"})
}

// isTransientError checks if the error is likely to be transient (network related)
func isTransientError(errorMsg string) bool {
	// Common transient error indicators
	return containsAny(errorMsg, []string{
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"temporary failure",
		"broken pipe",
		"no such host",
		"i/o timeout",
	})
}

// containsAny checks if a string contains any of the provided substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			charS := s[i+j]
			charSub := substr[j]
			// Case insensitive comparison
			if toLower(charS) != toLower(charSub) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a byte to lowercase if it's an uppercase letter
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// FilterMembersByOptions filters members based on export options
func FilterMembersByOptions(members []GroupMember, options ExportOptions) []map[string]interface{} {
	var filteredMembers []map[string]interface{}

	for _, member := range members {
		filteredMember := make(map[string]interface{})

		if options.IncludeID {
			filteredMember["id"] = member.ID
		}
		if options.IncludeUsername {
			filteredMember["username"] = member.Username
		}
		if options.IncludeFirstName {
			filteredMember["first_name"] = member.FirstName
		}
		if options.IncludeLastName {
			filteredMember["last_name"] = member.LastName
		}
		if options.IncludeIsBot {
			filteredMember["is_bot"] = member.IsBot
		}
		if options.IncludeIsScam {
			filteredMember["is_scam"] = member.IsScam
		}
		if options.IncludeIsFake {
			filteredMember["is_fake"] = member.IsFake
		}
		if options.IncludePhoneNumber {
			filteredMember["phone_number"] = member.PhoneNumber
		}

		filteredMembers = append(filteredMembers, filteredMember)
	}

	return filteredMembers
}

// ExportToJSON writes GroupMember slice to JSON file with timestamp format
func ExportToJSON(members []GroupMember, filePath string, options ExportOptions) error {
	// Create the output directory if it doesn't exist
	outputDir := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			outputDir = filePath[:i]
			break
		}
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Filter members based on export options
	filteredMembers := FilterMembersByOptions(members, options)

	// Create the JSON data structure following the specification
	data := map[string]interface{}{
		"members":     filteredMembers,
		"memberCount": len(members), // Keep original count for reference
		"exportTime":  time.Now().Format(time.RFC3339),
	}

	// Marshal the data to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members to JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	return nil
}

// GenerateFilePath generates file path using group name and current date in format 'groupname_members_YYYYMMDD.json'
func GenerateFilePath(groupName string, outputDir string) string {
	// Sanitize group name by removing special characters that might cause issues in filenames
	sanitizedGroupName := sanitizeFileName(groupName)

	// Get current date in YYYYMMDD format
	currentDate := time.Now().Format("20060102")

	// Create filename
	filename := fmt.Sprintf("%s_members_%s.json", sanitizedGroupName, currentDate)

	// Create full path
	return fmt.Sprintf("%s/%s", outputDir, filename)
}

// sanitizeFileName removes or replaces characters that might be problematic in filenames
func sanitizeFileName(name string) string {
	// Replace problematic characters with underscores
	result := name
	invalidChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*", "."}

	for _, char := range invalidChars {
		result = replaceAll(result, char, "_")
	}

	return result
}

// replaceAll replaces all occurrences of old string with new string
func replaceAll(str, old, new string) string {
	result := ""
	i := 0
	for i < len(str) {
		if i <= len(str)-len(old) && str[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(str[i])
			i++
		}
	}
	return result
}

// ValidateGroupAccess checks if the user has access to the specified group
func ValidateGroupAccess(client *gotgproto.Client, groupID int64) error {
	_, err := GetGroupName(client, groupID)
	if err != nil {
		return fmt.Errorf("user does not have access to group with ID %d: %v", groupID, err)
	}
	return nil
}

// GetGroupName retrieves the name of a group or channel by its ID
func GetGroupName(client *gotgproto.Client, groupID int64) (string, error) {
	// Convert group ID to input peer
	var inputPeer telegram.InputPeer

	// Check if it's a supergroup/channel (negative ID)
	if groupID < 0 {
		inputPeer = &telegram.InputPeerChannel{
			ChannelId:  int32(-groupID),
			AccessHash: 0, // Will be retrieved if needed
		}
	} else {
		inputPeer = &telegram.InputPeerChat{
			ChatId: int32(groupID),
		}
	}

	// Get full channel/chat info to retrieve the name
	var err error
	var chats *telegram.MessagesChats
	var isChannel = groupID < 0

	if isChannel {
		channelsReq := &telegram.ChannelsGetChannels{
			ID: []telegram.InputChannel{
				&telegram.InputChannel{
					ChannelId:  int32(-groupID),
					AccessHash: 0,
				},
			},
		}
		resp, err := client.API().ChannelsGetChannels(channelsReq)
		if err != nil {
			return "", fmt.Errorf("failed to get channel info: %v", err)
		}
		chats = resp.(*telegram.MessagesChats)
	} else {
		chatsReq := &telegram.MessagesGetChats{
			ID: []int32{int32(groupID)},
		}
		resp, err := client.API().MessagesGetChats(chatsReq)
		if err != nil {
			return "", fmt.Errorf("failed to get chat info: %v", err)
		}
		chats = resp.(*telegram.MessagesChats)
	}

	if len(chats.Chats) == 0 {
		return "", fmt.Errorf("no chat found with ID: %d", groupID)
	}

	chat := chats.Chats[0]
	switch c := chat.(type) {
	case *telegram.Chat:
		return c.Title, nil
	case *telegram.Channel:
		return c.Title, nil
	default:
		return fmt.Sprintf("group_%d", groupID), nil
	}
}

// ExportWorkflow implements main export workflow connecting configuration → authentication → group members → JSON export
func ExportWorkflow(config *ExportConfiguration) (*ExportResult, error) {
	// Initialize Telegram client
	client, err := InitializeTelegramClient(config)
	if err != nil {
		return &ExportResult{
			GroupID:      config.GroupID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to initialize client: %v", err),
		}, err
	}
	defer client.Stop()

	// Authenticate with Telegram
	if err := AuthenticateTelegram(client, config); err != nil {
		return &ExportResult{
			GroupID:      config.GroupID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("Authentication failed: %v", err),
		}, err
	}

	// Get group name
	fmt.Printf("Getting group name for group ID: %d\n", config.GroupID)
	groupName, err := GetGroupName(client, config.GroupID)
	if err != nil {
		fmt.Printf("Warning: Could not get group name: %v. Using fallback name.\n", err)
		groupName = fmt.Sprintf("group_%d", config.GroupID)
	}
	fmt.Printf("Exporting members from group: %s\n", groupName)

	// Get group members
	fmt.Printf("Starting to retrieve members from group ID: %d\n", config.GroupID)

	members, err := GetGroupMembers(client, config.GroupID)
	if err != nil {
		return &ExportResult{
			GroupID:      config.GroupID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to get group members: %v", err),
		}, err
	}

	// Load export options from environment variables
	options := LoadExportOptions()

	// Generate output file path
	outputFilePath := GenerateFilePath(groupName, config.OutputDir)

	// Export members to JSON
	if err := ExportToJSON(members, outputFilePath, options); err != nil {
		return &ExportResult{
			GroupID:      config.GroupID,
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to export to JSON: %v", err),
		}, err
	}

	// Create successful export result
	result := &ExportResult{
		GroupID:        config.GroupID,
		GroupName:      groupName,
		MemberCount:    len(members),
		OutputFilePath: outputFilePath,
		ExportTime:     time.Now(),
		Success:        true,
	}

	return result, nil
}

// Logger provides a simple logging mechanism
type Logger struct {
	Verbose bool
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	return &Logger{Verbose: verbose}
}

// Log logs a message if verbose mode is enabled
func (l *Logger) Log(message string) {
	if l.Verbose {
		log.Printf("[DEBUG] %s\n", message)
	}
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	log.Printf("[INFO] %s\n", message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	log.Printf("[WARN] %s\n", message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	log.Printf("[ERROR] %s\n", message)
}

// GetAccessibleGroups retrieves all groups accessible to the authenticated user
func GetAccessibleGroups(client *gotgproto.Client) ([]*telegram.Chat, []*telegram.Channel, error) {
	// Get dialogs (chats and channels the user is part of)
	dialogs, err := client.API().MessagesGetDialogs(&telegram.MessagesGetDialogs{
		OffsetDate: 0,
		OffsetID:   0,
		OffsetPeer: &telegram.InputPeerEmpty{},
		Limit:      100, // Limit for the first request
		Hash:       0,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dialogs: %v", err)
	}

	var chats []*telegram.Chat
	var channels []*telegram.Channel

	switch d := dialogs.(type) {
	case *telegram.MessagesDialogs:
		for _, chat := range d.Chats {
			switch c := chat.(type) {
			case *telegram.Chat:
				// Only include chats that are groups (not one-to-one conversations)
				if c.GetParticipantsCount() > 2 {
					chats = append(chats, c)
				}
			case *telegram.Channel:
				// Include channels and supergroups
				if c.GetMegaGroup() || !c.GetBroadcast() { // MegaGroup = supergroup, !Broadcast = not a channel
					channels = append(channels, c)
				}
			}
		}
	}

	return chats, channels, nil
}

// FormatChatName formats chat name for display
func FormatChatName(chat interface{}) string {
	switch c := chat.(type) {
	case *telegram.Chat:
		return fmt.Sprintf("%s (ID: %d)", c.Title, c.Id)
	case *telegram.Channel:
		return fmt.Sprintf("%s (ID: %d)", c.Title, c.Id)
	default:
		return "Unknown Chat"
	}
}

// SelectGroupByUser prompts user to select a group from available groups
func SelectGroupByUser(client *gotgproto.Client) (int64, error) {
	fmt.Println("Fetching accessible groups...")
	chats, channels, err := GetAccessibleGroups(client)
	if err != nil {
		return 0, fmt.Errorf("failed to get accessible groups: %v", err)
	}

	if len(chats) == 0 && len(channels) == 0 {
		return 0, fmt.Errorf("no groups found")
	}

	fmt.Println("\nAvailable groups:")
	groupCounter := 0

	// Print chats
	for i, chat := range chats {
		fmt.Printf("%d. %s\n", groupCounter, FormatChatName(chat))
		groupCounter++
	}

	// Print channels
	for i, channel := range channels {
		fmt.Printf("%d. %s\n", groupCounter, FormatChatName(channel))
		groupCounter++
	}

	// Get user selection
	fmt.Printf("\nEnter the number of the group you want to export from (0-%d): ", groupCounter-1)
	var selection int
	fmt.Scanln(&selection)

	if selection < 0 || selection >= groupCounter {
		return 0, fmt.Errorf("invalid selection")
	}

	// Determine which group was selected
	if selection < len(chats) {
		return int64(chats[selection].Id), nil
	} else {
		channelIndex := selection - len(chats)
		return -int64(channels[channelIndex].Id), nil // Negative for channels/supergroups
	}
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Printf("[ERROR] Error loading configuration: %v\n", err)
		return
	}

	// Create logger instance
	logger := NewLogger(config.Verbose)
	logger.Info("Starting Telegram group members export application")

	// Validate configuration (skip GroupID validation if we're going to let user select it)
	// We'll validate other config values but allow GroupID to be 0 for now
	if config.APIID == 0 || config.APIHash == "" || config.PhoneNumber == "" {
		log.Printf("[ERROR] Configuration validation error: API_ID, API_HASH, and PHONE are required\n")
		return
	}

	logger.Info("Initializing Telegram client")
	// Initialize Telegram client
	client, err := InitializeTelegramClient(config)
	if err != nil {
		log.Printf("[ERROR] Error initializing client: %v\n", err)
		return
	}
	defer client.Stop()

	logger.Info("Starting authentication process")
	// Authenticate with Telegram
	if err := AuthenticateTelegram(client, config); err != nil {
		log.Printf("[ERROR] Authentication failed: %v\n", err)
		return
	}

	var groupID int64
	if config.GroupID == 0 {
		// If no group ID was specified, let user select
		fmt.Println("No group ID specified in config. Available groups will be shown for selection.")
		logger.Info("Fetching accessible groups for user selection")
		selectedGroupID, err := SelectGroupByUser(client)
		if err != nil {
			log.Printf("[ERROR] Error selecting group: %v\n", err)
			return
		}
		groupID = selectedGroupID
	} else {
		groupID = config.GroupID
		logger.Info(fmt.Sprintf("Using group ID from configuration: %d", groupID))
	}

	// Validate that user has access to the selected group
	logger.Info(fmt.Sprintf("Validating access to group ID: %d", groupID))
	if err := ValidateGroupAccess(client, groupID); err != nil {
		log.Printf("[ERROR] Error validating group access: %v\n", err)
		return
	}

	// Update config with selected group ID
	config.GroupID = groupID

	logger.Info("Starting export workflow")

	// Track performance timing
	startTime := time.Now()
	result, err := ExportWorkflow(config)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[ERROR] Export failed: %v\n", err)
	} else {
		if result.Success {
			log.Printf("[INFO] Export completed successfully in %v!", duration)
			fmt.Printf("Group: %s\n", result.GroupName)
			fmt.Printf("Members exported: %d\n", result.MemberCount)
			fmt.Printf("Output file: %s\n", result.OutputFilePath)
			fmt.Printf("Export duration: %v\n", duration)

			// Alert if export took longer than expected
			if duration > 5*time.Minute && result.MemberCount > 10000 {
				logger.Warn(fmt.Sprintf("Export took %v which is longer than expected for %d members", duration, result.MemberCount))
			}
		} else {
			log.Printf("[ERROR] Export failed: %s\n", result.ErrorMessage)
		}
	}

	logger.Info("Application finished")
}