package go_ora

import "testing"

// A VECTOR holds as many elements as its header counts; a count past its bytes is
// vectorTypeError, never a panic nor values read from bytes the vector does not hold.
func TestVectorCountStaysInTheValue(t *testing.T) {
	// magic 219, version 0, flag 1 (a 1-byte count), then the format and the count
	header := func(format, count byte) []byte { return []byte{219, 0, 0, 1, format, count} }
	testScenarios := []struct {
		description string
		value       []byte
		expectError bool
		count       int
	}{
		{description: "float32, all elements", value: append(header(2, 2), 0x3F, 0x80, 0, 0, 0x40, 0, 0, 0), count: 2},
		{description: "float32, one element short", value: append(header(2, 2), 0x3F, 0x80, 0, 0), expectError: true},
		{description: "float64, one element short", value: append(header(3, 2), 0x3F, 0xF0, 0, 0, 0, 0, 0, 0), expectError: true},
		{description: "int8, all elements", value: append(header(4, 3), 1, 2, 3), count: 3},
		{description: "int8, one element short", value: append(header(4, 3), 1, 2), expectError: true},
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("decode panicked: %v", r)
				}
			}()
			value := scenario.value[:len(scenario.value):len(scenario.value)]
			var v Vector
			err := v.decode(value)
			if scenario.expectError {
				if err == nil {
					t.Errorf("decode = %v, want an error", v.Data)
				}
				return
			}
			if err != nil || v.Count != scenario.count {
				t.Errorf("decode: count %d, %v; want %d", v.Count, err, scenario.count)
			}
		})
	}
}
