package go_ora

import (
	"testing"

	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
)

// bfileLocator is a BFILE locator: 16 bytes, then the directory and the file name, each
// a 2-byte length and its bytes.
func bfileLocator(dir, file string) []byte {
	loc := make([]byte, 16)
	loc = append(loc, byte(len(dir)>>8), byte(len(dir)))
	loc = append(loc, dir...)
	loc = append(loc, byte(len(file)>>8), byte(len(file)))
	return append(loc, file...)
}

// A BFILE's directory and file name are read from its locator, and only from it: a
// name whose length points past the locator stays empty, never a panic nor bytes the
// locator does not hold.
func TestBFileNamesStayInTheLocator(t *testing.T) {
	testScenarios := []struct {
		description string
		locator     []byte
		dir, file   string
	}{
		{description: "both names", locator: bfileLocator("DATA_DIR", "a.txt"), dir: "DATA_DIR", file: "a.txt"},
		{description: "a file name past the locator", locator: bfileLocator("DATA_DIR", "a.txt")[:30], dir: "DATA_DIR"},
		{description: "a directory past the locator", locator: append(make([]byte, 16), 0, 10)},
		{description: "half a length", locator: make([]byte, 17)},
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("decoding the locator panicked: %v", r)
				}
			}()
			// decodePrimValue reads a value, then, for a BFILE, the locator
			input := append([]byte{1, 1, byte(len(scenario.locator))}, scenario.locator...)
			conn := &Connection{session: network.NewSessionWithInputBufferForDebug(input), sStrConv: converters.NewStringConverter(873)}
			par := ParameterInfo{DataType: OCIFileLocator}
			var temporaryLobs [][]byte
			if err := par.decodePrimValue(conn, &temporaryLobs, false); err != nil {
				t.Fatal(err)
			}
			file, ok := par.oPrimValue.(BFile)
			if !ok {
				t.Fatalf("%T, want BFile", par.oPrimValue)
			}
			if file.dirName != scenario.dir || file.fileName != scenario.file {
				t.Errorf("names %q and %q, want %q and %q", file.dirName, file.fileName, scenario.dir, scenario.file)
			}
		})
	}
}

// GetLobId is bytes 10 to 20 of the locator, or nil for a locator too short to hold it.
func TestGetLobIdStaysInTheLocator(t *testing.T) {
	var lob Lob
	locator := make([]byte, 24)
	for i := range locator {
		locator[i] = byte(i)
	}
	if got := lob.GetLobId(locator); len(got) != 10 || got[0] != 10 || got[9] != 19 {
		t.Errorf("GetLobId = %v, want bytes 10 to 19", got)
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("GetLobId of a 15-byte locator panicked: %v", r)
			}
		}()
		for _, n := range []int{15, 19} {
			if got := lob.GetLobId(locator[:n:n]); got != nil {
				t.Errorf("GetLobId of a %d-byte locator = %v, want nil", n, got)
			}
		}
		if got := lob.GetLobId(locator[:20:20]); len(got) != 10 {
			t.Errorf("GetLobId of a 20-byte locator = %v, want 10 bytes", got)
		}
	}()
}
