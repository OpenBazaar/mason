package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	shell "github.com/placer14/go-shell"
)

var ErrBinaryNotFound = errors.New("binary not found")

// OpenBazaarRunner is reponsible for the runtime operations of the
// openbazaar-go binary
type OpenBazaarRunner struct {
	proc           *shell.Process
	tee            io.WriteCloser
	binaryPath     string
	configPath     string
	additionalArgs []string
}

// FromBinaryPath will return an OpenBazaarRunner which uses the binary
// located at the path provided.
func FromBinaryPath(path string) (*OpenBazaarRunner, error) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil, ErrBinaryNotFound
	}
	return &OpenBazaarRunner{binaryPath: path}, nil
}

// WithArgs adds additional arguments for the running binary to recieve
func (r *OpenBazaarRunner) WithArgs(args []string) {
	if args == nil {
		return
	}
	// copy all args, then remove and process important ones
	var additionalArgs = append([]string(nil), args...)
	r.additionalArgs = r.filterAndApplyArgs(additionalArgs)
}

func (r *OpenBazaarRunner) filterAndApplyArgs(args []string) []string {
	for i, arg := range args {
		if arg == "-d" {
			if i == len(args)-1 {
				continue
			}
			r.SetConfigPath(args[i+1])
			return r.filterAndApplyArgs(append(args[:i], args[i+2:]...))
		}
	}
	return args
}

// SetConfigPath will ensure the running binary starts using the config
// data found at the path provided.
func (r *OpenBazaarRunner) SetConfigPath(path string) error {
	r.configPath = path
	return nil
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
		pErr = r.proc.Kill()
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
	if r.configPath != "" {
		return shell.Cmd(r.binaryPath, "start", "-v", "-d", r.configPath, r.additionalArgsString()).Tee(r.tee)
	}
	return shell.Cmd(r.binaryPath, "start", "-v", r.additionalArgsString()).Tee(r.tee)
}

func (r *OpenBazaarRunner) additionalArgsString() string {
	return strings.Join(r.additionalArgs, " ")
}

// AsyncStart will return immediately to allow other tasks to continue while
// running.
func (r *OpenBazaarRunner) AsyncStart() *OpenBazaarRunner {
	r.proc = r.startCmd().Start()
	return r
}

// RunStart will run synchronously and will return when the process finishes
// running.
func (r *OpenBazaarRunner) RunStart() *OpenBazaarRunner {
	r.proc = r.startCmd().Run()
	return r
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
