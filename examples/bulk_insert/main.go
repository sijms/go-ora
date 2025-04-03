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
	NAME		VARCHAR2(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	MAJOR     VARCHAR2(100),
	DATA      BLOB,
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
func bulkInsert3(db *sql.DB, rowNum int) error {
	t := time.Now()
	sqlText := `INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE, major, DATA) VALUES(:id, :name, :val, :dat, :major, :data)`
	id := make([]int, rowNum)
	name := make([]*string, rowNum)
	val := make([]*float64, rowNum)
	date := make([]interface{}, rowNum)
	major := make([]sql.NullString, rowNum)
	data := make([]interface{}, rowNum)
	initalVal := 0.1
	value := "test"
	//dateVal := time.Now()
	for x := 0; x < rowNum; x++ {
		id[x] = x + 1
		if x%2 == 0 {
			name[x] = nil
			val[x] = nil
		} else {
			name[x] = &value
			val[x] = &initalVal
			//date[x] = dateVal
			//data[x] = go_ora.Blob{Data: []byte("this is a test"), Valid: true}
		}
		date[x] = nil
		data[x] = nil
		initalVal += 0.1
		if x == 0 {
			major[x] = sql.NullString{"M-13", true}
		} else {
			if x%2 == 0 {
				major[x] = sql.NullString{"", false}
			} else {
				major[x] = sql.NullString{"SP-17", true}
			}

		}
	}
	_, err := db.Exec(sqlText, sql.Named("id", id),
		sql.Named("name", name),
		sql.Named("val", val),
		sql.Named("dat", date),
		sql.Named("major", major),
		sql.Named("data", data))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert ", rowNum, " rows: ", time.Now().Sub(t))
	return nil
}
func bulkInsert2(db *sql.DB, rowNum int) error {
	t := time.Now()
	sqlText := `INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE, major, DATA) VALUES(:id, :name, :val, :dat, :major, :data)`
	id := make([]int, rowNum)
	name := make([]*string, rowNum)
	val := make([]*float64, rowNum)
	date := make([]*time.Time, rowNum)
	major := make([]sql.NullString, rowNum)
	data := make([]interface{}, rowNum)
	initalVal := 0.1
	value := "test"
	dateVal := time.Now()
	for x := 0; x < rowNum; x++ {
		id[x] = x + 1
		if x%2 == 0 {
			name[x] = nil
			val[x] = nil
			date[x] = nil
			data[x] = nil
		} else {
			name[x] = &value
			val[x] = &initalVal
			date[x] = &dateVal
			data[x] = go_ora.Blob{Data: []byte("this is a test"), Valid: true}
		}
		initalVal += 0.1
		if x == 0 {
			major[x] = sql.NullString{"M-13", true}
		} else {
			if x%2 == 0 {
				major[x] = sql.NullString{"", false}
			} else {
				major[x] = sql.NullString{"SP-17", true}
			}

		}
	}
	_, err := db.Exec(sqlText, sql.Named("id", id),
		sql.Named("name", name),
		sql.Named("val", val),
		sql.Named("dat", date),
		sql.Named("major", major),
		sql.Named("data", data))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert ", rowNum, " rows: ", time.Now().Sub(t))
	return nil
}
func bulkInsert(databaseUrl string, rowNum int) error {
	conn, err := go_ora.NewConnection(databaseUrl, nil)
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
	sqlText := `INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE, major) VALUES(:1, :2, :3, :4, :5)`
	//rowNum := 100
	visitID := make([]driver.Value, rowNum)
	nameText := make([]driver.Value, rowNum)
	val := make([]driver.Value, rowNum)
	date := make([]driver.Value, rowNum)
	major := make([]driver.Value, rowNum)
	initalVal := 0.1
	for index := 0; index < rowNum; index++ {
		visitID[index] = index + 1
		nameText[index] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		val[index] = initalVal
		date[index] = time.Now()
		initalVal += 0.1
		if index == 0 {
			major[index] = "M-13"
		} else {
			major[index] = "SP-17"
		}
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
	result, err := conn.BulkInsert(sqlText, rowNum, visitID, nameText, val, date, major)
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
	//var (
	//	server string
	//)
	//flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	//flag.Parse()
	//
	//connStr := os.ExpandEnv(server)
	//if connStr == "" {
	//	fmt.Println("Missing -server option")
	//	usage()
	//	os.Exit(1)
	//}
	//fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
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

	//err = createTable(conn)
	//if err != nil {
	//	fmt.Println("Can't create table: ", err)
	//	return
	//}

	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("Can't drop table: ", err)
	//	}
	//}()

	//err = insertData(conn)
	//if err != nil {
	//	fmt.Println("Can't insert data: ", err)
	//	return
	//}
	//
	err = deleteData(conn)
	if err != nil {
		fmt.Println("Can't delete data: ", err)
		return
	}

	err = bulkInsert3(conn, 10)
	if err != nil {
		fmt.Println("Can't insert: ", err)
		return
	}

	//err = bulkInsert(os.Getenv("DSN"), 1000000)
	//if err != nil {
	//	fmt.Println("Can't bulkInsert: ", err)
	//	return
	//}
}
