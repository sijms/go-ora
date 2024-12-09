package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/sijms/go-ora/dbms"
	_ "github.com/sijms/go-ora/v2"
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

func addOutput(conn *sql.DB, data string) error {
	return exec_simple_conn(conn, fmt.Sprintf(`--sql
		BEGIN
			DBMS_OUTPUT.PUT_LINE('%s');
		END;
	`, data))
}

func withoutContext(conn *sql.DB) {
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
	err = addOutput(conn, "test1")
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
	err = addOutput(conn, "test2")
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

func withContext(conn *sql.DB, ctx context.Context) {
	err := dbms.EnableOutput(ctx, conn)
	if err != nil {
		fmt.Println("can't init DBMS_OUTPUT: ", err)
		return
	}
	defer func() {
		err = dbms.DisableOutput(ctx)
		if err != nil {
			fmt.Println("can't end dbms_output: ", err)
		}
	}()
	err = addOutput(conn, "test1")
	if err != nil {
		fmt.Println("can't write output: ", err)
		return
	}
	output, err := dbms.GetOutput(ctx)
	if err != nil {
		fmt.Println("can't get output: ", err)
		return
	}
	fmt.Print(output)
	err = addOutput(conn, "test2")
	if err != nil {
		fmt.Println("can't write output: ", err)
		return
	}
	err = dbms.PrintOutput(ctx, os.Stdout)
	if err != nil {
		fmt.Println("can't print output: ", err)
		return
	}
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

	withoutContext(conn)
	ctx := context.Background()
	withContext(conn, ctx)
}
