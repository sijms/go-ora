package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
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
		if index%5 == 0 {
			_, err = stmt.Exec(index, nameText, val, nil)
		} else {
			_, err = stmt.Exec(index, nameText, val, time.Now())
		}
		if err != nil {
			return err
		}
		val += 1.1
	}
	fmt.Println("100 rows inserted: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *sql.DB) error {
	t := time.Now()
	rows, err := conn.Query("SELECT VISIT_ID, NAME, VAL, VISIT_DATE FROM GOORA_TEMP_VISIT")
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close dataset: ", err)
		}
	}()
	var (
		id   int64
		name string
		val  float32
		date sql.NullTime
	)
	for rows.Next() {
		err = rows.Scan(&id, &name, &val, &date)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
	}
	fmt.Println("Finish query rows: ", time.Now().Sub(t))
	return nil
}

func updateData(conn *sql.DB) error {
	t := time.Now()
	updStmt, err := conn.Prepare(`UPDATE GOORA_TEMP_VISIT SET NAME = :1 WHERE VISIT_ID = :2`)
	if err != nil {
		return err
	}
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defer func() {
		_ = updStmt.Close()
	}()
	for index := 1; index <= 100; index++ {
		_, err = updStmt.Exec(nameText[:101-index], index)
		if err != nil {
			return err
		}
	}
	fmt.Println("Finish update: ", time.Now().Sub(t))
	return nil
}

func deleteData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("delete from GOORA_TEMP_VISIT")
	if err != nil {
		return err
	}
	fmt.Println("Finish delete: ", time.Now().Sub(t))
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

func usage() {
	fmt.Println()
	fmt.Println("crud")
	fmt.Println("  a complete code of create table insert, update, query and delete then drop table.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  curd -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  crud -server "oracle://user:pass@server/service_name"`)
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
	err = updateData(conn)
	if err != nil {
		fmt.Println("Can't update data: ", err)
		return
	}
	err = queryData(conn)
	if err != nil {
		fmt.Println("Can't query data: ", err)
		return
	}
	err = deleteData(conn)
	if err != nil {
		fmt.Println("Can't delete data: ", err)
		return
	}

}
