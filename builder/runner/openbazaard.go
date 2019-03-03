package runner

import (
	"errors"
	"fmt"
	"os"

	shell "github.com/placer14/go-shell"
)

var ErrBinaryNotFound = errors.New("binary not found")

type OpenBazaarRunner struct {
	binaryPath string
}

func FromBinaryPath(path string) (*OpenBazaarRunner, error) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil, ErrBinaryNotFound
	}
	return &OpenBazaarRunner{binaryPath: path}, nil
}

func (r *OpenBazaarRunner) Version() (string, error) {
	var proc = shell.Cmd(r.binaryPath, "-v").Run()
	if proc.ExitStatus != 0 {
		return "", fmt.Errorf("getting version: %s", proc.Error())
	}
	return proc.String(), nil
}
