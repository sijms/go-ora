package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/sijms/go-ora/v2/advanced_nego"
	"log"
	"os"
)

type KerberosAuth struct{}

func (kerb KerberosAuth) Authenticate(server, service string) ([]byte, error) {
	conf, err := config.Load("/etc/krb5.conf")
	if err != nil {
		return nil, err
	}
	ccache, err := credentials.LoadCCache("/tmp/krb5cc_1000")
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
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection: ", err)
		return
	}
}
