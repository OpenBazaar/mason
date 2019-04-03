package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/OpenBazaar/samulator/builder/blueprints"
	"github.com/OpenBazaar/samulator/builder/runner"
	"github.com/op/go-logging"
	shell "github.com/placer14/go-shell"
)

const GO_BUILD_VERSION = "1.11"

var log = logging.MustGetLogger("builder")

type openBazaarBuilder struct {
	cachePath        string
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
		cachePath:        fmt.Sprintf("~/.%s", "samulator"),
	}
}

func (b *openBazaarBuilder) Build() (*runner.OpenBazaarRunner, error) {
	c, err := cacher.Open(b.cachePath)
	if err != nil {
		log.Warningf("failed opening cache (at %s): %s", b.cachePath, err.Error())
	}
	if runnerPath, err := c.Get("openbazaard", b.versionReference); err == nil {
		return runner.FromBinaryPath(runnerPath)
	}

	b.workDir = generateTempPath(b.friendlyLabel)
	log.Infof("building at %s", b.workDir)

	src, err := blueprints.InflateOpenBazaarDaemon(b.workDir)
	if err != nil {
		return nil, fmt.Errorf("inflating source: %s", err.Error())
	}

	if err := src.CheckoutVersion(b.versionReference); err != nil {
		return nil, fmt.Errorf("checkout version: %s", err.Error())
	}

	runner, err := generateOSSpecificBuild(src)
	if err != nil {
		return nil, fmt.Errorf("building for %s: %s", runtime.GOOS, err.Error())
	}
	return runner, nil

	if err := cacher.Cache("openbazaard", b.versionReference, binaryPath(src)); err != nil {
		log.Warningf("failed caching build for %s (%s): %s", "openbazaard", v.versionReference, err.Error())
	}
	return runner.FromBinaryPath(b.binaryPath())
}

func generateOSSpecificBuild(src *blueprints.OpenBazaarSource) (*runner.OpenBazaarRunner, error) {
	var (
		getXGo      = shell.Cmd("go", "get", "github.com/karalabe/xgo")
		buildBinary = shell.Cmd(
			fmt.Sprintf("GOPATH=%s", src.WorkDir()),
			"xgo", "-v", "-targets", getXGoBuildTarget(), // build arch/OS targets
			"-dest=./dest",             // build destination path
			"-out", src.BinaryPrefix(), // binary name prefix
			"-go", GO_BUILD_VERSION, // specific go build version
			filepath.Join(src.WorkDir(), "src", "github.com", "OpenBazaar", "openbazaar-go"),
		)
		buildCommands = []*shell.Command{getXGo, buildBinary}
	)
	for _, cmd := range buildCommands {
		var proc = cmd.SetWorkDir(src.WorkDir()).Run()
		if proc.ExitStatus != 0 {
			return nil, fmt.Errorf("non-zero build exit: %s", proc.Error())
		}
	}
	return runner.FromBinaryPath(binaryPath(src))
}

func binaryPath(src *blueprints.OpenBazaarSource) string {
	var (
		targets        = strings.Split(getXGoBuildTarget(), "/")
		os, arch       = targets[0], targets[1]
		binaryFilename = fmt.Sprintf("%s-%s-10.6-%s", src.BinaryPrefix(), os, arch)
	)
	return filepath.Join(src.WorkDir(), "dest", binaryFilename)
}

func (b *openBazaarBuilder) MustClean() {
	if err := os.RemoveAll(b.workDir); err != nil {
		log.Errorf("cleaning (%s): %s", b.workDir, err.Error())
		panic(err.Error())
	}
}
