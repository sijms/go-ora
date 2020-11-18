package main

import (
	"database/sql/driver"
	"flag"
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

func usage() {
	fmt.Println()
	fmt.Println("query")
	fmt.Println("  query data from oracle.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  query -server server_url sql_query`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  query -server "oracle://user:pass@server/service_name" "select * from my_table"`)
	fmt.Println()
}

func main() {
	var (
		server string
		query  string
	)
	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Missing query")
		usage()
		os.Exit(1)
	}

	query = flag.Arg(0)
	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}

	DB, err := go_ora.NewConnection(connStr)
	dieOnError("Can't open the driver:", err)
	err = DB.Open()
	dieOnError("Can't open the connection:", err)

	defer DB.Close()

	stmt := go_ora.NewStmt(query, DB)

	defer stmt.Close()

	rows, err := stmt.Query(nil)
	dieOnError("Can't query", err)
	defer rows.Close()

	columns := rows.Columns()

	values := make([]driver.Value, len(columns))

	Header(columns)
	for {
		err = rows.Next(values)
		if err != nil {
			break
		}
		Record(columns, values)
	}
	if err != io.EOF {
		dieOnError("Can't Next", err)
	}
}

func Header(columns []string) {

}

func Record(columns []string, values []driver.Value) {
	for i, c := range values {
		fmt.Printf("%-25s: %v\n", columns[i], c)
	}
	fmt.Println()
}
