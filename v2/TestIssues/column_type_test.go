package TestIssues

import (
	"database/sql"
	"testing"
)

func TestColumnType(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE GOORA_T_CHAR ( 
    col1 char(32), 
    col2 nchar(32), 
    col3 varchar2(32), 
    col4 nvarchar2(32))`)
	}

	dropTable := func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_T_CHAR purge")
	}

	insert := func(db *sql.DB) error {
		return execCmd(db, `INSERT INTO GOORA_T_CHAR VALUES('char','nchar','varchar2','nvarchar2')`)
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
	rows, err := db.Query("SELECT col1, col2, col3, col4 FROM GOORA_T_CHAR")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	for rows.Next() {
		columns, err := rows.ColumnTypes()
		if err != nil {
			t.Error(err)
			return
		}
		for key, val := range columns {
			if (key == 0 || key == 1) && val.DatabaseTypeName() != "CHAR" {
				t.Errorf("expected: %s and got: %s", "CHAR", val.DatabaseTypeName())
			}
			if (key == 2 || key == 3) && val.DatabaseTypeName() != "NCHAR" {
				t.Errorf("expected: %s and got: %s", "NCHAR", val.DatabaseTypeName())
			}
		}
	}
}
