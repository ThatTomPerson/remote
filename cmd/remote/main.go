package main

import (
	"os"

	_ "github.com/ThatTomPerson/remote/internal/cli/bastion"
	_ "github.com/ThatTomPerson/remote/internal/cli/batch"
	_ "github.com/ThatTomPerson/remote/internal/cli/run"

	"github.com/ThatTomPerson/remote/internal/cli/root"
)

var version = "master"

func main() {
	root.Cmd.Version(version)
	_, err := root.Cmd.Parse(os.Args[1:])

	if err == nil {
		return
	}
}
