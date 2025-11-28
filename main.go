package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Config holds the application configuration
type Config struct {
	Phone      string
	APIId      int
	APIHash    string
	GroupNames []string
}

// Member represents a Telegram group member
type Member struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	config := loadConfig()

	ctx := context.Background()

	// Setup client options
	opts := telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	}

	client := telegram.NewClient(config.APIId, config.APIHash, opts)

	return client.Run(ctx, func(ctx context.Context) error {
		// Authenticate
		if err := authenticate(ctx, client, config.Phone); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Export group members
		if err := exportGroupMembers(ctx, client, config.GroupNames); err != nil {
			return fmt.Errorf("failed to export members: %w", err)
		}

		fmt.Println("\nExport completed successfully!")
		return nil
	})
}

// loadConfig loads configuration - modify these values
func loadConfig() *Config {
	return &Config{
		// TODO: Set your Telegram credentials here
		Phone:   "+1234567890",           // Your phone number
		APIId:   123456789,               // Your API ID from https://my.telegram.org
		APIHash: "your_api_hash_here",    // Your API Hash from https://my.telegram.org

		// Define groups to export
		GroupNames: []string{
			"ZingPlay Game Studios",
			"ZPS HCM",
			"ZPS HCM - Xin nghỉ (phép/đi trễ)",
		},
	}
}

// authenticate handles the Telegram authentication flow
func authenticate(ctx context.Context, client *telegram.Client, phone string) error {
	api := client.API()

	flow := auth.NewFlow(
		&terminalAuth{phone: phone},
		auth.SendCodeOptions{},
	)

	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		return err
	}

	// Get current user to confirm authentication
	user, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	if len(user) > 0 {
		if u, ok := user[0].(*tg.User); ok {
			fmt.Printf("Authenticated as: %s %s\n", u.FirstName, u.LastName)
		}
	}

	return nil
}

// terminalAuth implements auth flow for terminal input
type terminalAuth struct {
	phone string
}

func (a *terminalAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a *terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	password, err := readLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

func (a *terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	code, err := readLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (a *terminalAuth) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (a *terminalAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("signup not supported")
}

// readLine reads a line from stdin
func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// exportGroupMembers exports members from specified groups
func exportGroupMembers(ctx context.Context, client *telegram.Client, groupNames []string) error {
	api := client.API()

	// Create timestamped directory for exports
	timestamp := time.Now().Format("2006-01-02 15-04-05")
	exportDir := timestamp
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	fmt.Printf("\nExporting to directory: %s\n\n", exportDir)

	// Get all dialogs
	dialogs, err := getDialogs(ctx, api)
	if err != nil {
		return fmt.Errorf("failed to get dialogs: %w", err)
	}

	// Filter and process groups
	groupsExported := 0
	for _, dialog := range dialogs {
		// Check if it's a channel/group
		peer, ok := dialog.Peer.(*tg.PeerChannel)
		if !ok {
			continue
		}

		// Get channel info
		var channelName string
		for _, chat := range dialog.Chats {
			if channel, ok := chat.(*tg.Channel); ok && channel.ID == peer.ChannelID {
				channelName = channel.Title
				break
			}
		}

		// Check if this group should be exported
		if !contains(groupNames, channelName) {
			continue
		}

		fmt.Printf("Exporting group: %s\n", channelName)

		// Get members
		members, err := getChannelMembers(ctx, api, peer.ChannelID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get members for %s: %v\n", channelName, err)
			continue
		}

		// Sort members by ID
		sort.Slice(members, func(i, j int) bool {
			return members[i].ID < members[j].ID
		})

		// Export to CSV
		filename := sanitizeFilename(channelName) + ".csv"
		filepath := filepath.Join(exportDir, filename)
		if err := exportToCSV(filepath, members); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export %s: %v\n", channelName, err)
			continue
		}

		fmt.Printf("  -> Exported %d members to %s\n", len(members), filename)
		groupsExported++
	}

	if groupsExported == 0 {
		fmt.Println("No matching groups found to export.")
	}

	return nil
}

// getDialogs retrieves all dialogs (conversations)
func getDialogs(ctx context.Context, api *tg.Client) ([]*tg.Dialog, error) {
	var allDialogs []*tg.Dialog
	var offsetDate int
	var offsetID int
	var offsetPeer tg.InputPeerClass = &tg.InputPeerEmpty{}

	for {
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetDate: offsetDate,
			OffsetID:   offsetID,
			OffsetPeer: offsetPeer,
			Limit:      100,
		})
		if err != nil {
			return nil, err
		}

		var dialogsList []*tg.Dialog
		switch d := dialogs.(type) {
		case *tg.MessagesDialogs:
			dialogsList = d.Dialogs
			allDialogs = append(allDialogs, d.Dialogs...)
			// Last page reached
			return allDialogs, nil
		case *tg.MessagesDialogsSlice:
			dialogsList = d.Dialogs
			allDialogs = append(allDialogs, d.Dialogs...)
			if len(d.Dialogs) == 0 {
				return allDialogs, nil
			}
		default:
			return allDialogs, nil
		}

		// Update offset for next iteration
		if len(dialogsList) > 0 {
			lastDialog := dialogsList[len(dialogsList)-1]
			offsetDate = lastDialog.TopMessage
			offsetID = lastDialog.TopMessage
			offsetPeer = lastDialog.Peer
		} else {
			break
		}
	}

	return allDialogs, nil
}

// getChannelMembers retrieves all members from a channel
func getChannelMembers(ctx context.Context, api *tg.Client, channelID int64) ([]Member, error) {
	// Get channel input
	channel := &tg.InputChannel{
		ChannelID: channelID,
	}

	var members []Member
	var offset int

	for {
		participants, err := api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: channel,
			Filter:  &tg.ChannelParticipantsRecent{},
			Offset:  offset,
			Limit:   200,
		})
		if err != nil {
			return nil, err
		}

		channelParticipants, ok := participants.(*tg.ChannelsChannelParticipants)
		if !ok {
			break
		}

		if len(channelParticipants.Users) == 0 {
			break
		}

		// Extract member data
		for _, u := range channelParticipants.Users {
			if user, ok := u.(*tg.User); ok {
				member := Member{
					ID:        user.ID,
					FirstName: user.FirstName,
					LastName:  user.LastName,
				}
				if user.Username != "" {
					member.Username = user.Username
				}
				members = append(members, member)
			}
		}

		offset += len(channelParticipants.Users)

		// Check if we've reached the end
		if len(channelParticipants.Users) < 200 {
			break
		}
	}

	return members, nil
}

// exportToCSV exports members to a CSV file
func exportToCSV(filename string, members []Member) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "username", "first_name", "last_name"}); err != nil {
		return err
	}

	// Write member data
	for _, member := range members {
		record := []string{
			strconv.FormatInt(member.ID, 10),
			member.Username,
			member.FirstName,
			member.LastName,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Remove characters that are invalid in Windows filenames
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	return invalidChars.ReplaceAllString(name, "")
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
