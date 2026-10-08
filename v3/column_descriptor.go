package go_ora

import (
	"math"
	"reflect"
	"strings"

	"github.com/sijms/go-ora/v3/configurations"
	types "github.com/sijms/go-ora/v3/types"
)

// columnDescriptor is an immutable snapshot of a result-set column's metadata.
// It is built once per result set, after the column definitions are loaded,
// and fully copies every value needed by the database/sql ColumnType* methods.
// It retains no reference to ParameterInfo or Connection, so the reported
// metadata stays stable even when writeDefine mutates bind/define state or
// connection LOB options change later.
type columnDescriptor struct {
	name             string
	schemaName       string
	typeName         string
	dataType         uint16 // semantic type (originalDataType when staged)
	wireDataType     uint16 // DataType observed at build time
	charsetForm      int
	charsetID        int
	maxLen           int64
	maxCharLen       int64
	isXMLType        bool
	isJSON           bool
	lobFetch         configurations.LobFetch
	lobFetchKnown    bool
	lobReadMode      configurations.LobReadMode
	lobReadModeKnown bool
	vectorDim        int
	vectorFormat     uint8
	vectorFlag       uint8
	vectorType       types.VectorType
	udtScanType      reflect.Type

	// normalized outputs read by the ColumnType* methods
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

type columnLengthKind uint8

const (
	columnLengthNone columnLengthKind = iota
	columnLengthChars
	columnLengthBytes
	columnLengthUnbounded
)

type columnTypeBase struct {
	databaseTypeName string
	lengthKind       columnLengthKind
	scanType         reflect.Type
}

var columnTypeBaseByDataType = map[uint16]columnTypeBase{
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

// newColumnDescriptor snapshots the metadata of a result-set column. The
// input is a one-time read of ParameterInfo state (ordinary fields plus the
// staging captured by load()/captureColumnMetadata); the returned descriptor
// holds copies only and never retains a *ParameterInfo or *Connection.
func newColumnDescriptor(col *ParameterInfo) columnDescriptor {
	if col == nil {
		return columnDescriptor{}
	}
	desc := columnDescriptor{
		name:             col.Name,
		schemaName:       col.SchemaName,
		typeName:         col.TypeName,
		wireDataType:     col.DataType,
		charsetForm:      col.CharsetForm,
		charsetID:        col.CharsetID,
		maxLen:           col.MaxLen,
		maxCharLen:       col.MaxCharLen,
		isXMLType:        col.IsXmlType,
		isJSON:           col.IsJson,
		lobFetch:         col.lobFetch,
		lobFetchKnown:    col.lobFetchKnown,
		lobReadMode:      col.lobReadMode,
		lobReadModeKnown: col.lobReadModeKnown,
		vectorDim:        col.VectorDim,
		vectorFormat:     col.VectorFormat,
		vectorFlag:       col.VectorFlag,
		vectorType:       col.VectorType,
		udtScanType:      col.udtScanType,
		nullable:         col.AllowNull,
		nullabilityKnown: col.nullabilityKnown,
	}
	desc.dataType = desc.wireDataType
	if col.originalDataTypeKnown {
		desc.dataType = col.originalDataType
	}

	base := columnTypeBaseByDataType[desc.dataType]
	desc.databaseTypeName = base.databaseTypeName
	desc.scanType = base.scanType

	switch desc.dataType {
	case types.CHAR, types.CHARZ:
		if desc.charsetForm == 2 {
			desc.databaseTypeName = "NCHAR"
		}
	case types.NCHAR, types.VARCHAR:
		if desc.charsetForm == 2 {
			desc.databaseTypeName = "NVARCHAR2"
		}
	case types.OCIClobLocator:
		if desc.charsetForm == 2 {
			desc.databaseTypeName = "NCLOB"
		}
	}
	if desc.isUDT() {
		desc.databaseTypeName = strings.ToUpper(desc.typeName)
		desc.scanType = desc.udtScanType
	} else if desc.isXMLType || strings.EqualFold(desc.typeName, "XMLTYPE") {
		desc.databaseTypeName = "XMLTYPE"
	}
	if desc.isJSON {
		desc.databaseTypeName = "JSON"
	}

	switch base.lengthKind {
	case columnLengthChars:
		if desc.maxCharLen > 0 {
			desc.length = desc.maxCharLen
			desc.lengthKnown = true
		} else if desc.maxLen > 0 {
			desc.length = desc.maxLen
			desc.lengthKnown = true
		}
	case columnLengthBytes:
		if desc.maxLen > 0 {
			desc.length = desc.maxLen
			desc.lengthKnown = true
		}
	case columnLengthUnbounded:
		desc.length = math.MaxInt64
		desc.lengthKnown = true
	}
	if desc.isJSON {
		desc.length = math.MaxInt64
		desc.lengthKnown = true
	}
	if desc.isUDT() {
		desc.length = 0
		desc.lengthKnown = false
	}

	if desc.dataType == types.NUMBER && col.precisionScaleKnown {
		if col.unconstrainedNumber {
			desc.precision = math.MaxInt64
			desc.scale = math.MaxInt64
		} else {
			desc.precision = col.rawPrecision
			desc.scale = col.rawScale
		}
		desc.precisionScaleOK = true
	}

	desc.resolveScanType()
	return desc
}

func (desc *columnDescriptor) isUDT() bool {
	return len(desc.typeName) > 0 && !strings.EqualFold(desc.typeName, "XMLTYPE") &&
		(desc.dataType == types.XMLType || desc.dataType == types.OCIXMLType)
}

// resolveScanType adjusts the base scan type for LOB-like columns whose
// runtime Go type depends on the LOB fetch/read options captured at build
// time. All inputs are descriptor copies; no live state is consulted.
func (desc *columnDescriptor) resolveScanType() {
	switch desc.dataType {
	case types.OCIClobLocator:
		if desc.inlineLob() || desc.automaticLob() {
			desc.scanType = types.TyString
		} else {
			desc.scanType = reflect.TypeOf((*types.Clob)(nil))
		}
	case types.OCIBlobLocator:
		if desc.inlineLob() || desc.automaticLob() {
			desc.scanType = types.TyBytes
		} else {
			desc.scanType = reflect.TypeOf((*types.Blob)(nil))
		}
	case types.JSON:
		if desc.inlineLob() {
			desc.scanType = types.TyBytes
		} else {
			desc.scanType = reflect.TypeOf((*types.Json)(nil))
		}
	case types.VECTOR:
		if desc.inlineLob() {
			desc.scanType = types.TyBytes
		} else if desc.automaticLob() {
			switch desc.vectorFormat {
			case 2:
				desc.scanType = reflect.TypeOf((*[]float32)(nil)).Elem()
			case 3:
				desc.scanType = reflect.TypeOf((*[]float64)(nil)).Elem()
			case 4:
				desc.scanType = types.TyBytes
			default:
				desc.scanType = nil
			}
		} else {
			desc.scanType = reflect.TypeOf((*types.Vector)(nil))
		}
	case types.OCIFileLocator:
		desc.scanType = reflect.TypeOf((*types.BFile)(nil))
	}
}

// inlineLob reports whether the column is decoded inline. When the staged LOB
// fetch option is absent (synthetic columns), it falls back to the wire type
// observed at build time, which reflects the post-writeDefine define type.
func (desc *columnDescriptor) inlineLob() bool {
	if desc.dataType == types.OCIFileLocator || desc.isJSON {
		return false
	}
	if desc.lobFetchKnown {
		return desc.lobFetch == configurations.INLINE
	}
	switch desc.wireDataType {
	case types.LongVarChar:
		return desc.dataType == types.OCIClobLocator
	case types.LongRaw:
		return desc.dataType == types.OCIBlobLocator || desc.dataType == types.VECTOR || desc.dataType == types.JSON
	}
	return false
}

func (desc *columnDescriptor) automaticLob() bool {
	return !desc.lobReadModeKnown || desc.lobReadMode == configurations.LobReadMode_AUTO
}
