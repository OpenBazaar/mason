package builder_test

import (
	"testing"

	"github.com/OpenBazaar/samulator/builder"
)

func TestOpenBazaarBuilder(t *testing.T) {
	node := builder.NewOpenBazaarDaemon("label", "HEAD^")

	_, err := node.Build()
	if err != nil {
		t.Fatal(err)
	}
	defer node.MustClean()
}
