package aq

import (
	"errors"
	"time"

	//go_ora "github.com/sijms/go-ora/v2"
	"github.com/sijms/go-ora/v2/network"
)

type Queue struct {
	Name               string
	version            int
	session            *network.Session
	messageType        MessageType
	udtName            string
	toid               []byte
	AutoCommit         bool
	processTTCResponse func(msgCode uint8) error
	encodeData         func(data interface{}) ([]byte, error)
	decodeData         func(data []byte, messageType MessageType, udtName string) (interface{}, error)
}

//func NewQueue(session *network.Session, name string, messageType MessageType, udtTypeName string, udt_toid []byte) *Queue {
//	output := &Queue{
//		session:     session,
//		Name:        name,
//		version:     1,
//		messageType: messageType,
//		udtName:     udtTypeName,
//		AutoCommit:  false,
//	}
//
//	return output
//}

func (queue *Queue) NewMessage(data interface{}) (*Message, error) {
	var err error
	message := &Message{
		messageID:        nil,
		Delay:            0,
		Expiration:       -1,
		correlation:      "",
		exceptionQueue:   "",
		deqAttempts:      0,
		DeliveryMode:     DeliveryModePersistent,
		VisibilityMode:   VisibilityOnCommit,
		recipients:       nil,
		sender:           OracleAgent{},
		EnqTime:          time.Time{},
		State:            0,
		Priority:         0,
		ID:               "",
		propModified:     false,
		shareNum:         0xFFFF,
		senderName:       "",
		senderAddress:    "",
		senderProtocol:   0,
		transactionGroup: "",
		Payload:          data,
		extensions:       nil,
	}
	switch queue.messageType {
	case RAW:
		message.payloadInBytes, err = queue.encodeData(data)
	case UDT:
		message.payloadInBytes, err = queue.encodeData(data)
		//par.Value = go_ora.NewObject("", queue.UdtTypeName, data)
	default:
		return nil, errors.New("unsupported message type")
	}

	return message, err
}
func (queue *Queue) Enqueue(message *Message) error {
	session := queue.session
	session.ResetBuffer()
	session.PutBytes(3, 0x79, 0)
	if session.TTCVersion >= 18 {
		session.PutBytes(0)
	}
	queueNameBytes := session.StrConv.Encode(queue.Name)
	if len(queueNameBytes) > 0 {
		session.PutBytes(1)
		session.PutInt(len(queueNameBytes), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	// message.marshal()
	message.write(queue.session)
	if len(message.recipients) > 0 {
		session.PutBytes(1)
		session.PutInt(len(message.recipients)*3, 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutInt(int(message.VisibilityMode), 4, true, true)
	if len(message.messageID) > 0 {
		session.PutBytes(1)
		session.PutInt(len(message.messageID), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}

	session.PutBytes(0, 1)
	session.PutInt(0x10, 2, true, true)
	session.PutUint(queue.version, 2, true, true)
	switch queue.messageType {
	case RAW:
		session.PutBytes(0, 1)
		session.PutUint(len(message.payloadInBytes), 4, true, true)
	case JSON:
		session.PutBytes(0, 0, 0)
	default:
		session.PutBytes(1, 0, 0)
		//session.PutUint(len(message.payloadInBytes), 4, true, true)
	}
	// retrieve message id in response
	if len(message.messageID) == 0 {
		session.PutBytes(1)
		session.PutInt(0x10, 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}

	num := 0
	if queue.AutoCommit {
		num = 32
	}
	if message.DeliveryMode == DeliveryModeBuffered {
		num |= 2
	} else if message.DeliveryMode == DeliveryModeAny {
		num |= 16
	}
	session.PutUint(num, 4, true, true)
	session.PutBytes(0, 0)
	if len(message.extensions) > 0 {
		session.PutBytes(1)
		session.PutInt(len(message.extensions), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutBytes(0, 0, 0, 0, 0)
	if session.TTCVersion >= 4 {
		session.PutBytes(0, 0, 0, 0, 0, 0, 0, 0)
	}
	if session.TTCVersion >= 14 {
		if queue.messageType == JSON {
			session.PutBytes(1)
		} else {
			session.PutBytes(0)
		}
	}
	if len(queueNameBytes) > 0 {
		session.PutClr(queueNameBytes)
	}
	for _, recipient := range message.recipients {
		session.PutKeyValString(recipient.Key, recipient.Value, recipient.Num)
	}
	if len(message.messageID) > 0 {
		session.PutBytes(message.messageID...)
	}
	session.PutBytes(queue.toid...)
	switch queue.messageType {
	case RAW:
		session.PutBytes(message.payloadInBytes...)
	case UDT:
		session.PutBytes(0, 0, 0, 0)
		size := len(message.payloadInBytes)
		session.PutUint(size, 4, true, true)
		session.PutBytes(1, 1)
		session.PutClr(message.payloadInBytes)
	//if (!this.isRawQueue)
	//{
	//	if (!this.isJsonQueue)
	//	{
	//		this.toh.Init(this.payloadToid, this.messageDataLength);
	//		this.toh.Marshal(this.m_marshallingEngine);
	//		if (this.messageData != null)
	//		{
	//			this.m_marshallingEngine.MarshalCLR(this.messageData, 0, this.messageDataLength);
	//		}
	//	}
	//}
	default:
		return errors.New("unsupported message type")
	}
	for _, extension := range message.extensions {
		session.PutKeyVal(extension.key, extension.value, extension.num)
	}
	err := session.Write()
	if err != nil {
		return err
	}
	//if (this.isJsonQueue)
	//{
	//	int num2 = ((this.messageData != null) ? this.messageData.Length : 0);
	//	byte[] array = TTCLob.CreateQuasiLocator((long)num2);
	//	this.m_marshallingEngine.MarshalUB4((long)array.Length);
	//	this.m_marshallingEngine.MarshalB1Array(array);
	//	this.m_marshallingEngine.MarshalCLR(this.messageData, 0, num2);
	//}
	// read
	// 08 3a b0 a2 6a 94 91 15 12 e0 63 02 00 11 ac e0 6d 00 09 01 02 01 06
	return queue.readEnqueueResponse(message)
}

func (queue *Queue) Dequeue(options *DequeueOptions) (*Message, error) {
	outMsg := &Message{}
	session := queue.session
	session.ResetBuffer()
	session.PutBytes(3, 0x7A, 0)
	if session.TTCVersion >= 18 {
		session.PutBytes(0)
	}
	queueNameBytes := session.StrConv.Encode(queue.Name)
	if len(queueNameBytes) > 0 {
		session.PutBytes(1)
		session.PutInt(len(queueNameBytes), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutBytes(1, 1, 1, 1)
	consumer := session.StrConv.Encode(options.Consumer)
	if len(consumer) > 0 {
		session.PutBytes(1)
		session.PutUint(len(consumer), 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutInt(int(options.Mode), 4, true, true)
	session.PutInt(int(options.Navigation), 4, true, true)
	session.PutInt(int(options.Visibility), 4, true, true)
	session.PutInt(options.Wait, 4, true, true)
	if len(outMsg.messageID) > 0 {
		session.PutBytes(1)
		session.PutInt(len(outMsg.messageID), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	correlation := session.StrConv.Encode(options.Correlation)
	if len(correlation) > 0 {
		session.PutBytes(1)
		session.PutInt(len(correlation), 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutBytes(1)
	session.PutInt(len(queue.toid), 4, true, true)
	//session.PutInt(0x10, 2, true, true)
	session.PutUint(queue.version, 2, true, true)
	session.PutBytes(1)
	// retrieve message id in response
	if len(outMsg.messageID) == 0 {
		session.PutBytes(1)
		session.PutInt(0x10, 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	num := 0
	if queue.AutoCommit {
		num = 32
	}
	if outMsg.DeliveryMode == DeliveryModeBuffered {
		num |= 2
	} else if outMsg.DeliveryMode == DeliveryModeAny {
		num |= 16
	}
	session.PutUint(num, 4, true, true)
	condition := session.StrConv.Encode(options.Condition)
	if len(condition) > 0 {
		session.PutBytes(1)
		session.PutInt(len(condition), 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	if len(outMsg.extensions) > 0 {
		session.PutBytes(1)
		session.PutInt(len(outMsg.extensions), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	if session.TTCVersion >= 14 {
		session.PutBytes(0)
	}
	if session.TTCVersion >= 16 {
		session.PutInt(0xFFFF, 4, true, true)
	}
	if len(queueNameBytes) > 0 {
		session.PutClr(queueNameBytes)
	}
	if len(consumer) > 0 {
		session.PutClr(consumer)
	}
	if len(outMsg.messageID) > 0 {
		session.PutBytes(outMsg.messageID...)
	}
	if len(correlation) > 0 {
		session.PutClr(correlation)
	}
	session.PutBytes(queue.toid...)
	if len(condition) > 0 {
		session.PutClr(condition)
	}
	for _, extension := range outMsg.extensions {
		session.PutKeyVal(extension.key, extension.value, extension.num)
	}
	err := session.Write()
	if err != nil {
		return nil, err
	}
	err = queue.readDequeueResponse(outMsg)
	return outMsg, err
}

func (queue *Queue) readEnqueueResponse(message *Message) error {
	session := queue.session
	loop := true
	for loop {
		msg, err := session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 8:
			message.messageID, err = session.GetBytes(16)
			_, err = session.GetInt(2, true, true)
		default:
			err = queue.processTTCResponse(msg)
			if err != nil {
				return err
			}
			if msg == 4 || msg == 9 {
				loop = false
			}
		}
	}
	return nil
}

func (queue *Queue) readDequeueResponse(message *Message) error {
	session := queue.session
	loop := true
	for loop {
		msg, err := session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 8:
			var num int
			var err error
			num, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if num > 0 {
				err = message.read(session)
				if err != nil {
					return err
				}
			}
			_, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			// unmarshal toc
			num, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if num > 0 {
				queue.toid, err = session.GetClr()
				if err != nil {
					return err
				}
			}
			num, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if num > 0 {
				/*oid*/ _, err = session.GetClr()
				if err != nil {
					return err
				}
			}
			num, err = session.GetInt(4, true, true)
			if num > 0 {
				/*snapshot*/ _, err = session.GetClr()
			}
			//if session.TTCVersion >= 8 {
			//	this.ksnp.Unmarshal(meg)
			//} else {
			//}
			queue.version, err = session.GetInt(2, true, true)
			if err != nil {
				return err
			}
			var imageLen int
			imageLen, err = session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			/*flag*/ _, err = session.GetInt(2, true, true)
			if err != nil {
				return err
			}
			if imageLen > 0 {
				message.payloadInBytes, err = session.GetClr()
				if err != nil {
					return err
				}
				switch queue.messageType {
				case RAW:
					fallthrough
				case JSON:
					if len(message.payloadInBytes) > 4 {
						message.Payload = message.payloadInBytes[4:]
					} else {
						message.Payload = message.payloadInBytes
					}
				default:
					message.Payload, err = queue.decodeData(message.payloadInBytes, queue.messageType, queue.udtName)
					if err != nil {
						return nil
					}
				}
			}
			message.messageID, err = session.GetBytes(16)
			//_, err = session.GetInt(2, true, true)
		default:
			err = queue.processTTCResponse(msg)
			if err != nil {
				return err
			}
			if msg == 4 || msg == 9 {
				loop = false
			}
		}
	}
	return nil
}
