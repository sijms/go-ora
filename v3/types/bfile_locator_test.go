package types

import (
	"testing"

	"github.com/sijms/go-ora/v3/converters"
)

// locatorStream is a LobStreamer that only knows its locator.
type locatorStream struct {
	LobStreamer
	locator Locator
}

func (s locatorStream) GetLocator() Locator { return s.locator }

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
// name whose length points past the locator is empty, never a panic.
func TestBFileNamesStayInTheLocator(t *testing.T) {
	testScenarios := []struct {
		description string
		locator     []byte
		dir, file   string
	}{
		{description: "both names", locator: bfileLocator("DATA_DIR", "a.txt"), dir: "DATA_DIR", file: "a.txt"},
		{description: "a file name past the locator", locator: bfileLocator("DATA_DIR", "a.txt")[:30:30], dir: "DATA_DIR"},
		{description: "a directory past the locator", locator: append(make([]byte, 16), 0, 10)},
		{description: "half a length", locator: make([]byte, 17)},
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("reading the names panicked: %v", r)
				}
			}()
			file := BFile{lobBase: lobBase{stream: locatorStream{locator: scenario.locator}}, Conv: converters.NewStringConverter(873)}
			if dir, name := file.GetDirName(), file.GetFileName(); dir != scenario.dir || name != scenario.file {
				t.Errorf("names %q and %q, want %q and %q", dir, name, scenario.dir, scenario.file)
			}
		})
	}
}
