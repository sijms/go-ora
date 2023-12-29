package TestIssues

import (
	"context"
	"testing"
	"time"
)

func TestIssue213(t *testing.T) {
	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	ctx1 := context.Background()

	ctx2, cancel := context.WithTimeout(ctx1, 1*time.Second)
	if err = db.PingContext(ctx2); err != nil {
		t.Log(err)
		cancel()
		return
	}
	cancel()
	t.Log("Successfully connected.")

	conn, err := db.Conn(ctx1)
	defer func() {
		err = conn.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	val := 0

	r := conn.QueryRowContext(ctx1, "select 1 from dual where 1 = 1")
	err = r.Scan(&val)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("1 done")
	time.Sleep(2 * time.Second)

	r = conn.QueryRowContext(ctx1, "select 1 from dual where 1 = 1")
	err = r.Scan(&val)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("2 done")
}
