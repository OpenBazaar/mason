package subcommands

import (
	"fmt"
	"os"

	"github.com/OpenBazaar/samulator/builder"
	"github.com/op/go-logging"
)

type PrepareCommand struct {
	Args struct {
		Version string `description:"specify the git reference to prepare" positional-arg-name:"version"`
	} `positional-args:"yes" required:"yes"`
}

func (p *PrepareCommand) Execute(args []string) error {
	fmt.Println(os.Args)
	var log = logging.MustGetLogger("")

	if p.Args.Version == "" {
		return fmt.Errorf("must specify build version")
	}

	var _, err = builder.NewOpenBazaarDaemon("prepare-build", p.Args.Version).Build()
	if err != nil {
		return fmt.Errorf("building (%s): %s", p.Args.Version, err.Error())
	}

	log.Infof("version %s is prepared", p.Args.Version)
	return nil
}
