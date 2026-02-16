package type_coder

import "github.com/sijms/go-ora/v3/types"

type FileCoder struct {
	TypeInfo
}

func NewFileCoder(file *types.BFile) *FileCoder {
	ret := &FileCoder{}
	ret.DataType = types.OCIFileLocator
	ret.MaxLen = 4000
	return ret
}
