package go_ora

import (
	"database/sql/driver"
)

type Transaction struct {
	conn *Connection
}

// Commit and Rollback run without a context watcher. database/sql rolls a transaction back from Tx.awaitDone
// because the transaction's context is done, so a watcher on that context breaks the rollback it is running, or,
// when the break arrives after the rollback has finished, the next call on the connection (ORA-01013).
func (tx *Transaction) Commit() error {
	if tx.conn.State != Opened {
		return driver.ErrBadConn
	}
	return tx.conn.endTransaction(0xE)
}

func (tx *Transaction) Rollback() error {
	if tx.conn.State != Opened {
		return driver.ErrBadConn
	}
	return tx.conn.endTransaction(0xF)
}
