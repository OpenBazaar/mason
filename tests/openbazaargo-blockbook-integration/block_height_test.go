// +build integration btc

package openbazaargo_blockbook_integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/OpenBazaar/mason/builder"
)

func versionUnderTest() string {
	if obgoVersion := os.Getenv("OPENBAZAARGO_VERSION"); obgoVersion != "" {
		return obgoVersion
	}
	return "master"
}

func blockbookUnderTest() string {
	if apiEndpoint := os.Getenv("BLOCKBOOK_ENDPOINT"); apiEndpoint != "" {
		return apiEndpoint
	}
	return "https://btc.dev.ob1.io/api"
}

func TestBlockHeight(t *testing.T) {
	var obBuilder = builder.NewOpenBazaarDaemon(c.Args.Version, c.Args.Version)
	defer obBuilder.MustClean()

	var obProc, err = obBuilder.Build()
	if err != nil {
		return fmt.Errorf("building: %s", err.Error())
	}
	defer obProc.Cleanup()

	pr := obProc.SplitOutput()
	go logNodeOutput(pr, c.Args.Version)

	obProc.WithArgs(c.Args.StartParams)
	exit, err := obProc.RunStart().ExitCodeAndErr()
	if err != nil {
		return fmt.Errorf("returned (%d): %s", exit, err.Error())
	}
	return nil

}
