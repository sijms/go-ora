package v2_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"os"
	"sync/atomic"
	"testing"

	go_ora "github.com/sijms/go-ora/v3"
)

// The tests below cover issue #767: a transaction whose context is cancelled, and a connection closed while its
// transaction is open. They need ORACLE_DSN and are meant to run under -race.

func txTestTable(t *testing.T, db *sql.DB) {
	t.Helper()
	_, _ = db.Exec(`DROP TABLE GOORA_TX_TEST PURGE`)
	if _, err := db.Exec(`CREATE TABLE GOORA_TX_TEST (ID NUMBER)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DROP TABLE GOORA_TX_TEST PURGE`) })
}

func txTestRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM GOORA_TX_TEST`).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return n
}

// failingDialer dials connections whose next write fails, once armed, without sending anything.
type failingDialer struct {
	failNextWrite int32
}

func (d *failingDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := (&net.Dialer{}).DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return &failingConn{Conn: c, dialer: d}, nil
}

type failingConn struct {
	net.Conn
	dialer *failingDialer
}

func (c *failingConn) Write(b []byte) (int, error) {
	if atomic.CompareAndSwapInt32(&c.dialer.failNextWrite, 1, 0) {
		return 0, errors.New("write failed by the test")
	}
	return c.Conn.Write(b)
}

// database/sql rolls back a transaction whose context is cancelled from Tx.awaitDone and, as the driver validates
// its connections, keeps the connection. The rollback must complete, and the connection must serve the next
// statement: a break sent for the cancelled context made it fail with ORA-01013.
func TestTransaction_CancelledContextRollsBackAndKeepsTheConnectionUsable(t *testing.T) {
	db, err := newDbConnection()
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Skip("ORACLE_DSN is not set")
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	txTestTable(t, db)

	const iterations = 50
	for i := 0; i < iterations; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			cancel()
			t.Fatalf("iteration %d: begin: %v", i, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO GOORA_TX_TEST (ID) VALUES (:1)`, i); err != nil {
			cancel()
			t.Fatalf("iteration %d: insert: %v", i, err)
		}
		cancel()
		// The pool holds one connection: this statement runs on it once Tx.awaitDone has rolled back.
		if _, err = db.Exec(`UPDATE GOORA_TX_TEST SET ID = ID WHERE 1 = 0`); err != nil {
			t.Fatalf("iteration %d: statement after the cancelled transaction: %v", i, err)
		}
	}
	if n := txTestRows(t, db); n != 0 {
		t.Fatalf("%d rows of %d cancelled transactions were committed", n, iterations)
	}
}

// Oracle commits the open transaction of a session that logs off. Closing a connection with its transaction open
// must not commit it.
func TestTransaction_CloseWithAnOpenTransactionDoesNotCommitIt(t *testing.T) {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		t.Skip("ORACLE_DSN is not set")
	}
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	txTestTable(t, db)

	ctx := context.Background()
	conn, err := go_ora.NewConnector(dsn).Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{}); err != nil {
		t.Fatal(err)
	}
	stmt, err := conn.Prepare(`INSERT INTO GOORA_TX_TEST (ID) VALUES (1)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = stmt.(driver.StmtExecContext).ExecContext(ctx, nil); err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
	if err = conn.Close(); err != nil {
		t.Logf("close: %v", err)
	}
	if n := txTestRows(t, db); n != 0 {
		t.Fatalf("closing the connection committed its open transaction (%d row)", n)
	}
}

// A rollback that fails leaves the transaction's state on the server unknown. The connection must not be reused,
// and closing it must not commit the transaction.
func TestTransaction_AFailedRollbackIsNotCommittedWhenTheConnectionCloses(t *testing.T) {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		t.Skip("ORACLE_DSN is not set")
	}
	check, err := sql.Open("oracle", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = check.Close() }()
	txTestTable(t, check)

	connector := go_ora.NewConnector(dsn).(*go_ora.OracleConnector)
	dialer := &failingDialer{}
	connector.Dialer(dialer)
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(`INSERT INTO GOORA_TX_TEST (ID) VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	atomic.StoreInt32(&dialer.failNextWrite, 1)
	if err = tx.Rollback(); err == nil {
		t.Fatal("the rollback was expected to fail")
	}
	// An autocommitted statement on the connection would commit the transaction with it.
	if _, err = db.Exec(`UPDATE GOORA_TX_TEST SET ID = ID WHERE 1 = 0`); err != nil {
		t.Fatalf("statement after the failed rollback: %v", err)
	}
	if n := txTestRows(t, check); n != 0 {
		t.Fatalf("the transaction whose rollback failed was committed by the next statement (%d row)", n)
	}
	if err = db.Close(); err != nil {
		t.Logf("close: %v", err)
	}
	if n := txTestRows(t, check); n != 0 {
		t.Fatalf("the transaction whose rollback failed was committed when its connection closed (%d row)", n)
	}
}
