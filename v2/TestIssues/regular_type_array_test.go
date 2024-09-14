package TestIssues

import (
	"bytes"
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
	"time"
)

func TestRegularTypeArray(t *testing.T) {
	type Customer struct {
		Id        int            `udt:"ID"`
		Name      string         `udt:"NAME"`
		VISITS    []sql.NullTime `udt:"VISITS"`
		PRICES    []float64      `udt:"PRICES"`
		BirthDate time.Time      `udt:"BIRTH_DATE"`
	}
	var createTypes = func(db *sql.DB) error {
		return execCmd(db, "create or replace TYPE SLICE AS TABLE OF varchar2(500)",
			"create or replace type IntArray as table of number(10, 2)",
			"create or replace type DateArray as table of DATE",
			"create or replace type ZoneArray as table of timestamp with time zone",
			"create or replace type SLICE2 as table of nvarchar2(500)",
			"create or replace type ByteArray as table of raw(500)",
			"create or replace type BlobArray as table of BLOB",
			"create or replace type ClobArray as table of CLOB",
			"create or replace type NClobArray as table of NCLOB",
			`create or replace type Customer as object (
ID number(10),
NAME varchar2(100),
VISITS DateArray,
PRICES IntArray,
BIRTH_DATE DATE
)`,
		)
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db, "DROP TYPE Customer",
			"drop type SLICE", "drop type IntArray", "drop type DateArray",
			"drop type ZoneArray", "drop type SLICE2", "DROP TYPE ByteArray",
			"DROP TYPE BlobArray", "DROP TYPE ClobArray", "DROP TYPE NClobArray",
		)
	}

	var outputParString = func(db *sql.DB) error {
		var output []sql.NullString
		var output2 []sql.NullString
		var output3 []sql.NullInt64
		var output4 []sql.NullTime
		//var output5 []sql.NullTime
		var output6 [][]byte
		_, err := db.Exec(`
DECLARE
	p_array slice := slice();
	p_int_array IntArray := IntArray();
	p_date_array DateArray := DateArray();
-- 	p_zone_array ZoneArray := ZoneArray();
	p_raw_array ByteArray := ByteArray();
	p_array2 SLICE2 := SLICE2();
BEGIN
	for x in 1..10 loop
		p_array.extend;
		p_array2.extend;
		p_int_array.extend;
		p_date_array.extend;
		p_raw_array.extend;
		if mod(x, 5) = 0 then 
		   p_int_array(x) := null;
		   p_array(x) := null;
		   p_array2(x) := null;
		   p_date_array(x) := null;
		   p_raw_array(x) := null;
		else
			p_int_array(x) := x;
			p_array(x) := 'varchar_' || x;
			p_array2(x) := '안녕하세요AAA_' || x;
		    p_date_array(x) := sysdate;
		    p_raw_array(x) := utl_raw.cast_to_raw('raw_' || x);
		END IF;
	end loop;
	:output := p_array;
	:output2 := p_array2;
	:output3 := p_int_array;
	:output4 := p_date_array;
	:output6 := p_raw_array;
end;`, go_ora.Out{Dest: go_ora.Object{Name: "SLICE", Value: &output}},
			go_ora.Out{Dest: go_ora.Object{Name: "SLICE2", Value: &output2}},
			go_ora.Out{Dest: go_ora.Object{Name: "IntArray", Value: &output3}},
			go_ora.Out{Dest: go_ora.Object{Name: "DateArray", Value: &output4}},
			go_ora.Out{Dest: go_ora.Object{Name: "ByteArray", Value: &output6}},
		)
		if err != nil {
			return err
		}
		for index, temp := range output {
			if (index+1)%5 > 0 {
				expected := fmt.Sprintf("varchar_%d", index+1)
				if temp.String != expected {
					return fmt.Errorf("expected: %s and got: %s", expected, temp.String)
				}
			} else {
				if temp.Valid {
					return fmt.Errorf("expected null string and got: %s", temp.String)
				}
			}
		}
		for index, temp := range output2 {
			if (index+1)%5 > 0 {
				expected := fmt.Sprintf("안녕하세요AAA_%d", index+1)
				if temp.String != expected {
					return fmt.Errorf("expected: %s and got: %s", expected, temp.String)
				}
			} else {
				if temp.Valid {
					return fmt.Errorf("expected null string and got: %s", temp.String)
				}
			}
		}
		for index, temp := range output3 {
			if (index+1)%5 > 0 {
				if temp.Int64 != int64(index+1) {
					return fmt.Errorf("expected: %d and got: %d", index+1, temp.Int64)
				}
			} else {
				if temp.Valid {
					return fmt.Errorf("expected null int and got: %d", temp.Int64)
				}
			}
		}
		fmt.Println(output4)
		for index, temp := range output6 {
			if (index+1)%5 > 0 {
				expected := []byte(fmt.Sprintf("raw_%d", index+1))
				if !bytes.Equal(temp, expected) {
					return fmt.Errorf("expected: %v and got: %v", expected, temp)
				}
			} else {
				if temp != nil {
					return fmt.Errorf("expected null bytes and got: %v", temp)
				}
			}
		}
		return nil
	}

	var basicPars = func(db *sql.DB) error {
		var (
			length          = 10
			input1, output1 []sql.NullString
			input2, output2 []sql.NullString
			input3, output3 []sql.NullInt64
			input4, output4 []sql.NullTime
			//input5, output5 []sql.NullTime
			input6, output6 [][]byte
			input7, output7 []go_ora.Blob
			input8, output8 []go_ora.Clob
			input9, output9 []go_ora.NClob
		)
		// fill input arrays
		input1 = make([]sql.NullString, length)
		input2 = make([]sql.NullString, length)
		input3 = make([]sql.NullInt64, length)
		input4 = make([]sql.NullTime, length)
		//input5 = make([]sql.NullTime, length)
		input6 = make([][]byte, length)
		input7 = make([]go_ora.Blob, length)
		input8 = make([]go_ora.Clob, length)
		input9 = make([]go_ora.NClob, length)
		for x := 0; x < length; x++ {
			if x%5 == 0 {
				input1[x] = sql.NullString{"", false}
				input2[x] = sql.NullString{"", false}
				input3[x] = sql.NullInt64{0, false}
				input4[x] = sql.NullTime{Valid: false}
				//input5[x] = sql.NullTime{Valid: false}
				input6[x] = nil
				input7[x] = go_ora.Blob{Data: nil}
				input8[x] = go_ora.Clob{Valid: false}
				input9[x] = go_ora.NClob{Valid: false}
			} else {
				input1[x] = sql.NullString{"varchar_", true}
				input2[x] = sql.NullString{"안녕하세요AAA_", true}
				input3[x] = sql.NullInt64{int64(x), true}
				input4[x] = sql.NullTime{time.Now(), true}
				//input5[x] = sql.NullTime{Time: time.Now(), Valid: true}
				input6[x] = []byte("test_")
				input7[x] = go_ora.Blob{Data: []byte("BLOB")}
				input8[x] = go_ora.Clob{String: "CLOB_", Valid: true}
				input9[x] = go_ora.NClob{String: "NCLOB_안녕하세요AAA", Valid: true}
			}
		}
		_, err := db.Exec(`
DECLARE
	global_length number;
	-- define input arrays
	p1 SLICE;
	p2 SLICE2;
	p3 IntArray;
	p4 DateArray;
	p5 ZoneArray;
	p6 ByteArray;
	p7 BlobArray;
	p8 ClobArray;
	p9 NClobArray;
BEGIN
	global_length := :1;
	p1 := :input1;
	p2 := :input2;
	p3 := :input3;
	p4 := :input4;
-- 	p5 := :input5;
	p6 := :input6;
	p7 := :input7;
	p8 := :input8;
	p9 := :input9;
	for x in 1..global_length loop
		if p1(x) is not null then
			p1(x) := p1(x) || x;
		end if;
		if p2(x) is not null then
			p2(x) := p2(x) || x; 
		end if;
		if p3(x) is not null then
		   p3(x) := p3(x) + 10;
		end if;
		if p4(x) is not null then
		   p4(x) := add_Months(p4(x), x);
		end if;
		if p8(x) is not null then
		   p8(x) := TO_CLOB(TO_CHAR(p8(x)) || x);
		end if;
	end loop;
	:output1 := p1;
	:output2 := p2;
	:output3 := p3;
	:output4 := p4;
-- 	:output5 := p5;
	:output6 := p6;
	:output7 := p7;
	:output8 := p8;
	:output9 := p9;
END;`, length,
			go_ora.Object{Name: "SLICE", Value: input1},
			go_ora.Object{Name: "SLICE2", Value: input2},
			go_ora.Object{Name: "IntArray", Value: input3},
			go_ora.Object{Name: "DateArray", Value: input4},
			//go_ora.Object{Name: "ZoneArray", Value: input5},
			go_ora.Object{Name: "ByteArray", Value: input6},
			go_ora.Object{Name: "BlobArray", Value: input7},
			go_ora.Object{Name: "ClobArray", Value: input8},
			go_ora.Object{Name: "NClobArray", Value: input9},

			go_ora.Object{Name: "SLICE", Value: &output1},
			go_ora.Object{Name: "SLICE2", Value: &output2},
			go_ora.Object{Name: "IntArray", Value: &output3},
			go_ora.Object{Name: "DateArray", Value: &output4},
			//go_ora.Object{Name: "ZoneArray", Value: &output5},
			go_ora.Object{Name: "ByteArray", Value: &output6},
			go_ora.Object{Name: "BlobArray", Value: &output7},
			go_ora.Object{Name: "ClobArray", Value: &output8},
			go_ora.Object{Name: "NClobArray", Value: &output9},
		)
		if err != nil {
			return err
		}
		nullErrorFormat := "error in SLICE %s expected null and got: %v"
		errorFormat := "error in %s type expected: %v and got: %v"
		for x := 0; x < length; x++ {
			if x%5 == 0 {
				if output1[x].Valid {
					return fmt.Errorf(nullErrorFormat, "SLICE", output1[x].String)
				}
				if output2[x].Valid {
					return fmt.Errorf(nullErrorFormat, "SLICE2", output2[x].String)
				}
				if output3[x].Valid {
					return fmt.Errorf(nullErrorFormat, "IntArray", output3[x].Int64)
				}
				if output4[x].Valid {
					return fmt.Errorf(nullErrorFormat, "DateArray", output4[x].Time)
				}
				if output6[x] != nil {
					return fmt.Errorf(nullErrorFormat, "DateArray", output6)
				}
				if output7[x].Data != nil {
					return fmt.Errorf(nullErrorFormat, "BlobArray", output7[x].Data)
				}
				if output8[x].Valid {
					return fmt.Errorf(nullErrorFormat, "ClobArray", output8[x].String)
				}
				if output9[x].Valid {
					return fmt.Errorf(nullErrorFormat, "NClobArray", output9[x].String)
				}
			} else {
				if !strings.EqualFold(output1[x].String, fmt.Sprintf("varchar_%d", x+1)) {
					return fmt.Errorf(errorFormat, "SLICE", fmt.Sprintf("varchar_%d", x+1), output1[x].String)
				}
				if !strings.EqualFold(output2[x].String, fmt.Sprintf("안녕하세요AAA_%d", x+1)) {
					return fmt.Errorf(errorFormat, "SLICE2", fmt.Sprintf("안녕하세요AAA_%d", x+1), output2[x].String)
				}
				if input3[x].Int64+int64(10) != output3[x].Int64 {
					return fmt.Errorf(errorFormat, "IntArray", input3[x].Int64+int64(10), output3[x].Int64)
				}
				if !isEqualTime(output4[x].Time, input4[x].Time.AddDate(0, x+1, 0), false) {
					return fmt.Errorf(errorFormat, "DateArray", input4[x].Time.AddDate(0, x+1, 0), output4[x].Time)
				}
				if !bytes.Equal(input6[x], output6[x]) {
					return fmt.Errorf(errorFormat, "ByteArray", input6[x], output6[x])
				}
				if !bytes.Equal(input7[x].Data, output7[x].Data) {
					return fmt.Errorf(errorFormat, "BlobArray", input7[x].Data, output7[x].Data)
				}
				if !strings.EqualFold(fmt.Sprintf("%s%d", input8[x].String, x+1), output8[x].String) {
					return fmt.Errorf(errorFormat, "ClobArray", fmt.Sprintf("%s%d", input8[x].String, x+1), output8[x].String)
				}
				if !strings.EqualFold(input9[x].String, output9[x].String) {
					return fmt.Errorf(errorFormat, "NClobArray", input9[x].String, output9[x].String)
				}
			}
		}
		return nil
	}

	var getTempLobs = func(db *sql.DB) (int, error) {
		var r int
		err := db.QueryRow(`SELECT SUM(cache_lobs) + SUM(nocache_lobs) + SUM(abstract_lobs) 
	FROM v$temporary_lobs l, v$session s WHERE s.SID = l.SID AND s.sid = userenv('SID')`).Scan(&r)
		return r, err
	}
	var nestedRegularArray = func(db *sql.DB) error {
		refDate := time.Now()
		var customer = Customer{
			Id: 10, Name: "Name1", BirthDate: refDate,
		}
		var output Customer
		_, err := db.Exec(`
DECLARE
	t_customer Customer;
BEGIN
	t_customer := :1;
	t_customer.VISITS := DateArray();
	t_customer.PRICES := IntArray();
	for x in 1..5 loop
		t_customer.VISITS.extend;
		t_customer.PRICES.extend;
		t_customer.VISITS(x) := ADD_MONTHS(:2, x);
		t_customer.PRICES(x) := x + 1.25;
	end loop;
	:3 := t_customer;
END;`, go_ora.Object{Name: "Customer", Value: customer}, refDate,
			go_ora.Object{Name: "Customer", Value: &output})
		if err != nil {
			return err
		}
		if output.Id != 10 {
			return fmt.Errorf("expected Customer id: %d and got: %d", 10, output.Id)
		}
		if output.Name != "Name1" {
			return fmt.Errorf("expected Customer Name: %s and got: %s", "Name1", output.Name)
		}
		if !isEqualTime(output.BirthDate, refDate, false) {
			return fmt.Errorf("expected Customer BirthDate: %v and got: %v", refDate, output.BirthDate)
		}
		if len(output.VISITS) != 5 {
			return fmt.Errorf("expected to return 5 visits and got: %d", len(output.VISITS))
		}
		if len(output.PRICES) != 5 {
			return fmt.Errorf("expected to return 5 prices and got: %d", len(output.PRICES))
		}
		for x := 0; x < 5; x++ {
			if !isEqualTime(output.VISITS[x].Time, refDate.AddDate(0, x+1, 0), false) {
				return fmt.Errorf("expected visit date: %v and got %v", refDate.AddDate(0, x+1, 0), output.VISITS[x].Time)
			}
			if output.PRICES[x] != float64(x+1)+1.25 {
				return fmt.Errorf("expected visit price: %v and got: %v", float64(x)+1.25, output.PRICES[x])
			}
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

	err = go_ora.RegisterType(db, "varchar2", "SLICE", nil)
	if err != nil {
		fmt.Println("can't register SLICE: ", err)
		return
	}
	err = go_ora.RegisterType(db, "number", "IntArray", nil)
	if err != nil {
		fmt.Println("can't register IntArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "date", "DateArray", nil)
	if err != nil {
		fmt.Println("can't register DateArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "nvarchar2", "SLICE2", nil)
	if err != nil {
		fmt.Println("can't register SLICE2: ", err)
		return
	}
	err = go_ora.RegisterType(db, "raw", "ByteArray", nil)
	if err != nil {
		fmt.Println("can't register ByteArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "blob", "BlobArray", nil)
	if err != nil {
		fmt.Println("can't register BlobArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "clob", "ClobArray", nil)
	if err != nil {
		fmt.Println("can't register ClobArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "nclob", "NClobArray", nil)
	if err != nil {
		fmt.Println("can't register NClobArray: ", err)
		return
	}
	err = go_ora.RegisterType(db, "customer", "", Customer{})
	if err != nil {
		fmt.Println("can't register customer: ", err)
		return
	}
	err = outputParString(db)
	if err != nil {
		t.Error(err)
		return
	}
	lobsBefore, err := getTempLobs(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = basicPars(db)
	if err != nil {
		t.Error(err)
		return
	}
	lobsAfter, err := getTempLobs(db)
	if err != nil {
		t.Error(err)
		return
	}
	if lobsAfter > lobsBefore {
		t.Errorf("lobs leak before call: %d and after call: %d", lobsBefore, lobsAfter)
		return
	}
	err = nestedRegularArray(db)
	if err != nil {
		t.Error(err)
		return
	}
}
