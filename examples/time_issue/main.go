package main

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

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
	return execCmd(db, `
CREATE TABLE TTB_TIME(
    ID NUMBER,
    DATE1 DATE,
    DATE2 TIMESTAMP,
    DATE3 TIMESTAMP WITH TIME ZONE,
    DATE4 TIMESTAMP WITH LOCAL TIME ZONE
)`)
}

func dropTable(db *sql.DB) error {
	return execCmd(db, `DROP TABLE TTB_TIME PURGE`)
}

func insert(db *sql.DB) error {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	fmt.Println("time in Asia/Shanghai: ", time.Now().In(loc))
	_, err := db.Exec("INSERT INTO TTB_TIME(ID, DATE1, DATE2, DATE3, DATE4) VALUES(:1, :2, :3, :4, :5)",
		1, time.Now(), time.Now(), time.Now().In(loc), time.Now())
	return err
}

func insert2(db *sql.DB) error {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	fmt.Println("time in Asia/Shanghai: ", time.Now().In(loc))
	_, err := db.Exec("INSERT INTO TTB_TIME(ID, DATE1, DATE2, DATE3, DATE4) VALUES(:1, :2, :3, :4, :5)",
		1, time.Now(), go_ora.TimeStamp(time.Now()), go_ora.TimeStampTZ(time.Now().In(loc)),
		go_ora.TimeStamp(time.Now()))
	return err
}

func query(db *sql.DB) error {
	var (
		id                         int
		date1, date2, date3, date4 time.Time
	)
	err := db.QueryRow("SELECT ID, DATE1, DATE2, DATE3, DATE4 FROM TTB_TIME").Scan(&id, &date1, &date2, &date3, &date4)
	if err != nil {
		return err
	}
	fmt.Println("query with sql: ")
	fmt.Println("DATE: ", date1)
	fmt.Println("Timestamp: ", date2)
	fmt.Println("Timestamp TZ: ", date3)
	fmt.Println("Timestamp with local TZ: ", date4)
	return nil
}

func queryPars(db *sql.DB) error {
	var (
		id                  int
		date1, date2, date3 sql.NullTime
	)
	_, err := db.Exec("BEGIN SELECT ID, DATE1, DATE2, DATE3 INTO :1, :2, :3, :4 FROM TTB_TIME; END;",
		go_ora.Out{Dest: &id}, go_ora.Out{Dest: &date1}, go_ora.Out{Dest: &date2}, go_ora.Out{Dest: &date3})
	if err != nil {
		return err
	}
	fmt.Println("query with output parameters: ")
	fmt.Println("DATE: ", date1)
	fmt.Println("Timestamp: ", date2)
	fmt.Println("Timestamp TZ: ", date3)
	return nil
}
func main() {
	//fmt.Println(time.Now())
	//fmt.Println(time.Now().Local())
	//fmt.Println(time.Now().UTC())
	//return
	// DSN_MAALI_STORE
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open database: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close database: ", err)
		}
	}()
	err = createTable(db)
	if err != nil {
		fmt.Println("can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()
	err = execCmd(db, "delete from TTB_TIME")
	if err != nil {
		fmt.Println("can't delete data: ", err)
		return
	}
	err = insert(db)
	if err != nil {
		fmt.Println("can't insert: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	err = queryPars(db)
	if err != nil {
		fmt.Println("can't query par: ", err)
		return
	}
}
