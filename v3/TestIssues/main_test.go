//go:build !integration

package TestIssues

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if server == "" {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
