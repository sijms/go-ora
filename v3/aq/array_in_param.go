package aq

//type AQArrayInParam struct {
//	flag    int // equal to AQ.flag
//	message Message
//}

//func (param *AQArrayInParam) write(conn IConnection, messageType MessageType) error {
//	session := conn.GetSession()
//	session.PutBytes(7)
//	session.PutUint(param.flag, 4, true, true)
//	param.message.write(conn)
//	if len(param.message.recipients) > 0 {
//		session.PutInt(len(param.message.recipients)*3, 2, true, true)
//		for _, recipient := range param.message.recipients {
//			session.PutKeyValString(recipient.Key, recipient.Value, recipient.Num)
//		}
//	} else {
//		session.PutBytes(0)
//	}
//
//	switch messageType {
//	case RAW:
//		session.PutUint(len(param.message.bValue), 4, true, true)
//		session.PutBytes(param.message.bValue...)
//	case JSON:
//		quasi := utils.CreateQuasiLocator(uint64(len(param.message.bValue)))
//		session.PutUint(len(quasi), 4, true, true)
//		session.PutBytes(quasi...)
//		session.PutClr(param.message.bValue)
//	case UDT:
//		dataLen := len(param.message.bValue)
//		session.PutUint(0, 4, true, true)
//		session.PutUint(dataLen, 4, true, true)
//		session.PutBytes(1, 1)
//		if dataLen > 0 {
//			session.PutClr(param.message.bValue)
//		}
//	default:
//		return errors.New("unsupported message type for enqueue")
//	}
//	return nil
//}
