package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/sijms/go-ora/v2"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

// for example to connect to an Oracle Autonomous Database using an Oracle Wallet and go_orav2 :
// oracle://demo:Modem123mode@adb.us-ashburn-1.oraclecloud.com:1522/k8j2fvxbaujdcfy_daprdb_low.adb.oraclecloud.com /home/lucas/dapr-work/components-contrib/state/oracledatabase/Wallet_daprDB/

func main() {
	if len(os.Args) < 3 {
		fmt.Println("\nhello_ora_wallet")
		fmt.Println("\thello_ora_wallet check if it can connect to the given oracle database server using an Oracle Wallet, then print server banner.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("\thello_ora_wallet oracle://user:pass@server/service_name oracle_wallet_location")
		fmt.Println()
		os.Exit(1)
	}
	connStr := os.ExpandEnv(os.Args[1])
	walletpathStr := os.ExpandEnv(os.Args[2])
	connStr += "?TRACE FILE=trace.log&SSL=enable&SSL Verify=false&WALLET=" + url.QueryEscape(walletpathStr)
	db, err := sql.Open("oracle", connStr)
	dieOnError("error in sql.Open: %w", err)
	defer func() {
		err = db.Close()
		dieOnError("error in db.Close: %w", err)
	}()

	err = db.Ping()
	dieOnError("error in db.Ping: %w", err)
	var queryResultColumnOne string
	row := db.QueryRow("SELECT systimestamp FROM dual")
	err = row.Scan(&queryResultColumnOne)
	dieOnError("error in row.Scan: %w", err)
	fmt.Println("The time in the database ", queryResultColumnOne)

}
