package TestIssues

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"testing"
)

func TestLob(t *testing.T) {
	fileData, err := os.ReadFile("clob.json")
	if err != nil {
		t.Error(err)
		return
	}
	clob := strings.Repeat(string(fileData), 10)
	blob := bytes.Repeat(fileData, 10)
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE GOORA_TEMP_LOB(
	ID	number(10)	NOT NULL,
	DATA1 CLOB,
	DATA2 CLOB,
	DATA3 BLOB,
	DATA4 BLOB,
	PRIMARY KEY(ID)
	)`)
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_TEMP_LOB purge")
	}

	var insert = func(db *sql.DB) error {
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
		_, err = db.Exec("INSERT INTO GOORA_TEMP_LOB(ID, DATA1, DATA2, DATA3, DATA4) VALUES(:ID, :DATA1, :DATA2, :DATA3, :DATA4)", input)
		return err
	}

	var sqlQuery = func(db *sql.DB) error {
		sqlText := "SELECT ID, DATA1, DATA2, DATA3, DATA4 FROM GOORA_TEMP_LOB WHERE ID < 3"
		rows, err := db.Query(sqlText)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		var (
			id    int
			data1 sql.NullString
			data2 sql.NullString
			data3 []byte
			data4 []byte
		)
		for rows.Next() {
			err = rows.Scan(&id, &data1, &data2, &data3, &data4)
			if err != nil {
				return err
			}
			if id == 2 {
				if data1.Valid == true || data2.Valid == true || data3 != nil || data4 != nil {
					return errors.New("expected null and got a value")
				}
			} else {
				if data1.String != "this is a test" {
					return fmt.Errorf("expected: %s and got: %v", "this is a test", data1.String)
				}
				if !bytes.Equal(data3, []byte("this is a test")) {
					return fmt.Errorf("expected: %v and got: %v", []byte("this is a test"), data3)
				}
				if data2.String != clob {
					return errors.New("un-equal large clob data")
				}
				if !bytes.Equal(data4, blob) {
					return errors.New("un-equal large blob data")
				}
			}
		}
		return nil
	}
	var parameterQuery = func(db *sql.DB, id int) error {
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
		if id%2 == 0 {
			if temp.Data1.Valid == true || temp.Data2.Valid == true || temp.Data3 != nil || temp.Data4.Data != nil {
				return errors.New("expected null and got a value")
			}
		} else {
			if temp.Data1.String != "this is a test" {
				return fmt.Errorf("expected: %s and got: %v", "this is a test", temp.Data1.String)
			}
			if !bytes.Equal(temp.Data3, []byte("this is a test")) {
				return fmt.Errorf("expected: %v and got: %v", []byte("this is a test"), temp.Data3)
			}
			if temp.Data2.String != clob {
				return errors.New("un-equal large clob data")
			}
			if !bytes.Equal(temp.Data4.Data, blob) {
				return errors.New("un-equal large blob data")
			}
		}
		return nil
	}
	urlOptions["lob fetch"] = "post"

	defer func() {
		urlOptions["lob fetch"] = "pre"
	}()
	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	err = createTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = sqlQuery(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = parameterQuery(db, 1)
	if err != nil {
		t.Error(err)
		return
	}
	err = parameterQuery(db, 2)
	if err != nil {
		t.Error(err)
		return
	}
}
