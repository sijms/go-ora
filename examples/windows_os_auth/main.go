package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	go_ora "github.com/sijms/go-ora/v2"
)

func usage() {
	fmt.Println()
	fmt.Println("win_os_auth")
	fmt.Println("  a code that use windows os authentication.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  win_os_auth -server server -service service -user user -password password`)
	flag.PrintDefaults()
	fmt.Println()
}

func TestConn(conn *sql.DB) error {
	row := conn.QueryRow("SELECT 'Success' FROM DUAL")
	var val string
	err := row.Scan(&val)
	if err != nil {
		return err
	}
	fmt.Println(val)
	return row.Err()
}

func main() {
	var (
		server   string
		port     int
		service  string
		user     string
		password string
		domain   string
	)
	flag.StringVar(&server, "server", "", "Server name or ip address")
	flag.IntVar(&port, "port", 1521, "Oracle server port default 1521")
	flag.StringVar(&service, "service", "", "Oracle service name")
	flag.StringVar(&user, "user", "", "Windows os user name if omit the driver will use logon user")
	flag.StringVar(&password, "password", "", "Windows os password")
	flag.StringVar(&domain, "domain", "", "Windows machine domain name")
	flag.Parse()
	if server == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	if service == "" {
		fmt.Println("Missing -service option")
		usage()
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("Missing -password option")
		usage()
		os.Exit(1)
	}
	urlOptions := map[string]string{
		// automatically set if you pass an empty oracle user or password
		// otherwise you need to set it
		"AUTH TYPE": "OS",
		// operating system user if empty the driver will use logon user name
		"OS USER": user,
		// operating system password needed for os logon
		"OS PASS": password,
		// Windows system domain name
		"DOMAIN": domain,
		// NTS is the required for Windows os authentication
		// when you run the program from Windows machine it will be added automatically
		// otherwise you need to specify it
		"AUTH SERV": "NTS",
		// uncomment this option for debugging
		"TRACE FILE": "trace.log",
	}
	databaseUrl := go_ora.BuildUrl(server, port, service, "", "", urlOptions)
	conn, err := sql.Open("oracle", databaseUrl)
	if err != nil {
		fmt.Println("Can't open connection: ", err)
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
	err = TestConn(conn)
	if err != nil {
		fmt.Println("Can't test connection: ", err)
	}
}
