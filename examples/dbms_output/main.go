package main

import (
	"database/sql"
	"fmt"
	"github.com/sijms/go-ora/dbms"
	_ "github.com/sijms/go-ora/v2"
	"os"
)

func exec_simple_conn(conn *sql.DB, texts ...string) error {
	var err error
	for _, text := range texts {
		_, err = conn.Exec(text)
		if err != nil {
			return err
		}
	}
	return err
}
func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("error in connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("error in close: ", err)
		}
	}()
	output, err := dbms.NewOutput(conn, 0x7FFF)
	if err != nil {
		fmt.Println("can't init DBMS_OUTPUT: ", err)
		return
	}
	defer func() {
		err = output.Close()
		if err != nil {
			fmt.Println("can't end dbms_output: ", err)
		}
	}()
	err = exec_simple_conn(conn, `BEGIN
DBMS_OUTPUT.PUT_LINE('this is a test');
END;`)
	if err != nil {
		fmt.Println("can't write output: ", err)
		return
	}
	line, err := output.GetOutput()
	if err != nil {
		fmt.Println("can't get output: ", err)
		return
	}
	fmt.Print(line)
	err = exec_simple_conn(conn, `BEGIN
DBMS_OUTPUT.PUT_LINE('this is a test2');
END;`)
	if err != nil {
		fmt.Println("can't write output: ", err)
		return
	}
	err = output.Print(os.Stdout)
	if err != nil {
		fmt.Println("can't print: ", err)
		return
	}

}
