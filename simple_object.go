package go_ora

import (
	"errors"
	"fmt"
	"github.com/sijms/go-ora/network"
)

type simpleObject struct {
	session     *network.Session
	operationID uint8
	data        []byte
	err         error
}

func (obj *simpleObject) write() *simpleObject {
	//obj.session.ResetBuffer()
	obj.session.PutBytes(3, obj.operationID, 0)
	if obj.data != nil {
		obj.session.PutBytes(obj.data...)
	}
	obj.err = obj.session.Write()
	return obj
}

func (obj *simpleObject) read() error {
	if obj.err != nil {
		return obj.err
	}
	loop := true
	for loop {
		msg, err := obj.session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 4:
			obj.session.Summary, err = network.NewSummary(obj.session)
			if err != nil {
				return err
			}
			loop = false
		case 9:
			if obj.session.HasEOSCapability {
				if obj.session.Summary == nil {
					obj.session.Summary = new(network.SummaryObject)
				}
				obj.session.Summary.EndOfCallStatus, err = obj.session.GetInt(4, true, true)
				if err != nil {
					return err
				}
			}
			if obj.session.HasFSAPCapability {
				if obj.session.Summary == nil {
					obj.session.Summary = new(network.SummaryObject)
				}
				obj.session.Summary.EndToEndECIDSequence, err = obj.session.GetInt(2, true, true)
				if err != nil {
					return err
				}
			}
			loop = false
		default:
			return errors.New(fmt.Sprintf("message code error: received code %d and expected code is 4, 9", msg))
		}
	}
	if obj.session.HasError() {
		return errors.New(obj.session.GetError())
	}
	return nil
}
