package main

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

type test1 struct {
	Id       int64     `udt:"test_id"`
	Name     string    `udt:"test_name"`
	Data     string    `udt:"data"`
	CreateAt time.Time `udt:"created_at"`
}

func createPackage(conn *sql.DB) error {
	t := time.Now()
	sqlText := `create or replace package UDT_ARRAY_PKG AS
	-- type t_test1 is table of UDT_ARRAY_TABLE.DATA%type index by binary_integer;
	type t_id is table of UDT_ARRAY_TABLE.ID%type index by binary_integer;
	procedure test_get1(p_id t_id, p_test1 out test_type2);
    procedure test_get2(p_test test_type2);
end UDT_ARRAY_PKG;`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	sqlText = `create or replace PACKAGE BODY UDT_ARRAY_PKG AS
	procedure test_get1(p_id t_id, p_test1 out test_type2) as
		temp t_id := p_id;
		cursor tempCur is select id, DATA from UDT_ARRAY_TABLE 
			WHERE id in (select column_value from table(temp));
		tempRow tempCur%rowtype;
		idx number := 1;
	BEGIN
		p_test1 := test_type2();
		for tempRow in tempCur loop
			p_test1.extend;
			p_test1(idx) := tempRow.DATA;
			idx := idx + 1;
		end loop;
	END test_get1;
	procedure test_get2(p_test test_type2) as 
	BEGIN
		NULL;
	END test_get2;
end UDT_ARRAY_PKG;`
	_, err = conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("finish create package: ", time.Now().Sub(t))
	return nil
}

func dropPackage(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`drop package UDT_ARRAY_PKG`)
	if err != nil {
		return err
	}
	fmt.Println("Drop package: ", time.Now().Sub(t))
	return nil
}

func queryRow(conn *sql.DB) error {
	t := time.Now()
	test := test1{}
	err := conn.QueryRow(`SELECT DATA FROM UDT_ARRAY_TABLE WHERE ID=1`).Scan(&test)
	if err != nil {
		return err
	}
	fmt.Println("row: ", test)
	fmt.Println("finish query row: ", time.Now().Sub(t))
	return nil
}
func query(conn *sql.DB) error {
	t := time.Now()
	var data []test1
	_, err := conn.Exec(`BEGIN UDT_ARRAY_PKG.TEST_GET1(:1, :2); END;`, []int{1, 3, 5, 7}, go_ora.Out{Dest: &data, Size: 5})
	if err != nil {
		return err
	}
	fmt.Println("result: ", data)
	fmt.Println("finish query: ", time.Now().Sub(t))
	return nil
}

func get2(conn *sql.DB) error {
	t := time.Now()
	var data = []test1{
		{
			Id:       1,
			Name:     "name_1",
			Data:     "data",
			CreateAt: time.Now(),
		},
		{
			Id:       2,
			Name:     "name_2",
			Data:     "data",
			CreateAt: time.Now(),
		},
		{
			Id:       3,
			Name:     "name_3",
			Data:     "data",
			CreateAt: time.Now(),
		},
		{
			Id:       3,
			Name:     "name_4",
			Data:     "data",
			CreateAt: time.Now(),
		},
		{
			Id:       4,
			Name:     "name5",
			Data:     "data5",
			CreateAt: time.Now(),
		},
	}
	_, err := conn.Exec(`BEGIN UDT_ARRAY_PKG.TEST_GET2(:1); END;`, data)
	if err != nil {
		return err
	}
	fmt.Println("finish get2: ", time.Now().Sub(t))
	return nil
}
func insertData(conn *sql.DB) error {
	t := time.Now()
	sqlText := `INSERT INTO UDT_ARRAY_TABLE(ID, DATA) VALUES(:1, :2)`
	stmt, err := conn.Prepare(sqlText)
	if err != nil {
		return err
	}
	//data := make([]test1, 0, 10)
	//ids := make([]int, 0, 10)
	for x := 0; x < 10; x++ {
		_, err = stmt.Exec(x+1, test1{int64(x + 1),
			fmt.Sprintf("name_%d", x+1),
			"DATA",
			time.Now()})
		if err != nil {
			return err
		}
	}
	//_, err := conn.ExecContext(context.Background(), sqlText,
	//	[]driver.NamedValue{
	//		driver.NamedValue{"id", 0, ids},
	//		driver.NamedValue{"data", 0, data},
	//	})
	if err != nil {
		return err
	}
	fmt.Println("finish insert: ", time.Now().Sub(t))
	return nil
}
func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE UDT_ARRAY_TABLE
(
    ID  number(10, 0),
    DATA TEST_TYPE1
)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`DROP TABLE UDT_ARRAY_TABLE PURGE`)
	if err != nil {
		return err
	}
	fmt.Println("finish drop table: ", time.Now().Sub(t))
	return nil
}

func ceateUDTArray(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`CREATE or REPLACE TYPE TEST_TYPE2 AS TABLE of TEST_TYPE1`)
	if err != nil {
		return err
	}
	fmt.Println("Finish create UDT Array: ", time.Now().Sub(t))
	return nil
}
func dropUDTArray(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop type TEST_TYPE2")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop UDT Array: ", time.Now().Sub(t))
	return nil
}
func createUDT(conn *sql.DB) error {
	t := time.Now()
	sqlText := `create or replace TYPE TEST_TYPE1 IS OBJECT 
(
    TEST_ID NUMBER(10, 0),
    TEST_NAME VARCHAR2(10),
	DATA      CLOB,
    CREATED_AT DATE
)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create UDT: ", time.Now().Sub(t))
	return nil
}

func dropUDT(conn *sql.DB) error {
	t := time.Now()

	_, err := conn.Exec("drop type TEST_TYPE1")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop UDT: ", time.Now().Sub(t))
	return nil
}

func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()
	err = createUDT(conn)
	if err != nil {
		fmt.Println("can't create UDT: ", err)
		return
	}
	defer func() {
		err = dropUDT(conn)
		if err != nil {
			fmt.Println("can't drop UDT: ", err)
		}
	}()
	err = ceateUDTArray(conn)
	if err != nil {
		fmt.Println("can't create UDT array: ", err)
		return
	}
	defer func() {
		err = dropUDTArray(conn)
		if err != nil {
			fmt.Println("can't drop UDT array: ", err)
		}
	}()
	err = go_ora.RegisterType(conn, "TEST_TYPE1", "TEST_TYPE2", test1{})
	if err != nil {
		fmt.Println("can't register type: ", err)
		return
	}

	err = createTable(conn)
	if err != nil {
		fmt.Println("can't create table: ", err)
		return
	}

	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()

	//insert some data
	err = insertData(conn)
	if err != nil {
		fmt.Println("can't insert data: ", err)
		return
	}
	//create package
	err = createPackage(conn)
	if err != nil {
		fmt.Println("can't create package: ", err)
		return
	}
	defer func() {
		err = dropPackage(conn)
		if err != nil {
			fmt.Println("can't drop package: ", err)
		}
	}()
	err = query(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	//err = queryRow(conn)
	//if err != nil {
	//	fmt.Println("can't query row: ", err)
	//	return
	//}
	err = get2(conn)
	if err != nil {
		fmt.Println("can't get2: ", err)
		return
	}
}
