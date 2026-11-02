package TestIssues

import (
	"database/sql"
	"github.com/sijms/go-ora/v3/types"
	"testing"
)

func TestBulkinsertBlob(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_465(DATA BLOB)`)
	}
	var dropTable = func(db *sql.DB) error { return execCmd(db, `DROP TABLE TTB_465 PURGE`) }
	var insert = func(db *sql.DB, data []byte) error {
		datas := []types.Blob{{}, blobOf(data), blobOf(data)}
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
