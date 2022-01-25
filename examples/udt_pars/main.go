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
	Id   int64     `oracle:"name:test_id"`
	Name string    `oracle:"name:test_name"`
	Date time.Time `oracle:"name:test_date"`
}

func createTable(conn *go_ora.Connection) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	TEST_TYPE   TEST_TYPE1,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}

func insertData(conn *go_ora.Connection) error {
	t := time.Now()
	index := 1
	stmt := go_ora.NewStmt(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, TEST_TYPE) VALUES(:1, :2)`, conn)
	defer func() {
		_ = stmt.Close()
	}()
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for index = 1; index <= 100; index++ {
		var test test1
		test.Id = int64(index)
		test.Name = nameText[:index]
		test.Date = time.Now()
		_, err := stmt.Exec([]driver.Value{index, test})
		if err != nil {
			return err
		}
	}
	fmt.Println("100 rows inserted: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("SELECT VISIT_ID, TEST_TYPE FROM GOORA_TEMP_VISIT", conn)
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
	var (
		visitID int64
		test    test1
		count   int
	)
	for rows.Next_() {
		err = rows.Scan(&visitID, &test)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", visitID, "\tTest: ", test)
		count++
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	fmt.Printf("Finish query %d rows: %v\n", count, time.Now().Sub(t))
	return nil
}

func dropTable(conn *go_ora.Connection) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func createUDT(conn *go_ora.Connection) error {
	t := time.Now()
	sqlText := `create or replace TYPE TEST_TYPE1 IS OBJECT 
(
    TEST_ID NUMBER(10, 0),
    TEST_NAME VARCHAR2(200),
	TEST_DATE DATE
)`
	stmt := go_ora.NewStmt(sqlText, conn)
	defer func() {
		_ = stmt.Close()
	}()
	_, err := stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish create UDT: ", time.Now().Sub(t))
	return nil
}

func dropUDT(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("drop type TEST_TYPE1", conn)
	defer func() {
		_ = stmt.Close()
	}()
	_, err := stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish drop UDT: ", time.Now().Sub(t))
	return nil
}

func usage() {
	fmt.Println()
	fmt.Println("udt_par")
	fmt.Println("  a complete code for testing user define type (UDT) parameter [input and output].")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  udt_par -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  udt_par -server "oracle://user:pass@server/service_name"`)
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
		fmt.Println("Can't open driver", err)
		return
	}

	err = conn.Open()
	if err != nil {
		fmt.Println("Can't open connection", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()
	err = createUDT(conn)
	if err != nil {
		fmt.Println("Can't create UDT: ", err)
		return
	}
	defer func() {
		err = dropUDT(conn)
		if err != nil {
			fmt.Println("Can't drop UDT", err)
		}
	}()
	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table: ", err)
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table: ", err)
		}
	}()
	err = conn.RegisterType("TEST_TYPE1", test1{})
	if err != nil {
		fmt.Println("Can't register UDT: ", err)
		return
	}
	fmt.Println("UDT registered successfully")

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
	//err = inputPars(conn, 3, test1{
	//	Id:   4,
	//	Name: "this is a very big test4",
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//err = queryCustTypeTable(conn)
	//if err != nil {
	//	fmt.Println("Can't query: ", err)
	//	return
	//}

	//err = queryUDT(conn)
	//if err != nil {
	//	fmt.Println("Can't query UDT", err)
	//	return
	//}
	//conn, err := sql.Open("oracle", server)
	//if err != nil {
	//	fmt.Println("Can't open connection", err)
	//	return
	//}
	//defer func() {
	//	err = conn.Close()
	//	if err != nil {
	//		fmt.Println("Can't close connection", err)
	//	}
	//}()
	//
	//err = conn.Ping()
	//if err != nil {
	//	fmt.Println("Can't ping connection", err)
	//	return
	//}
	//
	//err = createTable(conn)
	//if err != nil {
	//	fmt.Println("Can't create table", err)
	//	return
	//}
	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("Can't drop table", err)
	//	}
	//}()
	//err = insertData(conn)
	//if err != nil {
	//	fmt.Println("Can't insert data", err)
	//	return
	//}
	//err = queryOutputPars(conn)
	//if err != nil {
	//	fmt.Println("Can't get output parameters", err)
	//	return
	//}

}
