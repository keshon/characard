package about

import (
	"context"

	"github.com/keshon/characard/internal/menu"

	"github.com/keshon/commandkit"
)

func init() {
	commandkit.DefaultRegistry.Register(&aboutCommand{})
}

type aboutCommand struct{}

func (c *aboutCommand) Name() string        { return "about" }
func (c *aboutCommand) Description() string { return "Show credits" }

func (c *aboutCommand) Run(ctx context.Context, inv *commandkit.Invocation) error {
	return menu.RunAbout()
}
