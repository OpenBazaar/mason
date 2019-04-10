package runner

import (
	"errors"
	"fmt"
	"io"
	"os"

	shell "github.com/placer14/go-shell"
)

var ErrBinaryNotFound = errors.New("binary not found")

type OpenBazaarRunner struct {
	tee        io.WriteCloser
	binaryPath string
	configPath string
}

// FromBinaryPath will return an OpenBazaarRunner which uses the binary
// located at the path provided.
func FromBinaryPath(path string) (*OpenBazaarRunner, error) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil, ErrBinaryNotFound
	}
	return &OpenBazaarRunner{binaryPath: path}, nil
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
	if r.tee != nil {
		err := r.tee.Close()
		r.tee = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *OpenBazaarRunner) startCmd() *shell.Command {
	if r.configPath != "" {
		return shell.Cmd(r.binaryPath, "start", "-v", "-d", r.configPath).Tee(r.tee)
	}
	return shell.Cmd(r.binaryPath, "start", "-v").Tee(r.tee)
}

// AsyncStart will return immediately to allow other tasks to continue while
// running.
func (r *OpenBazaarRunner) AsyncStart() *shell.Process {
	return r.startCmd().Start()
}

// RunStart will run synchronously and will return when the process finishes
// running.
func (r *OpenBazaarRunner) RunStart() *shell.Process {
	return r.startCmd().Run()
}

// Version returns the version of the running binary
func (r *OpenBazaarRunner) Version() (string, error) {
	var proc = shell.Cmd(r.binaryPath, "-v").Run()
	if proc.ExitStatus != 0 {
		return "", fmt.Errorf("getting version: %s", proc.Error())
	}
	return proc.String(), nil
}
