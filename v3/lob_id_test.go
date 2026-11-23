package go_ora

import "testing"

// GetLobId is bytes 10 to 20 of the locator, or nil for a locator too short to hold it.
func TestGetLobIdStaysInTheLocator(t *testing.T) {
	locator := make([]byte, 24)
	for i := range locator {
		locator[i] = byte(i)
	}
	if got := (&LobStream{sourceLocator: locator}).GetLobId(); len(got) != 10 || got[0] != 10 || got[9] != 19 {
		t.Errorf("GetLobId = %v, want bytes 10 to 19", got)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetLobId of a 15-byte locator panicked: %v", r)
		}
	}()
	for _, n := range []int{15, 19} {
		if got := (&LobStream{sourceLocator: locator[:n:n]}).GetLobId(); got != nil {
			t.Errorf("GetLobId of a %d-byte locator = %v, want nil", n, got)
		}
	}
	if got := (&LobStream{sourceLocator: locator[:20:20]}).GetLobId(); len(got) != 10 {
		t.Errorf("GetLobId of a 20-byte locator = %v, want 10 bytes", got)
	}
}
