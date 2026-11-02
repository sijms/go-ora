package TestIssues

import (
	"context"
	"testing"

	go_ora "github.com/sijms/go-ora/v3"
)

func TestDBMS_OUTPUT(t *testing.T) {
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
	ctx := context.Background()
	// DBMS_OUTPUT buffers per session, so every call has to use the same connection.
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.DISABLE; END;")
		err = conn.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	_, err = conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.ENABLE(32767); END;")
	if err != nil {
		t.Error(err)
		return
	}
	for _, expected := range []string{"this is a test", "this is a test2"} {
		_, err = conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.PUT_LINE(:1); END;", expected)
		if err != nil {
			t.Error(err)
			return
		}
		var (
			line   string
			status int
		)
		_, err = conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.GET_LINE(:1, :2); END;",
			go_ora.Out{Dest: &line, Size: 32767}, go_ora.Out{Dest: &status})
		if err != nil {
			t.Error(err)
			return
		}
		if status != 0 || line != expected {
			t.Errorf("expected: %s and got: %s (status %d)", expected, line, status)
		}
	}
}
