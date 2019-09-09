package main

import (
	"os"

	"github.com/OpenBazaar/mason/cmd/obr/subcommands"
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
)

type opts struct {
	*subcommands.PrepareCommand `command:"prepare" alias:"p" description:"Prepare a cached version of openbazaar-go" long-description:"Build and cache the specified version of openbazaar-go on the local machine. This ensures future executions do not require building this version again."`
	*subcommands.StartCommand   `command:"start" alias:"s" description:"Start a version of openbazaar-go" long-description:"Start a version of openbazaar-go which has been cached, or attempt to build it and then start it."`
}

func getStdoutBackend() logging.Backend {
	var (
		backend          = logging.NewLogBackend(os.Stdout, "", 0)
		formatter        = logging.MustStringFormatter(`%{color:reset}%{id:03x} %{module} ▶ %{message}`)
		backendFormatted = logging.NewBackendFormatter(backend, formatter)
	)
	return backendFormatted
}

func main() {
	logging.SetBackend(getStdoutBackend())

	var (
		options opts
		parser  = flags.NewParser(&options, flags.Default|flags.IgnoreUnknown)
	)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	return
}
