package main

import (
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func bulkInsert(databaseUrl string) error {
	conn, err := go_ora.NewConnection(databaseUrl)
	if err != nil {
		return err
	}
	err = conn.Open()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	t := time.Now()
	sqlText := `INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE) VALUES(:1, :2, :3, :4)`
	rowNum := 100
	visitID := make([]driver.Value, rowNum)
	nameText := make([]driver.Value, rowNum)
	val := make([]driver.Value, rowNum)
	date := make([]driver.Value, rowNum)
	initalVal := 1.1
	for index := 0; index < rowNum; index++ {
		visitID[index] = index + 1
		nameText[index] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		val[index] = initalVal
		date[index] = time.Now().AddDate(0, index, 0)
		initalVal += 1.1
		//if index%5 == 0 {
		//	_, err = stmt.Exec(index, nameText, val, nil)
		//} else {
		//	_, err = stmt.Exec(index, nameText, val, time.Now())
		//}
		//if err != nil {
		//	return err
		//}
		//val += 1.1
	}
	result, err := conn.BulkInsert(sqlText, rowNum, visitID, nameText, val, date)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("%d rows inserted: %v\n", rowsAffected, time.Now().Sub(t))
	return nil
}
func dropPackage(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`drop package GOORA_TEMP_PKG`)
	if err != nil {
		return err
	}
	fmt.Println("Drop package: ", time.Now().Sub(t))
	return nil
}
func createPackage(conn *sql.DB) error {
	t := time.Now()
	sqlText := `create or replace package GOORA_TEMP_PKG as
	type t_visit_id is table of GOORA_TEMP_VISIT.visit_id%type index by binary_integer;
    type t_visit_name is table of GOORA_TEMP_VISIT.name%type index by binary_integer;
	type t_visit_val is table of GOORA_TEMP_VISIT.val%type index by binary_integer;
    type t_visit_date is table of GOORA_TEMP_VISIT.visit_date%type index by binary_integer;
    
	procedure test_get1(p_visit_id t_visit_id, l_cursor out SYS_REFCURSOR);
    procedure test_get2(p_visit_id t_visit_id, p_visit_name out t_visit_name,
        p_visit_val out t_visit_val, p_visit_date out t_visit_date);
end GOORA_TEMP_PKG;
`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	sqlText = `create or replace PACKAGE BODY GOORA_TEMP_PKG as
	procedure test_get1(p_visit_id t_visit_id, l_cursor out SYS_REFCURSOR) as 
		temp t_visit_id := p_visit_id;
	begin
		OPEN l_cursor for select visit_id, name, val, visit_date from GOORA_TEMP_VISIT 
		    where visit_id in (select column_value from table(temp));
	end test_get1;
    
    procedure test_get2(p_visit_id t_visit_id, p_visit_name out t_visit_name,
        p_visit_val out t_visit_val, p_visit_date out t_visit_date) as
        temp t_visit_id := p_visit_id;
        cursor tempCur is select visit_id, name, val, visit_date from GOORA_TEMP_VISIT
            where visit_id in (select column_value from table(temp));
        tempRow tempCur%rowtype;
        idx number := 1;
    begin
        for tempRow in tempCur loop
            p_visit_name(idx) := tempRow.name;
            p_visit_val(idx) := tempRow.val;
            p_visit_date(idx) := tempRow.visit_date;
            idx := idx + 1;
        end loop;
    end test_get2;
end GOORA_TEMP_PKG;
`
	_, err = conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create package: ", time.Now().Sub(t))
	return nil
}

// function do something like "SELECT col1, col2, ... FROM table WHERE col1 in (values)"
func query1(conn *sql.DB) error {
	t := time.Now()
	var cursor go_ora.RefCursor
	// sql code take input array of integer and return a cursor that can be queried for result
	_, err := conn.Exec(`BEGIN GOORA_TEMP_PKG.TEST_GET1(:1, :2); END;`, []int64{1, 3, 5}, sql.Out{Dest: &cursor})
	if err != nil {
		return err
	}
	defer func() {
		err = cursor.Close()
		if err != nil {
			fmt.Println("Can't close RefCursor", err)
		}
	}()
	err = queryCursor(&cursor)
	if err != nil {
		return err
	}
	fmt.Println("Finish Query1: ", time.Now().Sub(t))
	return nil
}

// this function take one input array and return 3 output arraies of different types
func query2(conn *sql.DB) error {
	t := time.Now()
	var (
		nameArray []string
		valArray  []float32
		dateArray []sql.NullTime
	)
	// note size here is important and equal to max number of items that array can accommodate
	_, err := conn.Exec(`BEGIN GOORA_TEMP_PKG.TEST_GET2(:1, :2, :3, :4); END;`,
		[]int{1, 3, 5}, go_ora.Out{Dest: &nameArray, Size: 10},
		go_ora.Out{Dest: &valArray, Size: 10},
		go_ora.Out{Dest: &dateArray, Size: 10})
	if err != nil {
		return err
	}
	fmt.Println(nameArray)
	fmt.Println(valArray)
	fmt.Println(dateArray)
	fmt.Println("Finish Query2: ", time.Now().Sub(t))
	return nil
}

func queryCursor(cursor *go_ora.RefCursor) error {
	t := time.Now()
	rows, err := cursor.Query()
	if err != nil {
		return err
	}
	var (
		id   int64
		name string
		val  float32
		date sql.NullTime
	)
	for rows.Next_() {
		err = rows.Scan(&id, &name, &val, &date)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
	}
	fmt.Println("Finish query RefCursor: ", time.Now().Sub(t))
	return rows.Err()
}

func usage() {
	fmt.Println()
	fmt.Println("array")
	fmt.Println("  a complete code dealing with array.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  array -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  array -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}

func main() {
	var (
		server string
	)
	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		fmt.Println("Can't open the driver: ", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection: ", err)
		return
	}

	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table: ", err)
		return
	}

	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table: ", err)
		}
	}()

	err = bulkInsert(connStr)
	if err != nil {
		fmt.Println("Can't bulkInsert: ", err)
		return
	}

	// oracle assocuative array need type table indexed by binary_integer which can
	// only defined inside the package
	err = createPackage(conn)
	if err != nil {
		fmt.Println("Can't create package: ", err)
	}
	defer func() {
		err = dropPackage(conn)
		if err != nil {
			fmt.Println("Can't drop package: ", err)
		}
	}()

	err = query1(conn)
	if err != nil {
		fmt.Println("Can't call query1: ", err)
		return
	}
	err = query2(conn)
	if err != nil {
		fmt.Println("Can't call query2: ", err)
	}
}
