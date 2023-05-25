package go_ora

import (
	"encoding/hex"
	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
	"github.com/sijms/go-ora/v2/trace"
	"log"
	"strings"
	"testing"
)

var buffer = `00000000  00 00 01 93 06 00 00 00  00 00 10 17 30 77 d7 6c  |............0w.l|
00000010  cb 70 33 79 b1 a3 44 98  cb ad 02 fb 78 7a 09 0b  |.p3y..D.....xz..|
00000020  05 1c 2e 02 0f a0 01 01  5c d0 00 00 00 02 0f a0  |........\.......|
00000030  00 00 00 00 00 00 00 00  01 0a 01 0a 0a 43 4f 4c  |.............COL|
00000040  5f 55 52 4f 57 49 44 00  00 00 00 01 07 07 78 7a  |_UROWID.......xz|
00000050  09 0b 05 26 04 00 02 1f  e8 01 0a 01 0a 00 06 22  |...&..........."|
00000060  01 01 00 02 0f a0 00 00  00 07 01 0d 0d 01 00 00  |................|
00000070  00 7a 00 01 00 00 00 01  00 0a 07 01 0d 0d 01 00  |.z..............|
00000080  00 00 7a 00 01 00 00 00  01 00 01 15 00 00 07 15  |..z.............|
00000090  01 01 01 07 00 15 01 01  01 07 01 0d 0d 01 00 00  |................|
000000a0  00 7a 00 01 00 00 00 01  00 02 15 01 01 01 07 01  |.z..............|
000000b0  0d 0d 01 00 00 00 7a 00  01 00 00 00 01 00 03 15  |......z.........|
000000c0  01 01 01 07 01 0d 0d 01  00 00 00 7a 00 01 00 00  |...........z....|
000000d0  00 01 00 04 15 01 01 01  07 01 0d 0d 01 00 00 00  |................|
000000e0  7a 00 01 00 00 00 01 00  05 15 01 01 01 07 01 0d  |z...............|
000000f0  0d 01 00 00 00 7a 00 01  00 00 00 01 00 06 15 01  |.....z..........|
00000100  01 01 07 01 0d 0d 01 00  00 00 7a 00 01 00 00 00  |..........z.....|
00000110  01 00 07 15 01 01 01 07  01 0d 0d 01 00 00 00 7a  |...............z|
00000120  00 01 00 00 00 01 00 08  15 01 01 01 07 01 0d 0d  |................|
00000130  01 00 00 00 7a 00 01 00  00 00 01 00 09 08 01 06  |....z...........|
00000140  04 02 4f 3c 13 00 01 02  00 00 00 00 00 00 04 01  |..O<............|
00000150  01 02 07 b1 01 0c 02 05  7b 00 00 01 02 00 03 00  |........{.......|
00000160  01 20 00 00 00 00 00 00  00 00 00 00 00 00 01 01  |. ..............|
00000170  00 00 00 00 02 05 7b 01  0c 19 4f 52 41 2d 30 31  |......{...ORA-01|
00000180  34 30 33 3a 20 6e 6f 20  64 61 74 61 20 66 6f 75  |403: no data fou|
00000190  6e 64 0a                                          |nd.|`

func TestURowid(t *testing.T) {
	temp := extractBuffer(buffer)
	conn := new(Connection)
	conn.connOption = &network.ConnectionOption{}
	conn.connOption.Tracer = trace.NilTracer()
	conn.sStrConv = converters.NewStringConverter(871)
	conn.session = network.NewSessionWithInputBufferForDebug(temp[10:])
	conn.session.TTCVersion = 11
	conn.session.StrConv = conn.sStrConv
	stmt := new(Stmt)
	stmt.connection = conn
	stmt.scnForSnapshot = make([]int, 2)
	stmt._hasBLOB = false
	stmt._hasLONG = false
	stmt.disableCompression = false
	stmt.arrayBindCount = 0
	dataSet := new(DataSet)
	err := stmt.read(dataSet)
	if err != nil {
		t.Error(err)
	}
}

func extractBuffer(input string) []byte {
	buffer := make([]byte, 0, 500)
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		start := strings.Index(line, " ")
		end := strings.Index(line, "|")
		if start > 0 && end > 0 {
			words := strings.Split(strings.TrimSpace(line[start:end]), " ")
			if len(words) > 2 {
				temp, err := hex.DecodeString(strings.Join(words, ""))
				if err != nil {
					log.Fatalln(err)
				}
				buffer = append(buffer, temp...)
			}
		}
	}
	return buffer
}
