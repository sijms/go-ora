package network

import (
	"testing"
)

// TestGetInt64NegativeSize verifies that GetInt64 returns an error
// instead of panicking when size is negative. This can happen during
// RAC switchover when the connection stream is corrupted.
// See issue #714.
func TestGetInt64NegativeSize(t *testing.T) {
	session := NewSessionWithInputBufferForDebug(nil)

	_, err := session.GetInt64(-14, false, true)
	if err == nil {
		t.Fatal("expected error for negative size, got nil")
	}
}
