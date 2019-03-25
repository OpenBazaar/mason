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

	packageName     = "samulator"
	defaultTempPath = fmt.Sprintf("%s/.%s/", os.Getenv("HOME"), packageName)
)

func generateTempPath(buildName string) string {
	var tempPath = defaultTempPath
	if overrideTemp := os.Getenv("TEMP_PATH"); overrideTemp != "" {
		log.Infof("using alternative TEMP_PATH (%s)", overrideTemp)
		tempPath = overrideTemp
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return filepath.Join(tempPath, fmt.Sprintf("ob_build_%s_%d", buildName, r.Intn(9999)))
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
