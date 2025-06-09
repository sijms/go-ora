package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strconv"
	"time"
)

var tableName = "TTB_642"

type TTB_DATA struct {
	Id   int64     `db:"ID"`
	Name string    `db:"NAME"`
	Val  float64   `db:"VAL"`
	Date time.Time `db:"LDATE"`
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

func createTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, fmt.Sprintf(`CREATE TABLE %s(
	ID			number(10)	NOT NULL,
	NAME		VARCHAR2(200),
	VAL			number(10,2),
	LDATE	date,
	PRIMARY KEY(ID)
	)`, tableName))
	if err != nil {
		return err
	}
	fmt.Println("Finish create table :", time.Since(t))
	return nil
}

func dropTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, fmt.Sprintf("drop table %s purge", tableName))
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Since(t))
	return nil
}

func insert(db *sql.DB, count int) error {
	t := time.Now()
	data := make([]TTB_DATA, count)
	for x := 0; x < count; x++ {
		data[x].Id = int64(1 + x)
		data[x].Name = "test_ " + strconv.Itoa(x+1)
		data[x].Val = 100.23 + 1
		data[x].Date = time.Now()

	}
	_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)", tableName),
		go_ora.NewBatch(data))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Since(t))
	return nil
}

func update(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET VAL=:val WHERE ID=:id", tableName),
		sql.Named("val", go_ora.NewBatch([]float64{10.1, 10.1, 10.1})),
		sql.Named("id", go_ora.NewBatch([]int{1, 2, 3})))
	if err != nil {
		return err
	}
	fmt.Println("Finish update data: ", time.Since(t))
	return nil
}

func delete(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ID=:1", tableName), go_ora.NewBatch([]int{6, 7, 8, 9, 10}))
	if err != nil {
		return err
	}
	fmt.Println("Finish delete data: ", time.Since(t))
	return nil
}
func query(db *sql.DB) error {
	t := time.Now()
	rows, err := db.Query(fmt.Sprintf("SELECT ID, NAME, VAL, LDATE FROM %s", tableName))
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var data TTB_DATA
	count := 0
	for rows.Next() {
		err = rows.Scan(&data.Id, &data.Name, &data.Val, &data.Date)
		if err != nil {
			return err
		}
		fmt.Println(data)
		count++
	}
	fmt.Printf("Finish query (%d rows): %v\n", count, time.Since(t))
	return nil
}
func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't connect: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()

	err = createTable(db)
	if err != nil {
		fmt.Println("Can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			fmt.Println("Can't drop table: ", err)
		}
	}()
	err = insert(db, 10)
	if err != nil {
		fmt.Println("can't insert: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	err = update(db)
	if err != nil {
		fmt.Println("can't update: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	err = delete(db)
	if err != nil {
		fmt.Println("can't delete: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
}
