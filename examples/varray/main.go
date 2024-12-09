package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
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

func createType(db *sql.DB) error {
	return execCmd(db, `create type StringArray as VARRAY(10) of varchar2(20) not null`)
}

func dropType(db *sql.DB) error {
	return execCmd(db, `drop type StringArray`)
}

func outputPar(db *sql.DB) error {
	var output []string
	_, err := db.Exec(`
DECLARE
	l_array StringArray := StringArray();
BEGIN
	for x in 1..10 loop
		l_array.extend;
		l_array(x) := 'string_' || x;
	end loop;
	:1 := l_array;
END;`, go_ora.Object{Name: "StringArray", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println("output: ", output)
	return nil
}

func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
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
	err = createType(db)
	if err != nil {
		fmt.Println("can't create types: ", err)
		return
	}
	defer func() {
		err = dropType(db)
		if err != nil {
			fmt.Println("can't drop types: ", err)
		}
	}()
	err = go_ora.RegisterType(db, "varchar2", "StringArray", nil)
	if err != nil {
		fmt.Println("can't register string array: ", err)
		return
	}
	err = outputPar(db)
	if err != nil {
		fmt.Println("can't output pars: ", err)
		return
	}
}
