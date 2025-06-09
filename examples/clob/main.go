package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

var (
	server string
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_LOB(
	ID	number(10)	NOT NULL,
	DATA1 CLOB,
	DATA2 CLOB,
	DATA3 BLOB,
	DATA4 BLOB,
	PRIMARY KEY(ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table:", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_LOB purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

//func readWithOutputParameters2() error {
//	t := time.Now()
//	conn, err := go_ora.NewConnection(server)
//	if err != nil {
//		return err
//	}
//	err = conn.Open()
//	if err != nil {
//		return err
//	}
//	defer func() {
//		err = conn.Close()
//		if err != nil {
//			fmt.Println("Can't close 2nd connection: ", err)
//		}
//	}()
//	sqlText := `BEGIN
//SELECT VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4 INTO :1, :2, :3, :4 FROM GOORA_TEMP_VISIT WHERE VISIT_ID = 1;
//END;`
//	stmt := go_ora.NewStmt(sqlText, conn)
//	defer func() {
//		err = stmt.Close()
//		if err != nil {
//			fmt.Println("Can't close stmt: ", err)
//		}
//	}()
//	var (
//		data  go_ora.Clob
//		data2 go_ora.Clob
//		data3 go_ora.Blob
//		data4 go_ora.Blob
//	)
//
//	_, err = stmt.Exec([]driver.Value{go_ora.Out{Dest: data, Size: 10000},
//		go_ora.Out{Dest: &data2, Size: 10},
//		go_ora.Out{Dest: data3, Size: 10000},
//		go_ora.Out{Dest: &data4, Size: 10}})
//	if err != nil {
//		return err
//	}
//	if tempVal, ok := stmt.Pars[0].Value.(go_ora.Clob); ok {
//		printLargeString("Data1: ", tempVal.String)
//	}
//	printLargeString("Data2: ", data2.String)
//	if tempVal, ok := stmt.Pars[2].Value.(go_ora.Blob); ok {
//		printLargeString("Data3: ", string(tempVal.Data))
//	}
//	printLargeString("Data4: ", string(data4.Data))
//	//fmt.Println("Data2: ", data2.String)
//	//printLargeString("Data3: ", string(data3.Data))
//	//fmt.Println("Data4: ", string(data4.Data))
//	fmt.Println("Finish query output pars: ", time.Now().Sub(t))
//	return nil
//
//}
//func readWithOutputParameters(conn *sql.DB) error {
//	t := time.Now()
//	sqlText := `BEGIN
//SELECT VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4 INTO :1, :2, :3, :4 FROM GOORA_TEMP_VISIT WHERE VISIT_ID = 1;
//END;`
//	var (
//		data  go_ora.Clob
//		data2 go_ora.Clob
//		data3 go_ora.Blob
//		data4 go_ora.Blob
//	)
//	_, err := conn.Exec(sqlText, go_ora.Out{Dest: &data, Size: 100000},
//		go_ora.Out{Dest: &data2, Size: 10},
//		go_ora.Out{Dest: &data3, Size: 100000},
//		go_ora.Out{Dest: &data4, Size: 10})
//	if err != nil {
//		return err
//	}
//	printLargeString("Data1: ", data.String)
//	fmt.Println("Data2: ", data2.String)
//	printLargeString("Data3: ", string(data3.Data))
//	fmt.Println("Data4: ", string(data4.Data))
//	fmt.Println("Finish query output pars: ", time.Now().Sub(t))
//	return nil
//}

func readWithOutputPars(db *sql.DB, id int) error {
	t := time.Now()
	sqlText := `BEGIN
SELECT ID, DATA1, DATA2, DATA3, DATA4 INTO :ID, :DATA1, :DATA2, :DATA3, :DATA4 FROM GOORA_TEMP_LOB WHERE ID = :iid;
END;`
	var temp = struct {
		ID    int            `db:"ID,,,output"`
		Data1 sql.NullString `db:"DATA1,,500,output"`
		Data2 go_ora.Clob    `db:"DATA2,clob,100000000,output"`
		Data3 []byte         `db:"DATA3,,500,output"`
		Data4 go_ora.Blob    `db:"DATA4,blob,100000000,output"`
	}{}
	_, err := db.Exec(sqlText, &temp, sql.Named("iid", id))
	if err != nil {
		return err
	}
	fmt.Println("ID: ", temp.ID)
	fmt.Println("Data1: ", temp.Data1.String)
	if temp.Data2.Valid {
		printLargeString("Data2: ", temp.Data2.String)
	} else {
		printLargeString("Data2: ", "")
	}
	fmt.Println("Data3: ", string(temp.Data3))
	printLargeString("Data4: ", string(temp.Data4.Data))
	fmt.Println("Finish query with output parameters: ", time.Now().Sub(t))
	return nil
}
func readWithSql(conn *sql.DB) error {
	t := time.Now()
	rows, err := conn.Query("SELECT ID, DATA1, DATA2, DATA3, DATA4 FROM GOORA_TEMP_LOB WHERE ID < 3")
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close dataset: ", err)
		}
	}()
	var (
		id    int64
		data1 sql.NullString
		data2 sql.NullString
		data3 []byte
		data4 []byte
	)
	cnt := 0
	for rows.Next() {
		err := rows.Scan(&id, &data1, &data2, &data3, &data4)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id)
		fmt.Println("Data1: ", data1)
		if data2.Valid {
			printLargeString("Data2: ", data2.String)
		} else {
			printLargeString("Data2: ", "")
		}
		fmt.Println("Data3: ", string(data3))
		printLargeString("Data4: ", string(data4))
		cnt++
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	fmt.Printf("%d row readed: %v\n", cnt, time.Now().Sub(t))
	return nil
}
func printLargeString(prefix, data string) {
	if len(data) <= 25 {
		fmt.Println(prefix, data)
		return
	}
	temp := strings.ReplaceAll(data, "\r", "")
	temp = strings.ReplaceAll(temp, "\n", "\\n")
	fmt.Println(prefix, temp[:25], "...........", temp[len(temp)-25:], "\tsize: ", len(data))
}

func insertData(db *sql.DB) error {
	t := time.Now()
	fileData, err := os.ReadFile("clob.json")
	if err != nil {
		return err
	}
	clob := strings.Repeat(string(fileData), 10)
	blob := bytes.Repeat(fileData, 10)
	type TempStruct struct {
		ID    int            `db:"ID"`
		Data1 sql.NullString `db:"DATA1"`
		Data2 go_ora.Clob    `db:"DATA2"`
		Data3 []byte         `db:"DATA3"`
		Data4 go_ora.Blob    `db:"DATA4"`
	}
	input := make([]TempStruct, 10)
	for x := 1; x <= len(input); x++ {
		temp := TempStruct{}
		temp.ID = x
		if x%2 == 0 {
			temp.Data1.Valid = false
			temp.Data2.Valid = false
			temp.Data3 = nil
			temp.Data4.Data = nil
		} else {
			temp.Data1 = sql.NullString{"this is a test", true}
			temp.Data2.String, temp.Data2.Valid = clob, true

			temp.Data3 = []byte("this is a test")
			temp.Data4.Data = blob

		}
		input[x-1] = temp
	}
	_, err = db.Exec("INSERT INTO GOORA_TEMP_LOB(ID, DATA1, DATA2, DATA3, DATA4) VALUES(:ID, :DATA1, :DATA2, :DATA3, :DATA4)",
		go_ora.NewBatch(input))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}

//	func insertData2() error {
//		t := time.Now()
//		conn, err := go_ora.NewConnection(server)
//		if err != nil {
//			return err
//		}
//		err = conn.Open()
//		defer func() {
//			err = conn.Close()
//			if err != nil {
//				fmt.Println("Can't close connection2: ", err)
//			}
//		}()
//		if err != nil {
//			return err
//		}
//		val, err := os.ReadFile("clob.json")
//		if err != nil {
//			return err
//		}
//		val2 := go_ora.Clob{String: "string2"}
//		val1 := go_ora.Clob{String: string(val) + string(val) + string(val) + string(val) + string(val) + string(val) + string(val)}
//		stmt := go_ora.NewStmt(`INSERT INTO GOORA_TEMP_VISIT(VISIT_DATA, VISIT_DATA2, VISIT_ID) VALUES(:1, :2, :3)`, conn)
//		defer func() {
//			err = stmt.Close()
//			if err != nil {
//				fmt.Println("Can't close stmt: ", err)
//			}
//		}()
//		_, err = stmt.Exec([]driver.Value{val1, val2, 2})
//		if err != nil {
//			return err
//		}
//		fmt.Println("1 row inserted: ", time.Now().Sub(t))
//		return nil
//	}
//
//	func insertData3(conn *sql.DB) error {
//		t := time.Now()
//		val, err := ioutil.ReadFile("clob.json")
//		buffer := bytes.Buffer{}
//		for x := 0; x < 4; x++ {
//			buffer.Write(val)
//		}
//		_, err = conn.Exec("INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA, VISIT_DATA3) VALUES(3, :1, :2)",
//			go_ora.Clob{String: string(buffer.Bytes())}, go_ora.Blob{Data: buffer.Bytes()})
//		if err != nil {
//			return err
//		}
//		fmt.Println("1 row inserted: ", time.Now().Sub(t))
//		return nil
//	}
//
//	func insertData(conn *sql.DB) error {
//		t := time.Now()
//		val, err := os.ReadFile("clob.json")
//		if err != nil {
//			return err
//		}
//
//		_, err = conn.Exec(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA, VISIT_DATA2, VISIT_DATA3, VISIT_DATA4)
//
// VALUES(1, :1, :2, :3, :4)`,
//
//			string(val), go_ora.Clob{String: "😶‍🌫"}, val, go_ora.Blob{Data: []byte("string2")})
//		if err != nil {
//			return err
//		}
//		fmt.Println("1 row inserted: ", time.Now().Sub(t))
//		return nil
//	}
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
	server = os.Getenv("DSN")
	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}

	server = connStr
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", connStr)
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
	//err = insertData2()
	//if err != nil {
	//	fmt.Println("Can't make 2nd insert: ", err)
	//	return
	//}
	//err = insertData3(conn)
	//if err != nil {
	//	fmt.Println("Can't make 3rd insert: ", err)
	//	return
	//}
	err = readWithSql(conn)
	if err != nil {
		fmt.Println("Can't read data with sql: ", err)
		return
	}
	err = readWithOutputPars(conn, 1)
	if err != nil {
		fmt.Println("Can't read with output parameters: ", err)
		return
	}
	err = readWithOutputPars(conn, 2)
	if err != nil {
		fmt.Println("Can't read with output parameters: ", err)
		return
	}
	//err = readWithOutputParameters(conn)
	//if err != nil {
	//	fmt.Println("Can't read data with output1: ", err)
	//	return
	//}
	//err = readWithOutputParameters2()
	//if err != nil {
	//	fmt.Println("Can't read data with output2: ", err)
	//	return
	//}
}
