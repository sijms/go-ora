package go_ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/trace"
	types "github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/utils"

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
	if len(bitVector) > 0 && resultSet.cols != nil {
		for x := 0; x < len(bitVector); x++ {
			for i := 0; i < 8; i++ {
				if (x*8)+i < resultSet.columnCount {
					(*resultSet.cols)[(x*8)+i].getDataFromServer = bitVector[x]&(1<<i) > 0
				}
			}
		}
	} else if resultSet.cols != nil {
		for x := 0; x < len(*resultSet.cols); x++ {
			(*resultSet.cols)[x].getDataFromServer = true
		}
	}
}

func (resultSet *ResultSet) Close() error {
	if resultSet == nil || resultSet.parent == nil {
		return nil
	}
	if resultSet.parent.CanAutoClose() {
		return resultSet.parent.Close()
	}
	return nil
}

func (resultSet *ResultSet) Columns() []string {
	if resultSet == nil || resultSet.cols == nil || len(*resultSet.cols) == 0 {
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

type columnLengthKind uint8

const (
	columnLengthNone columnLengthKind = iota
	columnLengthChars
	columnLengthBytes
	columnLengthUnbounded
)

type columnTypeMetadata struct {
	databaseTypeName string
	length           int64
	lengthKnown      bool
	precision        int64
	scale            int64
	precisionScaleOK bool
	nullable         bool
	nullabilityKnown bool
	scanType         reflect.Type
}

type columnTypeBase struct {
	databaseTypeName string
	lengthKind       columnLengthKind
	scanType         reflect.Type
}

var columnTypeMetadataByType = map[uint16]columnTypeBase{
	types.NCHAR:            {databaseTypeName: "VARCHAR2", lengthKind: columnLengthChars, scanType: types.TyString},
	types.NUMBER:           {databaseTypeName: "NUMBER", scanType: types.TyString},
	types.FLOAT:            {databaseTypeName: "FLOAT"},
	types.LONG:             {databaseTypeName: "LONG", lengthKind: columnLengthUnbounded, scanType: types.TyString},
	types.VARCHAR:          {databaseTypeName: "VARCHAR2", lengthKind: columnLengthChars},
	types.ROWID:            {databaseTypeName: "ROWID", scanType: types.TyString},
	types.DATE:             {databaseTypeName: "DATE", scanType: types.TyTime},
	types.VarRaw:           {databaseTypeName: "RAW", lengthKind: columnLengthBytes},
	types.BFLOAT:           {databaseTypeName: "BINARY_FLOAT"},
	types.BDOUBLE:          {databaseTypeName: "BINARY_DOUBLE"},
	types.RAW:              {databaseTypeName: "RAW", lengthKind: columnLengthBytes, scanType: types.TyBytes},
	types.LongRaw:          {databaseTypeName: "LONG RAW", lengthKind: columnLengthUnbounded, scanType: types.TyBytes},
	types.LongVarChar:      {databaseTypeName: "LONG", lengthKind: columnLengthUnbounded, scanType: types.TyString},
	types.LongVarRaw:       {databaseTypeName: "LONG RAW", lengthKind: columnLengthUnbounded},
	types.CHAR:             {databaseTypeName: "CHAR", scanType: types.TyString},
	types.CHARZ:            {databaseTypeName: "CHAR"},
	types.IBFLOAT:          {databaseTypeName: "BINARY_FLOAT", scanType: types.TyFloat32},
	types.IBDOUBLE:         {databaseTypeName: "BINARY_DOUBLE", scanType: types.TyFloat64},
	types.REFCURSOR:        {databaseTypeName: "REF CURSOR"},
	types.OCIXMLType:       {databaseTypeName: "XMLTYPE", lengthKind: columnLengthUnbounded},
	types.XMLType:          {databaseTypeName: "XMLTYPE", lengthKind: columnLengthUnbounded},
	types.OCIClobLocator:   {databaseTypeName: "CLOB", lengthKind: columnLengthUnbounded},
	types.OCIBlobLocator:   {databaseTypeName: "BLOB", lengthKind: columnLengthUnbounded},
	types.OCIFileLocator:   {databaseTypeName: "BFILE", lengthKind: columnLengthUnbounded},
	types.RESULTSET:        {databaseTypeName: "RESULT SET"},
	types.JSON:             {databaseTypeName: "JSON", lengthKind: columnLengthUnbounded, scanType: reflect.TypeOf((*types.Json)(nil))},
	types.VECTOR:           {databaseTypeName: "VECTOR", lengthKind: columnLengthUnbounded},
	types.TimeStampDTY:     {databaseTypeName: "TIMESTAMP", scanType: types.TyTime},
	types.TimeStampTZ_DTY:  {databaseTypeName: "TIMESTAMP WITH TIME ZONE", scanType: types.TyTime},
	types.INTERVALYM_DTY:   {databaseTypeName: "INTERVAL YEAR TO MONTH", scanType: types.TyTime},
	types.INTERVALDS_DTY:   {databaseTypeName: "INTERVAL DAY TO SECOND", scanType: types.TyTime},
	types.TimeTZ:           {databaseTypeName: "TIME WITH TIME ZONE"},
	types.TIMESTAMP:        {databaseTypeName: "TIMESTAMP", scanType: types.TyTime},
	types.TIMESTAMPTZ:      {databaseTypeName: "TIMESTAMP WITH TIME ZONE", scanType: types.TyTime},
	types.IntervalYM:       {databaseTypeName: "INTERVAL YEAR TO MONTH"},
	types.IntervalDS:       {databaseTypeName: "INTERVAL DAY TO SECOND"},
	types.UROWID:           {databaseTypeName: "UROWID", lengthKind: columnLengthBytes, scanType: types.TyString},
	types.TimeStampLTZ_DTY: {databaseTypeName: "TIMESTAMP WITH LOCAL TIME ZONE", scanType: types.TyTime},
	types.TimeStampLTZ:     {databaseTypeName: "TIMESTAMP WITH LOCAL TIME ZONE", scanType: types.TyTime},
	types.BOOLEAN:          {databaseTypeName: "BOOLEAN", scanType: types.TyBool},
}

func (resultSet *ResultSet) columnMetadata(index int) columnTypeMetadata {
	if resultSet == nil || resultSet.cols == nil || index < 0 || index >= len(*resultSet.cols) {
		return columnTypeMetadata{}
	}
	return metadataForColumn(&(*resultSet.cols)[index])
}

func metadataForColumn(col *ParameterInfo) columnTypeMetadata {
	if col == nil {
		return columnTypeMetadata{}
	}
	dataType := col.DataType
	if col.originalDataTypeKnown {
		dataType = col.originalDataType
	}
	base := columnTypeMetadataByType[dataType]
	metadata := columnTypeMetadata{
		databaseTypeName: base.databaseTypeName,
		nullable:         col.AllowNull,
		nullabilityKnown: col.nullabilityKnown,
		scanType:         base.scanType,
	}

	switch dataType {
	case types.CHAR, types.CHARZ:
		if col.CharsetForm == 2 {
			metadata.databaseTypeName = "NCHAR"
		}
	case types.NCHAR, types.VARCHAR:
		if col.CharsetForm == 2 {
			metadata.databaseTypeName = "NVARCHAR2"
		}
	case types.OCIClobLocator:
		if col.CharsetForm == 2 {
			metadata.databaseTypeName = "NCLOB"
		}
	}
	if isUDTColumn(col, dataType) {
		metadata.databaseTypeName = strings.ToUpper(col.TypeName)
		metadata.scanType = col.udtScanType
	} else if col.IsXmlType || strings.EqualFold(col.TypeName, "XMLTYPE") {
		metadata.databaseTypeName = "XMLTYPE"
	}
	if col.IsJson {
		metadata.databaseTypeName = "JSON"
	}

	switch base.lengthKind {
	case columnLengthChars:
		if col.MaxCharLen > 0 {
			metadata.length = col.MaxCharLen
			metadata.lengthKnown = true
		} else if col.MaxLen > 0 {
			metadata.length = col.MaxLen
			metadata.lengthKnown = true
		}
	case columnLengthBytes:
		if col.MaxLen > 0 {
			metadata.length = col.MaxLen
			metadata.lengthKnown = true
		}
	case columnLengthUnbounded:
		metadata.length = math.MaxInt64
		metadata.lengthKnown = true
	}
	if col.IsJson {
		metadata.length = math.MaxInt64
		metadata.lengthKnown = true
	}
	if isUDTColumn(col, dataType) {
		metadata.length = 0
		metadata.lengthKnown = false
	}

	if dataType == types.NUMBER && col.precisionScaleKnown {
		if col.unconstrainedNumber {
			metadata.precision = math.MaxInt64
			metadata.scale = math.MaxInt64
		} else {
			metadata.precision = col.rawPrecision
			metadata.scale = col.rawScale
		}
		metadata.precisionScaleOK = true
	}

	metadata.scanType = metadataScanType(col, dataType, metadata.scanType)
	return metadata
}

func isUDTColumn(col *ParameterInfo, dataType uint16) bool {
	return len(col.TypeName) > 0 && !strings.EqualFold(col.TypeName, "XMLTYPE") &&
		(dataType == types.XMLType || dataType == types.OCIXMLType)
}

func metadataScanType(col *ParameterInfo, dataType uint16, scanType reflect.Type) reflect.Type {
	switch dataType {
	case types.OCIClobLocator:
		if inlineLobColumn(col, dataType) || automaticLobColumn(col) {
			return types.TyString
		}
		return reflect.TypeOf((*types.Clob)(nil))
	case types.OCIBlobLocator:
		if inlineLobColumn(col, dataType) || automaticLobColumn(col) {
			return types.TyBytes
		}
		return reflect.TypeOf((*types.Blob)(nil))
	case types.JSON:
		if inlineLobColumn(col, dataType) {
			return types.TyBytes
		}
		return reflect.TypeOf((*types.Json)(nil))
	case types.VECTOR:
		if inlineLobColumn(col, dataType) {
			return types.TyBytes
		}
		if automaticLobColumn(col) {
			switch col.VectorFormat {
			case 2:
				return reflect.TypeOf((*[]float32)(nil)).Elem()
			case 3:
				return reflect.TypeOf((*[]float64)(nil)).Elem()
			case 4:
				return types.TyBytes
			default:
				return nil
			}
		}
		return reflect.TypeOf((*types.Vector)(nil))
	case types.OCIFileLocator:
		return reflect.TypeOf((*types.BFile)(nil))
	default:
		return scanType
	}
}

func inlineLobColumn(col *ParameterInfo, dataType uint16) bool {
	if dataType == types.OCIFileLocator || col.IsJson {
		return false
	}
	if col.lobFetchKnown {
		return col.lobFetch == configurations.INLINE
	}
	switch col.DataType {
	case types.LongVarChar:
		return dataType == types.OCIClobLocator
	case types.LongRaw:
		return dataType == types.OCIBlobLocator || dataType == types.VECTOR || dataType == types.JSON
	}
	return false
}

func automaticLobColumn(col *ParameterInfo) bool {
	return !col.lobReadModeKnown || col.lobReadMode == configurations.LobReadMode_AUTO
}

// ColumnTypeDatabaseTypeName return Col DataType name
func (resultSet *ResultSet) ColumnTypeDatabaseTypeName(index int) string {
	return resultSet.columnMetadata(index).databaseTypeName
}

// ColumnTypeLength return length of column type
func (resultSet *ResultSet) ColumnTypeLength(index int) (int64, bool) {
	metadata := resultSet.columnMetadata(index)
	return metadata.length, metadata.lengthKnown
}

// ColumnTypeNullable return if column allow null or not
func (resultSet *ResultSet) ColumnTypeNullable(index int) (nullable, ok bool) {
	metadata := resultSet.columnMetadata(index)
	return metadata.nullable, metadata.nullabilityKnown
}

// ColumnTypePrecisionScale return the precision and scale for numeric types
func (resultSet *ResultSet) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	metadata := resultSet.columnMetadata(index)
	return metadata.precision, metadata.scale, metadata.precisionScaleOK
}

func (resultSet *ResultSet) ColumnTypeScanType(index int) reflect.Type {
	return resultSet.columnMetadata(index).scanType
}

func (resultSet *ResultSet) Err() error {
	return resultSet.lastErr
}

// set object value using currentRow[colIndex] return true if succeed or false
// for non-supported type
// error means error occur during operation
func (resultSet *ResultSet) setObjectValue(obj reflect.Value, colIndex int) error {
	//col := (*resultSet.cols)[colIndex]
	return types.RCopy(obj, resultSet.currentRow[colIndex])
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
				name, _, _, _ := utils.ExtractTag(field.Tag.Get("db"))
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
