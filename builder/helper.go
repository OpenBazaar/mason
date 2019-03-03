package builder

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
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
