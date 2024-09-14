// issue 589
package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
)

func TestRecallStoredProc(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `create or replace procedure SIMPLE_ADD_NUM(
	N1 in NUMBER,
	N2 in NUMBER,
	N3 out NUMBER) AS
BEGIN
	N3 := N1 + N2;
END;`, `create or replace procedure SIMPLE_ADD_STR(
	N1 in VARCHAR2,
	N2 in VARCHAR2,
	N3 out VARCHAR2) AS
BEGIN
	N3 := N1 || N2;
END;`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db, `drop procedure SIMPLE_ADD_NUM`, `drop procedure SIMPLE_ADD_STR`)
	}
	var callAddInt = func(db *sql.DB) error {
		var result int
		s, err := db.Prepare(`BEGIN SIMPLE_ADD_NUM(:1, :2, :3); END;`)
		if err != nil {
			return err
		}
		defer func() {
			err = s.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		_, err = s.Exec(1, 2, go_ora.Out{Dest: &result})
		if err != nil {
			return err
		}
		if result != 3 {
			return fmt.Errorf("expected 3 but got %d", result)
		}
		_, err = s.Exec(4, 5, go_ora.Out{Dest: &result})
		if err != nil {
			return err
		}
		if result != 9 {
			return fmt.Errorf("expected 9 but got %d", result)
		}
		return nil
	}
	var callAddString = func(db *sql.DB) error {
		var result string
		s, err := db.Prepare("BEGIN SIMPLE_ADD_STR(:1, :2, :3); END;")
		if err != nil {
			return err
		}
		defer func() {
			err = s.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		_, err = s.Exec("hello ", "world", go_ora.Out{Dest: &result, Size: 500})
		if err != nil {
			return err
		}
		if result != "hello world" {
			return fmt.Errorf(`expected "hello world" but got "%s"`, result)
		}
		_, err = s.Exec("helloHELLO ", "worldWORLD", go_ora.Out{Dest: &result, Size: 500})
		if err != nil {
			return err
		}
		if result != "helloHELLO worldWORLD" {
			return fmt.Errorf(`expected "helloHELLO worldWORLD" but got "%s"`, result)
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
	err = create(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = drop(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = callAddInt(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = callAddString(db)
	if err != nil {
		t.Error(err)
		return
	}
}
