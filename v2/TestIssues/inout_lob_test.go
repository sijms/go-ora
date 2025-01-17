package TestIssues

import (
	"database/sql"
	"errors"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
)

func TestInoutLob(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `create or replace procedure proc_626(
	par_01 in out clob,
	par_02 out number,
	par_03 out varchar2,
	par_04 out varchar2) AS
BEGIN
	par_01 := par_01 || ' output string';
	par_02 := 15;
	par_03 := 'this is a test1';
	par_04 := 'this is a test2';
END;`)
	}

	var drop = func(db *sql.DB) error {
		return execCmd(db, `drop procedure proc_626`)
	}

	var call = func(db *sql.DB) error {
		var (
			par_01 go_ora.Clob
			par_02 int
			par_03 string
			par_04 string
		)
		par_01 = go_ora.Clob{
			String: strings.Repeat("a", 0x8010),
			Valid:  true,
		}
		_, err := db.Exec("BEGIN proc_626(:par_01, :par_02, :par_03, :par_04); END;",
			sql.Named("par_01", go_ora.Out{Dest: &par_01, Size: 50000, In: true}),
			sql.Named("par_02", go_ora.Out{Dest: &par_02}),
			sql.Named("par_03", go_ora.Out{Dest: &par_03, Size: 10000}),
			sql.Named("par_04", go_ora.Out{Dest: &par_04, Size: 10000}))
		if err != nil {
			return err
		}
		if par_01.String != strings.Repeat("a", 0x8010)+" output string" {
			return errors.New("parameter par_01 return unexpected value")
		}
		if par_02 != 15 {
			return fmt.Errorf("par_02 expected: %d and got: %d", 15, par_02)
		}
		if par_03 != "this is a test1" {
			return fmt.Errorf(`par_03 expected: "this is a test1", got: %s`, par_03)
		}
		if par_04 != "this is a test2" {
			return fmt.Errorf(`par_04 expected: "this is a test2", got: %s`, par_04)
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
	err = call(db)
	if err != nil {
		t.Error(err)
		return
	}

}
