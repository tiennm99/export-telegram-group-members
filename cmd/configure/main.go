package main

import (
	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	configurecmd "github.com/tiennm99/export-telegram-group-members/internal/commands/configure"
)

func main() {
	cli.Main(configurecmd.Run)
}
