package builder

import (
	"fmt"
	"os"

	"github.com/OpenBazaar/samulator/builder/blueprints"
	"github.com/OpenBazaar/samulator/builder/runner"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("builder")

type openBazaarBuilder struct {
	friendlyLabel    string
	versionReference string
	sourceWorkdir    string
}

func NewOpenBazaarDaemon(label, version string) *openBazaarBuilder {
	return &openBazaarBuilder{
		friendlyLabel:    label,
		versionReference: version,
	}
}

func (b *openBazaarBuilder) Build() (*runner.OpenBazaarRunner, error) {
	b.sourceWorkdir = generateTempPath(b.friendlyLabel)
	src, err := blueprints.InflateOpenBazaarDaemon(b.sourceWorkdir)
	defer b.MustClean()
	if err != nil {
		return nil, fmt.Errorf("inflating source: %s", err.Error())
	}
	if err := src.CheckoutVersion(b.versionReference); err != nil {
		return nil, fmt.Errorf("checkout version: %s", err.Error())
	}
	return nil, nil
}

func (b *openBazaarBuilder) MustClean() {
	if err := os.RemoveAll(b.sourceWorkdir); err != nil {
		log.Errorf("cleaning (%s): %s", b.sourceWorkdir, err.Error())
	}
}
