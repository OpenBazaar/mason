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
		node := builder.NewOpenBazaarDaemon("label", "v0.13.0")

		ob, err := node.Build()
		if err != nil {
			t.Fatal(err)
		}

		version, err := ob.Version()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("current version: %s\n", version)
		defer node.MustClean()
		wg.Done()
	}()

	go func() {
		node := builder.NewOpenBazaarDaemon("label", "v0.12.0")

		ob, err := node.Build()
		if err != nil {
			t.Fatal(err)
		}

		version, err := ob.Version()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("current version: %s\n", version)
		defer node.MustClean()
		wg.Done()
	}()

	wg.Wait()
}
