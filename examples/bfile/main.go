package main

import (
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
)

var server string
var dirName string
var fileName string

/* before using this example do the following
*  connect to oracle database as sys and create directory object
*  put some file in the physical directory
*  run the program:
*      bfile -server serverUrl  -dir dirObjectName -file fileName
 */
func createTable(conn *go_ora.Connection) error {
	sqlText := `create table GOORA_TEST_BFILE (
    FILE_ID NUMBER(10) NOT NULL,
    FILE_DATA BFILE
)`
	_, err := conn.Exec(sqlText)
	return err
}
func dropTable(conn *go_ora.Connection) error {
	_, err := conn.Exec(`drop table GOORA_TEST_BFILE purge`)
	return err
}
func insertData(conn *go_ora.Connection) error {
	file, err := go_ora.NewBFile(conn, dirName, fileName)
	if err != nil {
		return err
	}
	sqlText := `INSERT INTO GOORA_TEST_BFILE(FILE_ID, FILE_DATA) VALUES(:1, :2)`
	_, err = conn.Exec(sqlText, 1, file)
	return err
}
func workWithBFile(file *go_ora.BFile) error {
	err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Println("Can't close file: ", err)
		}
	}()
	exists, err := file.Exists()
	if err != nil {
		return err
	}
	if exists {
		length, err := file.GetLength()
		if err != nil {
			return err
		}
		fmt.Println("File length: ", length)
		fmt.Println("Read all data: ")
		data, err := file.Read()
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		if length > 2 {
			fmt.Println("Read From Position 2 to the end of the file: ")
			data, err = file.ReadFromPos(2)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		}
		if length > 5 {
			fmt.Println("Read 5 bytes starting from position 2: ")
			data, err = file.ReadBytesFromPos(2, 5)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		}
	}
	return nil
}

func BFileOutputPar(conn *go_ora.Connection) error {
	sqlText := `BEGIN SELECT FILE_ID, FILE_DATA INTO :1, :2 FROM GOORA_TEST_BFILE WHERE FILE_ID = 1; END;`
	var (
		id   int64
		file go_ora.BFile
	)
	_, err := conn.Exec(sqlText, go_ora.Out{Dest: &id}, go_ora.Out{Dest: &file})
	if err != nil {
		return err
	}
	fmt.Println("ID: ", id)
	err = workWithBFile(&file)
	if err != nil {
		return err
	}
	return nil
}
func queryBFile(conn *go_ora.Connection) error {
	sqlText := "SELECT FILE_ID, FILE_DATA FROM GOORA_TEST_BFILE WHERE FILE_ID = 1"
	stmt := go_ora.NewStmt(sqlText, conn)
	var err error
	defer func() {
		err = stmt.Close()
		if err != nil {
			fmt.Println("Can't close stmt: ", err)
		}
	}()
	rows, err := stmt.Query_(nil)
	if err != nil {
		return err
	}
	var id int64
	var file go_ora.BFile
	for rows.Next_() {
		err = rows.Scan(&id, &file)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id)
		err = workWithBFile(&file)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}
func usage() {
	fmt.Println()
	fmt.Println("bfile")
	fmt.Println("  a code for dealing with BFile data type")
	fmt.Println("  before use the program create directory object in oracle that refer")
	fmt.Println("  some physical directory that contains some text file")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  bfile -server server_url -dir dirName -file fileName`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  bfile -server "oracle://user:pass@server/service_name" -dir "GOORA_TEMP_DIR" -file "test.bin"`)
	fmt.Println()
}
func main() {
	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.StringVar(&dirName, "dir", "", "oracle directory object name")
	flag.StringVar(&fileName, "file", "", "physical file name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	if dirName == "" {
		fmt.Println("Missing -dir option")
		usage()
		os.Exit(1)
	}
	if fileName == "" {
		fmt.Println("Missing -file option")
		usage()
		os.Exit(1)
	}

	fmt.Println("Connection string: ", connStr)
	conn, err := go_ora.NewConnection(connStr)
	if err != nil {
		fmt.Println("Can't create connection: ", err)
		return
	}
	err = conn.Open()
	if err != nil {
		fmt.Println("Can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()

	fmt.Println("Work with BFile object: ")
	bfile, err := go_ora.NewBFile(conn, dirName, fileName)
	if err != nil {
		fmt.Println("Can't create BFile object: ", err)
		return
	}
	err = workWithBFile(bfile)
	if err != nil {
		fmt.Println("BFile error: ", err)
	}
	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table: ", err)
		return
	}
	fmt.Println("Finish create table")
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table: ", err)
		}
		fmt.Println("Finish drop table")
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}
	fmt.Println("Finish insert one row: ")
	fmt.Println("Query BFile: ")
	err = queryBFile(conn)
	if err != nil {
		fmt.Println("Can't query: ", err)
		return
	}
	fmt.Println("BFile as output parameter: ")
	err = BFileOutputPar(conn)
	if err != nil {
		fmt.Println("Can't output: ", err)
		return
	}
}
