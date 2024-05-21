package main

import (
	"bytes"
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strings"
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
	t := time.Now()
	err := execCmd(db, `CREATE TABLE TTB_557 (
		ID NUMBER(10), 
		DATA_BLOB BLOB, 
		DATA_CLOB CLOB, 
		DATA_NCLOB NCLOB, 
		LDATE DATE
    ) NOCOMPRESS`)
	if err != nil {
		return err
	}
	fmt.Println("finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, `DROP TABLE TTB_557 PURGE`)
	if err != nil {
		return err
	}
	fmt.Println("finish drop table: ", time.Now().Sub(t))
	return nil
}
func raw(db *sql.DB, data []byte) error {
	t := time.Now()
	_, err := db.Exec(`INSERT INTO TTB_557(DATA_BLOB, ID, LDATE) VALUES(:1, :2, :3)`, data, 1, time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_BLOB, LDATE) VALUES(:1, :2, :3)`, 2, data, time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_BLOB) VALUES(:1, :2, :3)`, 3, time.Now(), data)
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT DATA_BLOB FROM TTB_557`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var output []byte
	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return err
		}
		if len(output) != len(data) {
			return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	err = execCmd(db, "DELETE FROM TTB_557")
	if err != nil {
		return err
	}
	fmt.Printf("finish insert %d raw at (begin, middle and end) of sql: %v\n", len(data), time.Now().Sub(t))
	return nil
}
func nvarchar(db *sql.DB, data string) error {
	t := time.Now()
	_, err := db.Exec(`INSERT INTO TTB_557(DATA_NCLOB, ID, LDATE) VALUES(:1, :2, :3)`, go_ora.NVarChar(data), 1, time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_NCLOB, LDATE) VALUES(:1, :2, :3)`, 2, go_ora.NVarChar(data), time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_NCLOB) VALUES(:1, :2, :3)`, 3, time.Now(), go_ora.NVarChar(data))
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT DATA_NCLOB FROM TTB_557`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var output string
	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return err
		}
		if len(output) != len(data) {
			return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	err = execCmd(db, "DELETE FROM TTB_557")
	if err != nil {
		return err
	}
	fmt.Printf("finish insert %d nvarchar string at (begin, middle and end) of sql: %v\n", len(data), time.Now().Sub(t))
	return nil
}
func varchar(db *sql.DB, data string) error {
	t := time.Now()
	_, err := db.Exec(`INSERT INTO TTB_557(DATA_CLOB, ID, LDATE) VALUES(:1, :2, :3)`, data, 1, time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_CLOB, LDATE) VALUES(:1, :2, :3)`, 2, data, time.Now())
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_CLOB) VALUES(:1, :2, :3)`, 3, time.Now(), data)
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT DATA_CLOB FROM TTB_557`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var output string
	for rows.Next() {
		err = rows.Scan(&output)
		if err != nil {
			return err
		}
		if len(output) != len(data) {
			return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	err = execCmd(db, "DELETE FROM TTB_557")
	if err != nil {
		return err
	}
	fmt.Printf("finish insert %d varchar string at (begin, middle and end) of sql: %v\n", len(data), time.Now().Sub(t))
	return nil
}

func main() {
	config, err := go_ora.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't parse config: ", err)
		return
	}
	go_ora.RegisterConnConfig(config)
	db, err := sql.Open("oracle", "")
	if err != nil {
		fmt.Println("can't open db: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close db: ", err)
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
			return
		}
	}()
	// insert short string
	err = varchar(db, strings.Repeat("*", 10000))
	if err != nil {
		fmt.Println("varchar short insert: ", err)
		return
	}
	// insert long string
	err = varchar(db, strings.Repeat("*", 40000))
	if err != nil {
		fmt.Println("varchar long insert: ", err)
		return
	}

	// insert short NVARCHAR
	err = nvarchar(db, strings.Repeat("早上好", 2000))
	if err != nil {
		fmt.Println("nvarchar short insert: ", err)
		return
	}
	// insert long NVARCHAR
	err = nvarchar(db, strings.Repeat("早上好", 40000))
	if err != nil {
		fmt.Println("nvarchar long insert: ", err)
		return
	}

	// insert short RAW
	err = raw(db, bytes.Repeat([]byte{3}, 10000))
	if err != nil {
		fmt.Println("raw short insert: ", err)
		return
	}

	// insert long RAW
	err = raw(db, bytes.Repeat([]byte{3}, 40000))
	if err != nil {
		fmt.Println("raw long insert: ", err)
		return
	}
}
