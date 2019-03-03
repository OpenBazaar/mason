package blueprints

import (
	"fmt"
	"os"
	"path/filepath"

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
	var source = &openBazaarSource{workingDir: targetDirectory}
	if err := source.inflate(); err != nil {
		return nil, err
	}
	return source, nil
}

func (s *openBazaarSource) inflate() error {
	var packagePath = s.PackagePath()
	if _, err := os.Stat(packagePath); err != nil && os.IsNotExist(err) {
		log.Infof("inflating openbazaard source")
		if mkerr := os.MkdirAll(packagePath, os.ModePerm); mkerr != nil {
			return fmt.Errorf("making source path: %s", mkerr.Error())
		}
		proc := shell.Cmd(openbazaardSource()).SetWorkDir(packagePath).Run()
		if proc.ExitStatus != 0 {
			return fmt.Errorf("cloning source: %s", proc.Error())
		}
	} else {
		log.Warningf("inflating openbazaard source skipped, source found at %s", packagePath)
	}
	return nil
}

func (s *openBazaarSource) PackagePath() string {
	return filepath.Join(s.workingDir, "src", "github.com", "OpenBazaar", "openbazaar-go")
}

func (s *openBazaarSource) CheckoutVersion(ref string) error {
	log.Infof("checkout openbazaard version %s", ref)
	var proc = shell.Cmd("git checkout", ref).SetWorkDir(s.PackagePath()).Run()
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
