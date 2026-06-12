package parameter_coder

import (
	"fmt"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type RowIDParameter struct {
	rowid types.RowID
	BasicParameter
}

func (param *RowIDParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	coder := &types.RowID{}
	return coder.SetValue(input, 0)
}

func (param *RowIDParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := &types.RowID{}
	*decoder = param.rowid
	decoder.SetBytes(param.BValue)
	return decoder.Value(param.DataType)
}

func (param *RowIDParameter) Write(session network.SessionWriter) error {
	return fmt.Errorf("cannot pass rowid as an input parameter")
}

func (param *RowIDParameter) Read(session network.SessionReader) error {
	switch param.DataType {
	case types.ROWID:
		length, err := session.GetByte()
		if err != nil {
			return err
		}
		if length == 0 {
			return nil
		}
		param.rowid.Rba, err = session.GetInt64(4, true, true)
		if err != nil {
			return err
		}
		param.rowid.PartitionID, err = session.GetInt64(2, true, true)
		if err != nil {
			return err
		}
		num, err := session.GetByte()
		if err != nil {
			return err
		}
		param.rowid.BlockNumber, err = session.GetInt64(4, true, true)
		if err != nil {
			return err
		}
		param.rowid.SlotNumber, err = session.GetInt64(2, true, true)
		if err != nil {
			return err
		}
		if param.rowid.Rba == 0 && param.rowid.PartitionID == 0 && num == 0 && param.rowid.BlockNumber == 0 && param.rowid.SlotNumber == 0 {
			return nil
		}
		return nil
	case types.UROWID:
		length, err := session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		if length > 0 {
			param.BValue, err = session.GetClr()
			if err != nil {
				return err
			}
		} else {
			param.BValue = nil
		}
		return nil
	}
	return fmt.Errorf("ROWID decoder called with unsupported data type: %d", param.DataType)
}
