package builder_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/OpenBazaar/mason/builder"
	shell "github.com/placer14/go-shell"
)

func init() {
	shell.Panic = false
}

func TestOpenBazaarBuildsInParallel(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		expectedVersion := "0.13.2"
		node := builder.NewOpenBazaarDaemon("label", fmt.Sprintf("v%s", expectedVersion))
		defer node.MustClean()

		ob, err := node.Build()
		if err != nil {
			t.Fatal(err)
		}

		version, err := ob.Version()
		if err != nil {
			t.Fatal(err)
		}

		if version != expectedVersion {
			t.Fatalf("expected version %s, got %s", expectedVersion, version)
		}
	}()

	go func() {
		defer wg.Done()
		expectedVersion := "0.13.1"
		node := builder.NewOpenBazaarDaemon("obParallelBuilds", fmt.Sprintf("v%s", expectedVersion))
		defer node.MustClean()

		ob, err := node.Build()
		if err != nil {
			t.Fatal(err)
		}

		version, err := ob.Version()
		if err != nil {
			t.Fatal(err)
		}
		if version != expectedVersion {
			t.Errorf("expected version %s, got %s", expectedVersion, version)
		}
	}()

	wg.Wait()
}
