package type_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type blobCoder struct {
	//data *types.blob
	LobCoder
}

//func NewBlobCoder(db *sql.DB, ctx context.Context, data []byte) (OracleTypeEncoder, error) {
//	temp, err := types.CreateBlob(db, ctx, data)
//	if err != nil {
//		return nil, err
//	}
//	ret := &BlobCoder{}
//	ret.bValue = temp.GetLocator()
//	return ret, nil
//}

func NewBlobEncoder(blob types.Blob) OracleTypeEncoder {
	ret := &blobCoder{}
	ret.BValue = blob.GetLocator()
	ret.TypeInfo.SetDefault()
	ret.DataType = types.OCIBlobLocator
	return ret
}

func NewBlobDecoder() OracleTypeDecoder {
	return &blobCoder{}
}

//func NewBlobCoder(blob types.Blob) *BlobCoder {
//	ret := &BlobCoder{}
//	ret.bValue, _ = blob.GetLocators()
//	return ret
//}

func (coder *blobCoder) Encode(_ converters.StringCoder, lobStream types.LobStreamer) error {
	//coder.TypeInfo.SetDefault()
	//coder.DataType = types.OCIBlobLocator
	return nil
}

func (coder *blobCoder) Decode(data []byte) (interface{}, error) {
	if coder.streamer.GetLocator() == nil {
		return nil, nil
	}
	return types.NewBlob(coder.streamer, data)
}

func (coder *blobCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.read(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}
