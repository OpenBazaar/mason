package blueprints

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
	shell "github.com/placer14/go-shell"
)

func init() {
	shell.Panic = false
}

const openbazaardDefaultSource = "https://github.com/OpenBazaar/openbazaar-go"

var log = logging.MustGetLogger("blueprints")

type openBazaarSource struct {
	workingDir string
}

func InflateOpenBazaarDaemon(targetDirectory string) (*openBazaarSource, error) {
	if _, err := os.Stat(targetDirectory); err != nil && os.IsNotExist(err) {
		log.Infof("inflating openbazaard source at %s", targetDirectory)
		if mkerr := os.MkdirAll(targetDirectory, os.ModePerm); mkerr != nil {
			return nil, fmt.Errorf("making source path: %s", mkerr.Error())
		}
		proc := shell.Cmd(openbazaardSource()).SetWorkDir(targetDirectory).Run()
		if proc.ExitStatus != 0 {
			return nil, fmt.Errorf("cloning source: %s", proc.Error())
		}
	} else {
		log.Warningf("inflating openbazaard source skipped, source found at %s", targetDirectory)
	}
	return &openBazaarSource{workingDir: targetDirectory}, nil
}

func (s *openBazaarSource) CheckoutVersion(ref string) error {
	var proc = shell.Cmd("git checkout", ref).SetWorkDir(s.workingDir).Run()
	if proc.ExitStatus != 0 {
		return fmt.Errorf("failed checkout version (%s): %s", ref, proc.Error())
	}
	return nil
}

func openbazaardSource() string {
	var source = openbazaardDefaultSource
	if altSource := os.Getenv("OPENBAZAARD_SOURCE"); altSource != "" {
		log.Infof("using alternative OPENBAZAARD_SOURCE (%s)", altSource)
		source = altSource
	}
	return fmt.Sprintf("git clone %s .", source)
}
