package go_ora

import (
	"errors"
	"fmt"
	"github.com/sijms/go-ora/network"
)

type DBVersion struct {
	Info               string
	Text               string
	Number             uint16
	MajorVersion       int
	MinorVersion       int
	PatchsetVersion    int
	isDb10gR20OrHigher bool
	isDb11gR10OrHigher bool
}

func GetDBVersion(session *network.Session) (*DBVersion, error) {
	session.ResetBuffer()
	session.PutBytes(3, 0x3B, 0)
	session.PutUint(1, 1, false, false)
	session.PutUint(0x100, 2, true, true)
	session.PutUint(1, 1, false, false)
	session.PutUint(1, 1, false, false)
	err := session.Write()
	if err != nil {
		return nil, err
	}
	msg, err := session.GetInt(1, false, false)
	if msg != 8 {
		return nil, errors.New(fmt.Sprintf("message code error: received code %d and expected code is 8", msg))
	}
	length, err := session.GetInt(2, true, true)
	if err != nil {
		return nil, err
	}
	info, err := session.GetBytes(int(length))
	if err != nil {
		return nil, err
	}
	number, err := session.GetInt(4, true, true)
	if err != nil {
		return nil, err
	}
	version := (number>>24&0xFF)*1000 + (number>>20&0xF)*100 + (number>>12&0xF)*10 + (number >> 8 & 0xF)
	text := fmt.Sprintf("%d.%d.%d.%d.%d", number>>24&0xFF, number>>20&0xF,
		number>>12&0xF, number>>8&0xF, number&0xFF)

	ret := &DBVersion{
		Info:            string(info),
		Text:            text,
		Number:          uint16(version),
		MajorVersion:    int(number >> 24 & 0xFF),
		MinorVersion:    int(number >> 20 & 0xF),
		PatchsetVersion: int(number >> 8 & 0xF),
	}
	if ret.MajorVersion > 10 || (ret.MajorVersion == 10 && ret.MinorVersion >= 2) {
		ret.isDb10gR20OrHigher = true
	}
	if ret.MajorVersion > 11 || (ret.MajorVersion == 11 && ret.MinorVersion >= 1) {
		ret.isDb11gR10OrHigher = true
	}
	return ret, nil
}
