package go_ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/network"
	"github.com/sijms/go-ora/v2/trace"
	"io"
	"reflect"
	"strings"
)

type Row []driver.Value
type ResultSet struct {
	columnCount     int
	rowCount        int
	uACBufferLength int
	maxRowSize      int
	cols            *[]ParameterInfo
	rows            []Row
	currentRow      Row
	index           int
	parent          StmtInterface
	lastErr         error
}

func (resultSet *ResultSet) load(session *network.Session) error {
	var err error
	_, err = session.GetByte()
	if err != nil {
		return err
	}
	var columnCount, num int
	columnCount, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	num, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	columnCount += num * 0x100
	if resultSet.columnCount == 0 {
		resultSet.columnCount = columnCount
	}

	if len(resultSet.currentRow) != resultSet.columnCount {
		resultSet.currentRow = make(Row, resultSet.columnCount)
	}
	resultSet.rowCount, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	resultSet.uACBufferLength, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	bitVector, err := session.GetDlc()
	if err != nil {
		return err
	}
	resultSet.setBitVector(bitVector)
	_, err = session.GetDlc()
	return nil
}

// setBitVector bit vector is an array of bit that defines which column needs to be read
// from network session
func (resultSet *ResultSet) setBitVector(bitVector []byte) {
	index := resultSet.columnCount / 8
	if resultSet.columnCount%8 > 0 {
		index++
	}
	if len(bitVector) > 0 {
		for x := 0; x < len(bitVector); x++ {
			for i := 0; i < 8; i++ {
				if (x*8)+i < resultSet.columnCount {
					(*resultSet.cols)[(x*8)+i].getDataFromServer = bitVector[x]&(1<<i) > 0
				}
			}
		}
	} else {
		if resultSet.cols != nil {
			for x := 0; x < len(*resultSet.cols); x++ {
				(*resultSet.cols)[x].getDataFromServer = true
			}
		}
	}
}

func (resultSet *ResultSet) Close() error {
	if resultSet.parent.CanAutoClose() {
		return resultSet.parent.Close()
	}
	return nil
}

func (resultSet *ResultSet) Columns() []string {
	if len(*resultSet.cols) == 0 {
		return nil
	}
	ret := make([]string, len(*resultSet.cols))
	for x := 0; x < len(*resultSet.cols); x++ {
		ret[x] = (*resultSet.cols)[x].Name
	}
	return ret
}

func (resultSet *ResultSet) Trace(t trace.Tracer) {
	for r, row := range resultSet.rows {
		if r > 25 {
			break
		}
		t.Printf("Row %d", r)
		for c, col := range *resultSet.cols {
			t.Printf("  %-20s: %v", col.Name, row[c])
		}
	}
}

// ColumnTypeDatabaseTypeName return Col DataType name
func (resultSet *ResultSet) ColumnTypeDatabaseTypeName(index int) string {
	return (*resultSet.cols)[index].DataType.String()
}

// ColumnTypeLength return length of column type
func (resultSet *ResultSet) ColumnTypeLength(index int) (int64, bool) {
	switch (*resultSet.cols)[index].DataType {
	case NCHAR, CHAR:
		return int64((*resultSet.cols)[index].MaxCharLen), true
	}
	return int64(0), false
}

// ColumnTypeNullable return if column allow null or not
func (resultSet *ResultSet) ColumnTypeNullable(index int) (nullable, ok bool) {
	return (*resultSet.cols)[index].AllowNull, true
}

// ColumnTypePrecisionScale return the precision and scale for numeric types
func (resultSet *ResultSet) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	col := (*resultSet.cols)[index]
	switch col.DataType {
	case NUMBER:
		return int64(col.Precision), int64(col.Scale), true
	}
	return int64(0), int64(0), false
}

func (resultSet *ResultSet) ColumnTypeScanType(index int) reflect.Type {
	col := (*resultSet.cols)[index]
	switch col.DataType {
	case NUMBER:
		if col.Precision > 0 {
			return tyFloat64
		} else {
			return tyInt64
		}
	case ROWID, UROWID:
		fallthrough
	case CHAR, NCHAR:
		fallthrough
	case OCIClobLocator:
		return tyString
	case RAW:
		fallthrough
	case OCIBlobLocator, OCIFileLocator:
		return tyBytes
	case DATE, TIMESTAMP:
		fallthrough
	case TimeStampDTY:
		fallthrough
	case TimeStampeLTZ, TimeStampLTZ_DTY:
		fallthrough
	case TIMESTAMPTZ, TimeStampTZ_DTY:
		return tyTime
	case IBFloat:
		return tyFloat32
	case IBDouble:
		return tyFloat64
	case IntervalDS_DTY, IntervalYM_DTY:
		return tyString
	default:
		return nil
	}
}

func (resultSet *ResultSet) Err() error {
	return resultSet.lastErr
}

// set object value using currentRow[colIndex] return true if succeed or false
// for non-supported type
// error means error occur during operation
func (resultSet *ResultSet) setObjectValue(obj reflect.Value, colIndex int) error {
	col := (*resultSet.cols)[colIndex]
	return setFieldValue(obj, col.cusType, resultSet.currentRow[colIndex])
}

// Scan act like scan in sql package return row values to dest variable pointers
func (resultSet *ResultSet) Scan(dest ...interface{}) error {
	if resultSet.lastErr != nil {
		return resultSet.lastErr
	}
	for srcIndex, destIndex := 0, 0; srcIndex < len(resultSet.currentRow); srcIndex, destIndex = srcIndex+1, destIndex+1 {
		if destIndex >= len(dest) {
			return errors.New("go-ora: mismatching between Scan function input count and column count")
		}
		if dest[destIndex] == nil {
			return fmt.Errorf("go-ora: argument %d is nil", destIndex)
		}
		destTyp := reflect.TypeOf(dest[destIndex])
		if destTyp.Kind() != reflect.Ptr {
			return errors.New("go-ora: argument in scan should be passed as pointers")
		}
		destTyp = destTyp.Elem()

		// if struct and tag
		if destTyp.Kind() == reflect.Struct {
			processedFields := 0
			for x := 0; x < destTyp.NumField(); x++ {
				if srcIndex+processedFields >= len(resultSet.currentRow) {
					continue
				}
				field := destTyp.Field(x)
				name, _, _, _ := extractTag(field.Tag.Get("db"))
				if len(name) == 0 {
					continue
				}
				colInfo := (*resultSet.cols)[srcIndex+processedFields]
				if !strings.EqualFold(colInfo.Name, name) {
					continue
				}
				err := resultSet.setObjectValue(reflect.ValueOf(dest[destIndex]).Elem().Field(x), srcIndex+processedFields)
				// err := setFieldValue(reflect.ValueOf(dest[destIndex]).Elem().Field(x), colInfo.cusType, dataSet.currentRow[srcIndex+processedFields])
				if err != nil {
					return err
				}
				processedFields++
			}
			if processedFields != 0 {
				srcIndex = srcIndex + processedFields - 1
				continue
			}
		}
		// else
		err := resultSet.setObjectValue(reflect.ValueOf(dest[destIndex]).Elem(), srcIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

// Next_ act like Next in sql package return false if no other rows in dataset
func (resultSet *ResultSet) Next_() bool {
	err := resultSet.Next(resultSet.currentRow)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false
		}
		resultSet.lastErr = err
		return false
	}
	return true
}

// Next implement method need for sql.Rows interface
func (resultSet *ResultSet) Next(dest []driver.Value) error {
	hasMoreRows := resultSet.parent.hasMoreRows()
	noOfRowsToFetch := len(resultSet.rows) // dataSet.parent.noOfRowsToFetch()
	// if noOfRowsToFetch == 0 {
	// 	return io.EOF
	// }
	hasBLOB := resultSet.parent.hasBLOB()
	hasLONG := resultSet.parent.hasLONG()
	if !hasMoreRows && noOfRowsToFetch == 0 {
		return io.EOF
	}
	if hasMoreRows && (hasBLOB || hasLONG) && resultSet.index == 0 {
		// dataSet.rows = make([]Row, 0, dataSet.parent.noOfRowsToFetch())
		if err := resultSet.parent.fetch(resultSet); err != nil {
			return err
		}
		noOfRowsToFetch = len(resultSet.rows)
		hasMoreRows = resultSet.parent.hasMoreRows()
		if !hasMoreRows && noOfRowsToFetch == 0 {
			return io.EOF
		}
	}
	if resultSet.index > 0 && resultSet.index%len(resultSet.rows) == 0 {
		if hasMoreRows {
			resultSet.rows = make([]Row, 0, resultSet.parent.noOfRowsToFetch())
			err := resultSet.parent.fetch(resultSet)
			if err != nil {
				return err
			}
			noOfRowsToFetch = len(resultSet.rows)
			hasMoreRows = resultSet.parent.hasMoreRows()
			resultSet.index = 0
			if !hasMoreRows && noOfRowsToFetch == 0 {
				return io.EOF
			}
		} else {
			return io.EOF
		}
	}

	if noOfRowsToFetch > 0 && resultSet.index%noOfRowsToFetch < len(resultSet.rows) {
		length := len(resultSet.rows[resultSet.index%noOfRowsToFetch])
		if len(dest) < length {
			length = len(dest)
		}
		for x := 0; x < length; x++ {
			dest[x] = resultSet.rows[resultSet.index%noOfRowsToFetch][x]
		}
		resultSet.index++
		return nil
	}
	return io.EOF
}
