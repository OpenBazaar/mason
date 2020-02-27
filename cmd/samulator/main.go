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
	OverridePostmanQAConfig bool `long:"postman-config" description:"override ports for each node to work with postman QA test suite"`
	EnableTestnet           bool `long:"testnet" short:"t" description:"start with testnet flag"`

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

const (
	buyer     = "buy"
	vendor    = "ven"
	moderator = "mod"
)

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

	var nodeOpts = nodeOptions{
		enableTestnet:         options.EnableTestnet,
		overridePostmanConfig: options.OverridePostmanQAConfig,
	}
	if options.BuyerConfigPath != "" {
		wg.Add(1)
		if options.BuyerVersion == "" {
			options.BuyerVersion = options.Version
		}
		nodeOpts.label = buyer
		nodeOpts.configPath = options.BuyerConfigPath
		nodeOpts.version = options.BuyerVersion
		err := runNode(nodeOpts)
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
		nodeOpts.label = vendor
		nodeOpts.configPath = options.VendorConfigPath
		nodeOpts.version = options.VendorVersion
		err := runNode(nodeOpts)
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
		nodeOpts.label = moderator
		nodeOpts.configPath = options.ModConfigPath
		nodeOpts.version = options.ModVersion
		err := runNode(nodeOpts)
		if err != nil {
			log.Errorf("running node: %s", err.Error())
			os.Exit(2)
		}
	}
	wg.Wait()
}

type nodeOptions struct {
	label      string
	version    string
	configPath string

	enableTestnet         bool
	overridePostmanConfig bool
}

func runNode(opts nodeOptions) error {
	var ob, err = builder.NewOpenBazaarDaemon(opts.label, opts.version).Build()
	if err != nil {
		return fmt.Errorf("building: %s", err.Error())
	}

	closeMutex.Lock()
	defer closeMutex.Unlock()

	pr := ob.SplitOutput()
	go logNodeOutput(pr, opts.label)

	ob.SetCustomDataPath(opts.configPath)
	ob.SetTestnetMode(opts.enableTestnet)

	if opts.overridePostmanConfig {
		ob.Init()
		switch opts.label {
		case buyer:
			err := ob.SetConfigValue("Addresses.Gateway", "/ip4/127.0.0.1/tcp/4002")
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Gateway: %s", err.Error())
			}
			err = ob.SetConfigValue("Addresses.Swarm", []string{
				"/ip4/0.0.0.0/tcp/4001",
				"/ip6/::/tcp/4001",
				"/ip4/0.0.0.0/tcp/9005/ws",
				"/ip6/::/tcp/9005/ws",
			})
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Swarm: %s", err.Error())
			}
		case vendor:
			err := ob.SetConfigValue("Addresses.Gateway", "/ip4/127.0.0.1/tcp/4102")
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Gateway: %s", err.Error())
			}
			err = ob.SetConfigValue("Addresses.Swarm", []string{
				"/ip4/0.0.0.0/tcp/4101",
				"/ip6/::/tcp/4101",
				"/ip4/0.0.0.0/tcp/9105/ws",
				"/ip6/::/tcp/9105/ws",
			})
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Swarm: %s", err.Error())
			}
		case moderator:
			err := ob.SetConfigValue("Addresses.Gateway", "/ip4/127.0.0.1/tcp/4202")
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Gateway: %s", err.Error())
			}
			err = ob.SetConfigValue("Addresses.Swarm", []string{
				"/ip4/0.0.0.0/tcp/4201",
				"/ip6/::/tcp/4201",
				"/ip4/0.0.0.0/tcp/9205/ws",
				"/ip6/::/tcp/9205/ws",
			})
			if err != nil {
				return fmt.Errorf("failed to set buyer Address.Swarm: %s", err.Error())
			}
		}
	}

	ob.AsyncStart()
	closeFn := func() {
		if err := ob.Cleanup(); err != nil {
			log.Errorf("cleanup process: %s", err.Error())
		}
		time.Sleep(1 * time.Second)
		wg.Done()
	}
	closeFns = append(closeFns, closeFn)
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
