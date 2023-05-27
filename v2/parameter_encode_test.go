package go_ora

import (
	"database/sql"
	"testing"
)

func TestEncodeValue(t *testing.T) {
	conn := &Connection{tcpNego: &TCPNego{ServernCharset: 870, ServerCharset: 0x230}}
	par := ParameterInfo{Direction: Output}
	t.Log("Encode sql.NullString[output]")
	err := par.encodeValue(&sql.NullString{}, 200, conn)
	if err != nil {
		t.Error(err)
	}
	if par.CharsetID != 0x230 {
		t.Errorf("charset id expected: %v and get %v", 0x230, par.CharsetID)
	}
	if par.BValue != nil {
		t.Error("binary value is not empty")
	}
	t.Log("Enocde sql.NullInt64[output]")
	err = par.encodeValue(&sql.NullInt64{}, 0, conn)
	if err != nil {
		t.Error(err)
	}
	if par.BValue != nil {
		t.Error("binary value is not empty")
	}
	t.Log("Enocde Int64[output]")
	temp := int64(5)
	err = par.encodeValue(&temp, 0, conn)
	if err != nil {
		t.Error(err)
	}
	if par.BValue != nil {
		t.Error("binary value is not empty")
	}

	t.Log("Enocde Clob[output]")
	err = par.encodeValue(&Clob{}, 0, conn)
	if err != nil {
		t.Error(err)
	}
	if par.BValue != nil {
		t.Error("binary value is not empty")
	}
}
