package builder_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/OpenBazaar/samulator/builder"
	shell "github.com/placer14/go-shell"
)

func init() {
	shell.Panic = false
}

func TestOpenBazaarBuildsInParallel(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		expectedVersion := "0.13.0"
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

		if version != "0.13.0" {
			t.Fatalf("expected version %s, got %s", expectedVersion, version)
		}
		wg.Done()
	}()

	go func() {
		expectedVersion := "0.12.0"
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
		if version != "0.12.0" {
			t.Fatalf("expected version %s, got %s", expectedVersion, version)
		}
		wg.Done()
	}()

	wg.Wait()
}
