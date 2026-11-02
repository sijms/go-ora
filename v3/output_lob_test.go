package go_ora

import (
	"context"
	"errors"
	"testing"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/types"
)

type fakeLob struct {
	reads int
	err   error
	null  bool
	mode  configurations.LobReadMode
}

func (f *fakeLob) GetLocator() types.Locator               { return nil }
func (f *fakeLob) Free() error                             { return nil }
func (f *fakeLob) GetLength() (int64, error)               { return 0, nil }
func (f *fakeLob) GetReadMode() configurations.LobReadMode { return f.mode }
func (f *fakeLob) Scan(interface{}) error                  { return nil }
func (f *fakeLob) Read(context.Context) error              { f.reads++; return f.err }
func (f *fakeLob) IsNil() bool                             { return f.null }
func (f *fakeLob) Value() (interface{}, error)             { return "lob text", nil }

func TestReadLobValue(t *testing.T) {
	cases := []struct {
		name        string
		lob         *fakeLob
		wantValue   interface{}
		wantReplace bool
		wantReads   int
	}{
		{"null lob becomes nil without a read", &fakeLob{null: true}, nil, true, 0},
		{"lob read mode none keeps the locator", &fakeLob{mode: configurations.LobReadMode_NONE}, nil, false, 0},
		{"lob read mode auto reads the content", &fakeLob{}, "lob text", true, 1},
	}
	for _, tc := range cases {
		value, replace, err := readLobValue(tc.lob)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if value != tc.wantValue || replace != tc.wantReplace || tc.lob.reads != tc.wantReads {
			t.Errorf("%s: got (%v, %v, reads=%d), want (%v, %v, reads=%d)",
				tc.name, value, replace, tc.lob.reads, tc.wantValue, tc.wantReplace, tc.wantReads)
		}
	}
}

func TestReadLobValueReturnsReadError(t *testing.T) {
	want := errors.New("read failed")
	if _, _, err := readLobValue(&fakeLob{err: want}); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func TestReadOutputLobsReadsNestedElements(t *testing.T) {
	single, first, nested := &fakeLob{}, &fakeLob{}, &fakeLob{}
	if err := readOutputLobs(single); err != nil {
		t.Fatal(err)
	}
	if err := readOutputLobs([]interface{}{first, "not a lob", nil, []interface{}{nested}}); err != nil {
		t.Fatal(err)
	}
	for name, lob := range map[string]*fakeLob{"single": single, "first": first, "nested": nested} {
		if lob.reads != 1 {
			t.Errorf("%s lob read %d times, want 1", name, lob.reads)
		}
	}
}

func TestReadOutputLobsReturnsReadError(t *testing.T) {
	want := errors.New("read failed")
	if err := readOutputLobs([]interface{}{&fakeLob{err: want}}); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
