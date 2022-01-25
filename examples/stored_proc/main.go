package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

//func dieOnError(msg string, err error) {
//	if err != nil {
//		fmt.Println(msg, err)
//		os.Exit(1)
//	}
//}

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

func insertData(conn *sql.DB) error {
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
		_, err = stmt.Exec(index, nameText, val, time.Now())
		if err != nil {
			return err
		}
		val += 1.1
	}
	fmt.Println("100 rows inserted: ", time.Now().Sub(t))
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

func createStoredProc(conn *sql.DB) error {
	sqlText := `CREATE OR REPLACE PROCEDURE GOORA_TEMP_GET_VISIT
(
	LVISIT_ID IN NUMBER,
	LNAME OUT VARCHAR2,
	LVAL OUT NUMBER,
	LVISIT_DATE OUT DATE
) AS
BEGIN
	SELECT NAME, VAL, VISIT_DATE INTO LNAME, LVAL, LVISIT_DATE
	FROM GOORA_TEMP_VISIT 
	WHERE VISIT_ID = LVISIT_ID;
END GOORA_TEMP_GET_VISIT;`
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create store procedure: ", time.Now().Sub(t))
	return nil
}

func dropStoredProcedure(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("DROP PROCEDURE GOORA_TEMP_GET_VISIT")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop store procedure: ", time.Now().Sub(t))
	return nil
}

func callStoredProcedure(conn *sql.DB) error {
	var (
		name string
		val  float64
		date time.Time
	)
	t := time.Now()
	name = strings.Repeat(" ", 200)
	_, err := conn.Exec(`BEGIN GOORA_TEMP_GET_VISIT(:1, :2, :3, :4); END;`, 1,
		sql.Out{Dest: &name},
		sql.Out{Dest: &val},
		sql.Out{Dest: &date})
	if err != nil {
		return err
	}
	fmt.Println("Finish call store procedure: ", time.Now().Sub(t))
	fmt.Println("Name: ", name, "\tVal: ", val, "\tDate: ", date)
	return nil
}
func usage() {
	fmt.Println()
	fmt.Println("store_proc")
	fmt.Println("  a complete code of using stored procedure by create temporary table GOORA_TEMP_VISIT and procedure GOORA_TEMP_GET_VISIT.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  store_proc -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  store_proc -server "oracle://user:pass@server/service_name"`)
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
	conn, err := sql.Open("oracle", server)
	if err != nil {
		fmt.Println("Can't open the driver: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close driver: ", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection: ", err)
		return
	}

	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table", err)
		return
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data", err)
	}
	err = createStoredProc(conn)
	if err != nil {
		fmt.Println("Can't create stored procedure", err)
		return
	}
	defer func() {
		err = dropStoredProcedure(conn)
		if err != nil {
			fmt.Println("Can't drop stored procedure", err)
		}
	}()
	err = callStoredProcedure(conn)
	if err != nil {
		fmt.Println("Can't call store procedure", err)
	}
}
