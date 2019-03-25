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

type OpenBazaarSource struct {
	workingDir          string
	checkedoutReference string
}

// InflateOpenBazaarDaemon creates a copy of the openbazaard source
// at the specified targetDirectory. Default version is `master` and can
// be set with *OpenBazaarSource.CheckoutVersion.
func InflateOpenBazaarDaemon(targetDirectory string) (*OpenBazaarSource, error) {
	var source = &OpenBazaarSource{
		workingDir:          targetDirectory,
		checkedoutReference: "master",
	}
	if err := source.inflate(); err != nil {
		return nil, err
	}
	return source, nil
}

func (s *OpenBazaarSource) inflate() error {
	if _, err := os.Stat(s.packagePath()); err != nil && os.IsNotExist(err) {
		log.Infof("inflating openbazaard source")
		if mkerr := os.MkdirAll(s.packagePath(), os.ModePerm); mkerr != nil {
			return fmt.Errorf("making source path: %s", mkerr.Error())
		}
		proc := shell.Cmd(openbazaardSource()).SetWorkDir(s.packagePath()).Run()
		if proc.ExitStatus != 0 {
			return fmt.Errorf("cloning source: %s", proc.Error())
		}
	} else {
		log.Warningf("inflating openbazaard source skipped, source found at %s", s.packagePath())
	}
	return nil
}

// WorkDir is the root of a GOPATH which contains the checked-out source
func (s *OpenBazaarSource) WorkDir() string { return s.workingDir }

func (s *OpenBazaarSource) packagePath() string {
	return filepath.Join(s.workingDir, "src", "github.com", "OpenBazaar", "openbazaar-go")
}

// CheckoutVersion sets the source state to match the files which were
// checked-in at the git commit `ref`
func (s *OpenBazaarSource) CheckoutVersion(ref string) error {
	log.Infof("checkout openbazaard version %s", ref)
	var proc = shell.Cmd("git checkout", ref).SetWorkDir(s.packagePath()).Run()
	if proc.ExitStatus != 0 {
		return fmt.Errorf("failed checkout version (%s): %s", ref, proc.Error())
	}
	s.checkedoutReference = ref
	return nil
}

func openbazaardSource() string {
	var source = openbazaardDefaultSource
	if altSource := os.Getenv("OPENBAZAARD_SOURCE"); altSource != "" {
		log.Infof("using alternative OPENBAZAARD_SOURCE (%s)", altSource)
		altPath, err := filepath.Abs(altSource)
		if err != nil {
			log.Warningf("can't find absolute path for OPENBAZAARD_SOURCE (%s)", altSource)
			altPath = source
		}
		source = altPath
	}
	return fmt.Sprintf("git clone %s .", source)
}

func (s *OpenBazaarSource) BinaryFilename() string {
	return fmt.Sprintf("openbazaard_%s", s.checkedoutReference)
}
