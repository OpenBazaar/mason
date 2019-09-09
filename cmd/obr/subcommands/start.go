package subcommands

import (
	"bufio"
	"fmt"
	"io"

	"github.com/OpenBazaar/mason/builder"
	"github.com/op/go-logging"
)

type StartCommand struct {
	Args struct {
		Version     string   `description:"specify the git reference to start" positional-arg-name:"version" required:"true"`
		StartParams []string `description:"provide params to be passed to daemon on start" positional-arg-name:"start params"`
	} `positional-args:"true"`
}

func (c *StartCommand) Execute(args []string) error {
	var log = logging.MustGetLogger("")
	log.Infof("starting %s...", c.Args.Version)
	log.Infof("args: %s", c.Args.StartParams)
	var obBuilder = builder.NewOpenBazaarDaemon(c.Args.Version, c.Args.Version)
	defer obBuilder.MustClean()

	var obProc, err = obBuilder.Build()
	if err != nil {
		return fmt.Errorf("building: %s", err.Error())
	}
	defer obProc.Cleanup()

	obProc.WithArgs(c.Args.StartParams)

	pr := obProc.SplitOutput()
	go logNodeOutput(pr, c.Args.Version)

	obProc.RunStart()
	if exit, err := obProc.ExitCodeAndErr(); err != nil {
		return fmt.Errorf("returned (%d): %s", exit, err.Error())
	}
	return nil
}

func logNodeOutput(r io.ReadCloser, prefix string) {
	defer r.Close()
	var (
		nodeLog = logging.MustGetLogger(prefix)
		scanner = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		nodeLog.Infof(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		nodeLog.Errorf("error reading node output (%s): %s", fmt.Sprintf("%010s", prefix), err.Error())
	}
}
