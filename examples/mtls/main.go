package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	go_ora "github.com/sijms/go-ora/v2"
)

func databaseConn(addr, serviceName string, tlsConfig *tls.Config) error {
	var driver go_ora.OracleDriver

	dsn := fmt.Sprintf(`oracle://%s/%s?SSL=enabled&AUTH TYPE=TCPS`, addr, serviceName)
	conn, err := driver.OpenConnector(dsn)
	if err != nil {
		return err
	}
	oc, ok := conn.(*go_ora.OracleConnector)
	if !ok {
		return fmt.Errorf("failed to cast to OracleConnector")
	}
	oc.WithTLSConfig(tlsConfig)

	dbConn := sql.OpenDB(conn)

	err = dbConn.Ping()
	if err != nil {
		return err
	}

	var queryResultColumnOne string
	row := dbConn.QueryRow("SELECT systimestamp FROM dual")
	err = row.Scan(&queryResultColumnOne)
	if err != nil {
		return err
	}

	fmt.Println("The time in the database:", queryResultColumnOne)

	return nil
}

func printHelp() {
	helpMessage := `
Usage: go run ./examples/mtls [OPTIONS]

Options:
  -addr string
        Database address (e.g., localhost:2484) (required)
  -service string
        Database service name (required)
  -cert string
        Path to the user certificate (e.g., alice.crt) (required)
  -key string
        Path to the user key (e.g., alice.key) (required)
  -server-ca-cert string
        Path to the server CA certificate file (optional)
  -insecure
        Skip TLS certificate verification (default: false)
  -help
        Display help information
`
	fmt.Print(helpMessage)
}

func main() {
	addr := flag.String("addr", "", "Database address (e.g., localhost:2484)")
	serviceName := flag.String("service", "", "Database service name")
	certFile := flag.String("cert", "", "Path to the user certificate (e.g., alice.crt)")
	keyFile := flag.String("key", "", "Path to the user key (e.g., alice.key)")
	serverCaCertFile := flag.String("server-ca-cert", "", "Path to the server CA certificate file (optional)")
	insecureSkipVerify := flag.Bool("insecure", false, "Skip TLS certificate verification (default: false)")
	help := flag.Bool("help", false, "Display help information")

	flag.Parse()

	// If help is requested, display the help message and exit.
	if *help || len(os.Args) == 1 {
		printHelp()
		return
	}

	// Expand any environment variables in the flags
	*addr = os.ExpandEnv(*addr)
	*serviceName = os.ExpandEnv(*serviceName)
	*certFile = os.ExpandEnv(*certFile)
	*keyFile = os.ExpandEnv(*keyFile)
	*serverCaCertFile = os.ExpandEnv(*serverCaCertFile)

	// Check for required flags
	if *addr == "" {
		log.Fatal("database address is required; flag missing or empty")
	}
	if *serviceName == "" {
		log.Fatal("database service name is required; flag missing or empty")
	}
	if *certFile == "" {
		log.Fatal("path to the user certificate is required; flag missing or empty")
	}
	if *keyFile == "" {
		log.Fatal("path to the user key is required; flag missing or empty")
	}

	// Load the user certificate
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("failed to load certificate: %v", err)
	}

	// Create a TLS config with the certificate
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: *insecureSkipVerify,
	}

	// Load server CA certificate if the file is provided
	if *serverCaCertFile != "" {
		caCert, err := os.ReadFile(*serverCaCertFile)
		if err != nil {
			log.Fatalf("failed to read server CA certificate file: %v", err)
		}

		certPool := x509.NewCertPool()

		if !certPool.AppendCertsFromPEM(caCert) {
			log.Fatalf("failed to append server CA certificate from file: %s", *serverCaCertFile)
		}

		tlsConfig.RootCAs = certPool
	}

	// Call the database connection function
	if err := databaseConn(*addr, *serviceName, tlsConfig); err != nil {
		log.Fatalf("database connection error: %v", err)
	}
}
