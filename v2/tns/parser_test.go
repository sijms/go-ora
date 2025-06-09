package tns

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestConsumeScalarValue(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		terminator rune
		expected   any
		expectErr  bool
	}{
		{
			name:       "simple string",
			input:      "abc)",
			terminator: ')',
			expected:   "abc",
			expectErr:  false,
		},
		{
			name:       "integer",
			input:      "123)",
			terminator: ')',
			expected:   123,
			expectErr:  false,
		},
		{
			name:       "float",
			input:      "123.45)",
			terminator: ')',
			expected:   123.45,
			expectErr:  false,
		},
		{
			name:       "unexpected end of string",
			input:      "abc",
			terminator: ')',
			expected:   "",
			expectErr:  true,
		},
		{
			name:       "whitespace around string",
			input:      "  abc  )",
			terminator: ')',
			expected:   "abc",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser{
				tns: []rune(tt.input),
				pos: 0,
			}
			result, err := p.consumeScalarValue(tt.terminator)
			if tt.expectErr {
				assert.Equal(t, err != nil, true)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, result, tt.expected)
			}
		})
	}
}

func TestConsumeObjectLiteral(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string]any
		expectErr bool
	}{
		{
			name:  "simple object",
			input: "(key=value)",
			expected: map[string]any{
				"key": "value",
			},
			expectErr: false,
		},
		{
			name:  "multiple keys",
			input: "(key1=value1)(key2=value2)",
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expectErr: false,
		},
		{
			name:  "repeated keys",
			input: "(key=value1)(key=value2)",
			expected: map[string]any{
				"key": []any{"value1", "value2"},
			},
			expectErr: false,
		},
		{
			name:  "nested objects",
			input: "(key=(nestedKey=nestedValue))",
			expected: map[string]any{
				"key": map[string]any{
					"nestedKey": "nestedValue",
				},
			},
			expectErr: false,
		},
		{
			name:  "spaces between key and value",
			input: "(key = value)",
			expected: map[string]any{
				"key": "value",
			},
			expectErr: false,
		},
		{
			name:  "spaces between key-value pairs",
			input: "(key1=value1) (key2=value2)",
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expectErr: false,
		},
		{
			name:  "spaces around key-value pairs",
			input: " ( key1 = value1 ) ( key2 = value2 ) ",
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expectErr: false,
		},
		{
			name:      "unexpected end of string",
			input:     "(key=value",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "empty key",
			input:     "(=value)",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser{
				tns: []rune(tt.input),
				pos: 0,
			}
			result, err := p.consumeObjectLiteral()
			if tt.expectErr {
				assert.Equal(t, err != nil, true)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, result, tt.expected)
			}
		})
	}
}
