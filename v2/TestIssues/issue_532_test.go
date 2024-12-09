package TestIssues

import (
	"database/sql"
	"strconv"
	"testing"
)

func TestIssue532(t *testing.T) {
	type TTB_DATA struct {
		ID   int    `db:"ID"`
		Name string `db:"NAME"`
	}
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_532(
		EMP_ID NUMBER, 
		EMP_NAME VARCHAR2(255), 
	 	PRIMARY KEY (EMP_ID)
	)`)
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table TTB_532 purge")
	}

	var insert = func(db *sql.DB, rowNum int) error {
		data := make([]TTB_DATA, rowNum)
		for index := range data {
			data[index].ID = index + 1
			data[index].Name = "NAME_" + strconv.Itoa(index+1)
		}
		_, err := db.Exec(`INSERT INTO TTB_532(EMP_ID, EMP_NAME) VALUES(:ID, :NAME)`, data)
		return err
	}

	var query = func(db *sql.DB) error {
		rows, err := db.Query("SELECT * FROM TTB_532 FOR UPDATE")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
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
	err = insert(db, 7)
	if err != nil {
		t.Error(err)
		return
	}
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
