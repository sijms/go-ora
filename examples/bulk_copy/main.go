package main

import (
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"log"
	"os"
	"time"
)

func createTable(conn *go_ora.Connection) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR2(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table :", time.Now().Sub(t))
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

func deleteData(conn *go_ora.Connection) error {
	t := time.Now()
	_, err := conn.Exec("TRUNCATE TABLE GOORA_TEMP_VISIT")
	if err != nil {
		return err
	}
	fmt.Println("Finish delete data: ", time.Now().Sub(t))
	return nil
}

func copyData(conn *go_ora.Connection, rowNum int) error {
	t := time.Now()
	bulk := go_ora.NewBulkCopy(conn, "GOORA_TEMP_VISIT")
	bulk.ColumnNames = []string{"VISIT_ID", "NAME", "VAL", "VISIT_DATE"}
	err := bulk.StartStream()
	if err != nil {
		return err
	}
	val := 1.1
	for x := 0; x < rowNum; x++ {
		err = bulk.AddRow(x+1, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			val, time.Now())
		if err != nil {
			_ = bulk.Abort()
			return err
		}
		val += 0.1
	}
	err = bulk.EndStream()
	if err != nil {
		_ = bulk.Abort()
		return err
	}
	err = bulk.Commit()
	if err != nil {
		_ = bulk.Abort()
		return err
	}
	fmt.Printf("%d rows copied: %v\n", rowNum, time.Now().Sub(t))
	return nil
}

func main() {
	conn, err := go_ora.NewConnection(os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("Err conn: %v", err)
	}
	err = conn.Open()
	if err != nil {
		log.Fatalf("Err open: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	//err = createTable(conn)
	//if err != nil {
	//	fmt.Println("Can't create table: ", err)
	//	return
	//}

	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("Can't drop table: ", err)
	//		return
	//	}
	//}()

	err = deleteData(conn)
	if err != nil {
		fmt.Println("Can't delete data: ", err)
		return
	}
	err = copyData(conn, 1000000)
	if err != nil {
		fmt.Println("Can't copy data: ", err)
	}

}
