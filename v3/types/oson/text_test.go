package oson

import (
	"encoding/hex"
	"testing"
)

// TestText checks the text of OSON images Oracle 23ai sent against what its
// JSON_SERIALIZE wrote for the same values.
func TestText(t *testing.T) {
	cases := map[string]struct {
		image string
		want  string
	}{
		"an object":                   {"ff4a5a012106010002000800002c00000161840101000521c102", "{\"a\":1}"},
		"a scalar":                    {"ff4a5a010016000321c12b", "42"},
		"a string":                    {"ff4a5a01001600020178", "\"x\""},
		"members in stored order":     {"ff4a5a012106060014002f00012c4a6f78adea0010000000070012000e000c066e65737465640464656570016b017a0161016d840402050104000e00260029002c8401030013c00200190020840106001e0176c0010024c00021c10221c10321c104", "{\"nested\":{\"deep\":[{\"k\":\"v\"},[[]]]},\"z\":1,\"a\":2,\"m\":3}"},
		"numbers":                     {"ff4a5a01200600000000370000c00a00160019001d002100240027002b002e0031003521800222005b6622ff645b21be0221c20222c1020b21d40b21d502223f33662080", "[1E-130,-1E+125,9.99E+125,0.000001,100,1.1,1000000000000000000000000000000000000000,1E+40,-0.5,0]"},
		"scalars without a JSON type": {"ff4a5a0131060a00190000007f00000013156e73999ba3addb0014000e000a00040000000c00070002001700110164017402743902747a0172016602626402796d026473017a840a050804070306020a01090020002800340040004e005300580061006700733c787c010f01010139787c010f0b1f010754d4c039787c010f0b1f01075bcd157c787c010f0601011dcd6500195a3a000200ff7fbfc0000036bfb999999999999a3d7fffffff3a3e80000004414243b50c52803e800000003c3c3c80000000", "{\"d\":\"2024-01-15T00:00:00\",\"t\":\"2024-01-15T10:30:00.123000\",\"t9\":\"2024-01-15T10:30:00.123456789\",\"tz\":\"2024-01-15T10:30:00.500000+05:30\",\"r\":\"00FF\",\"f\":1.5,\"bd\":0.1,\"ym\":\"-P1Y2M\",\"ds\":\"P4DT5H6M7.890000S\",\"z\":\"P0D\"}"},
		"escapes":                     {"ff4a5a013106010002000000120000820000017384010100050c71225c2fc3bc0a017fe282ac", "{\"s\":\"q\\\"\\\\/ü\\n\\u0001\\u007F€\"}"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			image, err := hex.DecodeString(c.image)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Text(image)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != c.want {
				t.Errorf("got  %s\nwant %s", got, c.want)
			}
		})
	}
}

func TestJsonNumber(t *testing.T) {
	cases := map[string]struct {
		text string
		want string
	}{
		"an integer":                {"42", "42"},
		"a fraction":                {"0.5", "0.5"},
		"no leading zero":           {".5", "0.5"},
		"negative, no leading zero": {"-.5", "-0.5"},
		"10^39 in full":             {"1" + zeros(39), "1" + zeros(39)},
		"10^40 in E notation":       {"1" + zeros(40), "1E+40"},
		"10^-40 in full":            {"0." + zeros(39) + "1", "0." + zeros(39) + "1"},
		"10^-41 in E notation":      {"0." + zeros(40) + "15", "1.5E-41"},
		"negative and large":        {"-999" + zeros(123), "-9.99E+125"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := jsonNumber(c.text); got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}

func zeros(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '0'
	}
	return string(b)
}
