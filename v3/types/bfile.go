package types

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/sijms/go-ora/v3/converters"
)

type BFile struct {
	Basic
	lobBase
	//dirName  string
	//fileName string
	Conv     converters.IStringConverter
	isOpened bool
}

func CreateBFile(db *sql.DB, dirName, fileName string) (*BFile, error) {
	var err error
	ret := &BFile{}
	//ret.dirName = dirName
	//ret.fileName = fileName
	err = ret.createStreamer(db)
	if err != nil {
		return nil, err
	}
	filePath := ""
	if len(dirName) != 0 && len(fileName) != 0 {
		filePath = dirName + "/" + fileName
	}
	return ret, ret.SetValue(filePath)
}
func (file *BFile) GetMaxLen() int64 {
	return MaxLenBFile
}
func (file *BFile) SetValue(input interface{}) error {
	if input == nil {
		file.bValue = nil
		return nil
	}
	var fileName, dirName []byte
	if file.Conv == nil {
		file.Conv = converters.NewStringConverter(0x7D0)
	}
	switch input := input.(type) {
	case BFile:
		*file = input
		return nil
	case *BFile:
		*file = *input
		return nil
	case string:
		if len(input) == 0 {
			file.bValue = nil
			return nil
		}
		index := strings.Index(input, "\\")
		if index < 0 {
			index = strings.Index(input, "/")
			if index < 0 {
				return fmt.Errorf("invalid file path: %s (should be dirname/filename)", input)
			}
		}
		dirName = file.Conv.Encode(input[:index])
		fileName = file.Conv.Encode(input[index+1:])
	case *string:
		if len(*input) == 0 {
			file.bValue = nil
			return nil
		}
		index := strings.Index(*input, "\\")
		if index < 0 {
			index = strings.Index(*input, "/")
			if index < 0 {
				return fmt.Errorf("invalid file path: %s (should be dirname/filename)", *input)
			}
		}
		dirName = file.Conv.Encode((*input)[:index])
		fileName = file.Conv.Encode((*input)[index+1:])
	default:
		return fmt.Errorf("cannot set value of type: %T into BFile", input)
	}
	totalLen := 16 + len(dirName) + len(fileName) + 4
	locatorBuffer := &bytes.Buffer{}
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
	file.bValue = locatorBuffer.Bytes()
	if file.stream != nil {
		file.stream.SetLocator(file.bValue)
	}
	return nil
}

//	func (file *BFile) filePath() string {
//		if len(file.dirName) == 0 || len(file.fileName) == 0 {
//			return ""
//		}
//		return file.dirName + "/" + file.fileName
//	}
func (file *BFile) SetStreamer(input LobStreamer) {
	file.stream = input
	if input != nil {
		// get dir and file name before change conversion
		dirName := file.GetDirName()
		fileName := file.GetFileName()
		file.Conv, _ = file.stream.GetStringCoder().GetDefaultStringCoder()
		filePath := ""
		if len(dirName) != 0 && len(fileName) != 0 {
			filePath = dirName + "/" + fileName
		}
		_ = file.SetValue(filePath)
	}
}
func (file *BFile) Value() (interface{}, error) {
	return file.bValue, nil
}

func (file *BFile) Scan(input interface{}) error {
	return file.SetValue(input)
}

func (file *BFile) CopyTo(dest driver.Value) error {
	if dst, ok := dest.(*[]byte); ok {
		*dst = file.bValue
		return nil
	}
	return fmt.Errorf("cannot copy BFile to type %T", dest)
}

func (file *BFile) IsOpen() bool {
	return file.isOpened
}

func (file *BFile) IsInit() bool {
	if file.stream != nil && file.stream.GetLocator() == nil && file.bValue != nil {
		file.stream.SetLocator(file.bValue)
	}
	return len(file.bValue) > 0 && file.stream != nil
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

func (file *BFile) GetDirName() string {
	dirName := ""
	locator := file.GetLocator()
	if len(locator) > 16 && file.Conv != nil {
		index := 16
		length := int(binary.BigEndian.Uint16(locator[index : index+2]))
		index += 2
		dirName = file.Conv.Decode(locator[index : index+length])
	}
	return dirName
}

func (file *BFile) GetFileName() string {
	fileName := ""
	locator := file.GetLocator()
	if len(locator) > 16 && file.Conv != nil {
		index := 16
		length := int(binary.BigEndian.Uint16(locator[index : index+2]))
		index += 2
		_ = file.Conv.Decode(locator[index : index+length])
		index += length
		length = int(binary.BigEndian.Uint16(locator[index : index+2]))
		index += 2
		fileName = file.Conv.Decode(locator[index : index+length])
	}
	return fileName
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
