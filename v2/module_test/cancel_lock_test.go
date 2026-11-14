package v2_test

import (
	"context"
	"testing"
	"time"
)

// TestCancelArrayDMLWaitingOnRowLock checks that an ExecContext with array binds returns
// when its context ends while the statement waits on a row lock. Needs ORACLE_DSN.
func TestCancelArrayDMLWaitingOnRowLock(t *testing.T) {
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if db == nil {
		t.Skip("ORACLE_DSN is not set")
	}
	defer db.Close()
	_, _ = db.Exec(`DROP TABLE GOORA_CANCEL_LOCK PURGE`)
	if _, err = db.Exec(`CREATE TABLE GOORA_CANCEL_LOCK (ID NUMBER PRIMARY KEY, V NUMBER)`); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DROP TABLE GOORA_CANCEL_LOCK PURGE`)
	if _, err = db.Exec(`INSERT INTO GOORA_CANCEL_LOCK VALUES (1, 0)`); err != nil {
		t.Fatal(err)
	}
	holder, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Rollback()
	if _, err = holder.Exec(`UPDATE GOORA_CANCEL_LOCK SET V = 1 WHERE ID = 1`); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := db.ExecContext(ctx, `MERGE INTO GOORA_CANCEL_LOCK t
USING (SELECT :1 ID, :2 V FROM dual) s ON (t.ID = s.ID)
WHEN MATCHED THEN UPDATE SET t.V = s.V`, []int64{1}, []int64{2})
		done <- err
	}()
	select {
	case err = <-done:
		if err == nil {
			t.Fatal("ExecContext succeeded while the row was locked")
		}
		t.Logf("ExecContext returned: %v", err)
	case <-time.After(60 * time.Second):
		t.Fatal("ExecContext did not return within 60 s after its context ended")
	}
}
