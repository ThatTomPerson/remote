package batch

import (
	"github.com/thattomperson/remote/internal/cli/root"
	"github.com/tj/kingpin"
)

func init() {
	cmd := root.Command("batch", "Run a command in a new container.")

	cmd.Action(func(_ *kingpin.ParseContext) error {
		return nil
	})
}
