package aq

import "github.com/sijms/go-ora/v2/network"

type AQArrayInParam struct {
	flag    int // equal to AQ.flag
	message Message
}

func (param *AQArrayInParam) write(session *network.Session) error {
	session.PutBytes(7)
	session.PutUint(param.flag, 4, true, true)
	param.message.write(session)
	if len(param.message.recipients) > 0 {
		session.PutInt(len(param.message.recipients)*3, 2, true, true)
		for _, recipient := range param.message.recipients {
			session.PutKeyValString(recipient.Key, recipient.Value, recipient.Num)
		}
	} else {
		session.PutBytes(0)
	}
	//if param.message.isJsonQueue() {
	//
	//}
	//param.message.payloadInBytes
	return nil
}

//func MarshalData() {
//
//
//	int num;
//	if (this.messageData != null) {
//		num = this.messageData.Length;
//	} else
//	{
//		num = 0;
//	}
//	if (this.isJSONQueue) {
//		int num2 = ((this.messageData != null) ? this.messageData.Length : 0);
//		byte[] array2 = TTCLob.CreateQuasiLocator((long) num2);
//		this.m_marshallingEngine.MarshalUB4((long)array2.Length);
//		this.m_marshallingEngine.MarshalB1Array(array2);
//		this.m_marshallingEngine.MarshalCLR(this.messageData, 0, num2);
//		return;
//	}
//	if (!this.isRawQueue) {
//		this.toh.Init(this.messageOid, num);
//		this.toh.Marshal(this.m_marshallingEngine);
//		if (this.messageData != null) {
//			this.m_marshallingEngine.MarshalCLR(this.messageData, 0, num);
//			return;
//		}
//	} else
//	{
//		this.m_marshallingEngine.MarshalUB4((long)num);
//		this.m_marshallingEngine.MarshalB1Array(this.messageData);
//	}
//}
