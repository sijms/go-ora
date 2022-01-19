package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

// create sequence
func createSeq(conn *sql.DB) error {
	sqlText := `CREATE SEQUENCE GOORA_TEMP_VISIT_SEQ 
		MINVALUE 1 MAXVALUE 999 
		INCREMENT BY 1 
		START WITH 1 
		NOCACHE  NOCYCLE`
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create sequence: ", time.Now().Sub(t))
	return nil
}

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

func createTrigger(conn *sql.DB) error {
	sqlText := `CREATE OR REPLACE TRIGGER GOORA_TEMP_VISIT_TRG
BEFORE INSERT ON GOORA_TEMP_VISIT 
FOR EACH ROW 
BEGIN
	SELECT GOORA_TEMP_VISIT_SEQ.NEXTVAL INTO :NEW.VISIT_ID FROM DUAL;
END;`
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create trigger: ", time.Now().Sub(t))
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

func dropSeq(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop sequence GOORA_TEMP_VISIT_SEQ")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop sequence: ", time.Now().Sub(t))
	return nil
}

func dropTrigger(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop trigger GOORA_TEMP_VISIT_TRG")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop trigger: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	stmt, err := conn.Prepare(`INSERT INTO GOORA_TEMP_VISIT(NAME, VAL, VISIT_DATE) 
VALUES(:1, :2, :3) RETURNING VISIT_ID INTO :4`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	val := 1.1
	var id int64
	for index := 0; index < 100; index++ {
		_, err = stmt.Exec(nameText, val, time.Now(), sql.Out{Dest: &id})
		if err != nil {
			return err
		}
		fmt.Printf("Insert row# %d and returning id value = %d\n", index, id)
		id = 3
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}
func usage() {
	fmt.Println()
	fmt.Println("insert_return")
	fmt.Println("  a complete code of insert row with returning new id in output parameter.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  insert_return -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  insert_return -server "oracle://user:pass@server/service_name"`)
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

	err = createSeq(conn)
	if err != nil {
		fmt.Println("Can't create sequence", err)
		return
	}
	defer func() {
		err = dropSeq(conn)
		if err != nil {
			fmt.Println("Can't drop sequence: ", err)
		}
	}()

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

	err = createTrigger(conn)
	if err != nil {
		fmt.Println("Can't create trigger", err)
		return
	}
	defer func() {
		err = dropTrigger(conn)
		if err != nil {
			fmt.Println("Can't drop tirgger", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data", err)
	}
}
