package aq

import (
	"database/sql/driver"
	"github.com/sijms/go-ora/v2/network"
)

type QueueHolder struct {
	queue Queue
}

func (qh *QueueHolder) Columns() []string {
	return []string{qh.queue.Name}
}

func (qh *QueueHolder) Close() error {
	return nil
}

func (qh *QueueHolder) Next(dest []driver.Value) error {
	if len(dest) > 0 {
		dest[0] = qh.queue
	}
	return nil
}

func NewQueueHolder(session *network.Session, name string, messageType MessageType, udtName string, toid []byte,
	processTTCResponse func(msgCode uint8) error,
	encodeData func(data interface{}) ([]byte, error),
	decodeData func(data []byte, messageType MessageType, udtName string) (interface{}, error)) *QueueHolder {
	queue := Queue{
		session:            session,
		Name:               name,
		version:            1,
		AutoCommit:         false,
		messageType:        messageType,
		udtName:            udtName,
		processTTCResponse: processTTCResponse,
		encodeData:         encodeData,
		decodeData:         decodeData,
	}
	switch messageType {
	case RAW:
		queue.toid = RAW_TOID
	case JSON:
		queue.toid = JSON_TOID
	case XML:
		queue.toid = XMLTYPE_TOID
	default:
		queue.toid = toid
	}
	return &QueueHolder{queue}
}
