package go_ora

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/trace"
	"github.com/sijms/go-ora/v3/types"
)

var (
	conn      = newTestConnection()
	expNilPar = ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NCHAR,
			Flag:     3,
			MaxLen:   1,
		},
	}
)

func checkParInfo(par *ParameterInfo, expPar *ParameterInfo) error {
	if par.CharsetForm != expPar.CharsetForm {
		return fmt.Errorf("expected charset form %v and get %v", expPar.CharsetForm, par.CharsetForm)
	}
	if par.CharsetID != expPar.CharsetID {
		return fmt.Errorf("expected charset id %v and get %v", expPar.CharsetID, par.CharsetID)
	}
	if par.DataType != expPar.DataType {
		return fmt.Errorf("expected data type %v and get %v", expPar.DataType, par.DataType)
	}
	if par.Flag != expPar.Flag {
		return fmt.Errorf("expected flag %v and get %v", expPar.Flag, par.Flag)
	}
	if par.ContFlag != expPar.ContFlag {
		return fmt.Errorf("exptected cont flag %v and get %v", expPar.ContFlag, par.ContFlag)
	}
	if par.MaxLen != expPar.MaxLen {
		return fmt.Errorf("expected max len %v and get %v", expPar.MaxLen, par.MaxLen)
	}
	if par.MaxCharLen != expPar.MaxCharLen {
		return fmt.Errorf("expected max char len %v and get %v", expPar.MaxCharLen, par.MaxCharLen)
	}

	if !reflect.DeepEqual(par.iPrimValue, expPar.iPrimValue) {
		return fmt.Errorf("expected primary values %v and get %v", expPar.iPrimValue, par.iPrimValue)
	}
	if bytes.Compare(par.BValue, expPar.BValue) != 0 {
		return fmt.Errorf("expected binary value %v and get %v", expPar.BValue, par.BValue)
	}
	return nil
}

//func testEncodeValue(t *testing.T, title string, par *ParameterInfo, value interface{}, expType TNSType, flag, contFlag, charsetID, charsetForm, maxLen, maxCharLen int) error {
//	t.Log(title)
//	err := par.encodeValue(value, -1, conn)
//	if err != nil {
//		return err
//	}
//	err = checkParInfo(par, expType, flag, contFlag, charsetID, charsetForm, maxLen, maxCharLen)
//	if err != nil {
//		return err
//	}
//	t.Logf("value: %v", par.Value)
//	t.Logf("primitive value: %v", par.iPrimValue)
//	t.Logf("network value: %v", par.BValue)
//	t.Log()
//	return nil
//}

func TestEncodeValue(t *testing.T) {
	// test input parameters
	// test number
	par := &ParameterInfo{Direction: Input}
	var err error
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType:    types.NCHAR,
			Flag:        3,
			ContFlag:    16,
			CharsetID:   0x230,
			CharsetForm: 1,
			MaxLen:      1,
			MaxCharLen:  1,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = 5
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{193, 6},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = 10.9
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{193, 11, 91},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = true
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{193, 2},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = false
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{128},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullBool{false, true}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{128},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullBool{true, false}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullInt32{25, true}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
			BValue:   []byte{193, 26},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullInt32{25, false}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.NUMBER,
			Flag:     3,
			MaxLen:   22,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	stringVal := "this is a test"
	par = &ParameterInfo{Direction: Input}
	par.Value = stringVal
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType:    types.NCHAR,
			Flag:        3,
			ContFlag:    16,
			CharsetID:   0x230,
			CharsetForm: 1,
			MaxCharLen:  int64(len(stringVal)),
			MaxLen:      int64(len(stringVal)),
			BValue:      []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullString{stringVal, false}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType:    types.NCHAR,
			Flag:        3,
			ContFlag:    16,
			CharsetID:   0x230,
			CharsetForm: 1,
			MaxLen:      1,
			MaxCharLen:  1,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = newTestNString(t, stringVal)
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType:    types.NCHAR,
			Flag:        3,
			ContFlag:    16,
			CharsetID:   870,
			CharsetForm: 2,
			MaxCharLen:  int64(len(stringVal)),
			MaxLen:      int64(len(stringVal)),
			BValue:      []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = newTestNString(t, sql.NullString{String: stringVal, Valid: false})
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType:    types.NCHAR,
			Flag:        3,
			ContFlag:    16,
			CharsetID:   870,
			CharsetForm: 2,
			MaxLen:      1,
			MaxCharLen:  1,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	timeVal := time.Date(2023, 5, 28, 23, 38, 11, 500, time.FixedZone("", 0))
	par = &ParameterInfo{Direction: Input}
	par.Value = timeVal
	conn.dataNego = &DataTypeNego{
		clientTZVersion: 1,
		serverTZVersion: 1,
	}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.DATE,
			Flag:     3,
			ContFlag: 0,
			MaxLen:   7,
			BValue:   []byte{120, 123, 5, 28, 24, 39, 12},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = sql.NullTime{timeVal, false}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.DATE,
			Flag:     3,
			MaxLen:   7,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = types.NewTimeStamp(timeVal)
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.TIMESTAMP,
			Flag:     3,
			ContFlag: 0,
			MaxLen:   11,
			BValue:   []byte{120, 123, 5, 28, 24, 39, 12, 0, 0, 1, 244},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = newTestNullTimeStamp(t)
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.TIMESTAMP,
			Flag:     3,
			MaxLen:   11,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = TimeStampTZ(timeVal)
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.TimeStampTZ_DTY,
			Flag:     3,
			ContFlag: 0,
			MaxLen:   13,
			BValue:   []byte{120, 123, 5, 28, 24, 39, 12, 0, 0, 1, 244, 20, 60},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	par = &ParameterInfo{Direction: Input}
	par.Value = NullTimeStampTZ{TimeStampTZ(time.Now()), false}
	err = par.encodeValue(-1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		BasicParameter: parameter_coder.BasicParameter{
			DataType: types.TimeStampTZ_DTY,
			Flag:     3,
			MaxLen:   13,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
}

// newTestNString builds the national-charset string value that v3 uses for nvarchar (command.go).
func newTestNString(t *testing.T, value interface{}) *types.String {
	t.Helper()
	ret := &types.String{}
	ret.UseNCharset = true
	if err := ret.SetValue(value); err != nil {
		t.Fatal(err)
	}
	return ret
}

func newTestNullTimeStamp(t *testing.T) *types.Date {
	t.Helper()
	ret := types.NewTimeStamp(time.Now())
	if err := ret.SetValue(sql.NullTime{}); err != nil {
		t.Fatal(err)
	}
	return ret
}

// newTestConnection wires the driver's parameter coders and a tracer the same way a real connection does.
func newTestConnection() *Connection {
	ret := &Connection{
		tcpNego:  &TCPNego{ServerNCharset: 870, ServerCharset: 0x230},
		dataNego: &DataTypeNego{},
		tracer:   trace.NilTracer(),
	}
	ret.sStrConv = converters.NewStringConverter(ret.tcpNego.ServerCharset)
	ret.nStrConv = converters.NewStringConverter(ret.tcpNego.ServerNCharset)
	ret.buildParameterCoderMap(GetDefaultDriver())
	return ret
}

// A plain time.Time is bound as DATE: Oracle converts a DATE column to compare it with a
// TIMESTAMP WITH TIME ZONE bind, which stops it from using an index on that column.
// Use TimeStampTZ to send fractional seconds and the zone.
func TestEncodeTimeAsDate(t *testing.T) {
	timeVal := time.Date(2023, 5, 28, 23, 38, 11, 500000000, time.UTC)
	for _, value := range []interface{}{timeVal, sql.NullTime{Time: timeVal, Valid: true}} {
		par := &ParameterInfo{Direction: Input, Value: value}
		if err := par.encodeValue(-1, newTestConnection()); err != nil {
			t.Fatal(err)
		}
		if par.DataType != types.DATE || int64(len(par.BValue)) != types.MaxLenDate {
			t.Errorf("%T: expected data type %d with %d bytes and got %d with %d bytes",
				value, types.DATE, types.MaxLenDate, par.DataType, len(par.BValue))
		}
	}
}
