package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	_ "github.com/sijms/go-ora"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("\nhello_sql")
		fmt.Println("\thello_sql check if it can connect to the given oracle server using sql interface, then print server banner.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("\thello_sql oracle://user:pass@server/service_name")
		fmt.Println()
		os.Exit(1)
	}
	connStr := os.ExpandEnv(os.Args[1])

	conn, err := sql.Open("oracle", connStr)
	dieOnError("Can't open connection:", err)

	defer conn.Close()

	err = conn.Ping()
	dieOnError("Can't ping connection:", err)

	fmt.Println("\nSuccessfully connected.\n")
	stmt, err := conn.Prepare("SELECT * FROM v$version")
	dieOnError("Can't prepare query:", err)

	defer stmt.Close()

	rows, err := stmt.Query()
	dieOnError("Can't create query:", err)

	defer rows.Close()

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			break
		}
		fmt.Println(s)
	}
	if rows.Err() != nil && rows.Err() != io.EOF {
		dieOnError("Can't fetch row:", rows.Err())
	}
}
