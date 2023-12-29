package go_ora

import (
	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
	"testing"
)

var inputBuffer = []byte{8, 1, 1, 1, 128, 0, 0, 1, 20, 0, 0, 0, 0, 2, 3, 105, 1, 1,
	20, 0, 1, 0, 1, 4, 4, 78, 65, 77, 69, 0, 0, 0, 0, 0, 1, 14, 2, 1, 144,
	2, 1, 144, 3, 1, 0, 0, 1, 5, 1, 1, 0, 0, 1, 32, 1, 22, 0, 0, 0, 0, 0, 4,
	1, 7, 1, 6, 0, 0, 0, 0, 1, 5, 1, 39, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func TestBulkCopyRead(t *testing.T) {

	session := network.NewSessionWithInputBufferForDebug(inputBuffer)
	session.TTCVersion = 12
	conn := &Connection{
		session: session,
	}
	conn.sStrConv = converters.NewStringConverter(873)
	session.StrConv = conn.sStrConv
	bulk := BulkCopy{
		conn:          conn,
		TableName:     "",
		SchemaName:    "",
		PartitionName: "",
		ColumnNames:   nil,
		columns:       nil,
		tableCursor:   0,
		sdbaBits:      0,
		dbaBits:       0,
	}
	err := bulk.readPrepareResponse()
	if err != nil {
		t.Error(err)
	}
}
