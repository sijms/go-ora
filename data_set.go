package go_ora

import (
	"database/sql/driver"
	"github.com/sijms/go-ora/network"
	"io"
)

type Row []driver.Value

type DataSet struct {
	ColumnCount     int
	RowCount        int
	UACBufferLength int
	MaxRowSize      int
	Cols            []ParameterInfo
	Rows            []Row
	index           int
	parent          *Stmt
}

func (dataSet *DataSet) read(session *network.Session) error {
	_, err := session.GetByte()
	if err != nil {
		return err
	}
	dataSet.ColumnCount, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	num, err := session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	dataSet.ColumnCount += num * 0x100
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
		for x := 0; x < index; x++ {
			for i := 0; i < 8; i++ {
				if x*8+i < dataSet.ColumnCount {
					dataSet.Cols[(x*8)+i].getDataFromServer = bitVector[x]>>i&1 > 0
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
	if dataSet.parent.hasMoreRows && dataSet.index > 0 && dataSet.index%dataSet.parent.noOfRowsToFetch == 0 {
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
