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
func usage() {
	fmt.Println()
	fmt.Println("bulk_insert")
	fmt.Println("  a complete code comparing regular insert with bulk insert.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  bulk_insert -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  bulk_insert -server "oracle://user:pass@server/service_name"`)
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
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}

	err = deleteData(conn)
	if err != nil {
		fmt.Println("Can't delete data: ", err)
		return
	}

	err = bulkInsert(connStr)
	if err != nil {
		fmt.Println("Can't bulkInsert: ", err)
		return
	}
}
