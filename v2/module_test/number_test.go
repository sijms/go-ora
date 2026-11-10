package v2_test

import (
	"strings"
	"testing"
)

// TestSmallNumberIsNotZero checks that values from 1E-130 to below 1E-128, which are stored
// starting with the zero byte 0x80, are not read as 0. Needs ORACLE_DSN.
func TestSmallNumberIsNotZero(t *testing.T) {
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if db == nil {
		t.Skip("ORACLE_DSN is not set")
	}
	defer db.Close()
	cases := []struct{ literal, want string }{
		{"0", "0"},
		{"1E-130", "0." + strings.Repeat("0", 129) + "1"},
		{"1E-129", "0." + strings.Repeat("0", 128) + "1"},
		{"1.5E-129", "0." + strings.Repeat("0", 128) + "15"},
		{"-1E-130", "-0." + strings.Repeat("0", 129) + "1"},
	}
	for _, c := range cases {
		var got string
		if err := db.QueryRow(`SELECT ` + c.literal + ` FROM dual`).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.literal, err)
		}
		if got != c.want {
			t.Errorf("%s scanned as %q, want %q", c.literal, got, c.want)
		}
	}
}
