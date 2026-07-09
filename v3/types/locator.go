package types

import "encoding/binary"

type Locator []byte

func NewQuasiLocator(dataLen uint64) Locator {
	var ret Locator = make([]byte, 40)
	ret[1] = 38
	ret[3] = 4
	ret[4] = 97
	ret[5] = 8
	ret[9] = 1
	binary.BigEndian.PutUint64(ret[10:], dataLen)
	return ret
}
func (loc Locator) IsQuasi() bool {
	return len(loc) > 3 && loc[3] == 4
}
func (loc Locator) IsValueBased() bool {
	return len(loc) > 4 && loc[4]&0x20 == 0x20
}
func (loc Locator) IsTemporary() bool {
	return len(loc) > 7 && loc[7]&1 == 1 || len(loc) > 4 && loc[4]&0x40 == 0x40 || loc.IsValueBased()
}
func (loc Locator) IsVarWidthChar() bool {
	return len(loc) > 6 && loc[6]&0x80 == 0x80
}

func (loc Locator) IsLittleEndian() bool {
	return len(loc) > 7 && loc[7]&0x40 == 0x40
}

func (loc Locator) IsReadOnly() bool {
	return len(loc) > 6 && loc[6]&1 == 1
}
