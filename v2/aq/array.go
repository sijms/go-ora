package aq

import "github.com/sijms/go-ora/v2/network"

type VisibilityMode int

const (
	VisibilityImmediate VisibilityMode = 1
	VisibilityOnCommit  VisibilityMode = 2
)

type AQ struct {
	name       string
	version    int
	flag       int
	visibility VisibilityMode
}
type QAExtension struct {
	key   []byte
	value []byte
	num   uint8
}

type QAEnqueue struct {
	opt int
}
type AQArray struct {
	opt     int
	flag    int
	xiters  int
	qName   string
	data    []AQArrayInParam
	context *AQ
	message *Message
}

func (array *AQArray) Write(session *network.Session) error {
	if len(array.data) > 0 {
		session.PutBytes(1)
		session.PutInt(len(array.data), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutUint(array.flag, 4, true, true)
	if array.opt == 1 {
		session.PutBytes(1, 0)
	} else {
		session.PutBytes(1, 1)
	}
	session.PutInt(array.opt, 4, true, true)
	if array.opt == 1 {
		session.PutBytes(1)
	} else {
		session.PutBytes(0)
	}
	if session.TTCVersion >= 16 {
		session.PutUint(0xFFFF, 4, true, true)
	}
	if array.opt == 1 {
		session.PutUint(array.xiters, 4, true, true)
	}
	if len(array.data) > 0 {
		if array.opt == 1 {
			// Marshal Propagation
			session.PutBytes(0)

			// Marshal Header
			session.PutBytes(6)
			temp := session.StrConv.Encode(array.qName)
			if len(temp) > 0 {
				session.PutInt(len(temp), 2, true, true)
				session.PutBytes(temp...)
			} else {
				session.PutBytes(0)
			}
			session.PutBytes(array.message.messageID...)
			session.PutUint(array.context.version, 2, true, true)
			session.PutUint(array.context.flag, 4, true, true)
			session.PutInt(array.context.visibility, 4, true, true)
			session.PutBytes(0, 0)
			if len(array.message.extensions) > 0 {
				session.PutInt(len(array.message.extensions), 2, true, true)
				for _, extension := range array.message.extensions {
					session.PutKeyVal(extension.key, extension.value, extension.num)
				}
			} else {
				session.PutBytes(0)
			}
			session.PutBytes(0)

			// Marshal all parameters
			//for _, item := range array.data {
			//	item.MarshalData()
			//}

			// Marshal Done
			session.PutBytes(9)
		} else {
			// Marshal all parameters
			//for _, item := range array.data {
			//item.MarshalData()
			//}
		}
	}
	return session.Write()
}
