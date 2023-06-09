package go_ora

import (
	"bytes"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

var conn = &Connection{tcpNego: &TCPNego{ServernCharset: 870, ServerCharset: 0x230}}
var expNilPar = ParameterInfo{
	DataType: NCHAR,
	Flag:     3,
	MaxLen:   1,
}

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

	if par.iPrimValue != expPar.iPrimValue {
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
	err = par.encodeValue(nil, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: NCHAR,
		Flag:     3,
		MaxLen:   1,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(5, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: int64(5),
		BValue:     []byte{193, 6},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(10.9, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: float64(10.9),
		BValue:     []byte{193, 11, 91},
	})
	if err != nil {
		t.Error(err)
		return
	}

	// test bool = true
	err = par.encodeValue(true, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: int64(1),
		BValue:     []byte{193, 2},
	})
	if err != nil {
		t.Error(err)
		return
	}

	// test bool = true
	err = par.encodeValue(false, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: int64(0),
		BValue:     []byte{128},
	})
	if err != nil {
		t.Error(err)
		return
	}

	// NullBool = false
	err = par.encodeValue(sql.NullBool{false, true}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: int64(0),
		BValue:     []byte{128},
	})
	if err != nil {
		t.Error(err)
		return
	}

	// NullBool = null
	err = par.encodeValue(sql.NullBool{true, false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: NUMBER,
		Flag:     3,
		MaxLen:   22,
	})
	if err != nil {
		t.Error(err)
		return
	}

	// NullInt32
	err = par.encodeValue(sql.NullInt32{25, true}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:   NUMBER,
		Flag:       3,
		MaxLen:     22,
		iPrimValue: int64(25),
		BValue:     []byte{193, 26},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(sql.NullInt32{25, false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: NUMBER,
		Flag:     3,
		MaxLen:   22,
	})
	if err != nil {
		t.Error(err)
		return
	}

	stringVal := "this is a test"
	err = par.encodeValue(stringVal, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:    NCHAR,
		Flag:        3,
		ContFlag:    16,
		CharsetID:   0x230,
		CharsetForm: 1,
		MaxCharLen:  len(stringVal),
		MaxLen:      len(stringVal),
		iPrimValue:  stringVal,
		BValue:      []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(sql.NullString{stringVal, false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:    NCHAR,
		Flag:        3,
		ContFlag:    16,
		CharsetID:   0x230,
		CharsetForm: 1,
		MaxLen:      1,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(NVarChar(stringVal), -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:    NCHAR,
		Flag:        3,
		ContFlag:    16,
		CharsetID:   870,
		CharsetForm: 2,
		MaxCharLen:  len(stringVal),
		MaxLen:      len(stringVal),
		iPrimValue:  stringVal,
		BValue:      []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(NullNVarChar{NVarChar(stringVal), false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType:    NCHAR,
		Flag:        3,
		ContFlag:    16,
		CharsetID:   870,
		CharsetForm: 2,
		MaxLen:      1,
	})
	if err != nil {
		t.Error(err)
		return
	}

	timeVal := time.Date(2023, 5, 28, 23, 38, 11, 500, time.Local)
	err = par.encodeValue(timeVal, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		DataType:   DATE,
		Flag:       3,
		ContFlag:   0,
		MaxLen:     11,
		iPrimValue: timeVal,
		BValue:     []byte{120, 123, 5, 28, 24, 39, 12},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(sql.NullTime{time.Now(), false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: DATE,
		Flag:     3,
		MaxLen:   11,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(TimeStamp(timeVal), -1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		DataType:   TIMESTAMP,
		Flag:       3,
		ContFlag:   0,
		MaxLen:     11,
		iPrimValue: timeVal,
		BValue:     []byte{120, 123, 5, 28, 24, 39, 12, 0, 0, 1, 244},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(NullTimeStamp{TimeStamp(time.Now()), false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: TIMESTAMP,
		Flag:     3,
		MaxLen:   11,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(TimeStampTZ(timeVal), -1, conn)
	if err != nil {
		t.Error(err)
		return
	}

	err = checkParInfo(par, &ParameterInfo{
		DataType:   TimeStampTZ_DTY,
		Flag:       3,
		ContFlag:   0,
		MaxLen:     13,
		iPrimValue: timeVal,
		BValue:     []byte{120, 123, 5, 28, 24, 39, 12, 0, 0, 1, 244, 23, 60},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = par.encodeValue(NullTimeStampTZ{TimeStampTZ(time.Now()), false}, -1, conn)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkParInfo(par, &ParameterInfo{
		DataType: TimeStampTZ_DTY,
		Flag:     3,
		MaxLen:   13,
	})
	if err != nil {
		t.Error(err)
		return
	}
}
