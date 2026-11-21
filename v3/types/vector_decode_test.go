package types

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
)

// TestVectorValue decodes vector images Oracle 23ai sent, one of each format.
func TestVectorValue(t *testing.T) {
	cases := map[string]struct {
		image string
		want  interface{}
	}{
		"float32":        {"db00001202000000030000000000000000bf800000c0200000c0400000", []float32{1, 2.5, 3}},
		"float64":        {"db00001203000000020000000000000000bfb999999999999a01c81bc377ff8a63", []float64{0.1, -1e300}},
		"int8, signed":   {"db00001204000000030000000000000000" + "01fe7f", []int8{1, -2, 127}},
		"binary":         {"db01001005000000108000000000000000ff00", []uint8{255, 0}},
		"sparse float32": {"db0200320200000005c00752e50db3a3a2000200000001000000" + "03bfc00000c0200000", SparseVector{Dimensions: 5, Indices: []uint32{1, 3}, Values: []float32{1.5, 2.5}}},
		"sparse int8":    {"db02003204000000" + "03c01c00000000000000010000000007", SparseVector{Dimensions: 3, Indices: []uint32{0}, Values: []int8{7}}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			image, err := hex.DecodeString(c.image)
			if err != nil {
				t.Fatal(err)
			}
			vector := &Vector{}
			vector.SetBytes(image)
			got, err := vector.Value()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %#v, want %#v", got, c.want)
			}
			// a value reads the same each time
			again, _ := vector.Value()
			if !reflect.DeepEqual(again, c.want) {
				t.Errorf("read again: %#v", again)
			}
		})
	}
}

func TestVectorInt8RoundTrip(t *testing.T) {
	vector := &Vector{}
	if err := vector.SetValue([]int8{1, -2, 127}); err != nil {
		t.Fatal(err)
	}
	got, err := vector.Value()
	if err != nil || !reflect.DeepEqual(got, []int8{1, -2, 127}) {
		t.Errorf("got %#v, %v", got, err)
	}
}

// TestNumberSmallest checks NUMBER's smallest magnitudes, whose first byte is
// zero's.
func TestNumberSmallest(t *testing.T) {
	tiny := "0." + strings.Repeat("0", 129) + "1"
	cases := map[string]struct {
		data []byte
		want string
	}{
		"zero":     {[]byte{128}, "0"},
		"1e-130":   {[]byte{128, 2}, tiny},
		"-1e-130":  {[]byte{127, 100, 102}, "-" + tiny},
		"1.1e-129": {[]byte{128, 12}, "0." + strings.Repeat("0", 128) + "11"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			n := &Number{}
			n.SetBytes(c.data)
			got, err := n.String()
			if err != nil || got != c.want {
				t.Errorf("got %s, %v, want %s", got, err, c.want)
			}
		})
	}
}
