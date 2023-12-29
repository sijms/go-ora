package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
)

func TestIssue264(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_264(
    ID number(10) NOT NULL,
    SVGTEMPLATE BLOB
    )`)
	}
	var dropTable = func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_264 PURGE`)
	}
	var insert = func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO TTB_264(ID, SVGTEMPLATE) VALUES(:1, :2)", 1, []byte("123456789012345678901234567890"))
		if err != nil {
			return err
		}
		return nil
	}
	var update = func(db *sql.DB) error {
		blob := go_ora.Blob{Data: []byte("123456789012345678901234567890 123456789012345678901234567890")}
		stmt, err := db.Prepare("UPDATE TTB_264 SET SVGTEMPLATE=:1 WHERE ID=1")
		if err != nil {
			return err
		}
		_, err = stmt.Exec(blob)
		if err != nil {
			return err
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
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = update(db)
	if err != nil {
		t.Error(err)
		return
	}
}
