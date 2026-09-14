package TestIssues

import (
	"context"
	"testing"
)

// TestIssue732 verifies that the FastLogin cookie optimization does not cause
// EOF errors on the second (and subsequent) physical connections.
// See: https://github.com/sijms/go-ora/issues/732
func TestIssue732(t *testing.T) {
	// FastLogin is on by default; we explicitly ensure it is not disabled
	// so this test exercises the cookie-replay path.
	urlOptions["FAST+LOGIN"] = "true"
	defer delete(urlOptions, "FAST+LOGIN")

	db, err := getDB()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()

	// Set MaxIdleConns to 0 so each ping opens a new physical connection
	// (the pool won't retain idle connections between calls).
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	// With the bug, the second ping would fail with io.EOF because
	// the cached cookie is replayed and the 23c server rejects it.
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		err = db.PingContext(ctx)
		if err != nil {
			t.Fatalf("ping #%d: %v", i, err)
		}
	}
}
