package go_ora

import (
	"database/sql/driver"
	"errors"
	"github.com/sijms/go-ora/v2/network"
)

type simpleObject struct {
	connection *Connection
	//session     *network.Session
	operationID uint8
	data        []byte
	err         error
}

func (obj *simpleObject) write() *simpleObject {
	//obj.session.ResetBuffer()
	session := obj.connection.session
	session.PutBytes(3, obj.operationID, 0)
	if obj.data != nil {
		session.PutBytes(obj.data...)
	}
	obj.err = session.Write()
	return obj
}

func (obj *simpleObject) read() error {
	if obj.err != nil {
		return obj.err
	}
	return obj.connection.read()
	//loop := true
	//for loop {
	//	msg, err := session.GetByte()
	//	if err != nil {
	//		return err
	//	}
	//	err = obj.connection.readMsg(msg)
	//	if err != nil {
	//		return err
	//	}
	//	if msg == 4 || msg == 9 {
	//		loop = false
	//	}
	//}
	//if session.HasError() {
	//	return session.GetError()
	//}
	//return nil
}

func (obj *simpleObject) exec() error {
	conn := obj.connection
	tracer := conn.tracer
	obj.write()
	if obj.err != nil {
		return obj.err
	}
	err := conn.read()
	if errors.Is(err, network.ErrConnReset) {
		err = conn.read()
	}
	if err != nil {
		if isBadConn(err) {
			obj.connection.setBad()
			tracer.Print("Error: ", err)
			return driver.ErrBadConn
		}
		return err
	}
	return nil
	//var reconnect bool
	//for writeTrials := 0; writeTrials < failOver; writeTrials++ {
	//	reconnect, err = obj.connection.reConnect(err, writeTrials)
	//	if err != nil {
	//		tracer.Print("Error: ", err)
	//		if !reconnect {
	//			return err
	//		}
	//		continue
	//	}
	//	break
	//}
	//if reconnect {
	//	return &network.OracleError{ErrCode: 3135}
	//}
	//return err
}
