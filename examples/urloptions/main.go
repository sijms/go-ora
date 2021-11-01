package main

import (
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strconv"
)

func usage() {
	fmt.Println()
	fmt.Println("url options")
	fmt.Println("  a tool show url-options available for this client and result databaseURL.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  urloptions -options`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  urloptions -server "server" -service "service" port 1521 user "user" password "pass"`)
	fmt.Println()
}
func main() {
	var (
		server, service, user, password, traceFile, ssl, sslVerify, wallet, servers, sid string
		port, prefetchRows                                                               int
	)
	flag.StringVar(&server, "server", "", "Server's name or ip")
	flag.IntVar(&port, "port", 1521, "Server port number default 1521")
	flag.StringVar(&service, "service", "", "Service name")
	flag.StringVar(&user, "user", "", "User or schema name")
	flag.StringVar(&password, "password", "", "Password for user(schema)")
	flag.StringVar(&traceFile, "trace", "trace.log", "File name for trace log")
	flag.StringVar(&sid, "sid", "", "SID can be used instead of service")
	flag.StringVar(&ssl, "ssl", "disable", "[enable|disable] or [true|false] enable or disable using ssl")
	flag.StringVar(&sslVerify, "ssl-verify", "disable", "[enable|disable] or [true|false] enable ssl verify for the server")
	flag.StringVar(&wallet, "wallet", "", "Path for auto-login oracle wallet")
	flag.StringVar(&servers, "servers", "", `Add more servers (nodes) in case of RAC in form of "srv1:port,srv2:port"`)
	flag.IntVar(&prefetchRows, "prefetch-rows", 25, "Number of rows in one network fetch")
	flag.Parse()

	if server == "" {
		fmt.Println("Missing server option")
		usage()
		os.Exit(1)
	}
	if user == "" {
		fmt.Println("Missing user option")
		usage()
		os.Exit(1)
	}
	if service == "" && sid == "" {
		fmt.Println("You need to pass either service or sid option")
		usage()
		os.Exit(1)
	}
	urloptions := map[string]string{
		"TRACE FILE":    traceFile,
		"server":        servers,
		"SSL":           ssl,
		"SSL Verify":    sslVerify,
		"SID":           sid,
		"wallet":        wallet,
		"PREFETCH_ROWS": strconv.Itoa(prefetchRows),
	}
	databaseURL := go_ora.BuildUrl(server, port, service, user, password, urloptions)
	fmt.Println("Database URL: ", databaseURL)
}
