package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
	phoneNumber, isExist := os.LookupEnv("TG_PHONE")
	if !isExist {
		log.Fatal("TG_PHONE not set")
	}
	rawAppId, isExist := os.LookupEnv("APP_ID")
	if !isExist {
		log.Fatal("APP_ID not set!")
		return
	}
	appId, err := strconv.Atoi(rawAppId)
	if err != nil {
		log.Fatal(err)
		return
	}
	appHash, isExist := os.LookupEnv("APP_HASH")
	if !isExist {
		log.Fatal("APP_HASH not set")
		return
	}
	rawGroupIds, isExist := os.LookupEnv("GROUP_IDS")
	if !isExist {
		log.Fatal("GROUP_IDS not set")
		return
	}
	groupIds, err := strToInts(rawGroupIds)
	if err != nil {
		log.Fatal(err)
		return
	}

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		appId,
		// Get ApiHash from https://my.telegram.org/apps
		appHash,
		// ClientType, as we defined above
		gotgproto.ClientTypePhone(phoneNumber),
		// Optional parameters of client
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open("session.sqlite3")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	groupId := groupIds[0]
	ctx := client.CreateContext()
	// client.API().ChannelsGetParticipants() // TODO
	group, err := ctx.GetChat(groupId)
	fmt.Printf("About: %s", group.GetAbout())

	client.Idle()
}

func echo(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage
	_, err := ctx.Reply(update, ext.ReplyTextString(msg.Text), nil)
	return err
}

func strToInts(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	ints := make([]int64, 0, len(parts))

	for _, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		ints = append(ints, v)
	}

	return ints, nil
}
