// Package cli provides a CLI adapter for commandkit: parses os.Args,
// resolves commands (including subcommands via "cmd sub" -> "cmd-sub"),
// parses flags, and invokes the command with Invocation.Data set to CLIData.
package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/keshon/commandkit"
)

// CLIData is stored in Invocation.Data when dispatching from the CLI adapter.
// Commands can type-assert inv.Data.(*CLIData) to access parsed flags and args.
type CLIData struct {
	Args  []string
	Flags *flag.FlagSet
}

// CLIFlagger is an optional interface for commands that accept flags.
// If a command implements it, the adapter creates a FlagSet and parses args before Run.
type CLIFlagger interface {
	commandkit.Command
	Flags(fs *flag.FlagSet)
}

// DispatchCLI parses os.Args, resolves the command (including "auth create" -> "auth-create"),
// parses flags when the command implements CLIFlagger, and runs the command.
func DispatchCLI(registry *commandkit.Registry) error {
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	cmdName := args[0]
	var runArgs []string

	var cmd commandkit.Command
	if len(args) >= 2 {
		subName := args[1]
		cmd = registry.Get(cmdName + "-" + subName)
		if cmd != nil {
			runArgs = args[2:]
		}
	}
	if cmd == nil {
		cmd = registry.Get(cmdName)
		runArgs = args[1:]
	}
	if cmd == nil {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	ctx := context.Background()
	inv := &commandkit.Invocation{Args: runArgs}

	if flagger, ok := cmd.(CLIFlagger); ok {
		fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
		flagger.Flags(fs)
		if err := fs.Parse(runArgs); err != nil {
			return err
		}
		inv.Args = fs.Args()
		inv.Data = &CLIData{Args: fs.Args(), Flags: fs}
	} else {
		inv.Args = runArgs
	}

	return cmd.Run(ctx, inv)
}

// CLIDataFrom returns the CLI payload from inv.Data, or nil if not set.
func CLIDataFrom(inv *commandkit.Invocation) *CLIData {
	if inv == nil || inv.Data == nil {
		return nil
	}
	d, _ := inv.Data.(*CLIData)
	return d
}
