package TestIssues

import (
	"database/sql"
	"errors"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"reflect"
	"strconv"
)

// var db *sql.DB
var server = os.Getenv("SERVER")
var port int
var service = os.Getenv("SERVICE")
var username = os.Getenv("USER")
var password = os.Getenv("PASSWORD")
var urlOptions = map[string]string{
	"TRACE FILE": "trace.log",
	"lob fetch":  "pre",
}

type Execuer interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
}

func createMainTable(db *sql.DB) error {
	sqlText := `CREATE TABLE TTB_MAIN(
	ID	number(10)	NOT NULL,
	NAME		VARCHAR(500),
	VAL			number(10,2),
	LDATE   		date,
	DATA			RAW(100),
	PRIMARY KEY(ID)
	)`
	return execCmd(db, sqlText)
}

func dropMainTable(db *sql.DB) error {
	return execCmd(db, "drop table TTB_MAIN purge")
}

func getDB() (*sql.DB, error) {
	url := go_ora.BuildUrl(server, port, service, username, password, urlOptions)
	return sql.Open("oracle", url)
}

func init() {
	temp := os.Getenv("PORT")
	tempInt, err := strconv.ParseInt(temp, 10, 32)
	if err != nil {
		port = 1521
	} else {
		port = int(tempInt)
	}
	ssl_value := os.Getenv("SSL")
	wallet := os.Getenv("WALLET")
	if ssl_value == "TRUE" {
		urlOptions["SSL"] = "true"
		urlOptions["SSL VERIFY"] = "false"
		urlOptions["wallet"] = wallet
	}
}

func execCmd(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			if len(stmts) > 1 {
				return fmt.Errorf("error: %v in execuation of stmt: %s", err, stmt)
			} else {
				return err
			}
		}
	}
	return nil
}

func queryStruct(row *sql.Row, s any) error {
	sValue := reflect.ValueOf(s)
	if sValue.Kind() != reflect.Ptr {
		return errors.New("you should pass pointer")
	}
	if sValue.IsNil() {
		return errors.New("you should pass non nil value")
	}
	if sValue.Elem().Kind() != reflect.Struct {
		return errors.New("only accept pointer to struct")
	}
	sValue = sValue.Elem()
	fieldCount := sValue.NumField()
	scanValue := make([]any, fieldCount)
	for x := 0; x < fieldCount; x++ {
		fieldValue := sValue.Field(x)
		scanValue[x] = fieldValue.Addr().Interface()
	}
	return row.Scan(scanValue...)
}
