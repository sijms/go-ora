package main

import (
	"database/sql"
	"fmt"
	"os"
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
	db, err := sql.Open("oracle", connStr)
	dieOnError("Can't create connection:", err)

	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	err = db.Ping()
	dieOnError("Can't ping connection:", err)

	fmt.Println("\nSuccessfully connected.\n")
	rows, err := db.Query("SELECT * FROM v$version")
	dieOnError("Can't create query:", err)
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close rows: ", err)
		}
	}()
	var version string
	for rows.Next() {
		err = rows.Scan(&version)
		dieOnError("Can't scan row:", err)
	}
}
