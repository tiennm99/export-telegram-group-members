package main

import (
	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	exportcmd "github.com/tiennm99/export-telegram-group-members/internal/commands/export"
)

func main() {
	cli.Main(exportcmd.Run)
}
