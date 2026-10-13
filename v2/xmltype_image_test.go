package go_ora

import (
	"bytes"
	"testing"

	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
)

const al32utf8 = 873

func newXMLTypeTestConnection() *Connection {
	return &Connection{
		session:  network.NewSessionWithInputBufferForDebug(nil),
		tcpNego:  &TCPNego{ServerCharset: al32utf8},
		sStrConv: converters.NewStringConverter(al32utf8),
	}
}

func TestDecodeObject_XMLTypeString(t *testing.T) {
	largeXML := "<a>" + string(bytes.Repeat([]byte("x"), 5000)) + "</a>"
	tests := []struct {
		name  string
		image []byte
		want  string
	}{
		{
			name:  "flag 0x14",
			image: append([]byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00, 0x00, 0x14}, []byte("<a>1</a>")...),
			want:  "<a>1</a>",
		},
		{
			name:  "flag 0x8014",
			image: append([]byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00, 0x80, 0x14}, []byte("<a>1</a>")...),
			want:  "<a>1</a>",
		},
		{
			name:  "long image length",
			image: append([]byte{0x85, 0x01, 0xFE, 0x00, 0x00, 0x13, 0x9B, 0x01, 0x00, 0x00, 0x80, 0x14}, []byte(largeXML)...),
			want:  largeXML,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newXMLTypeTestConnection()
			par := &ParameterInfo{DataType: XMLType, BValue: tt.image}
			var temporaryLobs [][]byte
			if err := decodeObject(conn, par, &temporaryLobs); err != nil {
				t.Fatalf("decodeObject() returned error: %v", err)
			}
			got, ok := par.oPrimValue.(string)
			if !ok {
				t.Fatalf("decodeObject() value = %#v (%T), want string", par.oPrimValue, par.oPrimValue)
			}
			if got != tt.want {
				t.Errorf("decodeObject() value has length %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestDecodeObject_XMLTypeUnknownFlag(t *testing.T) {
	conn := newXMLTypeTestConnection()
	image := append([]byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00, 0x80, 0x10}, []byte("<a>1</a>")...)
	par := &ParameterInfo{DataType: XMLType, BValue: image}
	var temporaryLobs [][]byte
	if err := decodeObject(conn, par, &temporaryLobs); err == nil {
		t.Errorf("decodeObject() returned nil error and value %#v, want error", par.oPrimValue)
	}
}

func TestParseXMLTypeImage(t *testing.T) {
	largeXML := append(append([]byte("<a>"), bytes.Repeat([]byte("x"), 5000)...), []byte("</a>")...)
	tests := []struct {
		name        string
		image       []byte
		wantFlag    uint32
		wantPayload []byte
	}{
		{
			// oracle 23ai free
			name:        "string",
			image:       append([]byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00, 0x00, 0x14}, []byte("<a>1</a>")...),
			wantFlag:    0x14,
			wantPayload: []byte("<a>1</a>"),
		},
		{
			// autonomous database 26ai
			name:        "string with extra flag bits",
			image:       append([]byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00, 0x80, 0x14}, []byte("<a>1</a>")...),
			wantFlag:    0x8014,
			wantPayload: []byte("<a>1</a>"),
		},
		{
			// autonomous database 26ai, 5007 byte document
			name:        "string with long image length",
			image:       append([]byte{0x85, 0x01, 0xFE, 0x00, 0x00, 0x13, 0x9B, 0x01, 0x00, 0x00, 0x80, 0x14}, largeXML...),
			wantFlag:    0x8014,
			wantPayload: largeXML,
		},
		{
			name:        "skip next 4 flag",
			image:       append([]byte{0x85, 0x01, 0x14, 0x01, 0x00, 0x10, 0x00, 0x14, 0xAA, 0xBB, 0xCC, 0xDD}, []byte("<a>1</a>")...),
			wantFlag:    0x100014,
			wantPayload: []byte("<a>1</a>"),
		},
		{
			name:        "prefix segment",
			image:       append([]byte{0x81, 0x01, 0x13, 0x02, 0xEE, 0xEE, 0x01, 0x00, 0x00, 0x00, 0x14}, []byte("<a>1</a>")...),
			wantFlag:    0x14,
			wantPayload: []byte("<a>1</a>"),
		},
		{
			name:        "lob locator",
			image:       []byte{0x85, 0x01, 0x0B, 0x01, 0x00, 0x00, 0x80, 0x11, 0xDE, 0xAD, 0xBE},
			wantFlag:    0x8011,
			wantPayload: []byte{0xDE, 0xAD, 0xBE},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlag, gotPayload, err := parseXMLTypeImage(tt.image)
			if err != nil {
				t.Fatalf("parseXMLTypeImage() returned error: %v", err)
			}
			if gotFlag != tt.wantFlag {
				t.Errorf("parseXMLTypeImage() flag = 0x%X, want 0x%X", gotFlag, tt.wantFlag)
			}
			if !bytes.Equal(gotPayload, tt.wantPayload) {
				t.Errorf("parseXMLTypeImage() payload = %q, want %q", gotPayload, tt.wantPayload)
			}
		})
	}
}

func TestParseXMLTypeImageErrors(t *testing.T) {
	tests := []struct {
		name  string
		image []byte
	}{
		{name: "empty", image: nil},
		{name: "truncated before xml flag", image: []byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x00}},
		{name: "truncated long length", image: []byte{0x85, 0x01, 0xFE, 0x00, 0x00}},
		{name: "truncated skip next 4", image: []byte{0x85, 0x01, 0x10, 0x01, 0x00, 0x10, 0x00, 0x14, 0xAA}},
		{name: "truncated prefix segment", image: []byte{0x81, 0x01, 0x10, 0x09, 0xEE}},
		{name: "degenerate image", image: []byte{0x95, 0x01, 0x10, 0x01, 0x00, 0x00, 0x00, 0x14}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := parseXMLTypeImage(tt.image); err == nil {
				t.Errorf("parseXMLTypeImage(% X) returned nil error, want error", tt.image)
			}
		})
	}
}
