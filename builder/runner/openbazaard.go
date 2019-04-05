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
	binaryPath string
	configPath string
}

func FromBinaryPath(path string) (*OpenBazaarRunner, error) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil, ErrBinaryNotFound
	}
	return &OpenBazaarRunner{binaryPath: path}, nil
}

func (r *OpenBazaarRunner) SetConfigPath(path string) error {
	r.configPath = path
	return nil
}

func (r *OpenBazaarRunner) startCmd(tee io.Writer) *shell.Command {
	if r.configPath != "" {
		return shell.Cmd(r.binaryPath, "start", "-v", "-d", r.configPath).Tee(tee)
	}
	return shell.Cmd(r.binaryPath, "start", "-v").Tee(tee)
}

func (r *OpenBazaarRunner) AsyncStart(tee io.Writer) *shell.Process {
	return r.startCmd(tee).Start()
}

func (r *OpenBazaarRunner) RunStart(tee io.Writer) *shell.Process {
	return r.startCmd(tee).Run()
}

func (r *OpenBazaarRunner) Version() (string, error) {
	var proc = shell.Cmd(r.binaryPath, "-v").Run()
	if proc.ExitStatus != 0 {
		return "", fmt.Errorf("getting version: %s", proc.Error())
	}
	return proc.String(), nil
}
