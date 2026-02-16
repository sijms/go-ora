package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v3"
	"strings"
	"testing"
)

func TestClientCharset(t *testing.T) {
	var createPackage = func(db *sql.DB) error {
		return execCmd(db,
			// create package
			`CREATE OR REPLACE PACKAGE GOORA_TEMP IS
	TYPE VARCHAR2TABLE_T IS TABLE OF VARCHAR2(32767) INDEX BY BINARY_INTEGER;
	PROCEDURE TEST_PROC(
		STRING_ARRAY_IN IN VARCHAR2TABLE_T,
		STRING_OUT OUT VARCHAR2
	);
END GOORA_TEMP;`,

			// create package body
			`CREATE OR REPLACE PACKAGE BODY GOORA_TEMP IS
	PROCEDURE TEST_PROC(
		STRING_ARRAY_IN IN VARCHAR2TABLE_T,
		STRING_OUT OUT VARCHAR2
	) IS
	BEGIN
		FOR i IN STRING_ARRAY_IN.FIRST..STRING_ARRAY_IN.LAST LOOP
			STRING_OUT := STRING_OUT || STRING_ARRAY_IN(i);
		END LOOP;
	END;
END GOORA_TEMP;`,
		)
	}

	var dropPackage = func(db *sql.DB) error {
		return execCmd(db, "DROP PACKAGE GOORA_TEMP")
	}

	var callProc = func(db *sql.DB, strings_in []string) error {
		var string_out string
		_, err := db.Exec(`BEGIN GOORA_TEMP.TEST_PROC(:1, :2); END;`, strings_in,
			go_ora.Out{&string_out, 256, false})
		if err != nil {
			return err
		}
		expected := strings.Join(strings_in, "")
		if string_out != expected {
			return fmt.Errorf("Expected %s and got %s", expected, string_out)
		}
		return nil
	}
	urlOptions["charset"] = "UTF8"
	defer func() {
		delete(urlOptions, "charset")
	}()
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
	err = createPackage(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropPackage(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = callProc(db, []string{"a", "b", "c"})
	if err != nil {
		t.Error(err)
		return
	}
	err = callProc(db, []string{"a", "b", "中国人中国人中国人"})
	if err != nil {
		t.Error(err)
		return
	}

}
