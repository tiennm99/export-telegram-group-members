package main

import (
	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	comparegroupscmd "github.com/tiennm99/export-telegram-group-members/internal/commands/comparegroups"
)

func main() {
	cli.Main(comparegroupscmd.Run)
}
