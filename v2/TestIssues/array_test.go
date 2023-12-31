package TestIssues

import (
	"bytes"
	"context"
	"database/sql"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
	"time"
)

func TestArray(t *testing.T) {
	var insert = func(db *sql.DB) error {
		sqlText := `INSERT INTO TTB_MAIN(ID, NAME, VAL, LDATE, DATA) VALUES(:ID, :NAME, :VAL, :LDATE, :DATA)`
		length := 10
		type TempStruct struct {
			Id   int             `db:"ID"`
			Name sql.NullString  `db:"NAME"`
			Val  sql.NullFloat64 `db:"VAL"`
			Date sql.NullTime    `db:"LDATE"`
			Data []byte          `db:"DATA"`
		}
		args := make([]TempStruct, length)
		for x := 0; x < length; x++ {

			args[x] = TempStruct{
				Id:   x + 1,
				Name: sql.NullString{strings.Repeat("*", 20), true},
				Val:  sql.NullFloat64{float64(length) / float64(x+1), true},
				Date: sql.NullTime{time.Now(), true},
				Data: bytes.Repeat([]byte{55}, 20),
			}
			if x == 2 {
				args[x].Name.Valid = false
				args[x].Val.Valid = false
				args[x].Date.Valid = false
			}
		}
		_, err := db.Exec(sqlText, args)
		return err
	}

	var createPackage = func(db *sql.DB) error {
		sqlText := `create or replace package GOORA_TEMP_PKG as
	type t_visit_id is table of TTB_MAIN.id%type index by binary_integer;
    type t_visit_name is table of TTB_MAIN.name%type index by binary_integer;
	type t_visit_val is table of TTB_MAIN.val%type index by binary_integer;
    type t_visit_date is table of TTB_MAIN.ldate%type index by binary_integer;
    
	procedure test_get1(p_visit_id t_visit_id, l_cursor out SYS_REFCURSOR);
    procedure test_get2(p_visit_id t_visit_id, p_visit_name out t_visit_name,
        p_visit_val out t_visit_val, p_visit_date out t_visit_date);
end GOORA_TEMP_PKG;
`
		err := execCmd(db, sqlText)
		if err != nil {
			return err
		}
		sqlText = `create or replace PACKAGE BODY GOORA_TEMP_PKG as
	procedure test_get1(p_visit_id t_visit_id, l_cursor out SYS_REFCURSOR) as 
		temp t_visit_id := p_visit_id;
	begin
		OPEN l_cursor for select id, name, val, ldate from TTB_MAIN 
		    where id in (select column_value from table(temp));
	end test_get1;
    
    procedure test_get2(p_visit_id t_visit_id, p_visit_name out t_visit_name,
        p_visit_val out t_visit_val, p_visit_date out t_visit_date) as
        temp t_visit_id := p_visit_id;
        cursor tempCur is select id, name, val, ldate from TTB_MAIN
            where id in (select column_value from table(temp));
        tempRow tempCur%rowtype;
        idx number := 1;
    begin
        for tempRow in tempCur loop
            p_visit_name(idx) := tempRow.name;
            p_visit_val(idx) := tempRow.val;
            p_visit_date(idx) := tempRow.ldate;
            idx := idx + 1;
        end loop;
    end test_get2;
end GOORA_TEMP_PKG;
`
		return execCmd(db, sqlText)
	}

	var dropPackage = func(db *sql.DB) error {
		return execCmd(db, "drop package GOORA_TEMP_PKG")
	}

	var query1 = func(db *sql.DB) error {
		var cursor go_ora.RefCursor
		// sql code take input array of integer and return a cursor that can be queried for result
		_, err := db.Exec(`BEGIN GOORA_TEMP_PKG.TEST_GET1(:1, :2); END;`, []int64{1, 3, 5}, sql.Out{Dest: &cursor})
		if err != nil {
			return err
		}
		rows, err := go_ora.WrapRefCursor(context.Background(), db, &cursor)
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
		}()
		var (
			id   int64
			name sql.NullString
			val  sql.NullFloat64
			date sql.NullTime
		)
		for rows.Next() {
			err = rows.Scan(&id, &name, &val, &date)
			if err != nil {
				return err
			}
			t.Log("ID: ", id, "\tName: ", name.String, "\tVal: ", val.Float64, "\tDate: ", date.Time)
		}
		return rows.Err()
	}

	var query2 = func(db *sql.DB) error {
		var (
			nameArray []sql.NullString
			valArray  []sql.NullFloat64
			dateArray []sql.NullTime
		)
		// note size here is important and equal to max number of items that array can accommodate
		_, err := db.Exec(`BEGIN GOORA_TEMP_PKG.TEST_GET2(:1, :2, :3, :4); END;`,
			[]int{1, 3, 5}, go_ora.Out{Dest: &nameArray, Size: 10},
			go_ora.Out{Dest: &valArray, Size: 10},
			go_ora.Out{Dest: &dateArray, Size: 10})
		if err != nil {
			return err
		}
		t.Log(nameArray)
		t.Log(valArray)
		t.Log(dateArray)
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
	err = createMainTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropMainTable(db)
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
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = query1(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = query2(db)
	if err != nil {
		t.Error(err)
		return
	}
}
