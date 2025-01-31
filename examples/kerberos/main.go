package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"C" // Import cgo to avoid project-wide go fmt failures.

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/spnego"

	go_ora "github.com/sijms/go-ora/v2"
	"github.com/sijms/go-ora/v2/advanced_nego"
)

type KerberosAuth struct {
	ccache string
}

func (kerb *KerberosAuth) Authenticate(server, service string) ([]byte, error) {
	krb5conf := os.Getenv("KRB5_CONFIG")
	if krb5conf == "" {
		krb5conf = "/etc/krb5.conf"
	}
	conf, err := config.Load(krb5conf)
	if err != nil {
		return nil, err
	}
	ccache, err := credentials.LoadCCache(kerb.ccache)
	if err != nil {
		return nil, err
	}
	cl, err := client.NewFromCCache(ccache, conf)
	if err != nil {
		return nil, err
	}

	ticket, key, err := cl.GetServiceTicket(service + "/" + server)
	if err != nil {
		return nil, err
	}
	token, err := spnego.NewKRB5TokenAPREQ(cl, ticket, key, []int{gssapi.ContextFlagMutual}, []int{})
	if err != nil {
		return nil, err
	}
	return token.APReq.Marshal()
}

func usage() {
	fmt.Println()
	fmt.Println("kerberos")
	fmt.Println("  a program to test kerberos authentication.")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  kerberos -host mydb.example.com`)
	fmt.Println()
}

type options struct {
	host              string
	port              int
	serviceName       string
	kerberosCacheFile string
	useGlobalAuth     bool
}

func (o *options) validate() error {
	if o.host == "" {
		return errors.New("-host option is missing")
	}
	if o.port <= 0 || o.port > 65535 {
		return errors.New("-port option is missing")
	}
	if o.serviceName == "" {
		return errors.New("-service option is missing")
	}
	return nil
}

func parseOptions() *options {
	var opts options

	flag.StringVar(&opts.host, "host", "", "Oracle server host. REQUIRED.")
	flag.IntVar(&opts.port, "port", 1521, "Oracle server port.")
	flag.StringVar(&opts.serviceName, "service", "", "Database service name. REQUIRED.")
	flag.StringVar(&opts.kerberosCacheFile, "ccache", "/tmp/krb5cc_1000", "Kerberos ticket cache file.")
	flag.BoolVar(&opts.useGlobalAuth, "global_auth", false, "Configure Kerberos authentication via global variable.")

	flag.Parse()

	return &opts
}

func main() {
	opts := parseOptions()
	err := opts.validate()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		usage()
		os.Exit(1)
	}

	urlOpts := map[string]string{"AUTH TYPE": "KERBEROS"}
	connString := go_ora.BuildUrl(opts.host, opts.port, opts.serviceName, "", "", urlOpts)

	fmt.Printf("Connection string: %s\n", connString)

	auth := &KerberosAuth{ccache: opts.kerberosCacheFile}

	// open connection in preferred way
	var sqlConn *sql.DB
	if opts.useGlobalAuth {
		advanced_nego.SetKerberosAuth(auth)
		sqlConn, err = sql.Open("oracle", connString)
		if err != nil {
			log.Fatalf("Cannot connect: %v", err)
		}
	} else {
		connector := go_ora.NewConnector(connString).(*go_ora.OracleConnector)
		connector.WithKerberosAuth(auth)
		sqlConn = sql.OpenDB(connector)
	}
	defer sqlConn.Close()

	// verify the connection
	err = sqlConn.Ping()
	if err != nil {
		log.Fatalf("Can't ping connection: %v", err)
	} else {
		fmt.Println("PING ok.")
	}
	// report username as seen by the database.
	row := sqlConn.QueryRow("SELECT USER FROM DUAL")
	var user string
	err = row.Scan(&user)
	if err != nil {
		fmt.Println("Query read error:", err)
	} else {
		fmt.Println("Reported user:", user)
	}
}
