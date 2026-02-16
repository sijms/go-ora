package TestIssues

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v3"
	"strings"
	"testing"
	"time"
)

func TestArray(t *testing.T) {
	var expectedTime = time.Now()
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
				Date: sql.NullTime{expectedTime, true},
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
		procedure test_empty_array(p_visit_id t_visit_id, p_visit_name out t_visit_name,
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
	
		procedure test_empty_array(p_visit_id t_visit_id, p_visit_name out t_visit_name,
			p_visit_val out t_visit_val, p_visit_date out t_visit_date) as
		BEGIN
			NULL;
		END test_empty_array;
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
			if id == 1 {
				if name.String != strings.Repeat("*", 20) {
					return fmt.Errorf("expected name to be %s, got %s at id: %d", strings.Repeat("*", 20), name.String, id)
				}
				if val.Float64 != 10.0 {
					return fmt.Errorf("expected val to be %f, got %f at id: %d", 10.0, val.Float64, id)
				}
				if !isEqualTime(date.Time, expectedTime, false) {
					return fmt.Errorf("expected date to be %v, and got %v at id: %d", expectedTime, date.Time, id)
				}
			}
			if id == 3 {
				if name.Valid {
					return fmt.Errorf("expected name to be null, got %s at id: %d", name.String, id)
				}
				if val.Valid {
					return fmt.Errorf("expected val to be null, got %f at id: %d", val.Float64, id)
				}
				if date.Valid {
					return fmt.Errorf("expected date to be null, and got %v at id: %d", date.Time, id)
				}
			}
			if id == 5 {
				if name.String != strings.Repeat("*", 20) {
					return fmt.Errorf("expected name to be %s, got %s at id: %d", strings.Repeat("*", 20), name.String, id)
				}
				if val.Float64 != 2.0 {
					return fmt.Errorf("expected val to be %f, got %f at id: %d", 2.0, val.Float64, id)
				}
				if !isEqualTime(date.Time, expectedTime, false) {
					return fmt.Errorf("expected date to be %v, and got %v at id: %d", expectedTime, date.Time, id)
				}
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
			[]int{1, 3, 4, 5, 7, 8}, go_ora.Out{Dest: &nameArray, Size: 10},
			go_ora.Out{Dest: &valArray, Size: 10},
			go_ora.Out{Dest: &dateArray, Size: 10})
		if err != nil {
			return err
		}
		expectedString := sql.NullString{strings.Repeat("*", 20), true}

		if nameArray[0] != expectedString {
			return fmt.Errorf("expected name %s, got %s at position 0", expectedString.String, nameArray[0].String)
		}
		if nameArray[1].Valid {
			return fmt.Errorf("expected name to be null at position 1, got %s", nameArray[1].String)
		}
		if nameArray[2] != expectedString {
			return fmt.Errorf("expected name %s, got %s at position 2", expectedString.String, nameArray[2].String)
		}
		if nameArray[3] != expectedString {
			return fmt.Errorf("expected name %s, got %s at position 3", expectedString.String, nameArray[3].String)
		}
		if nameArray[4] != expectedString {
			return fmt.Errorf("expected name %s, got %s at position 4", expectedString.String, nameArray[4].String)
		}
		if nameArray[5] != expectedString {
			return fmt.Errorf("expected name %s, got %s at position 5", expectedString.String, nameArray[5].String)
		}

		if valArray[0].Float64 != 10.0 {
			return fmt.Errorf("expected val %v and got %v at position: 0", 10.0, valArray[0].Float64)
		}
		if valArray[1].Valid {
			return fmt.Errorf("expected null val at position 1 got: %v", valArray[1].Valid)
		}
		if valArray[2].Float64 != 2.5 {
			return fmt.Errorf("expected val %v and got %v at position: 2", 2.5, valArray[2].Float64)
		}
		if valArray[3].Float64 != 2.0 {
			return fmt.Errorf("expected val %v and got %v at position: 3", 2.0, valArray[3].Float64)
		}
		if valArray[4].Float64 != 1.43 {
			return fmt.Errorf("expected val %v and got %v at position: 4", 1.43, valArray[4].Float64)
		}
		if valArray[5].Float64 != 1.25 {
			return fmt.Errorf("expected val %v and got %v at position: 5", 1.25, valArray[5].Float64)
		}

		if !isEqualTime(dateArray[0].Time, expectedTime, false) {
			return fmt.Errorf("expected date %v, got %v at position: 0", expectedTime, dateArray[0].Time)
		}
		if dateArray[1].Valid {
			return fmt.Errorf("expected null date at position 1 got: %v", dateArray[1].Valid)
		}
		if !isEqualTime(dateArray[2].Time, expectedTime, false) {
			return fmt.Errorf("expected date: %v, got %v at position: 2", expectedTime, dateArray[2].Time)
		}
		if !isEqualTime(dateArray[3].Time, expectedTime, false) {
			return fmt.Errorf("expected date: %v, got %v at position: 3", expectedTime, dateArray[3].Time)
		}
		if !isEqualTime(dateArray[4].Time, expectedTime, false) {
			return fmt.Errorf("expected date: %v, got %v at position: 4", expectedTime, dateArray[4].Time)
		}
		if !isEqualTime(dateArray[5].Time, expectedTime, false) {
			return fmt.Errorf("expected date: %v, got %v at position: 5", expectedTime, dateArray[5].Time)
		}
		return nil
	}

	var query3 = func(db *sql.DB) error {
		var (
			nameArray []sql.NullString
			valArray  []sql.NullFloat64
			dateArray []sql.NullTime
		)
		// note size here is important and equal to max number of items that array can accommodate
		_, err := db.Exec(`BEGIN GOORA_TEMP_PKG.test_empty_array(:1, :2, :3, :4); END;`,
			[]int{1, 3, 5}, go_ora.Out{Dest: &nameArray, Size: 10},
			go_ora.Out{Dest: &valArray, Size: 10},
			go_ora.Out{Dest: &dateArray, Size: 10})
		if err != nil {
			return err
		}
		if len(nameArray) != 0 {
			return fmt.Errorf("expected empty array for name value and got: %d array size", len(nameArray))
		}
		if len(valArray) != 0 {
			return fmt.Errorf("expected empty array for numeric value and got: %d array size", len(valArray))
		}
		if len(dateArray) != 0 {
			return fmt.Errorf("expected empty array for date value and got: %d array size", len(dateArray))
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
	err = query3(db)
	if err != nil {
		t.Error(err)
		return
	}
}
