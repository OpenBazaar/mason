package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenBazaar/mason/util"
	"github.com/jessevdk/go-flags"
	"github.com/otiai10/copy"
	shell "github.com/placer14/go-shell"
)

var (
	ErrBinaryNotFound                          = errors.New("binary not found")
	ErrCannotStartStateTransactionWhileRunning = errors.New("state transaction cannot begin when running")
	ErrInitNodeBeforeConfigValueSet            = errors.New("node must be initialized before setting config values")
)

// OpenBazaarRunner is reponsible for the runtime operations of the
// openbazaar-go binary
type (
	runnerState      int
	OpenBazaarRunner struct {
		additionalArgs []string
		binaryPath     string
		proc           *shell.Process
		state          runnerState
		tee            io.WriteCloser

		enableTestnet bool
		dataPath      string
		txDataPath    string
	}
)

const (
	stateReady runnerState = iota
	stateInitialized
	stateRunning
)

// FromBinaryPath will return an OpenBazaarRunner which uses the binary
// located at the path provided.
func FromBinaryPath(path string) (*OpenBazaarRunner, error) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil, ErrBinaryNotFound
	}
	return &OpenBazaarRunner{binaryPath: path}, nil
}

// SetConfigValue will follow a dot-separated string pointing to the nested
// config key to change, and change it to the provided string value
func (r *OpenBazaarRunner) SetConfigValue(path string, value interface{}) error {
	if r.state < stateInitialized {
		return ErrInitNodeBeforeConfigValueSet
	}
	configPath := filepath.Join(r.dataPath, "config")
	fi, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("could not find config at (%s)", configPath)
	} else if err != nil {
		return fmt.Errorf("stat config at (%s): %s", configPath, err.Error())
	}
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %s", err.Error())
	}
	var config map[string]interface{}
	if err := json.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("unmarshal config: %s", err.Error())
	}

	// set value and write new config
	if err := walkAndSetJSON(strings.Split(path, "."), value, config); err != nil {
		return fmt.Errorf("setting json value: %s", err.Error())
	}

	newBytes, err := json.Marshal(config)
	err = ioutil.WriteFile(configPath, newBytes, fi.Mode())
	if err != nil {
		return fmt.Errorf("writing config changes: %s", err.Error())
	}
	return nil
}

func walkAndSetJSON(path []string, value interface{}, jsonMap map[string]interface{}) error {
	if len(path) > 1 {
		jsonMsgRemain, ok := jsonMap[path[0]]
		if !ok {
			return fmt.Errorf("path segment (%s) not found", path[0])
		}
		newJSONMap, ok := jsonMsgRemain.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to cast json fragment")
		}
		return walkAndSetJSON(path[1:], value, newJSONMap)
	}
	jsonMap[path[0]] = value
	return nil
}

// SetCustomDataPath will ensure the running binary starts using the state
// data found at the path provided.
func (r *OpenBazaarRunner) SetCustomDataPath(path string) error {
	if _, err := os.Stat(path); err == nil {
		r.state = stateInitialized
	}
	r.dataPath = path
	return nil
}

// SetTestnetMode will ensure the running binary starts using the testnet
// flag
func (r *OpenBazaarRunner) SetTestnetMode(enabled bool) error {
	r.enableTestnet = enabled
	return nil
}

func (r *OpenBazaarRunner) BeginNodeStateTransaction() error {
	switch r.state {
	case stateRunning:
		return ErrCannotStartStateTransactionWhileRunning
	}
	var tempStatePath = util.GenerateTempPath("openbazaard_state")
	if err := copy.Copy(r.dataPath, tempStatePath); err != nil {
		return fmt.Errorf("copying node state: %s", err.Error())
	}
	r.txDataPath = tempStatePath

	return nil
}

//func (r *OpenBazaarRunner) RollbackNodeStateTransaction() error {
//}

// WithArgs adds additional arguments for the running binary to recieve
func (r *OpenBazaarRunner) WithArgs(args []string) *OpenBazaarRunner {
	if args == nil {
		return r
	}
	// copy all args, then remove and process important ones
	r.additionalArgs = r.filterAndApplyArgs(append([]string{}, args...))
	return r
}

func (r *OpenBazaarRunner) filterAndApplyArgs(args []string) []string {
	var startOpts struct {
		DataPath      string `short:"d"`
		TestnetEnable bool   `short:"t" long:"testnet"`
	}

	remainingArgs, _ := flags.ParseArgs(&startOpts, args)

	// handle testnet
	r.SetTestnetMode(startOpts.TestnetEnable)

	// handle custom data path
	if startOpts.DataPath != "" {
		r.SetCustomDataPath(startOpts.DataPath)
	}

	return remainingArgs
}

func (r *OpenBazaarRunner) additionalArgsString() string {
	return strings.Join(r.additionalArgs, " ")
}

// SplitOutput returns a io.ReadCloser which has the stdout and stderr
// streams being sent to its in-memory pipe buffer immediately after
// being started.
func (r *OpenBazaarRunner) SplitOutput() io.ReadCloser {
	pr, pw := io.Pipe()
	r.tee = pw
	return pr
}

// Cleanup ensures all resources which require cleaning are given an
// opportunity. It is the responsibility of the consumer to ensure
// Cleanup is called when the runner is no longer used.
func (r *OpenBazaarRunner) Cleanup() error {
	var pErr, tErr error
	if r.proc != nil {
		pErr = r.Kill()
		defer func() { r.proc = nil }()
	}
	if r.tee != nil {
		tErr = r.tee.Close()
		r.tee = nil
	}
	if pErr != nil {
		return fmt.Errorf("proc cleanup: %s (%s)", pErr.Error(), r.proc.Error())
	}
	if tErr != nil {
		return fmt.Errorf("tee cleanup: %s", tErr.Error())
	}
	return nil
}

func (r *OpenBazaarRunner) startCmd() *shell.Command {
	var cmd = []interface{}{r.binaryPath, "start", "-v"}
	if r.dataPath != "" {
		cmd = append(cmd, "-d", r.dataPath)
	}
	if r.enableTestnet {
		cmd = append(cmd, "-t")
	}
	for _, a := range r.additionalArgs {
		cmd = append(cmd, a)
	}
	return shell.Cmd(cmd...).Tee(r.tee)
}

func (r *OpenBazaarRunner) initCmd() *shell.Command {
	var cmd = []interface{}{r.binaryPath, "init", "-v"}
	if r.dataPath != "" {
		cmd = append(cmd, "-d", r.dataPath)
	}
	if r.enableTestnet {
		cmd = append(cmd, "-t")
	}
	return shell.Cmd(cmd...).Tee(r.tee)
}

// Init will synchronously initialize the node
func (r *OpenBazaarRunner) Init() *OpenBazaarRunner {
	if r.state >= stateInitialized {
		return r
	}
	r.proc = r.initCmd().Run()
	if r.proc.ExitStatus == 0 {
		r.state = stateInitialized
	}
	return r
}

// AsyncStart will return immediately to allow other tasks to continue while
// running.
func (r *OpenBazaarRunner) AsyncStart() *OpenBazaarRunner {
	r.proc = r.startCmd().Start()
	r.state = stateRunning
	return r
}

// RunStart will run synchronously and will return when the process finishes
// running.
func (r *OpenBazaarRunner) RunStart() *OpenBazaarRunner {
	r.proc = r.startCmd().Run()
	return nil
}

// Kill will ensure the binary process is stopped
func (r *OpenBazaarRunner) Kill() error {
	r.state = stateReady
	return r.proc.Kill()
}

// Version returns the version of the running binary
func (r *OpenBazaarRunner) Version() (string, error) {
	var proc = shell.Cmd(r.binaryPath, "-v").Run()
	if proc.ExitStatus != 0 {
		return "", fmt.Errorf("getting version: %s", proc.Error())
	}
	return proc.String(), nil
}

// ExitCodeAndErr returns the exit code and error state of the executed binary
func (r *OpenBazaarRunner) ExitCodeAndErr() (int, error) {
	if r.proc == nil {
		return -65535, nil
	}
	return r.proc.ExitStatus, r.proc.Error()
}
