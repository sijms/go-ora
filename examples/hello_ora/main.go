package main

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"os"

	go_ora "github.com/sijms/go-ora"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("\nhello_ora")
		fmt.Println("\thello_ora check if it can connect to the given oracle server, then print server banner.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("\thello_ora oracle://user:pass@server/service_name")
		fmt.Println()
		os.Exit(1)
	}
	connStr := os.ExpandEnv(os.Args[1])

	conn, err := go_ora.NewConnection(connStr)
	dieOnError("Can't create connection:", err)

	err = conn.Open()
	dieOnError("Can't open connection:", err)

	defer conn.Close()

	err = conn.Ping(context.Background())
	dieOnError("Can't ping connection:", err)

	fmt.Println("\nSuccessfully connected.\n")
	stmt := go_ora.NewStmt("SELECT * FROM v$version", conn)
	defer stmt.Close()

	rows, err := stmt.Query(nil)
	dieOnError("Can't create query:", err)

	defer rows.Close()

	values := make([]driver.Value, 1)
	for {
		err = rows.Next(values)
		if err != nil {
			break
		}
		fmt.Println(values[0])
	}
	if err != nil && err != io.EOF {
		dieOnError("Can't fetch row:", err)
	}
}
