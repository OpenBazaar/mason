package builder

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func generateTempPath(buildName string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return filepath.Join(os.TempDir(), fmt.Sprintf("ob_build_%s_%d", buildName, r.Intn(9999)))
}
