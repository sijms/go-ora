package main

import (
	"database/sql"
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

func structsInsert(databaseUrl string) error {

	type GooraTempVisit struct {
		VisitId   int       `db:"VISIT_ID"`
		Name      string    `db:"NAME"`
		Val       float32   `db:"VAL"`
		VisitDate time.Time `db:"VISIT_DATE"`
	}

	var values []interface{}
	values = append(values, GooraTempVisit{
		VisitId:   1,
		Name:      "jack",
		Val:       2.2,
		VisitDate: time.Now(),
	})
	values = append(values, GooraTempVisit{
		VisitId:   2,
		Name:      "mike",
		Val:       4.2,
		VisitDate: time.Now(),
	})
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

	result, err := conn.StructsInsert(sqlText, values)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("%d rows inserted: %v\n", rowsAffected, time.Now().Sub(t))
	return nil
}
func usage() {
	fmt.Println()
	fmt.Println("structs_insert")
	fmt.Println("  a complete code comparing regular insert with structs.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  structs_insert -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  structs_insert -server "oracle://user:pass@server/service_name"`)
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

	err = structsInsert(connStr)
	if err != nil {
		fmt.Println("Can't structs_insert: ", err)
		return
	}
}
