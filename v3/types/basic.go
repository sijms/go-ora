package types

type Basic struct {
	bValue   []byte
	dataType uint16
}

func (b *Basic) Bytes() []byte {
	return b.bValue
}

func (b *Basic) SetBytes(input []byte) {
	b.bValue = input
}

func (b *Basic) SetDataType(dt uint16) {
	if b.dataType == 0 {
		b.dataType = dt
	}
}

func (b *Basic) GetDataType() uint16 {
	return b.dataType
}

func (b *Basic) GetMaxLen() int64 { return 1 }
