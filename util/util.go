package util

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/op/go-logging"
)

var (
	targetOS, targetArch string

	log              = logging.MustGetLogger("util")
	projectDirectory = ".mason"

	homeDir = os.Getenv("HOME")
)

func workDir() string {
	var workDir = os.Getenv("BUILD_PATH")
	if workDir == "" {
		workDir = os.Getenv("HOME")
	}
	if workDir == "" {
		workAbsPath, err := filepath.Abs(".")
		if err == nil {
			workDir = workAbsPath
		}
	}
	return filepath.Join(workDir, projectDirectory)
}

// GenerateTempPath provides a safe and well-labeled location
// to store temporary files which contains state data
func GenerateTempPath(label string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return filepath.Join(workDir(), "tmp", fmt.Sprintf("mason%s_%d", label, r.Intn(9999)))
}

// GenerateTempBuildPath provides a safe and well-labeled location
// to store temporary files which contains build data
func GenerateTempBuildPath(label string) string {
	return GenerateTempPath(fmt.Sprintf("build_%s", label))
}

// GetXGoBuildTarget returns the appropriate os/arch for the current system's use
func GetXGoBuildTarget() string {
	if targetOS == "" {
		switch runtime.GOOS {
		case "darwin", "linux", "windows":
			targetOS = runtime.GOOS
		default:
			log.Errorf("unsupported OS")
		}
	}

	if targetArch == "" {
		switch runtime.GOARCH {
		case "386", "amd64":
			targetArch = runtime.GOARCH
		default:
			log.Errorf("unsupported architecture")
		}
	}

	return fmt.Sprintf("%s/%s", targetOS, targetArch)
}
