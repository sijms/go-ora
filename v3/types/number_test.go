package types

import (
	"bytes"
	"database/sql"
	"testing"
)

func TestNumberSetValueBool(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  []byte
	}{
		{"true", true, []byte{193, 2}},
		{"false", false, []byte{128}},
		{"valid NullBool", sql.NullBool{Bool: true, Valid: true}, []byte{193, 2}},
		{"null NullBool", sql.NullBool{}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var num Number
			if err := num.SetValue(c.input); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(num.bValue, c.want) {
				t.Errorf("expected %v and got %v", c.want, num.bValue)
			}
		})
	}
}
