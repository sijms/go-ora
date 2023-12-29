package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
)

func TestCascadeType(t *testing.T) {
	type HkGoraIdObj struct {
		Id int64 `udt:"ID"`
	}

	type HkGoraTestSubObj struct {
		SubId   int64  `udt:"SUB_ID"`
		SubData string `udt:"SUB_DATA"`
	}
	type HkGoraTestObj struct {
		Id      int64              `udt:"ID"`
		Name    string             `udt:"NAME"`
		SubInfo []HkGoraTestSubObj `udt:"SUB_INFO"`
	}
	var createTypes = func(db *sql.DB) error {
		err := execCmd(db,
			"create or replace type HK_GORA_ID_OBJ as object(ID number(22,0));",
			"create or replace type HK_GORA_ID_COLL as table of HK_GORA_ID_OBJ;",
			"create or replace type HK_GORA_TEST_SUB_OBJ as object(SUB_ID number(22,0), SUB_DATA varchar2(2000))",
			"create or replace type HK_GORA_TEST_SUB_COLL as table of HK_GORA_TEST_SUB_OBJ",
			"create or replace type HK_GORA_TEST_OBJ as object(ID number(22,0), NAME varchar2(2000), SUB_INFO HK_GORA_TEST_SUB_COLL);",
			"create or replace type HK_GORA_TEST_COLL as table of HK_GORA_TEST_OBJ;")
		if err != nil {
			return err
		}
		return nil
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db,
			"drop type HK_GORA_TEST_COLL",
			"drop type HK_GORA_TEST_OBJ",
			"drop type HK_GORA_TEST_SUB_COLL",
			"drop type HK_GORA_TEST_SUB_OBJ",
			"drop type HK_GORA_ID_COLL",
			"drop type HK_GORA_ID_OBJ")
	}
	var registerTypes = func(db *sql.DB) error {
		err := go_ora.RegisterType(db, "HK_GORA_ID_OBJ", "HK_GORA_ID_COLL", HkGoraIdObj{})
		if err != nil {
			return err
		}
		err = go_ora.RegisterType(db, "HK_GORA_TEST_SUB_OBJ", "HK_GORA_TEST_SUB_COLL", HkGoraTestSubObj{})
		if err != nil {
			return err
		}
		err = go_ora.RegisterType(db, "HK_GORA_TEST_OBJ", "HK_GORA_TEST_COLL", HkGoraTestObj{})
		if err != nil {
			return err
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
			t.Error()
		}
	}()

	err = createTypes(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTypes(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = registerTypes(db)
	if err != nil {
		t.Error(err)
	}
}
