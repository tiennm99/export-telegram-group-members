package main

import (
	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	comparecmd "github.com/tiennm99/export-telegram-group-members/internal/commands/compare"
)

func main() {
	cli.Main(comparecmd.Run)
}
