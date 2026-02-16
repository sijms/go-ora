package types

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"errors"

	"github.com/sijms/go-ora/v3/converters"
)

type BFile struct {
	dirName  string
	fileName string
	Valid    bool
	isOpened bool
	Locator  []byte
	lobBase
}

func CreateNullBFile() *BFile {
	return &BFile{
		Valid: false,
	}
}

func CreateBFile(db *sql.DB, dirName, fileName string) (*BFile, error) {
	output := &BFile{
		fileName: fileName,
		dirName:  dirName,
		Valid:    true,
	}
	_, err := db.Exec("SELECT :1 FROM DUAL", output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func CreateBFile2(coder converters.StringCoder, dirName, fileName string) (*BFile, error) {
	output := &BFile{
		fileName: fileName,
		dirName:  dirName,
		Valid:    true,
		//TypeInfo: type_coder.TypeInfo{
		//	DataType: OCIFileLocator,
		//	MaxLen:   4000,
		//},
	}
	strConv, err := coder.GetDefaultStringCoder()
	if err != nil {
		return nil, err
	}
	err = output.Init(strConv)
	return output, err
}

//
//func CreateBFileFromStream(stream LobStreamer, dirName, fileName string) *BFile {
//	return &BFile{
//		stream:   stream,
//		dirName:  dirName,
//		fileName: fileName,
//		isOpened: false,
//		Valid:    len(stream.GetLocator()) > 0,
//	}
//}

func (file *BFile) Init(strConv converters.IStringConverter) error {
	if file.Valid {
		dirName := strConv.Encode(file.dirName)
		fileName := strConv.Encode(file.fileName)
		totalLen := 16 + len(dirName) + len(fileName) + 4
		locatorBuffer := new(bytes.Buffer)
		err := binary.Write(locatorBuffer, binary.BigEndian, uint16(totalLen-2))
		if err != nil {
			return err
		}
		locatorBuffer.Write([]byte{0, 1, 8, 8, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0})
		err = binary.Write(locatorBuffer, binary.BigEndian, uint16(len(dirName)))
		if err != nil {
			return err
		}
		if len(dirName) > 0 {
			locatorBuffer.Write(dirName)
		}
		err = binary.Write(locatorBuffer, binary.BigEndian, uint16(len(fileName)))
		if err != nil {
			return err
		}
		if len(fileName) > 0 {
			locatorBuffer.Write(fileName)
		}
		file.Locator = locatorBuffer.Bytes()
	}
	return nil
}

func (file *BFile) GetDirName() string {
	return file.dirName
}

func (file *BFile) GetFileName() string {
	return file.fileName
}

func (file *BFile) IsOpen() bool {
	return file.isOpened
}

func (file *BFile) IsInit() bool {
	return len(file.Locator) > 0
}

func (file *BFile) Open(ctx context.Context) error {
	if !file.IsInit() {
		return errors.New("BFile is not initialized")
	}
	if file.isOpened {
		return nil
	}
	err := file.stream.Open(0xB, 0x100)
	if err != nil {
		return err
	}
	file.isOpened = true
	return nil
}

func (file *BFile) Close() error {
	if !file.IsInit() {
		return errors.New("BFile is not initialized")
	}
	if !file.isOpened {
		return nil
	}
	err := file.stream.Close(0x200)
	if err != nil {
		return err
	}
	file.isOpened = false
	return nil
}

func (file *BFile) Exists() (bool, error) {
	if !file.isOpened {
		return false, errors.New("invalid operation on closed object")
	}
	if !file.IsInit() {
		return false, errors.New("BFile is not initialized")
	}
	return file.stream.Exists()
}

func (file *BFile) Scan(value interface{}) error {
	if value == nil {
		file.Valid = false
		file.fileName = ""
		file.dirName = ""
		file.Locator = nil
		return nil
	}
	switch temp := value.(type) {
	case *BFile:
		*file = *temp
	case BFile:
		*file = temp
	default:
		return errors.New("BFILE column type require BFile value")
	}
	return nil
}

//
//func (file *BFile) Encode() ([]byte, error) {
//	return nil, nil
//}
//func (file *BFile) Decode(data []byte, _ uint16) (interface{}, error) {
//	return nil, nil
//}
//
//func (file *BFile) Read(session network.SessionReader, tnsType uint16, isUDTPar bool) error {
//	return nil
//}
//func (file *BFile) Write(session network.SessionWriter) error {
//	return nil
//}
