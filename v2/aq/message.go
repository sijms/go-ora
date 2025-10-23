package aq

import (
	"time"

	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
)

var XMLTYPE_TOID = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 2, 1, 0}
var ANYDATA_TOID = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 2, 0, 17}
var JSON_TOID = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 71}
var RAW_TOID = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 23}

type OracleAgent struct {
	Key   string
	Value string
	Num   uint8
}
type DeliveryMode int

const (
	DeliveryModePersistent DeliveryMode = 1
	DeliveryModeBuffered   DeliveryMode = 2
	DeliveryModeAny        DeliveryMode = 3
)

type MessageType int

const (
	RAW  MessageType = 1
	UDT  MessageType = 2
	XML  MessageType = 3
	JSON MessageType = 4
)

type MessageState int

const (
	MessageStateReady MessageState = iota
	MessageStateWaiting
	MessageStateProcessed
	MessageStateExpired
)

type Message struct {
	ID             string
	messageID      []byte
	Delay          int
	Expiration     int
	Priority       int
	correlation    string
	exceptionQueue string
	deqAttempts    int
	State          MessageState
	DeliveryMode   DeliveryMode
	VisibilityMode VisibilityMode
	recipients     []OracleAgent
	sender         OracleAgent
	EnqTime        time.Time

	propModified   bool
	Payload        interface{}
	payloadInBytes []byte
	shareNum       int

	senderName       string
	senderAddress    string
	senderProtocol   uint8
	transactionGroup string

	extensions []QAExtension
}

func (message *Message) read(session *network.Session) error {
	var err error
	message.Priority, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	message.Delay, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	message.Expiration, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	var num int
	num, err = session.GetInt(4, true, true)
	if num > 0 {
		var temp []byte
		temp, err = session.GetClr()
		if err != nil {
			return err
		}
		message.correlation = session.StrConv.Decode(temp)
	} else {
		message.correlation = ""
	}
	message.deqAttempts, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	num, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	if num > 0 {
		var temp []byte
		temp, err = session.GetClr()
		if err != nil {
			return err
		}
		message.exceptionQueue = session.StrConv.Decode(temp)
	} else {
		message.exceptionQueue = ""
	}
	num, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	message.State = MessageState(num)
	num, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	if num > 0 {
		var temp []byte
		temp, err = session.GetClr()
		if err != nil {
			return err
		}
		message.EnqTime, err = converters.DecodeDate(temp)
		if err != nil {
			return err
		}
		// oracle date and timestamp

		//this.aqmsta = this.mEngine.UnmarshalSB4(false);
		//if (this.mEngine.UnmarshalSB4(false) > 0) {
		//	this.mEngine.UnmarshalCLR(this.aqmeqtBuffer, 0, this.retInt, 7, false);
		//	OracleDate oracleDate = new OracleDate(this.aqmeqtBuffer);
		//	this.aqmeqt = new OracleTimeStamp(oracleDate.Value);
		//}
	}
	if session.TTCVersion >= 3 {
		num, err = session.GetInt(2, true, true)
		if err != nil {
			return err
		}
		if num > 0 {
			var temp []byte
			temp, err = session.GetClr()
			if err != nil {
				return err
			}
			message.transactionGroup = session.StrConv.Decode(temp)
		} else {
			message.transactionGroup = ""
		}
	}
	var length int
	length, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	_, err = session.GetByte()
	if err != nil {
		return err
	}
	for i := 0; i < length; i++ {
		var key, val []byte
		key, val, num, err = session.GetKeyVal()
		if err != nil {
			return err
		}
		switch num {
		case 64:
			if len(key) > 0 {
				message.senderName = session.StrConv.Decode(key)
			}
		case 65:
			if len(key) > 0 {
				message.senderAddress = session.StrConv.Decode(key)
			}
		case 66:
			if len(val) > 0 {
				message.senderProtocol = val[0]
			}
		case 69:
			if len(val) > 0 {
				message.ID = session.StrConv.Decode(val)
			}
		}
	}
	if session.TTCVersion >= 3 {
		_, err = session.GetInt(2, true, true)
		if err != nil {
			return err
		}
		/*csn*/ _, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		/*dsn*/ _, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
	}
	if session.TTCVersion >= 4 {
		var flag int
		flag, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		if flag == 0 {
			flag = 1
		}
		message.DeliveryMode = DeliveryMode(flag)
	}
	if session.TTCVersion >= 16 {
		message.shareNum, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
	}
	return nil
}
func (message *Message) write(session *network.Session) {
	session.PutInt(message.Priority, 4, true, true)
	session.PutInt(message.Delay, 4, true, true)
	session.PutInt(message.Expiration, 4, true, true)
	if len(message.correlation) > 0 {
		temp := session.StrConv.Encode(message.correlation)
		session.PutInt(len(temp), 2, true, true)
		session.PutClr(temp)
	} else {
		session.PutBytes(0)
	}
	session.PutBytes(0)
	if len(message.exceptionQueue) > 0 {
		temp := session.StrConv.Encode(message.exceptionQueue)
		session.PutInt(len(message.exceptionQueue), 2, true, true)
		session.PutClr(temp)
	} else {
		session.PutBytes(0)
	}
	session.PutInt(int(message.State), 4, true, true)
	session.PutBytes(0)
	if session.TTCVersion >= 3 {
		temp := session.StrConv.Encode(message.transactionGroup)
		if len(temp) > 0 {
			session.PutInt(len(temp), 2, true, true)
			session.PutClr(temp)
		} else {
			session.PutBytes(0)
		}
	}
	session.PutInt(4, 2, true, true)
	session.PutBytes(14)
	session.PutKeyValString(message.senderName, "", 64)
	session.PutKeyValString(message.senderAddress, "", 65)
	session.PutKeyVal(nil, []byte{message.senderProtocol}, 66)
	session.PutKeyValString("", message.ID, 69)
	if session.TTCVersion >= 3 {
		session.PutBytes(0, 0, 0)
	}
	if session.TTCVersion >= 4 {
		session.PutBytes(0)
	}
	if session.TTCVersion >= 16 {
		session.PutUint(message.shareNum, 2, true, true)
	}
}

//func NewMessage(data interface{}, messageType MessageType) (*Message, error) {
//	// suppose now payload is []byte
//
//}
