package aq

import (
	"database/sql/driver"
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

func NewQueueHolder(conn IConnection, name string, messageType MessageType, udtName string, toid []byte) *QueueHolder {
	queue := Queue{
		conn:        conn,
		Name:        name,
		version:     1,
		AutoCommit:  false,
		messageType: messageType,
		udtName:     udtName,
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
