package go_ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/network"
	"github.com/sijms/go-ora/v2/trace"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Compile time Sentinels for implemented Interfaces.
var _ = driver.Rows((*DataSet)(nil))
var _ = driver.RowsColumnTypeDatabaseTypeName((*DataSet)(nil))
var _ = driver.RowsColumnTypeLength((*DataSet)(nil))
var _ = driver.RowsColumnTypeNullable((*DataSet)(nil))

// var _ = driver.RowsColumnTypePrecisionScale((*DataSet)(nil))
// var _ = driver.RowsColumnTypeScanType((*DataSet)(nil))
// var _ = driver.RowsNextResultSet((*DataSet)(nil))

type Row []driver.Value

type DataSet struct {
	columnCount     int
	rowCount        int
	uACBufferLength int
	maxRowSize      int
	Cols            []ParameterInfo
	rows            []Row
	currentRow      Row
	lasterr         error
	index           int
	parent          StmtInterface
}

// load Loading dataset information from network session
func (dataSet *DataSet) load(session *network.Session) error {
	_, err := session.GetByte()
	if err != nil {
		return err
	}
	columnCount, err := session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	num, err := session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	columnCount += num * 0x100
	if columnCount > dataSet.columnCount {
		dataSet.columnCount = columnCount
	}
	if len(dataSet.currentRow) != dataSet.columnCount {
		dataSet.currentRow = make(Row, dataSet.columnCount)
	}
	dataSet.rowCount, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	dataSet.uACBufferLength, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	bitVector, err := session.GetDlc()
	if err != nil {
		return err
	}
	dataSet.setBitVector(bitVector)
	_, err = session.GetDlc()
	return nil
}

// setBitVector bit vector is an array of bit that define which column need to be read
// from network session
func (dataSet *DataSet) setBitVector(bitVector []byte) {
	index := dataSet.columnCount / 8
	if dataSet.columnCount%8 > 0 {
		index++
	}
	if len(bitVector) > 0 {
		for x := 0; x < len(bitVector); x++ {
			for i := 0; i < 8; i++ {
				if (x*8)+i < dataSet.columnCount {
					dataSet.Cols[(x*8)+i].getDataFromServer = bitVector[x]&(1<<i) > 0
				}
			}
		}
	} else {
		for x := 0; x < len(dataSet.Cols); x++ {
			dataSet.Cols[x].getDataFromServer = true
		}
	}

}

func (dataSet *DataSet) Close() error {
	return nil
}

// Next_ act like Next in sql package return false if no other rows in dataset
func (dataSet *DataSet) Next_() bool {
	err := dataSet.Next(dataSet.currentRow)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false
		}
		dataSet.lasterr = err
		return false
	}

	return true
}

// Scan act like scan in sql package return row values to dest variable pointers
func (dataSet *DataSet) Scan(dest ...interface{}) error {
	if dataSet.lasterr != nil {
		return dataSet.lasterr
	}
	//if len(dest) != len(dataSet.currentRow) {
	//	return fmt.Errorf("go-ora: expected %d destination arguments in Scan, not %d",
	//		len(dataSet.currentRow), len(dest))
	//}
	//destIndex := 0
	for srcIndex, destIndex := 0, 0; srcIndex < len(dataSet.currentRow); srcIndex, destIndex = srcIndex+1, destIndex+1 {
		if destIndex >= len(dest) {
			return errors.New("go-or: mismatching between Scan function input count and column count")
		}
		if dest[destIndex] == nil {
			return fmt.Errorf("go-ora: argument %d is nil", destIndex)
		}
		destTyp := reflect.TypeOf(dest[destIndex])
		if destTyp.Kind() != reflect.Ptr {
			return errors.New("go-ora: argument in scan should be passed as pointers")
		}
		col := dataSet.currentRow[srcIndex]
		switch destTyp.Elem().Kind() {
		case reflect.String:
			reflect.ValueOf(dest[destIndex]).Elem().SetString(getString(col))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			temp, err := getInt(col)
			if err != nil {
				return fmt.Errorf("go-ora: column %d require an integer", srcIndex)
			}
			reflect.ValueOf(dest[destIndex]).Elem().SetInt(temp)
			//if temp, ok := col.(int64); ok {
			//	reflect.ValueOf(dest[destIndex]).Elem().SetInt(temp)
			//} else if temp, ok := col.(float64); ok {
			//	reflect.ValueOf(dest[destIndex]).Elem().SetInt(int64(temp))
			//} else if temp, ok := col.(string); ok {
			//	tempInt, err := strconv.ParseInt(temp, 10, 64)
			//	if err != nil {
			//		return fmt.Errorf("go-ora: column %d require an integer", srcIndex)
			//	}
			//	reflect.ValueOf(dest[destIndex]).Elem().SetInt(tempInt)
			//} else {
			//	return fmt.Errorf("go-ora: column %d require an integer", srcIndex)
			//}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			temp, err := getInt(col)
			if err != nil {
				return fmt.Errorf("go-ora: column %d require an integer", srcIndex)
			}
			reflect.ValueOf(dest[destIndex]).Elem().SetUint(uint64(temp))
		case reflect.Float32, reflect.Float64:
			temp, err := getFloat(col)
			if err != nil {
				return fmt.Errorf("go-ora: column %d require type float", srcIndex)
			}
			reflect.ValueOf(dest[destIndex]).Elem().SetFloat(temp)
		default:
			if destTyp.Elem() == reflect.TypeOf(time.Time{}) {
				if _, ok := col.(time.Time); ok {
					reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
				} else {
					return fmt.Errorf("go-ora: column %d require type time.Time", srcIndex)
				}
			} else if destTyp.Elem() == reflect.TypeOf([]byte{}) {
				if _, ok := col.([]byte); ok {
					reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
				} else {
					return fmt.Errorf("go-ora: column %d require type []byte", srcIndex)
				}
			} else if destTyp.Elem().Kind() == reflect.Struct {
				for x := 0; x < destTyp.Elem().NumField(); x++ {
					col := dataSet.currentRow[srcIndex]
					f := destTyp.Elem().Field(x)
					tag := f.Tag.Get("db")
					if len(tag) == 0 {
						continue
					}
					tag = strings.Trim(tag, "\"")
					parts := strings.Split(tag, ",")
					for _, part := range parts {
						subs := strings.Split(part, ":")
						if len(subs) != 2 {
							continue
						}
						if strings.TrimSpace(strings.ToLower(subs[0])) == "name" {
							fieldID := strings.TrimSpace(strings.ToUpper(subs[1]))
							colInfo := dataSet.Cols[srcIndex]
							if strings.ToUpper(colInfo.Name) != fieldID {
								return fmt.Errorf(
									"go-ora: column %d name %s is mismatching with tag name %s of structure field",
									srcIndex, colInfo.Name, fieldID)
							}
							reflect.ValueOf(dest[destIndex]).Elem().Field(x).Set(reflect.ValueOf(col))
							srcIndex++
						}
					}
				}
				srcIndex--
			} else {
				return fmt.Errorf("go-ora: column %d require type %v", srcIndex, reflect.TypeOf(col))
			}
		}
		//if destTyp.Elem().Kind() == reflect.Struct && destTyp.Elem() != reflect.TypeOf(time.Time{}) {
		//	for x := 0; x < destTyp.Elem().NumField(); x++ {
		//		col := dataSet.currentRow[srcIndex]
		//		f := destTyp.Elem().Field(x)
		//		tag := f.Tag.Get("db")
		//		if len(tag) == 0 {
		//			continue
		//		}
		//		tag = strings.Trim(tag, "\"")
		//		parts := strings.Split(tag, ",")
		//		for _, part := range parts {
		//			subs := strings.Split(part, ":")
		//			if len(subs) != 2 {
		//				continue
		//			}
		//			if strings.TrimSpace(strings.ToLower(subs[0])) == "name" {
		//				fieldID := strings.TrimSpace(strings.ToUpper(subs[1]))
		//				colInfo := dataSet.Cols[srcIndex]
		//				if strings.ToUpper(colInfo.Name) != fieldID {
		//					return fmt.Errorf(
		//						"go-ora: column %d name %s is mismatching with tag name %s of structure field",
		//						srcIndex, colInfo.Name, fieldID)
		//				}
		//				reflect.ValueOf(dest[destIndex]).Elem().Field(x).Set(reflect.ValueOf(col))
		//				srcIndex++
		//			}
		//		}
		//	}
		//	srcIndex--
		//	//continue
		//} else {
		//	col := dataSet.currentRow[srcIndex]
		//	switch col.(type) {
		//	case string:
		//		switch destTyp.Elem().Kind() {
		//		case reflect.String:
		//			reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
		//		default:
		//			return fmt.Errorf("go-ora: column %d require type string", srcIndex)
		//		}
		//	case int64:
		//		switch destTyp.Elem().Kind() {
		//		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		//			reflect.ValueOf(dest[destIndex]).Elem().SetInt(reflect.ValueOf(col).Int())
		//		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//			reflect.ValueOf(dest[destIndex]).Elem().SetUint(uint64(reflect.ValueOf(col).Int()))
		//		case reflect.Float32, reflect.Float64:
		//			reflect.ValueOf(dest[destIndex]).Elem().SetFloat(float64(reflect.ValueOf(col).Int()))
		//		default:
		//			return fmt.Errorf("go-ora: column %d require an integer", srcIndex)
		//		}
		//	case float64:
		//		switch destTyp.Elem().Kind() {
		//		case reflect.Float32, reflect.Float64:
		//			reflect.ValueOf(dest[destIndex]).Elem().SetFloat(reflect.ValueOf(col).Float())
		//		default:
		//			return fmt.Errorf("go-ora: column %d require type float", srcIndex)
		//		}
		//	case time.Time:
		//		if destTyp.Elem() == reflect.TypeOf(time.Time{}) {
		//			reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
		//		} else {
		//			return fmt.Errorf("go-ora: column %d require type time.Time", srcIndex)
		//		}
		//	case []byte:
		//		if destTyp.Elem() == reflect.TypeOf([]byte{}) {
		//			reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
		//		} else {
		//			return fmt.Errorf("go-ora: column %d require type []byte", srcIndex)
		//		}
		//	default:
		//		if reflect.TypeOf(col) == destTyp.Elem() {
		//			reflect.ValueOf(dest[destIndex]).Elem().Set(reflect.ValueOf(col))
		//		} else {
		//			return fmt.Errorf("go-ora: column %d require type %v", srcIndex, reflect.TypeOf(col))
		//		}
		//	}
		//}
	}
	//for i, col := range dataSet.currentRow {
	//	if dest[i] == nil {
	//		return fmt.Errorf("go-ora: argument %d is nil", i)
	//	}
	//	destTyp := reflect.TypeOf(dest[i])
	//	if destTyp.Kind() != reflect.Ptr {
	//		return errors.New("go-ora: argument in scan should be passed as pointers")
	//	}
	//
	//	switch col.(type) {
	//	case string:
	//		switch destTyp.Elem().Kind() {
	//		case reflect.String:
	//			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	//		default:
	//			return fmt.Errorf("go-ora: column %d require type string", i)
	//		}
	//	case int64:
	//		switch destTyp.Elem().Kind() {
	//		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	//			reflect.ValueOf(dest[i]).Elem().SetInt(reflect.ValueOf(col).Int())
	//		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	//			reflect.ValueOf(dest[i]).Elem().SetUint(uint64(reflect.ValueOf(col).Int()))
	//		case reflect.Float32, reflect.Float64:
	//			reflect.ValueOf(dest[i]).Elem().SetFloat(float64(reflect.ValueOf(col).Int()))
	//		default:
	//			return fmt.Errorf("go-ora: column %d require an integer", i)
	//		}
	//	case float64:
	//		switch destTyp.Elem().Kind() {
	//		case reflect.Float32, reflect.Float64:
	//			reflect.ValueOf(dest[i]).Elem().SetFloat(reflect.ValueOf(col).Float())
	//		default:
	//			return fmt.Errorf("go-ora: column %d require type float", i)
	//		}
	//	case time.Time:
	//		if destTyp.Elem() == reflect.TypeOf(time.Time{}) {
	//			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	//		} else {
	//			return fmt.Errorf("go-ora: column %d require type time.Time", i)
	//		}
	//	case []byte:
	//		if destTyp.Elem() == reflect.TypeOf([]byte{}) {
	//			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	//		} else {
	//			return fmt.Errorf("go-ora: column %d require type []byte", i)
	//		}
	//	default:
	//		if reflect.TypeOf(col) == destTyp.Elem() {
	//			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	//		} else {
	//			return fmt.Errorf("go-ora: column %d require type %v", i, reflect.TypeOf(col))
	//		}
	//	}
	//}
	return nil
}
func getString(col interface{}) string {
	if temp, ok := col.(string); ok {
		return temp
	} else {
		return fmt.Sprintf("%v", col)
	}
}
func getFloat(col interface{}) (float64, error) {
	if temp, ok := col.(float64); ok {
		return temp, nil
	} else if temp, ok := col.(int64); ok {
		return float64(temp), nil
	} else if temp, ok := col.(string); ok {
		tempFloat, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			return 0, err
		}
		return tempFloat, nil
	} else {
		return 0, errors.New("unkown type")
	}
}
func getInt(col interface{}) (int64, error) {
	if temp, ok := col.(int64); ok {
		return temp, nil
	} else if temp, ok := col.(float64); ok {
		return int64(temp), nil
	} else if temp, ok := col.(string); ok {
		tempInt, err := strconv.ParseInt(temp, 10, 64)
		if err != nil {
			return 0, err
		}
		return tempInt, nil
	} else {
		return 0, errors.New("unkown type")
	}
}

//func setInt(left interface{}, right interface{} ) bool {
//	if temp, ok := right.(int64); ok {
//		reflect.ValueOf(left).Elem().SetInt(temp)
//	} else if temp, ok := right.(float64); ok {
//		reflect.ValueOf(left).Elem().SetInt(int64(temp))
//	} else if temp, ok := right.(string); ok {
//		tempInt, err := strconv.ParseInt(temp, 10, 64)
//		if err != nil {
//			return false
//		}
//		reflect.ValueOf(left).Elem().SetInt(tempInt)
//	} else {
//		return false
//	}
//	return true
//}
// Err return last error
func (dataSet *DataSet) Err() error {
	return dataSet.lasterr
}

// Next implement method need for sql.Rows interface
func (dataSet *DataSet) Next(dest []driver.Value) error {
	hasMoreRows := dataSet.parent.hasMoreRows()
	noOfRowsToFetch := len(dataSet.rows) // dataSet.parent.noOfRowsToFetch()
	hasBLOB := dataSet.parent.hasBLOB()
	hasLONG := dataSet.parent.hasLONG()
	if !hasMoreRows && noOfRowsToFetch == 0 {
		return io.EOF
	}
	if dataSet.index > 0 && dataSet.index%len(dataSet.rows) == 0 {
		if hasMoreRows {
			dataSet.rows = make([]Row, 0, dataSet.parent.noOfRowsToFetch())
			err := dataSet.parent.fetch(dataSet)
			if err != nil {
				return err
			}
			noOfRowsToFetch = len(dataSet.rows)
			hasMoreRows = dataSet.parent.hasMoreRows()
			dataSet.index = 0
			if !hasMoreRows && noOfRowsToFetch == 0 {
				return io.EOF
			}
		} else {
			return io.EOF
		}
	}
	if hasMoreRows && (hasBLOB || hasLONG) && dataSet.index == 0 {
		//dataSet.rows = make([]Row, 0, dataSet.parent.noOfRowsToFetch())
		if err := dataSet.parent.fetch(dataSet); err != nil {
			return err
		}
	}
	if dataSet.index%noOfRowsToFetch < len(dataSet.rows) {
		for x := 0; x < len(dataSet.rows[dataSet.index%noOfRowsToFetch]); x++ {
			dest[x] = dataSet.rows[dataSet.index%noOfRowsToFetch][x]
		}
		dataSet.index++
		return nil
	}
	return io.EOF
}

//func (dataSet *DataSet) NextRow(args... interface{}) error {
//	var values = make([]driver.Value, len(args))
//	err := dataSet.Next(values)
//	if err != nil {
//		return err
//	}
//	for index, arg := range args {
//		*arg = values[index]
//		//if val, ok := values[index].(t); !ok {
//		//
//		//}
//	}
//	return nil
//}

// Columns return a string array that represent columns names
func (dataSet *DataSet) Columns() []string {
	if len(dataSet.Cols) == 0 {
		return nil
	}
	ret := make([]string, len(dataSet.Cols))
	for x := 0; x < len(dataSet.Cols); x++ {
		ret[x] = dataSet.Cols[x].Name
	}
	return ret
}

func (dataSet DataSet) Trace(t trace.Tracer) {
	for r, row := range dataSet.rows {
		if r > 25 {
			break
		}
		t.Printf("Row %d", r)
		for c, col := range dataSet.Cols {
			t.Printf("  %-20s: %v", col.Name, row[c])
		}
	}
}

// ColumnTypeDatabaseTypeName return Col DataType name
func (dataSet DataSet) ColumnTypeDatabaseTypeName(index int) string {
	return dataSet.Cols[index].DataType.String()
}

// ColumnTypeLength return length of column type
func (dataSet DataSet) ColumnTypeLength(index int) (length int64, ok bool) {
	switch dataSet.Cols[index].DataType {
	case NCHAR, CHAR:
		return int64(dataSet.Cols[index].MaxCharLen), true
	case NUMBER:
		return int64(dataSet.Cols[index].Precision), true
	}
	return int64(0), false

}

// ColumnTypeNullable return if column allow null or not
func (dataSet DataSet) ColumnTypeNullable(index int) (nullable, ok bool) {
	return dataSet.Cols[index].AllowNull, true
}
