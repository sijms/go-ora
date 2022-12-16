package main

import "C"
import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/sijms/go-ora/v2/advanced_nego"
	"log"
	"os"
)

type KerberosAuth struct{}

//export Authenticate
func (kerb KerberosAuth) Authenticate(server, service string) ([]byte, error) {
	// run a c++ function Authenticate
	return nil, nil
}
func usage() {
	fmt.Println()
	fmt.Println("kerberos")
	fmt.Println("  a program to test kerberos5 authentication.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  kerberos -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  kerberos -server "oracle://user:pass@server/service_name"`)
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
	advanced_nego.SetKerberosAuth(&KerberosAuth{})
	//options := map[string]string{
	//	"TRACE FILE": "trace.log",
	//	"AUTH TYPE":  "KERBEROS",
	//}
	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		log.Fatalln("cannot connect: ", err)
	}
	defer func() { _ = conn.Close() }()
	err = conn.Ping()
	if err != nil {
		log.Fatalln("cannot ping: ", err)
	}
	return
}
