package go_ora

import (
	"database/sql/driver"
	"github.com/sijms/go-ora/trace"
	"io"

	"github.com/sijms/go-ora/network"
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
	ColumnCount     int
	RowCount        int
	UACBufferLength int
	MaxRowSize      int
	Cols            []ParameterInfo
	Rows            []Row
	//currentRow      Row
	index  int
	parent StmtInterface
}

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
	if columnCount > dataSet.ColumnCount {
		dataSet.ColumnCount = columnCount
	}
	//if len(dataSet.currentRow) != dataSet.ColumnCount {
	//	dataSet.currentRow = make(Row, dataSet.ColumnCount)
	//}
	dataSet.RowCount, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	dataSet.UACBufferLength, err = session.GetInt(2, true, true)
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

func (dataSet *DataSet) setBitVector(bitVector []byte) {
	index := dataSet.ColumnCount / 8
	if dataSet.ColumnCount%8 > 0 {
		index++
	}
	if len(bitVector) > 0 {
		for x := 0; x < len(bitVector); x++ {
			for i := 0; i < 8; i++ {
				if (x*8)+i < dataSet.ColumnCount {
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

func (dataSet *DataSet) Next(dest []driver.Value) error {
	//fmt.Println("has more row: ", dataSet.parent.hasMoreRows)
	//fmt.Println("row length: ", len(dataSet.Rows))
	//fmt.Println("cursor id: ", dataSet.parent.cursorID)
	//if dataSet.parent.hasMoreRows && dataSet.index == len(dataSet.Rows) && len(dataSet.Rows) < dataSet.parent.noOfRowsToFetch {
	//	fmt.Println("inside first fetch")
	//	oldFetchCount := dataSet.parent.noOfRowsToFetch;
	//	dataSet.parent.noOfRowsToFetch = oldFetchCount - len(dataSet.Rows)
	//	err := dataSet.parent.fetch(dataSet)
	//	if err != nil {
	//		return err
	//	}
	//	dataSet.parent.noOfRowsToFetch = oldFetchCount
	//	fmt.Println("row count after first fetch: ", len(dataSet.Rows))
	//}
	hasMoreRows := dataSet.parent.hasMoreRows()
	noOfRowsToFetch := len(dataSet.Rows) // dataSet.parent.noOfRowsToFetch()
	hasBLOB := dataSet.parent.hasBLOB()
	hasLONG := dataSet.parent.hasLONG()
	if !hasMoreRows && noOfRowsToFetch == 0 {
		return io.EOF
	}
	if dataSet.index > 0 && dataSet.index%len(dataSet.Rows) == 0 {
		if hasMoreRows {
			dataSet.Rows = make([]Row, 0, dataSet.parent.noOfRowsToFetch())
			err := dataSet.parent.fetch(dataSet)
			if err != nil {
				return err
			}
			noOfRowsToFetch = len(dataSet.Rows)
			hasMoreRows = dataSet.parent.hasMoreRows()
			dataSet.index = 0
			if !hasMoreRows && noOfRowsToFetch == 0 {
				return io.EOF
			}
		} else {
			return io.EOF
		}
	}
	//if hasMoreRows && dataSet.index != 0 && dataSet.index%noOfRowsToFetch == 0 {
	//
	//}
	if hasMoreRows && (hasBLOB || hasLONG) && dataSet.index == 0 {
		if err := dataSet.parent.fetch(dataSet); err != nil {
			return err
		}
	}
	if dataSet.index%noOfRowsToFetch < len(dataSet.Rows) {
		for x := 0; x < len(dataSet.Rows[dataSet.index%noOfRowsToFetch]); x++ {
			dest[x] = dataSet.Rows[dataSet.index%noOfRowsToFetch][x]
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
	for r, row := range dataSet.Rows {
		if r > 25 {
			break
		}
		t.Printf("Row %d", r)
		for c, col := range dataSet.Cols {
			t.Printf("  %-20s: %v", col.Name, row[c])
		}
	}
}

func (dataSet DataSet) ColumnTypeDatabaseTypeName(index int) string {
	return dataSet.Cols[index].DataType.String()
}

func (dataSet DataSet) ColumnTypeLength(index int) (length int64, ok bool) {
	switch dataSet.Cols[index].DataType {
	case NCHAR, CHAR:
		return int64(dataSet.Cols[index].MaxCharLen), true
	case NUMBER:
		return int64(dataSet.Cols[index].Precision), true
	}
	return int64(0), false

}

func (dataSet DataSet) ColumnTypeNullable(index int) (nullable, ok bool) {
	return dataSet.Cols[index].AllowNull, true
}
