package type_coder

import (
	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type LobCoder struct {
	streamer types.LobStreamer
	TypeInfo
}

func (coder *LobCoder) SetLobStreamer(lobStreamer types.LobStreamer) {
	coder.streamer = lobStreamer
}
func (coder *LobCoder) Write(session network.SessionWriter) error {
	if len(coder.BValue) > 0 {
		session.PutUint(len(coder.BValue), 4, true, true)
	}
	session.PutClr(coder.BValue)
	return nil
}

func (coder *LobCoder) read(session network.SessionReader) (bValue []byte, err error) {
	var locator []byte
	if coder.streamer.GetLobStreamMode() == configurations.INLINE {
		// the following code is working when the lob is inline (default)
		var maxSize int64

		maxSize, err = session.GetInt64(4, true, true)
		if err != nil {
			return
		}
		if maxSize > 0 {
			/*size*/ _, err = session.GetInt64(8, true, true)
			if err != nil {
				return
			}
			/*chunkSize*/ _, err = session.GetInt(4, true, true)
			if err != nil {
				return
			}
			bValue, err = session.GetClr()
			if err != nil {
				return
			}
			locator, err = session.GetClr()
			if err != nil {
				return
			}
			coder.streamer.SetLocator(locator)
		} else {
			// set value nil
			coder.streamer.SetLocator(nil)
			bValue = nil
		}
	} else {
		// the following code is working when the lob is not inline or part of UDT
		var temp []byte
		temp, err = coder.basicRead(session)
		if err != nil {
			return nil, err
		}
		if len(temp) == 0 {
			coder.streamer.SetLocator(nil)
			return nil, nil
		}
		if coder.IsUDTPar {
			locator = temp
		} else {
			locator, err = session.GetClr()
		}
		if err != nil {
			return
		}
		coder.streamer.SetLocator(locator)
	}
	return
	//return bValue, nil
}

//func (lob *baseLob) write(session network.SessionWriter, data []byte, quasiLocator bool) error {
//	var err error
//	if len(data) > 0 {
//		if quasiLocator {
//			lob.locator = lob.createQuasiLocator(uint64(len(data)))
//			return nil
//		}
//		if lob.streamer == nil {
//			return errors.New("lob streaming object should be set before write")
//		}
//		// create temporary lob
//		lob.locator, err = lob.streamer.CreateTemporaryLocator()
//		if err != nil {
//			return err
//		}
//		// upload data to the server
//		err = lob.streamer.Write(context.Background(), data)
//		if err != nil {
//			return err
//		}
//	} else {
//		lob.locator = nil
//	}
//	return nil
//}
