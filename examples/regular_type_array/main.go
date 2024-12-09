package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
)

type Customer struct {
	Id        int            `udt:"ID"`
	Name      string         `udt:"NAME"`
	VISITS    []sql.NullTime `udt:"VISITS"`
	PRICES    []float64      `udt:"PRICES"`
	BirthDate time.Time      `udt:"BIRTH_DATE"`
}

func execCmd(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			if len(stmts) > 1 {
				return fmt.Errorf("error: %v in execuation of stmt: %s", err, stmt)
			} else {
				return err
			}
		}
	}
	return nil
}

func createTypes(db *sql.DB) error {
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

func dropTypes(db *sql.DB) error {
	return execCmd(db, "DROP TYPE Customer",
		"drop type SLICE", "drop type IntArray", "drop type DateArray",
		"drop type ZoneArray", "drop type SLICE2", "DROP TYPE ByteArray",
		"DROP TYPE BlobArray", "DROP TYPE ClobArray", "DROP TYPE NClobArray",
	)
}

func nestedRegularArray(db *sql.DB) error {
	customer := Customer{
		Id: 10, Name: "Name1", BirthDate: time.Now(),
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
		t_customer.VISITS(x) := ADD_MONTHS(sysdate, x);
		t_customer.PRICES(x) := x + 1.25;
	end loop;
	:2 := t_customer;
END;`, go_ora.Object{Name: "Customer", Value: customer},
		go_ora.Object{Name: "Customer", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func inputPar(db *sql.DB) error {
	input := []int{1, 2, 3, 4, 5}
	input2 := []sql.NullString{
		{"test", true},
		{"test", true},
		{"", false},
		{"", false},
		{"test", true},
	}

	var output []int
	var output2 []string
	_, err := db.Exec(`
DECLARE 
	p_input IntArray;
	p_input2 SLICE;
	p_output IntArray := IntArray();
	p_output2 SLICE := SLICE();
BEGIN
	p_input := :input;
	p_input2 := :input2;
    for x in p_input.first..p_input.last loop
    	p_output.extend;
    	p_output2.extend;
    	if p_input2(x) is null then
    	   p_output2(x) := 'null';
    	else
    		p_output2(x) := p_input2(x) || p_input(x);
    	end if;
    	p_output(x) := p_input(x) + 5;
    	
	end loop;
	:output := p_output;
	:output2 := p_output2;
END;`,
		go_ora.Object{Name: "IntArray", Value: input},
		go_ora.Object{Name: "SLICE", Value: input2},
		go_ora.Out{Dest: go_ora.Object{Name: "IntArray", Value: &output}},
		go_ora.Out{Dest: go_ora.Object{Name: "SLICE", Value: &output2}},
	)
	if err != nil {
		return err
	}
	fmt.Println(output)
	fmt.Println(output2)
	return nil
}

func blobInput(db *sql.DB) error {
	array := make([]go_ora.Blob, 3)
	for x := 0; x < 3; x++ {
		array[x].Valid = true
		array[x].Data = []byte("test")
	}
	_, err := db.Exec(`
DECLARE
	p_array BlobArray;
BEGIN
	p_array := :1;
END;`, go_ora.Object{Name: "BlobArray", Value: &array})
	return err
}

func blobOutput(db *sql.DB) error {
	var array []go_ora.Blob
	_, err := db.Exec(`
DECLARE
	p_array BlobArray := BlobArray();
BEGIN
	for x in 1..3 loop
		p_array.extend;
		p_array(x) := utl_raw.cast_to_raw('TEST');
	end loop;
	:1 := p_array;
END;`, go_ora.Out{Dest: go_ora.Object{Name: "BlobArray", Value: &array}})
	if err != nil {
		return err
	}
	fmt.Println(array)
	return nil
}

func basicPars(db *sql.DB) error {
	var (
		length          int = 10
		input1, output1 []sql.NullString
		input2, output2 []sql.NullString
		input3, output3 []sql.NullInt64
		input4, output4 []sql.NullTime
		input5, output5 []sql.NullTime
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
	input5 = make([]sql.NullTime, length)
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
			input5[x] = sql.NullTime{Valid: false}
			input6[x] = nil
			input7[x] = go_ora.Blob{Valid: false}
			input8[x] = go_ora.Clob{Valid: false}
			input9[x] = go_ora.NClob{Valid: false}
		} else {
			input1[x] = sql.NullString{"varchar_", true}
			input2[x] = sql.NullString{"안녕하세요AAA_", true}
			input3[x] = sql.NullInt64{int64(x), true}
			input4[x] = sql.NullTime{time.Now(), true}
			input5[x] = sql.NullTime{Time: time.Now(), Valid: true}
			input6[x] = []byte("test_")
			input7[x] = go_ora.Blob{Data: []byte("BLOB"), Valid: true}
			input8[x] = go_ora.Clob{String: "CLOB_", Valid: true}
			input9[x] = go_ora.NClob{String: "NCLOB_안녕하세요AAA", Valid: true}
		}
	}
	fmt.Println("input for slice: ", input1)
	fmt.Println("input for slice2: ", input2)
	fmt.Println("input for IntArray: ", input3)
	fmt.Println("input for DateArray: ", input4)
	fmt.Println("input for ZoneArray: ", input5)
	fmt.Println("input for ByteArray: ", input6)
	fmt.Println("input for BlobArray: ", input7)
	fmt.Println("input for ClobArray: ", input8)
	fmt.Println("input for NClobArray: ", input9)
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
		// go_ora.Object{Name: "ZoneArray", Value: input5},
		go_ora.Object{Name: "ByteArray", Value: input6},
		go_ora.Object{Name: "BlobArray", Value: input7},
		go_ora.Object{Name: "ClobArray", Value: input8},
		go_ora.Object{Name: "NClobArray", Value: input9},

		go_ora.Object{Name: "SLICE", Value: &output1},
		go_ora.Object{Name: "SLICE2", Value: &output2},
		go_ora.Object{Name: "IntArray", Value: &output3},
		go_ora.Object{Name: "DateArray", Value: &output4},
		// go_ora.Object{Name: "ZoneArray", Value: &output5},
		go_ora.Object{Name: "ByteArray", Value: &output6},
		go_ora.Object{Name: "BlobArray", Value: &output7},
		go_ora.Object{Name: "ClobArray", Value: &output8},
		go_ora.Object{Name: "NClobArray", Value: &output9},
	)
	if err != nil {
		return err
	}
	fmt.Println("output for slice: ", output1)
	fmt.Println("output for slice2: ", output2)
	fmt.Println("output for IntArray: ", output3)
	fmt.Println("output for DateArray: ", output4)
	fmt.Println("output for ZoneArray: ", output5)
	fmt.Println("output for ByteArray: ", output6)
	fmt.Println("output for BlobArray: ", output7)
	fmt.Println("output for ClobArray: ", output8)
	fmt.Println("output for NClobArray: ", output9)
	return nil
}

func outputParString(db *sql.DB) error {
	var output []sql.NullString
	var output2 []sql.NullString
	var output3 []sql.NullInt64
	var output4 []sql.NullTime
	var output5 []sql.NullTime
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
-- 		p_zone_array.extend;
		p_raw_array.extend;
		if mod(x, 5) = 0 then 
		   p_int_array(x) := null;
		   p_array(x) := null;
		   p_array2(x) := null;
		   p_date_array(x) := null;
-- 		   p_zone_array(x) := null;
		   p_raw_array(x) := null;
		else
			p_int_array(x) := x;
			p_array(x) := 'varchar_' || x;
			p_array2(x) := '안녕하세요AAA_' || x;
		    p_date_array(x) := sysdate;
-- 		    p_zone_array(x) := systimestamp;
		    p_raw_array(x) := utl_raw.cast_to_raw('raw_' || x);
		END IF;
	end loop;
	:output := p_array;
	:output2 := p_array2;
	:output3 := p_int_array;
	:output4 := p_date_array;
-- 	:output5 := p_zone_array;
	:output6 := p_raw_array;
end;`, go_ora.Out{Dest: go_ora.Object{Name: "SLICE", Value: &output}},
		go_ora.Out{Dest: go_ora.Object{Name: "SLICE2", Value: &output2}},
		go_ora.Out{Dest: go_ora.Object{Name: "IntArray", Value: &output3}},
		go_ora.Out{Dest: go_ora.Object{Name: "DateArray", Value: &output4}},
		// go_ora.Out{Dest: go_ora.Object{Name: "ZoneArray", Value: &output5}},
		go_ora.Out{Dest: go_ora.Object{Name: "ByteArray", Value: &output6}},
	)
	if err != nil {
		return err
	}
	fmt.Println(output)
	fmt.Println(output2)
	fmt.Println(output3)
	fmt.Println(output4)
	fmt.Println(output5)
	fmt.Println(output6)
	return nil
}

func getTemporaryLobs(db *sql.DB) (int, error) {
	var r int
	err := db.QueryRow(`SELECT SUM(cache_lobs) + SUM(nocache_lobs) + SUM(abstract_lobs) 
	FROM v$temporary_lobs l, v$session s WHERE s.SID = l.SID AND s.sid = userenv('SID')`).Scan(&r)
	return r, err
}

func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open db: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close db: ", err)
		}
	}()

	err = createTypes(db)
	if err != nil {
		fmt.Println("can't create types: ", err)
		return
	}

	defer func() {
		err = dropTypes(db)
		if err != nil {
			fmt.Println("can't drop types: ", err)
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
	//err = go_ora.RegisterType(db, "timestamp with time zone", "ZoneArray", nil)
	//if err != nil {
	//	fmt.Println("can't register ZoneArray: ", err)
	//	return
	//}
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
		fmt.Println("can't output par: ", err)
		return
	}
	fmt.Println()
	err = inputPar(db)
	if err != nil {
		fmt.Println("can't input par: ", err)
		return
	}
	// get temporary lobs before
	temp, err := getTemporaryLobs(db)
	if err != nil {
		fmt.Println("can't get temporay lob before: ", err)
		return
	}
	fmt.Println()
	fmt.Println("temporary lob before: ", temp)
	err = basicPars(db)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	err = blobOutput(db)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	err = blobInput(db)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	temp, err = getTemporaryLobs(db)
	if err != nil {
		fmt.Println("can't get temporay lob after: ", err)
		return
	}
	fmt.Println("temporary lob after: ", temp)
	fmt.Println()
	err = nestedRegularArray(db)
	if err != nil {
		fmt.Println(err)
		return
	}
}
