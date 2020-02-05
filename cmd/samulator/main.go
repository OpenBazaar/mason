package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/OpenBazaar/mason/builder"
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
	"github.com/placer14/go-shell"
)

type opts struct {
	BuyerConfigPath  string `short:"b" long:"buyer" description:"path to buyer configuration"`
	BuyerVersion     string `long:"buyer-version" description:"SHA to use for buyer node"`
	VendorConfigPath string `short:"v" long:"vendor" description:"path to vendor configuration"`
	VendorVersion    string `long:"vendor-version" description:"SHA to use for vendor node"`
	ModConfigPath    string `short:"m" long:"mod" description:"path to mod configuration"`
	ModVersion       string `long:"moderator-version" description:"SHA to use for moderator node"`
}

func (o opts) empty() bool {
	return o.BuyerConfigPath == "" && o.VendorConfigPath == "" && o.ModConfigPath == ""
}

var (
	wg         sync.WaitGroup
	closeMutex sync.RWMutex
	closeFns   = make([]func(), 0)
	log        = logging.MustGetLogger("samulator")
)

func getStdoutBackend() logging.Backend {
	var (
		backend          = logging.NewLogBackend(os.Stdout, "", 0)
		formatter        = logging.MustStringFormatter(`%{color:reset}%{id:03x} %{module} ▶ %{message}`)
		backendFormatted = logging.NewBackendFormatter(backend, formatter)
	)
	return backendFormatted
}

func main() {
	shell.Panic = true

	var (
		options opts

		parser         = flags.NewParser(&options, flags.Default)
		heardInterrupt = make(chan os.Signal)
	)
	signal.Notify(heardInterrupt, syscall.SIGTERM)
	signal.Notify(heardInterrupt, syscall.SIGINT)
	go func() {
		<-heardInterrupt
		log.Infof("interrupted, killing nodes...")
		closeNodes()
	}()
	logging.SetBackend(getStdoutBackend())

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if options.empty() {
		log.Errorf("no config paths provided, exiting")
		os.Exit(3)
	}

	if options.BuyerConfigPath != "" {
		wg.Add(1)
		if options.BuyerVersion == "" {
			options.BuyerVersion = "ethereum-master"
		}
		err := runNode("buyer", options.BuyerVersion, options.BuyerConfigPath)
		if err != nil {
			fmt.Println(err)
		}
		if err != nil {
			log.Errorf("running node: %s", err.Error())
			os.Exit(2)
		}
	}

	if options.VendorConfigPath != "" {
		wg.Add(1)
		if options.VendorVersion == "" {
			options.VendorVersion = "ethereum-master"
		}
		err := runNode("vendor", options.VendorVersion, options.VendorConfigPath)
		if err != nil {
			fmt.Println(err)
		}
		if err != nil {
			log.Errorf("running node: %s", err.Error())
			os.Exit(2)
		}
	}

	if options.ModConfigPath != "" {
		wg.Add(1)
		if options.ModVersion == "" {
			options.ModVersion = "ethereum-master"
		}
		err := runNode("moderator", options.ModVersion, options.ModConfigPath)
		if err != nil {
			fmt.Println(err)
		}
		if err != nil {
			log.Errorf("running node: %s", err.Error())
			os.Exit(2)
		}
	}
	wg.Wait()
}

func runNode(label, version, configPath string) error {
	var ob, err = builder.NewOpenBazaarDaemon(label, version).Build()
	if err != nil {
		return fmt.Errorf("building: %s", err.Error())
	}

	ob.SetCustomDataPath(configPath)

	closeMutex.Lock()
	defer closeMutex.Unlock()

	pr := ob.SplitOutput()
	go logNodeOutput(pr, label)

	ob.AsyncStart()
	close := func() {
		if err := ob.Cleanup(); err != nil {
			log.Errorf("cleanup process: %s", err.Error())
		}
		time.Sleep(1 * time.Second)
		wg.Done()
	}
	closeFns = append(closeFns, close)
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

func closeNodes() {
	closeMutex.RLock()
	defer closeMutex.RUnlock()
	for _, c := range closeFns {
		go c()
	}
}
