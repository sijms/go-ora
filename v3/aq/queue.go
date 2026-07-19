package aq

import (
	"errors"
	"fmt"
	"time"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/utils"
)

type Queue interface {
	NewMessage(data interface{}) (*Message, error)
	Enqueue(message *Message) error
	Dequeue(options *DequeueOptions) (*Message, error)
	SetAutoCommit(autoCommit bool)
}
type queue struct {
	Name        string
	version     int
	conn        IConnection
	messageType MessageType
	udtName     string
	toid        []byte
	AutoCommit  bool
}

func CreateQueue(db utils.Execuer, name string, messageType MessageType, udtName string) (Queue, error) {
	ret := &queue{
		Name:        name,
		messageType: messageType,
		udtName:     udtName,
	}
	_, err := db.Exec("--GET-CONNECTION-REF--", &ret.conn)
	if err != nil {
		return nil, err
	}

	switch messageType {
	case RAW:
		ret.toid = RAW_TOID
	case JSON:
		ret.toid = JSON_TOID
	case XML:
		ret.toid = XMLTYPE_TOID
	default:
		if len(udtName) > 0 {
			coder, err := ret.conn.GetParameterCoder(udtName)
			if err != nil {
				return nil, err
			}
			ret.toid = coder.GetParameterInfo().ToID
		} else {
			return nil, fmt.Errorf("message type udt should have udtName")
		}
	}
	return ret, nil
}

func (qu *queue) SetAutoCommit(autoCommit bool) {
	qu.AutoCommit = autoCommit
}

func (qu *queue) NewMessage(data interface{}) (*Message, error) {
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
	var encoder parameter_coder.OracleParameterCoder
	switch qu.messageType {
	case RAW:
		encoder, err = qu.conn.GetParameterCoder(types.RAW)
		//message.payloadInBytes, err = queue.encodeData(data)
	case UDT:
		encoder, err = qu.conn.GetParameterCoder(qu.udtName)
	case JSON:
		encoder, err = qu.conn.GetParameterCoder(types.JSON)
		if err != nil {
			return nil, err
		}
		encoder.SetAQMessage()
	default:
		err = fmt.Errorf("unsupported message type: %v", qu.messageType)
	}
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(data, qu.conn)
	if err != nil {
		return nil, err
	}
	ms := network.NewMemorySession(nil, nil, qu.conn.GetSession().GetProperties())
	err = encoder.Write(ms)
	message.bValue = ms.GetWriteBuffer()
	return message, err
}
func (qu *queue) Enqueue(message *Message) error {
	session := qu.conn.GetSession()
	session.ResetBuffer()
	session.PutTTCFunc(0x3, 0x79)

	queueNameBytes := qu.conn.GetServerStringCoder().Encode(qu.Name)
	if len(queueNameBytes) > 0 {
		session.PutBytes(1)
		session.PutInt(len(queueNameBytes), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	// message.marshal()
	message.write(qu.conn)
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
	session.PutUint(qu.version, 2, true, true)
	switch qu.messageType {
	case RAW:
		session.PutBytes(0, 1)
		session.PutUint(len(message.bValue), 4, true, true)
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
	if qu.AutoCommit {
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
	if qu.conn.TTCVersion() >= 4 {
		session.PutBytes(0, 0, 0, 0, 0, 0, 0, 0)
	}
	if qu.conn.TTCVersion() >= 14 {
		if qu.messageType == JSON {
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
	session.PutBytes(qu.toid...)
	switch qu.messageType {
	case JSON:
	case RAW:
		session.PutBytes(message.bValue...)
	case UDT:
		session.PutBytes(message.bValue...)
		//session.PutBytes(0, 0, 0, 0)
		//size := len(message.payloadInBytes)
		//session.PutUint(size, 4, true, true)
		//session.PutBytes(1, 1)
		//session.PutClr(message.payloadInBytes)
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
	if qu.messageType == JSON {
		session.PutBytes(message.bValue...)
		//// create quasi locator
		//quasi := utils.CreateQuasiLocator(uint64(len(message.payloadInBytes)))
		//// pass uint4 length of locator
		//session.PutUint(len(quasi), 4, true, true)
		//// pass locator as clr
		//session.PutBytes(quasi...)
		//// pass data as clr
		//session.PutClr(message.payloadInBytes)
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
	return qu.readEnqueueResponse(message)
}

func (qu *queue) Dequeue(options *DequeueOptions) (*Message, error) {
	outMsg := &Message{}
	session := qu.conn.GetSession()
	session.ResetBuffer()
	session.PutTTCFunc(0x3, 0x7A)
	queueNameBytes := qu.conn.GetServerStringCoder().Encode(qu.Name)
	if len(queueNameBytes) > 0 {
		session.PutBytes(1)
		session.PutInt(len(queueNameBytes), 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutBytes(1, 1, 1, 1)
	consumer := qu.conn.GetServerStringCoder().Encode(options.Consumer)
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
	correlation := qu.conn.GetServerStringCoder().Encode(options.Correlation)
	if len(correlation) > 0 {
		session.PutBytes(1)
		session.PutInt(len(correlation), 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	session.PutBytes(1)
	session.PutInt(len(qu.toid), 4, true, true)
	//session.PutInt(0x10, 2, true, true)
	session.PutUint(qu.version, 2, true, true)
	session.PutBytes(1)
	// retrieve message id in response
	if len(outMsg.messageID) == 0 {
		session.PutBytes(1)
		session.PutInt(0x10, 2, true, true)
	} else {
		session.PutBytes(0, 0)
	}
	num := 0
	if qu.AutoCommit {
		num = 32
	}
	if outMsg.DeliveryMode == DeliveryModeBuffered {
		num |= 2
	} else if outMsg.DeliveryMode == DeliveryModeAny {
		num |= 16
	}
	session.PutUint(num, 4, true, true)
	condition := qu.conn.GetServerStringCoder().Encode(options.Condition)
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
	if qu.conn.TTCVersion() >= 14 {
		session.PutBytes(0)
	}
	if qu.conn.TTCVersion() >= 16 {
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
	session.PutBytes(qu.toid...)
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
	err = qu.readDequeueResponse(outMsg)
	return outMsg, err
}

func (qu *queue) readEnqueueResponse(message *Message) error {
	session := qu.conn.GetSession()
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
			err = qu.conn.ProcessTCCResponse(msg)
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

func (qu *queue) readDequeueResponse(message *Message) error {
	session := qu.conn.GetSession()
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
				err = message.read(qu.conn)
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
				qu.toid, err = session.GetClr()
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
			qu.version, err = session.GetInt(2, true, true)
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

				message.bValue, err = session.GetClr()
				if err != nil {
					return err
				}
				var decoder parameter_coder.OracleParameterDecoder
				switch qu.messageType {
				case RAW:
					decoder, err = qu.conn.GetParameterCoder(types.RAW)
				case JSON:
					decoder, err = qu.conn.GetParameterCoder(types.JSON)
				default:
					decoder, err = qu.conn.GetParameterCoder(qu.udtName)
				}
				if err != nil {
					return err
				}
				if qu.messageType == RAW || qu.messageType == JSON && len(message.bValue) > 4 {
					decoder.SetBytes(message.bValue[4:])
				} else {
					decoder.SetBytes(message.bValue)
				}
				message.Payload, err = decoder.Decode(qu.conn)
				if qu.messageType == JSON {
					if temp, ok := message.Payload.(*types.Json); ok {
						message.Payload, err = temp.Value()
						if err != nil {
							return err
						}
					}
				}
				if err != nil {
					return nil
				}
			}
			message.messageID, err = session.GetBytes(16)
			//_, err = session.GetInt(2, true, true)
		default:
			err = qu.conn.ProcessTCCResponse(msg)
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
