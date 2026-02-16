package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
)

func TestIssue408(t *testing.T) {
	type HkGoraIdObj struct {
		Id int64 `udt:"ID"`
	}

	type HkGoraTestObj struct {
		Id   int64  `udt:"ID"`
		Name string `udt:"NAME"`
	}

	var createTypes = func(db *sql.DB) error {
		return execCmd(db, "create or replace type HK_GORA_ID_OBJ as object(ID number(22,0));",
			"create or replace type HK_GORA_ID_COLL as table of HK_GORA_ID_OBJ;",
			"create or replace type HK_GORA_TEST_OBJ as object(ID number(22,0), NAME varchar2(2000));",
			"create or replace type HK_GORA_TEST_COLL as table of HK_GORA_TEST_OBJ;")
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db, "drop type HK_GORA_ID_COLL",
			"drop type HK_GORA_ID_OBJ",
			"drop type HK_GORA_TEST_COLL",
			"drop type HK_GORA_TEST_OBJ")
	}
	var createPackage = func(db *sql.DB) error {
		return execCmd(db, `
create or replace package HK_GORA_TEST_PCK as
   function F_GetData(P_Ids HK_GORA_ID_COLL, P_AnzChars number) return HK_GORA_TEST_COLL;
end;`, `
create or replace package body HK_GORA_TEST_PCK as
   function F_GetData(P_Ids HK_GORA_ID_COLL, P_AnzChars number) return HK_GORA_TEST_COLL is
      L_Coll HK_GORA_TEST_COLL := HK_GORA_TEST_COLL();
   begin
	  
      for r in (with data as(
                  select 1 as ID, rpad('a',P_AnzChars,'a') as DATA from dual union all
                  select 2      , 'Data2'                   from dual
               )
               select HK_GORA_TEST_OBJ(ID, DATA) as HK_GORA_DATA
                 from data 
                where ID in (select ID from table(P_Ids))) loop
         L_Coll.Extend;
         L_Coll(L_Coll.LAST) := r.HK_GORA_DATA;
      end loop;
      return L_Coll;
   end;
end;`)
	}
	var dropPackage = func(db *sql.DB) error {
		return execCmd(db, "DROP PACKAGE HK_GORA_TEST_PCK")
	}

	var registerTypes = func(db *sql.DB) error {
		err := go_ora.RegisterType(db, "HK_GORA_ID_OBJ", "HK_GORA_ID_COLL", HkGoraIdObj{})
		if err != nil {
			return err
		}
		err = go_ora.RegisterType(db, "HK_GORA_TEST_OBJ", "HK_GORA_TEST_COLL", HkGoraTestObj{})
		if err != nil {
			return err
		}
		return nil
	}
	var executeTest = func(db *sql.DB, anzChars int) error {
		var res []HkGoraTestObj
		_, err := db.Exec(`BEGIN :1 := HK_GORA_TEST_PCK.F_GetData(:2, :3); END;`,
			go_ora.Out{Dest: &res, Size: 2},
			[]HkGoraIdObj{{Id: 1}, {Id: 2}},
			anzChars)
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
			t.Error(err)
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
	err = executeTest(db, 238)
	err = executeTest(db, 255)
}
