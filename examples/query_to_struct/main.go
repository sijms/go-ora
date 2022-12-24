package main

import (
	"database/sql/driver"
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

type test1 struct {
	name string
}

func (t *test1) Scan(src any) error {
	t.name = fmt.Sprintf("%v", src)
	return nil
}

type visit struct {
	//Id   int64  `db:"name:visit_id"`
	// string replaced with new type that implement sql.Scanner interface
	Name test1   `db:"name:name"`
	Val  float32 `db:"name:val"`
	//Date time.Time	`db:"name:visit_date"`
}

func createTable(conn *go_ora.Connection) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	stmt := go_ora.NewStmt(sqlText, conn)
	_, err := stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}

func insertData(conn *go_ora.Connection) error {
	t := time.Now()
	index := 1
	stmt, err := conn.Prepare(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE) 
VALUES(:1, :2, :3, :4)`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	val := 1.1
	for index = 1; index <= 100; index++ {
		_, err = stmt.Exec([]driver.Value{index, nameText, val, time.Now()})
		if err != nil {
			return err
		}
		val += 1.1
	}
	fmt.Println("100 rows inserted: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("SELECT VISIT_ID, NAME, VAL, VISIT_DATE FROM GOORA_TEMP_VISIT", conn)
	rows, err := stmt.Query_(nil)
	if err != nil {
		return err
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	//var (
	//	id int64
	//	name string
	//	val float32
	//	date time.Time
	//)
	var vi visit
	var Id int
	var Date time.Time
	for rows.Next_() {
		err = rows.Scan(&Id, &vi, &Date)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", Id, "\tName: ", vi.Name, "\tval: ", vi.Val, "\tDate: ", Date)
	}
	fmt.Println("Finish query rows: ", time.Now().Sub(t))
	return nil
}

type db_TS struct {
	S1C1 string `db:"name:s1c1"`
	S1C2 string `db:"name:s1c2"`
	S2C1 string `db:"name:s2c1"`
	S2C2 string `db:"name:s2c2"`
}

func queryTest2(conn *go_ora.Connection) error {
	var DBID int = 0
	stmt := go_ora.NewStmt("select 1 as DBID from dual", conn) // It's an integer, an Oracle number to be more precise.
	defer func() {
		err := stmt.Close()
		if err != nil {
			fmt.Println("Can't close stmt: ", err)
		}
	}()

	rows, err := stmt.Query_(nil)
	if err != nil {
		return err
	}
	rows.Next_()
	err = rows.Scan(&DBID)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", DBID)
	return nil
}
func queryTest(conn *go_ora.Connection) error {
	queries := []string{
		"select 's1c1' as s1c1, 's1c2' as s1c2 from dual",
		"select 's2c1' as s2c1, 's2c2' as s2c2 from dual",
	}
	for _, query := range queries {
		stmt := go_ora.NewStmt(query, conn)
		defer func() {
			err := stmt.Close()
			if err != nil {
				fmt.Println("Can't close stmt: ", err)
			}
		}()

		rows, err := stmt.Query_(nil)
		if err != nil {
			return err
		}
		rowsCount := 0
		var db db_TS
		for rows.Next_() {
			err = rows.Scan(&db)
			if err != nil {
				return err
			}
			rowsCount += 1
		}
		fmt.Println("Rows count: ", rowsCount)
	}
	return nil
}
func dropTable(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("drop table GOORA_TEMP_VISIT purge", conn)
	_, err := stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func usage() {
	fmt.Println()
	fmt.Println("query_to_struct")
	fmt.Println("  a complete code of create table insert, query to struct then drop table.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  query_to_struct -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  query_to_struct -server "oracle://user:pass@server/service_name"`)
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
	conn, err := go_ora.NewConnection(connStr)
	if err != nil {
		fmt.Println("Can't create connection: ", err)
		return
	}
	err = conn.Open()
	if err != nil {
		fmt.Println("Can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
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

	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}

	err = queryData(conn)
	if err != nil {
		fmt.Println("Can't query data: ", err)
		return
	}

	err = queryTest(conn)
	if err != nil {
		fmt.Println("Can't query data: ", err)
		return
	}
	err = queryTest2(conn)
	if err != nil {
		fmt.Println("Can't query 2: ", err)
		return
	}
}
