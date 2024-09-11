package TestIssues

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestIssue466(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_466(ID varchar2(100))`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_466 purge`)
	}
	generateRows := func(startIndex, count int) []string {
		rows := make([]string, count)
		for i := 0; i < count; i++ {
			rows[i] = fmt.Sprintf("ID%02d", i+startIndex)
		}
		return rows
	}
	insert := func(stmt *sql.Stmt, rows []string) error {
		result, err := stmt.Exec(rows)
		if err != nil {
			return err
		}
		inserted, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if int(inserted) != len(rows) {
			return fmt.Errorf("expected to insert: %d but the actual insert: %d", len(rows), inserted)
		}
		return nil
	}
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
	err = createTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			t.Error(err)
		}
	}()
	stmt, err := db.Prepare("INSERT INTO TTB_466 (ID) VALUES (:1)")
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
	err = insert(stmt, generateRows(1, 10))
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(stmt, generateRows(11, 10))
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(stmt, generateRows(21, 5))
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(stmt, generateRows(26, 10))
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(stmt, generateRows(36, 10))
	if err != nil {
		t.Error(err)
		return
	}
}
