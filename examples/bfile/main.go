package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `create table GOORA_TEST_BFILE (
    FILE_ID NUMBER(10) NOT NULL,
    FILE_DATA BFILE
)`
	_, err := conn.Exec(sqlText)
	if err == nil {
		fmt.Println("Finish create table: ", time.Now().Sub(t))
	}
	return err
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`drop table GOORA_TEST_BFILE purge`)
	if err == nil {
		fmt.Println("Finish drop table: ", time.Now().Sub(t))
	}
	return err
}

func insertData(conn *sql.DB, id int, dirName, fileName string) error {
	t := time.Now()
	var file *go_ora.BFile
	var err error
	if len(dirName) > 0 {
		file, err = go_ora.CreateBFile(conn, dirName, fileName)
		if err != nil {
			return err
		}
	} else {
		file = go_ora.CreateNullBFile()
	}
	sqlText := `INSERT INTO GOORA_TEST_BFILE(FILE_ID, FILE_DATA) VALUES(:1, :2)`
	_, err = conn.Exec(sqlText, id, file)
	if err == nil {
		fmt.Println("Finish insert row: ", time.Now().Sub(t))
	}
	return err
}

func deleteData(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec("DELETE FROM GOORA_TEST_BFILE")
	if err != nil {
		return err
	}
	fmt.Println("Finish delete: ", time.Now().Sub(t))
	return nil
}

func query(db *sql.DB) error {
	t := time.Now()
	sqlText := "SELECT FILE_ID, FILE_DATA FROM GOORA_TEST_BFILE"
	rows, err := db.Query(sqlText)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close rows: ", err)
		}
	}()
	var id int64
	var file go_ora.BFile
	for rows.Next() {
		err = rows.Scan(&id, &file)
		if err != nil {
			return err
		}
		fmt.Println("id: ", id, "\tDir: ", file.GetDirName(), "\tFile: ", file.GetFileName(), "\tValid: ", file.Valid)
	}
	fmt.Println("Finish query: ", time.Now().Sub(t))
	return nil
}

func BFileOutputPar(db *sql.DB) error {
	sqlText := `BEGIN SELECT FILE_ID, FILE_DATA INTO :1, :2 FROM GOORA_TEST_BFILE WHERE FILE_ID = 2; END;`
	var (
		id   int64
		file go_ora.BFile
	)
	_, err := db.Exec(sqlText, go_ora.Out{Dest: &id}, go_ora.Out{Dest: &file})
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

func workWithBFile(file *go_ora.BFile) error {
	if !file.Valid {
		fmt.Println("BFile value is nil")
		return nil
	}
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

func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("Can't open database: ", err)
		return
	}

	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()
	//fmt.Println("Work with BFile object: ")
	//bfile, err := go_ora.NewBFile2(db, os.Getenv("DIR_NAME"), os.Getenv("FILE_NAME"))
	//if err != nil {
	//	fmt.Println("Can't create bfile object: ", err)
	//	return
	//}
	//err = workWithBFile(bfile)
	//if err != nil {
	//	fmt.Println("BFile error: ", err)
	//	return
	//}

	err = createTable(db)
	if err != nil {
		fmt.Println("Can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			fmt.Println("Can't drop table: ", err)
		}
	}()
	err = deleteData(db)
	if err != nil {
		fmt.Println("Can't delete data: ", err)
		return
	}
	err = insertData(db, 1, os.Getenv("DIR_NAME"), os.Getenv("FILE_NAME"))
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}
	err = insertData(db, 2, "", "")
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("Can't query: ", err)
		return
	}
	err = BFileOutputPar(db)
	if err != nil {
		fmt.Println("Can't output: ", err)
		return
	}
}
