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

	homeDir          = os.Getenv("HOME")
	defaultBuildPath = filepath.Join(homeDir, ".samulator", "build")
)

func generateTempPath(buildName string) string {
	var tempPath = defaultBuildPath
	if overrideTemp := os.Getenv("BUILD_PATH"); overrideTemp != "" {
		log.Infof("using alternative BUILD_PATH (%s)", overrideTemp)
		tempPath = overrideTemp
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return filepath.Join(tempPath, fmt.Sprintf("samulator_build_%s_%d", buildName, r.Intn(9999)))
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
