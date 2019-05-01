package builder

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	targetOS, targetArch string
	projectDirectory     = ".samulator"

	homeDir          = os.Getenv("HOME")
	defaultBuildPath = filepath.Join(homeDir, projectDirectory, "build")
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

func generateTempPath(buildName string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return filepath.Join(workDir(), fmt.Sprintf("samulator_build_%s_%d", buildName, r.Intn(9999)))
}

func getXGoBuildTarget() string {
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
