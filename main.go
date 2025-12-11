package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
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
	rawApiId, isExist := os.LookupEnv("APP_ID")
	if !isExist {
		log.Fatal("APP_ID not set!")
		return
	}
	appId, err := strconv.Atoi(rawApiId)
	if err != nil {
		log.Fatal(err)
		return
	}
	appHash, isExist := os.LookupEnv("APP_HASH")
	if !isExist {
		log.Fatal("APP_HASH not set")
		return
	}
	groupIdsStr, isExist := os.LookupEnv("GROUP_IDS")
	if !isExist {
		log.Fatal("GROUP_IDS not set")
		return
	}
	parts := strings.Split(groupIdsStr, ",")
	groupIds := make([]int64, 0, len(parts))

	for _, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			log.Fatal(err)
			return
		}
		groupIds = append(groupIds, v)
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
			Session: sessionMaker.SqlSession(sqlite.Open("echobot.sqlite3")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	ctx := client.CreateContext()
	groupId := groupIds[0]
	group, err := ctx.GetChat(groupId)
	group.GetAbout()

	clientDispatcher := client.Dispatcher

	clientDispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, echo), 1)

	client.Idle()
}

func echo(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage
	_, err := ctx.Reply(update, ext.ReplyTextString(msg.Text), nil)
	return err
}
