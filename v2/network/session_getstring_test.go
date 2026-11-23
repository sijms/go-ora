package network

import "testing"

// GetString reads a value and returns its first length bytes, and never more than the
// value holds: a length past the value returns the value. A value that cannot be read
// is an error, never a panic.
func TestGetString(t *testing.T) {
	testScenarios := []struct {
		description string
		input       []byte
		length      int
		expected    string
		expectError bool
	}{
		{description: "the whole value", input: []byte{5, 'h', 'e', 'l', 'l', 'o'}, length: 5, expected: "hello"},
		{description: "a prefix of the value", input: []byte{5, 'h', 'e', 'l', 'l', 'o'}, length: 4, expected: "hell"},
		{description: "a length past the value", input: []byte{5, 'h', 'e', 'l', 'l', 'o'}, length: 8, expected: "hello"},
		{description: "a NULL value with a length", input: []byte{0}, length: 1, expected: ""},
		{description: "a negative length", input: []byte{5, 'h', 'e', 'l', 'l', 'o'}, length: -1, expected: ""},
		{description: "no input", input: nil, length: 1, expectError: true},
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("GetString(%d) panicked: %v", scenario.length, r)
				}
			}()
			got, err := NewSessionWithInputBufferForDebug(scenario.input).GetString(scenario.length)
			if scenario.expectError {
				if err == nil {
					t.Errorf("GetString(%d) = %q, want an error", scenario.length, got)
				}
				return
			}
			if err != nil || got != scenario.expected {
				t.Errorf("GetString(%d) = %q, %v; want %q", scenario.length, got, err, scenario.expected)
			}
		})
	}
}
