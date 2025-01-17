package main

import (
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

func createProc(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, `create or replace procedure proc_626(
	par_01 in out clob,
	par_02 out number,
	par_03 out varchar2,
	par_04 out varchar2) AS
BEGIN
	par_01 := par_01 || ' + output string';
	par_02 := 15;
	par_03 := 'this is a test1';
	par_04 := 'this is a test2';
END;`)
	if err != nil {
		return err
	}
	fmt.Println("created proc successfully: ", time.Since(t))
	return nil
}

func dropProc(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, `drop procedure proc_626`)
	if err != nil {
		return err
	}
	fmt.Println("dropped proc successfully: ", time.Since(t))
	return nil
}

func callProc(db *sql.DB) error {
	t := time.Now()
	var (
		par_01 go_ora.Clob
		par_02 int
		par_03 string
		par_04 string
	)
	par_01 = go_ora.Clob{
		String: strings.Repeat("a", 0x8010),
		Valid:  true,
	}
	_, err := db.Exec("BEGIN proc_626(:par_01, :par_02, :par_03, :par_04); END;",
		sql.Named("par_01", go_ora.Out{Dest: &par_01, Size: 50000, In: true}),
		sql.Named("par_02", go_ora.Out{Dest: &par_02}),
		sql.Named("par_03", go_ora.Out{Dest: &par_03, Size: 10000}),
		sql.Named("par_04", go_ora.Out{Dest: &par_04, Size: 10000}))
	if err != nil {
		return err
	}
	fmt.Println("created proc successfully: ", time.Since(t))
	fmt.Println("Par 1: ", par_01.String)
	fmt.Println("Length of par1: ", len(par_01.String))
	fmt.Println("Par 2: ", par_02)
	fmt.Println("Par 3: ", par_03)
	fmt.Println("Par 4: ", par_04)
	return nil
}

func main() {
	db, err := sql.Open("oracle", os.Getenv("MAALI_DSN"))
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
	err = createProc(db)
	if err != nil {
		fmt.Println("can't create proc: ", err)
		return
	}
	//defer func() {
	//	err = dropProc(db)
	//	if err != nil {
	//		fmt.Println("can't drop proc: ", err)
	//	}
	//}()
	err = callProc(db)
	if err != nil {
		fmt.Println("can't call proc: ", err)
		return
	}
}
