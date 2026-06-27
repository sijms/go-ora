package parameter_coder

import (
	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type lobParameter struct {
	BasicParameter
	streamer types.LobStreamer
}

func (param *lobParameter) SetLobStreamer(streamer types.LobStreamer) {
	param.streamer = streamer
}

func (param *lobParameter) Write(session network.SessionWriter) error {
	if len(param.BValue) > 0 {
		session.PutUint(len(param.BValue), 4, true, true)
	}
	session.PutClr(param.BValue)
	return nil
}
func (param *lobParameter) read(session network.SessionReader) error {
	var locator types.Locator
	var err error
	if param.streamer.GetLobStreamMode() == configurations.INLINE {
		// the following code is working when the lob is inline (default)
		var maxSize int64
		maxSize, err = session.GetInt64(4, true, true)
		if err != nil {
			return err
		}
		if maxSize > 0 {
			/*size*/ _, err = session.GetInt64(8, true, true)
			if err != nil {
				return err
			}
			/*chunkSize*/ _, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			param.BValue, err = session.GetClr()
			if err != nil {
				return err
			}
			locator, err = session.GetClr()
			if err != nil {
				return err
			}
		} else {
			// set value nil
			locator = nil
			param.BValue = nil
		}
	} else {
		// the following code is working when the lob is not inline or part of UDT
		locator, err = param.BasicRead(session)
		if err != nil {
			return err
		}
		if len(locator) == 0 {
			locator = nil
			param.BValue = nil
		} else {
			if !param.IsUDTPar {
				locator, err = session.GetClr()
				if err != nil {
					return err
				}
			}
		}
	}
	param.streamer.SetLocator(locator)
	return nil
}
