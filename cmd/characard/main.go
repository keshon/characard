package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/keshon/buildinfo"
	"github.com/keshon/commandkit"

	"github.com/keshon/characard/internal/cards"
	"github.com/keshon/characard/internal/cli"
	_ "github.com/keshon/characard/internal/commands/about"
	_ "github.com/keshon/characard/internal/commands/help"
	_ "github.com/keshon/characard/internal/commands/settings"
	_ "github.com/keshon/characard/internal/commands/tui"
	"github.com/keshon/characard/internal/config"
	"github.com/keshon/characard/internal/menu"
)

func main() {
	if len(os.Args) < 2 {
		runInteractiveMenu()
		return
	}
	if err := cli.DispatchCLI(commandkit.DefaultRegistry); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runInteractiveMenu() {
	for {
		action, err := menu.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		switch action {
		case menu.ActionQuit:
			return
		case menu.ActionCards:
			dir := "."
			if s, err := config.GetSettings(); err == nil && s.CardsDir != "" {
				dir = s.CardsDir
			}
			if err := cards.Run(context.Background(), dir); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			continue
		case menu.ActionSettings:
			if err := menu.RunSettings(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			continue
		case menu.ActionHelp:
			if err := menu.RunHelp(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			continue
		case menu.ActionAbout:
			if err := menu.RunAbout(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			continue
		}
	}
}
