package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	server string
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	VISIT_DATA  CLOB,
	VISIT_DATA2 CLOB,
	VISIT_DATA3 BLOB,
	VISIT_DATA4 BLOB,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}
func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}
func readWithOutputParameters(conn *sql.DB) error {
	t := time.Now()
	sqlText := `BEGIN
SELECT VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4 INTO :1, :2, :3, :4 FROM GOORA_TEMP_VISIT WHERE VISIT_ID = 1;
END;`
	var (
		//data string
		data  go_ora.Clob
		data2 go_ora.Clob
		data3 go_ora.Blob
		data4 go_ora.Blob
		//data2 string
	)
	// give space for data
	data.String = strings.Repeat(" ", 9444)
	data2.String = strings.Repeat(" ", 10)
	data3.Data = make([]byte, 9444)
	data4.Data = make([]byte, 10)
	_, err := conn.Exec(sqlText, sql.Out{Dest: &data}, sql.Out{Dest: &data2},
		sql.Out{Dest: &data3}, sql.Out{Dest: &data4})
	if err != nil {
		return err
	}
	fmt.Println("Data1: ", data.String)
	fmt.Println("Data2: ", data2.String)
	fmt.Println("Data3: ", data3.Data)
	fmt.Println("Data4: ", data4.Data)
	fmt.Println("Finish query output pars: ", time.Now().Sub(t))
	return nil
}
func readWithSql(conn *sql.DB) error {
	t := time.Now()
	row := conn.QueryRow("SELECT VISIT_ID, VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4 FROM GOORA_TEMP_VISIT")
	if row != nil {
		var (
			visitID int64
			data1   string
			data2   string
			data3   []byte
			data4   []byte
		)
		err := row.Scan(&visitID, &data1, &data2, &data3, &data4)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", visitID)
		fmt.Println("Data1: ", data1)
		fmt.Println("Data2: ", data2)
		fmt.Println("Data3: ", data3)
		fmt.Println("Data4: ", data4)
		//fmt.Println("ID: ", visitID, "\t Data1: ", data1, "\t Data2: ", data2)
		//fmt.Println("ID: ", visitID, "\t Data2: ", data2)
	}
	fmt.Println("1 row readed: ", time.Now().Sub(t))
	return nil
}
func insertData2() error {
	t := time.Now()
	conn, err := go_ora.NewConnection(server)
	if err != nil {
		return err
	}
	err = conn.Open()
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection2: ", err)
		}
	}()
	if err != nil {
		return err
	}
	val, err := ioutil.ReadFile("clob.json")
	if err != nil {
		return err
	}
	val2 := "string2"
	stmt := go_ora.NewStmt(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA, VISIT_DATA2) VALUES(2, :1, :2)`, conn)
	defer func() {
		err = stmt.Close()
		if err != nil {
			fmt.Println("Can't close stmt: ", err)
		}
	}()
	stmt.AddParam(":1", string(val), len(string(val)), go_ora.Input)
	stmt.AddParam(":2", val2, len(val2), go_ora.Input)
	_, err = stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish second insert: ", time.Now().Sub(t))
	return nil
}
func insertData(conn *sql.DB) error {
	t := time.Now()
	val, err := ioutil.ReadFile("clob.json")
	if err != nil {
		return err
	}

	_, err = conn.Exec(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4)
 VALUES(1, :1, :2, :3, :4)`,
		string(val), "string2", val, []byte("string2"))
	if err != nil {
		return err
	}
	fmt.Println("1 row inserted: ", time.Now().Sub(t))
	return nil
}
func usage() {
	fmt.Println()
	fmt.Println("clob")
	fmt.Println("  a code for using clob by create table GOORA_TEMP_VISIT then insert then drop")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  clob -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  clob -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}
func main() {

	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", server)
	if err != nil {
		fmt.Println("Can't open the driver", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection", err)
		return
	}

	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table", err)
		return
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}
	err = insertData2()
	if err != nil {
		fmt.Println("Can't make second insert: ", err)
		return
	}
	err = readWithSql(conn)
	if err != nil {
		fmt.Println("Can't read data: ", err)
		return
	}
	err = readWithOutputParameters(conn)
	if err != nil {
		fmt.Println("Can't read data: ", err)
		return
	}
}
