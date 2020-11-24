package go_ora

import (
	"database/sql/driver"
	"io"

	"github.com/sijms/go-ora/trace"

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
	currentRow      Row
	index           int
	parent          *Stmt
}

func (dataSet *DataSet) read(session *network.Session) error {
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
	if len(dataSet.currentRow) != dataSet.ColumnCount {
		dataSet.currentRow = make(Row, dataSet.ColumnCount)
	}
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
	if dataSet.parent.hasMoreRows && dataSet.index%dataSet.parent.noOfRowsToFetch == 0 {
		dataSet.Rows = make([]Row, 0, dataSet.parent.noOfRowsToFetch)
		err := dataSet.parent.fetch(dataSet)
		if err != nil {
			return err
		}
	}
	if dataSet.index%dataSet.parent.noOfRowsToFetch < len(dataSet.Rows) {
		for x := 0; x < len(dataSet.Rows[dataSet.index%dataSet.parent.noOfRowsToFetch]); x++ {
			dest[x] = driver.Value(dataSet.Rows[dataSet.index%dataSet.parent.noOfRowsToFetch][x])
		}
		dataSet.index++
		return nil
	}
	return io.EOF
}

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
