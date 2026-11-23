package types

import "testing"

// A physical UROWID holds 13 bytes; a shorter one is an error, not a panic.
func TestShortPhysicalURowIDIsRefused(t *testing.T) {
	for _, n := range []int{5, 12, 13} {
		data := append([]byte{1}, make([]byte, n-1)...)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("a %d-byte physical UROWID panicked: %v", n, r)
				}
			}()
			id := RowID{Basic: Basic{bValue: data[:n:n], dataType: UROWID}}
			v, err := id.Value()
			switch {
			case n < 13 && err == nil:
				t.Errorf("Value of a %d-byte physical UROWID = %v, want an error", n, v)
			case n == 13 && err != nil:
				t.Errorf("Value of a 13-byte physical UROWID: %v", err)
			}
		}()
	}
}
