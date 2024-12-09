// call prepare stmt which contain lob after call query double time will get error
// fetch out of sequence.
// the issue occur because queryLobPrefetch should be called one time (at first) only
package TestIssues

import (
	"database/sql"
	"testing"
)

func TestIssue556(t *testing.T) {
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
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT TO_CLOB('this is a test') FROM DUAL")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		t.Error(err)
		return
	}
	var data string
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			t.Error(err)
			return
		}
		if data != "this is a test" {
			t.Errorf("expected %s and got %s", "this is a test", data)
			return
		}
	}
	rows, err = stmt.Query()
	if err != nil {
		t.Error(err)
		return
	}
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			t.Error(err)
			return
		}
		if data != "this is a test" {
			t.Errorf("expected %s and got %s", "this is a test", data)
			return
		}
	}
}
