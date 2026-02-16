package TestIssues

import (
	"database/sql"
	"errors"
	go_ora "github.com/sijms/go-ora/v3"
	"strings"
	"testing"
)

func TestIssue173(t *testing.T) {
	var createFunction = func(db *sql.DB) error {
		return execCmd(db, `CREATE OR REPLACE FUNCTION go_ora$text(p_param VARCHAR2) RETURN CLOB AS
	   v_txt CLOB;
	BEGIN
	   FOR i IN 1 .. 4000
	       LOOP
	           v_txt := v_txt || 'tttttttttttttttttttttttt';
	       END LOOP;
	   RETURN v_txt || '\n' || p_param;
	END;`)
	}
	var dropFunction = func(db *sql.DB) error { return execCmd(db, "DROP FUNCTION go_ora$text") }
	var run = func(db *sql.DB) error {
		var str go_ora.Clob
		_, err := db.Exec("BEGIN :1 := go_ora$text(:2);end;", go_ora.Out{Dest: &str, Size: 100000}, "ss")
		if err != nil {
			return err
		}
		expected := strings.Repeat("tttttttttttttttttttttttt", 4000) + "\\n" + "ss"
		if str.String != expected {
			return errors.New("return value is different from input value")
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
	err = createFunction(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropFunction(db)
		if err != nil {
			t.Error(err)
		}
	}()
	for i := 0; i < 100; i++ {
		err = run(db)
		if err != nil {
			t.Error(err)
			return
		}
	}

}
