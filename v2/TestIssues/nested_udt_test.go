package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
	"time"
)

func TestNestedUDT(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `
create or replace type subType2 as object(
	num number(10),
	LDate date
)`, `
create or replace type subType1 as object(
	num number,
	name varchar2(100),
	Sub2 subType2
)`, `
create or replace type mainType as object(
	num number,
	sub1 subType1,
    name varchar2(100)
)`, `
create or replace type mainTypeCol as table of mainType`, `
CREATE TABLE TTB_NESTED_UDT
(
    ID NUMBER(10, 0),
    DATA1 mainType,
    SEP1 VARCHAR2(100),
    DATA2 subType1,
    SEP2 VARCHAR2(100)
)`, `
CREATE OR REPLACE PROCEDURE TP_NESTED_UDT(p_id in number, p_array out mainTypeCol) AS
	cursor tempCur is select DATA1 FROM TTB_NESTED_UDT WHERE ID <= p_id ORDER BY ID;
	tempRow tempCur%rowtype;
	idx number := 1;
BEGIN
	p_array := mainTypeCol();
	for tempRow in tempCur loop
		p_array.extend;
		p_array(idx) := tempRow.DATA1;
		idx := idx + 1;
	end loop;
END TP_NESTED_UDT;`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db,
			"DROP PROCEDURE TP_NESTED_UDT",
			"DROP TABLE TTB_NESTED_UDT",
			"DROP TYPE mainTypeCol",
			"DROP TYPE mainType",
			"DROP TYPE subType1",
			"DROP TYPE subType2")
	}
	type sub2 struct {
		Num   int       `udt:"NUM"`
		LDate time.Time `udt:"LDate"`
	}
	type sub1 struct {
		Num  float64 `udt:"NUM"`
		Name string  `udt:"NAME"`
		Sub2 sub2    `udt:"SUB2"`
	}
	type mainS struct {
		Num  float64        `udt:"NUM"`
		Sub1 sub1           `udt:"SUB1"`
		S1   sql.NullString `udt:"NAME"`
	}
	now := time.Now()
	expectedTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local)
	var isEqual = func(s1, s2 mainS) bool {
		time1 := s1.Sub1.Sub2.LDate
		time2 := s2.Sub1.Sub2.LDate
		if time1.Year() == time2.Year() &&
			time1.Month() == time2.Month() &&
			time1.Day() == time2.Day() &&
			time1.Hour() == time2.Hour() &&
			time1.Minute() == time2.Minute() &&
			time1.Second() == time2.Second() {
			s1.Sub1.Sub2.LDate = time2
			if s1 != s2 {
				return false
			} else {
				return true
			}
		}
		return false
	}
	var insert = func(db *sql.DB) error {
		length := 100
		type insertData struct {
			Id    int    `db:"ID"`
			Data1 mainS  `db:"DATA1"`
			Sep1  string `db:"SEP1"`
			Data2 sub1   `db:"DATA2"`
			Sep2  string `db:"SEP2"`
		}
		baseValue := 1.1
		data := make([]insertData, length)
		for index, _ := range data {
			data[index].Id = index + 1
			data[index].Sep1 = strings.Repeat("-", 100)
			data[index].Sep2 = strings.Repeat("-", 100)
			data[index].Data1.Num = baseValue + float64(index)
			if index%5 == 0 {
				data[index].Data1.S1 = sql.NullString{"", false}
			} else {
				data[index].Data1.S1 = sql.NullString{"NAME", true}
			}
			data[index].Data1.Sub1.Num = baseValue + float64(index+1)
			data[index].Data1.Sub1.Name = "TEST"
			data[index].Data1.Sub1.Sub2.Num = index + 2
			data[index].Data1.Sub1.Sub2.LDate = time.Now()
			data[index].Data2.Num = baseValue + float64(index+3)
			data[index].Data2.Name = "TEST1"
			data[index].Data2.Sub2.Num = index + 3
			data[index].Data2.Sub2.LDate = time.Now()
		}
		_, err := db.Exec(`INSERT INTO TTB_NESTED_UDT(ID, DATA1, SEP1, DATA2, SEP2) 
			VALUES(:ID, :DATA1, :SEP1, :DATA2, :SEP2)`, data)
		return err
	}
	var queryTable = func(db *sql.DB) error {
		rows, err := db.Query("SELECT ID, DATA1, SEP1, DATA2, SEP2 FROM TTB_NESTED_UDT")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				fmt.Println("can't close rows: ", err)
			}
		}()
		var (
			id    int
			data1 mainS
			sep1  string
			data2 sub1
			sep2  string
		)
		for rows.Next() {
			err = rows.Scan(&id, &data1, &sep1, &data2, &sep2)
			if err != nil {
				return err
			}
			t.Log("id: ", id, "\tsep1: ", sep1, "\tsep2: ", sep2)
			t.Log("data1: ", data1, "\tdata2: ", data2)
		}
		return rows.Err()
	}
	var query = func(db *sql.DB) error {
		got := mainS{}
		err := db.QueryRow("SELECT mainType(5, subType1(1, 'test', subType2(33, :1)), 'NAME') from dual", expectedTime).Scan(&got)
		if err != nil {
			return err
		}
		expected := mainS{
			Num: 5,
			Sub1: sub1{
				Num:  1,
				Name: "test",
				Sub2: sub2{
					Num:   33,
					LDate: expectedTime,
				},
			},
			S1: sql.NullString{"NAME", true},
		}
		if !isEqual(expected, got) {
			return fmt.Errorf("expected: %v and got %v", expected, got)
		}
		return nil
	}
	var inputPar = func(db *sql.DB) error {
		var input = mainS{
			Num: 7,
			Sub1: sub1{
				Num:  8,
				Name: "NAME",
				Sub2: sub2{
					Num:   9,
					LDate: expectedTime,
				},
			},
			S1: sql.NullString{"test", true},
		}
		var output = mainS{}
		_, err := db.Exec(`
	DECLARE
		v_main mainType;
	BEGIN
		v_main := :1;
		v_main.Num := v_main.Num + 10;
		:2 := v_main;
    END;`, input, go_ora.Out{Dest: &output})
		if err != nil {
			return err
		}
		expected := mainS{
			Num: 17,
			Sub1: sub1{
				Num:  8,
				Name: "NAME",
				Sub2: sub2{
					Num:   9,
					LDate: expectedTime,
				},
			},
			S1: sql.NullString{"test", true},
		}
		if !isEqual(expected, output) {
			return fmt.Errorf("expected: %v and got %v", expected, output)
		}
		return nil
	}
	var callProc = func(db *sql.DB) error {
		var output []mainS
		_, err := db.Exec(`BEGIN TP_NESTED_UDT(10, :1); END;`, go_ora.Out{Dest: &output, Size: 100})
		if err != nil {
			return err
		}
		if len(output) != 10 {
			return fmt.Errorf("expected array size: %d and got: %d", 10, len(output))
		}
		for x, item := range output {
			t.Logf("item %d: %v", x, item)
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
	err = go_ora.RegisterType(db, "subType2", "", sub2{})
	if err != nil {
		t.Error(err)
		return
	}
	err = go_ora.RegisterType(db, "subType1", "", sub1{})
	if err != nil {
		t.Error(err)
		return
	}
	err = go_ora.RegisterType(db, "mainType", "mainTypeCol", mainS{})
	if err != nil {
		t.Error(err)
		return
	}
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = inputPar(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = queryTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = callProc(db)
	if err != nil {
		t.Error(err)
		return
	}
}
