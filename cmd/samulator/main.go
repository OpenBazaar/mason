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
	EnableTestnet bool `long:"testnet" short:"t" description:"start with testnet flag"`

	BuyerConfigPath  string `short:"b" long:"buyer" description:"path to buyer configuration"`
	VendorConfigPath string `short:"v" long:"vendor" description:"path to vendor configuration"`
	ModConfigPath    string `short:"m" long:"mod" description:"path to mod configuration"`

	Version       string `long:"version" description:"version of buyer, vendor, or mod if each version is not specified"`
	BuyerVersion  string `long:"bv" description:"set Buyer SHA to build (overrides --version)"`
	VendorVersion string `long:"vv" description:"set Vendor SHA to build (overrides --version)"`
	ModVersion    string `long:"mv" description:"set Buyer SHA to build (overrides --version)"`
}

func (o opts) pathEmpty() bool {
	return o.BuyerConfigPath == "" && o.VendorConfigPath == "" && o.ModConfigPath == ""
}
func (o opts) versionEmpty() bool {
	return o.Version == "" && o.BuyerVersion == "" && o.VendorVersion == "" && o.ModVersion == ""
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

	if options.pathEmpty() {
		log.Errorf("no config paths provided, exiting")
		os.Exit(3)
	}

	if options.versionEmpty() {
		log.Errorf("no build version provided, exiting")
		os.Exit(4)
	}

	if options.BuyerConfigPath != "" {
		wg.Add(1)
		if options.BuyerVersion == "" {
			options.BuyerVersion = options.Version
		}
		err := runNode("buyer", options.BuyerVersion, options.BuyerConfigPath, options.EnableTestnet)
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
			options.VendorVersion = options.Version
		}
		err := runNode("vendor", options.VendorVersion, options.VendorConfigPath, options.EnableTestnet)
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
			options.ModVersion = options.Version
		}
		err := runNode("moderator", options.ModVersion, options.ModConfigPath, options.EnableTestnet)
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

func runNode(label, version, configPath string, enableTestnet bool) error {
	var ob, err = builder.NewOpenBazaarDaemon(label, version).Build()
	if err != nil {
		return fmt.Errorf("building: %s", err.Error())
	}

	ob.SetCustomDataPath(configPath)
	ob.SetTestnetMode(enableTestnet)

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
