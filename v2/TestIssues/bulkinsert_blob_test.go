package TestIssues

import (
	"database/sql"
	"testing"

	go_ora "github.com/sijms/go-ora/v2"
)

func TestBulkinsertBlob(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_465(DATA BLOB)`)
	}
	dropTable := func(db *sql.DB) error { return execCmd(db, `DROP TABLE TTB_465 PURGE`) }
	insert := func(db *sql.DB, data []byte) error {
		datas := make([]go_ora.Blob, 3)
		datas[0].Data = nil
		datas[1].Data = data
		datas[2].Data = data
		_, err := db.Exec("INSERT INTO TTB_465 VALUES(:1)", datas)
		return err
	}

	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error()
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
	err = insert(db, []byte("this is a test"))
	if err != nil {
		t.Error(err)
		return
	}
}
