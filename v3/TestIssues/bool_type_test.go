package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
)

func TestBoolType(t *testing.T) {
	var createProc = func(db *sql.DB) error {
		return execCmd(db, `CREATE OR REPLACE PROCEDURE GO_ORA_TEMP_PROC(
		L_BOOL IN BOOLEAN,
		MESSAGE OUT VARCHAR2) AS
BEGIN
	IF L_BOOL THEN 
		MESSAGE := 'TRUE';
	ELSE
		MESSAGE := 'FALSE';
	END IF;
END GO_ORA_TEMP_PROC;`)
	}
	var dropProc = func(db *sql.DB) error {
		return execCmd(db, `drop procedure GO_ORA_TEMP_PROC`)
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
	err = createProc(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropProc(db)
		if err != nil {
			t.Error(err)
		}
	}()
	var message string
	_, err = db.Exec("BEGIN GO_ORA_TEMP_PROC(:1, :2); END;", go_ora.PLBool(false), go_ora.Out{Dest: &message, Size: 100})
	if err != nil {
		t.Error(err)
		return
	}
	if message != "FALSE" {
		t.Errorf("expected %s and got %s", "FALSE", message)
	}
}
