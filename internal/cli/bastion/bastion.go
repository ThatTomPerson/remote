package bastion

import (
	"github.com/ThatTomPerson/remote/internal/cli/root"
	"github.com/tj/kingpin"
)

func init() {
	cmd := root.Command("bastion", "Run a command in a new container.")

	cmd.Action(func(_ *kingpin.ParseContext) error {
		return nil
	})
}
