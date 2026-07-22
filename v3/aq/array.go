package aq

import (
	"errors"
	"fmt"
)

type VisibilityMode int

const (
	VisibilityImmediate VisibilityMode = 1
	VisibilityOnCommit  VisibilityMode = 2
)

//	type AQ struct {
//		name       string
//		version    int
//		flag       int
//		visibility VisibilityMode
//	}
type QAExtension struct {
	key   []byte
	value []byte
	num   uint8
}

func (qu *queue) writeArrayHeader(messages []*Message) {
	session := qu.conn.GetSession()
	session.PutBytes(6)
	queueNameBytes := qu.conn.GetServerStringCoder().Encode(qu.Name)
	if len(queueNameBytes) > 0 {
		session.PutInt(len(queueNameBytes), 2, true, true)
	}
	session.PutCHR(queueNameBytes)

	session.PutBytes(qu.toid...)
	session.PutUint(qu.version, 2, true, true)

	hdrFlag := 0
	if qu.AutoCommit {
		hdrFlag = 32
	}
	session.PutUint(hdrFlag, 4, true, true)
	session.PutInt(int(VisibilityOnCommit), 4, true, true)
	session.PutBytes(0, 0)

	if len(messages[0].extensions) > 0 {
		session.PutInt(len(messages[0].extensions), 2, true, true)
		for _, ext := range messages[0].extensions {
			session.PutKeyVal(ext.key, ext.value, ext.num)
		}
	} else {
		session.PutBytes(0x00)
	}
	session.PutBytes(0x00)
}

func (qu *queue) writeArrayData(messages []*Message) error {
	session := qu.conn.GetSession()
	for _, message := range messages {
		session.PutBytes(7)

		msgFlag := 0
		if qu.AutoCommit {
			msgFlag = 32
		}
		if message.DeliveryMode == DeliveryModeBuffered {
			msgFlag |= 2
		} else if message.DeliveryMode == DeliveryModeAny {
			msgFlag |= 16
		}
		session.PutUint(msgFlag, 4, true, true)

		message.write(qu.conn)

		if len(message.recipients) > 0 {
			session.PutInt(len(message.recipients)*3, 2, true, true)
			for _, recipient := range message.recipients {
				session.PutKeyValString(recipient.Key, recipient.Value, recipient.Num)
			}
		} else {
			session.PutBytes(0)
		}
		switch qu.messageType {
		case RAW:
			session.PutUint(len(message.bValue), 4, true, true)
			session.PutBytes(message.bValue...)
		case JSON:
			session.PutBytes(message.bValue...)
			//quasi := utils.CreateQuasiLocator(uint64(len(message.bValue)))
			//session.PutUint(len(quasi), 4, true, true)
			//session.PutBytes(quasi...)
			//session.PutClr(message.bValue)
		case UDT:
			session.PutBytes(message.bValue...)
			//dataLen := len(message.bValue)
			//session.PutUint(0, 4, true, true)
			//session.PutUint(dataLen, 4, true, true)
			//session.PutBytes(1, 1)
			//if dataLen > 0 {
			//	session.PutClr(message.bValue)
			//}
		default:
			return fmt.Errorf("unsupported message type for enqueue: %v", qu.messageType)
		}
	}
	return nil
}

func (qu *queue) EnqueueMessages(messages []*Message) error {
	var err error
	if len(messages) == 0 {
		return errors.New("no messages to enqueue")
	}
	session := qu.conn.GetSession()
	session.ResetBuffer()

	session.PutTTCFunc(0x3, 0x91)

	// Field 1: null pointer + SWORD(0) — no dequeue input array
	session.PutBytes(0x00, 0x00)

	// Field 2: flags (AQXDEF_ARR=1, AQXDEF_RETMID=2)
	flag := 1
	needMsgID := len(messages) > 0 && len(messages[0].messageID) == 0
	if needMsgID {
		flag |= 2
	}
	session.PutUint(flag, 4, true, true)

	// Field 3: pointer + UB4(0) — dequeue output array
	session.PutBytes(0x01, 0x00)

	session.PutInt(1, 4, true, true)

	// Field 5: pointer — enqueue message
	session.PutBytes(0x01)

	// Field 6: TTC >= 16: UB4(65535)
	if qu.conn.TTCVersion() >= 16 {
		session.PutUint(0xFFFF, 4, true, true)
	}

	// Field 7: iteration count
	session.PutUint(len(messages), 4, true, true)

	// Field 8: null propagation
	session.PutBytes(0x00)

	// ===== Header section (type 6) =====
	qu.writeArrayHeader(messages)

	// ===== Per-message data =====
	err = qu.writeArrayData(messages)
	if err != nil {
		return err
	}

	// ===== Done marker =====
	session.PutBytes(9)

	err = session.Write()
	if err != nil {
		return err
	}

	return qu.readArrayResponse(messages, true)
}

func (qu *queue) DequeueMessages(options *DequeueOptions, count int) ([]*Message, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	session := qu.conn.GetSession()
	session.ResetBuffer()

	session.PutTTCFunc(0x3, 0x91)

	// Field 1: pointer — dequeue input array
	session.PutBytes(1)
	// messages length
	session.PutUint(count, 2, true, true)

	// Field 2: flags (AQXDEF_ARR=1, AQXDEF_RETMID=2)
	flag := 1 | 2
	session.PutUint(flag, 4, true, true)

	// Field 3: pointer + UB4(1) — dequeue output array
	session.PutBytes(1, 1)

	// Field 4: opt = 2 (dequeue)
	session.PutInt(2, 4, true, true)

	// Field 5: pointer — no enqueue message for dequeue
	session.PutBytes(0x00)

	// Field 6: TTC >= 16: shareNum
	if qu.conn.TTCVersion() >= 16 {
		session.PutUint(0xFFFF, 4, true, true)
	}

	var err error
	messages := make([]*Message, 0, count)
	for _ = range count {
		message, err := qu.NewMessage(nil)
		if err != nil {
			return nil, err
		}

		// ===== Dequeue options section =====
		queueNameBytes := qu.conn.GetServerStringCoder().Encode(qu.Name)
		session.PutDlc(queueNameBytes)
		message.write(qu.conn)
		if len(message.recipients) > 0 {
			session.PutInt(len(message.recipients)*3, 2, true, true)
			for _, recipient := range message.recipients {
				session.PutKeyValString(recipient.Key, recipient.Value, recipient.Num)
			}
		} else {
			session.PutBytes(0)
		}
		// consumer name bytes
		consumer := qu.conn.GetServerStringCoder().Encode(options.Consumer)
		session.PutDlc(consumer)
		session.PutInt(int(options.Mode), 4, true, true)
		session.PutInt(int(options.Navigation), 4, true, true)
		session.PutInt(int(options.Visibility), 4, true, true)
		session.PutInt(options.Wait, 4, true, true)
		session.PutDlc(message.messageID)
		correlation := qu.conn.GetServerStringCoder().Encode(options.Correlation)
		session.PutDlc(correlation)
		condition := qu.conn.GetServerStringCoder().Encode(options.Condition)
		session.PutDlc(condition)
		session.PutBytes(0)
		session.PutDlc(nil)
		session.PutBytes(0)
		session.PutDlc(qu.toid)
		session.PutInt(qu.version, 2, true, true)
		session.PutBytes(0, 0, 0)

		flag := 0
		if qu.AutoCommit {
			flag = 32
		}
		if message.DeliveryMode == DeliveryModeBuffered {
			flag |= 2
		} else if message.DeliveryMode == DeliveryModeAny {
			flag |= 16
		}
		session.PutUint(flag, 4, true, true)
		if len(message.extensions) > 0 {
			session.PutInt(len(message.extensions), 4, true, true)
			for _, ext := range message.extensions {
				session.PutKeyVal(ext.key, ext.value, ext.num)
			}
		} else {
			session.PutBytes(0x0)
		}
		session.PutBytes(00)
		messages = append(messages, message)
	}

	err = session.Write()
	if err != nil {
		return nil, err
	}
	err = qu.readArrayResponse(messages, false)
	return messages, err
}
