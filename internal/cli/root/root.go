package root

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

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

	c := make(chan string, 1)

	Cmd.PreAction(func(ctx *kingpin.ParseContext) error {

		if *verbose {
			log.SetLevel(log.DebugLevel)
			log.Debugf("remote version %s (os: %s, arch: %s)", Cmd.GetVersion(), runtime.GOOS, runtime.GOARCH)
		}

		go func() {
			res, err := http.Get("https://api.github.com/repos/thattomperson/remote/releases/latest")
			if err != nil {
				c <- ""
				return
			}
			defer res.Body.Close()

			b := struct {
				TagName string `json:"tag_name"`
			}{}

			err = json.NewDecoder(res.Body).Decode(&b)
			if err != nil {
				c <- ""
				return
			}

			c <- b.TagName
		}()

		return nil
	})

	Cmd.PostAction(func(ctx *kingpin.ParseContext) error {
		select {
		case v := <-c:
			if v != Cmd.GetVersion() && v != "" {
				log.Infof("remote %s is out of date", Cmd.GetVersion())
				log.Infof("go to https://github.com/thattomperson/remote/releases/tag/%s to update to %s", v, v)
			}
		case <-time.After(2 * time.Second):
		}

		return nil
	})
}
