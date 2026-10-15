package main

import (
	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	crawlcmd "github.com/tiennm99/export-telegram-group-members/internal/commands/crawl"
)

func main() {
	cli.Main(crawlcmd.Run)
}
