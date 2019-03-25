package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/OpenBazaar/samulator/builder/blueprints"
	"github.com/OpenBazaar/samulator/builder/runner"
	"github.com/op/go-logging"
	shell "github.com/placer14/go-shell"
)

const GO_BUILD_VERION = "1.10"

var log = logging.MustGetLogger("builder")

type openBazaarBuilder struct {
	friendlyLabel    string
	versionReference string
	workDir          string
	targetOS         string
	targetArch       string
}

func NewOpenBazaarDaemon(label, version string) *openBazaarBuilder {
	return &openBazaarBuilder{
		friendlyLabel:    label,
		versionReference: version,
	}
}

func (b *openBazaarBuilder) Build() (*runner.OpenBazaarRunner, error) {
	b.workDir = generateTempPath(b.friendlyLabel)
	log.Infof("building at %s", b.workDir)

	src, err := blueprints.InflateOpenBazaarDaemon(b.workDir)
	if err != nil {
		return nil, fmt.Errorf("inflating source: %s", err.Error())
	}

	if err := src.CheckoutVersion(b.versionReference); err != nil {
		return nil, fmt.Errorf("checkout version: %s", err.Error())
	}

	if err := generateOSSpecificBuild(src); err != nil {
		return nil, fmt.Errorf("building for %s: %s", runtime.GOOS, err.Error())
	}
	return runner.FromBinaryPath(b.binaryPath())
}

func generateOSSpecificBuild(src *OpenBazaarSource) error {
	var (
		getXGo      = shell.Cmd("go", "get", "github.com/karalabe/xgo")
		buildBinary = shell.Cmd(
			fmt.Sprintf("GOPATH=%s", src.WorkDir()),
			"xgo", "-xv", "-targets", getXGoBuildTarget(), // build arch/OS targets
			"-dest ./dest",               // build destination path
			"-out", src.BinaryFilename(), // binary name prefix
			"-go", GO_BUILD_VERION, // specific go build version
			"github.com/OpenBazaar/openbazaar-go",
		)
		buildCommands = []*shell.Command{getXGo, buildBinary}
	)
	for _, cmd := range buildCommands {
		var proc = cmd.SetWorkDir(src.WorkDir()).Run()
		if proc.ExitStatus != 0 {
			return fmt.Errorf("non-zero build exit: %s", proc.Error())
		}
	}
	return nil
}

func (b *openBazaarBuilder) binaryPath() string {
	return filepath.Join(b.workDir, "dest", b.binaryFilename())
}

func (b *openBazaarBuilder) binaryFilename() string {
	return fmt.Sprintf("%s-%s-10.6-%s", b.xgoOutname(), b.targetOS, b.targetArch)
}

func (b *openBazaarBuilder) MustClean() {
	if err := os.RemoveAll(b.workDir); err != nil {
		log.Errorf("cleaning (%s): %s", b.workDir, err.Error())
		panic(err.Error())
	}
}
