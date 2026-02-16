package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
)

func TestBulkinsertBFile(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `create table GOORA_TEST_BFILE (
    FILE_ID NUMBER(10) NOT NULL,
    FILE_DATA BFILE
)`)
	}
	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_TEST_BFILE purge")
	}

	var insertEmpty = func(db *sql.DB) error {
		var files []interface{} = make([]interface{}, 2)
		files[0] = go_ora.CreateNullBFile()
		files[1] = go_ora.CreateNullBFile()
		ids := []int{1, 2}
		_, err := db.Exec("INSERT INTO GOORA_TEST_BFILE(FILE_ID, FILE_DATA) VALUES(:1, :2)", ids, files)
		return err
	}
	var insert = func(db *sql.DB, dirName, fileName string) error {
		var files = make([]interface{}, 3)
		var err error
		files[0], err = go_ora.CreateBFile(db, dirName, fileName)
		if err != nil {
			return err
		}
		files[1] = go_ora.CreateNullBFile()
		files[2], err = go_ora.CreateBFile(db, dirName, fileName)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO GOORA_TEST_BFILE(FILE_ID, FILE_DATA) VALUES(:1, :2)", []int{1, 2, 3}, files)
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
	err = insertEmpty(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(db, "dir", "file")
	if err != nil {
		t.Error(err)
		return
	}
}
