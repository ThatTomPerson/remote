package root

import (
	"runtime"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/tj/kingpin"
)

// Cmd is the root command.
var Cmd = kingpin.New("remote", "")

// Command registers a command.
var Command = Cmd.Command

func init() {
	log.SetHandler(cli.Default)
	verbose := Cmd.Flag("verbose", "Enable verbose log output.").Short('v').Bool()

	Cmd.PreAction(func(ctx *kingpin.ParseContext) error {

		if *verbose {
			log.SetLevel(log.DebugLevel)
			log.Debugf("remote version %s (os: %s, arch: %s)", Cmd.GetVersion(), runtime.GOOS, runtime.GOARCH)
		}

		return nil
	})
}
