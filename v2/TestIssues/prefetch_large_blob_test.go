package TestIssues

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPrefetchLargeBlob(t *testing.T) {
	refDate := time.Now()
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_PREFETCH_LARGE_BLOB (
	ID			number(10)	NOT NULL,
	DATA1		BLOB,
	SEP1		number(10, 3),
	DATA2       CLOB,
	SEP2        varchar2(100),
	DATA3       NCLOB,
	SEP3        DATE,
	DATA4       LONG,
	SEP4        nvarchar2(100),
	PRIMARY KEY(ID)
	) NOCOMPRESS`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "drop table TTB_PREFETCH_LARGE_BLOB purge")
	}
	insert := func(db *sql.DB) error {
		type Table struct {
			Id    int       `db:"ID"`
			Data1 []byte    `db:"DATA1,blob"`
			Sep1  float32   `db:"SEP1"`
			Data2 string    `db:"DATA2,clob"`
			Sep2  string    `db:"SEP2"`
			Data3 string    `db:"DATA3,nclob"`
			Sep3  time.Time `db:"SEP3"`
			Data4 string    `db:"DATA4"`
			Sep4  string    `db:"SEP4,nvarchar"`
		}
		length := 10
		input := make([]Table, 0, length)

		for x := 0; x < length; x++ {
			temp := Table{
				Id:    x + 1,
				Data1: bytes.Repeat([]byte{35, 36}, 50000),
				Sep1:  float32(x) + 1.12,
				Data2: strings.Repeat("AB", 50000),
				Sep2:  "test",
				Data3: strings.Repeat("안녕하세요", 50000),
				Sep3:  refDate,
				Data4: strings.Repeat("LONG", 50),
				Sep4:  "안녕하세요",
			}
			input = append(input, temp)
		}
		_, err := db.Exec(`INSERT INTO TTB_PREFETCH_LARGE_BLOB(ID, DATA1, SEP1, DATA2, SEP2, DATA3, SEP3, DATA4, SEP4) 
VALUES(:ID, :DATA1, :SEP1, :DATA2, :SEP2, :DATA3, :SEP3, :DATA4, :SEP4)`, input)
		return err
	}
	query := func(db *sql.DB) error {
		var (
			id    int
			data1 []byte
			sep1  float32
			data2 string
			sep2  string
			data3 string
			sep3  time.Time
			data4 string
			sep4  string
		)
		rows, err := db.Query(`SELECT ID, DATA1, SEP1, DATA2, SEP2, DATA3, SEP3, DATA4, SEP4 FROM TTB_PREFETCH_LARGE_BLOB`)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		index := 1
		expectedData1 := bytes.Repeat([]byte{35, 36}, 50000)
		expectedData2 := strings.Repeat("AB", 50000)
		expectedData3 := strings.Repeat("안녕하세요", 50000)
		expectedData4 := strings.Repeat("LONG", 50)
		expectedSep2 := "test"
		expectedSep4 := "안녕하세요"
		for rows.Next() {
			err = rows.Scan(&id, &data1, &sep1, &data2, &sep2, &data3, &sep3, &data4, &sep4)
			if err != nil {
				return err
			}
			if id != index {
				return fmt.Errorf("expected ID: %d and got: %d", index, id)
			}
			if !bytes.Equal(expectedData1, data1) {
				return errors.New("unequal values of data1")
			}
			if sep1 != float32(index)+0.12 {
				return fmt.Errorf("expected SEP1: %f and got: %f", float32(index)+1.12, sep1)
			}
			if expectedData2 != data2 {
				return errors.New("unequal values of data2")
			}
			if sep2 != expectedSep2 {
				return fmt.Errorf("expected SEP2: %s and got: %s", expectedSep2, sep2)
			}
			if data3 != expectedData3 {
				return errors.New("unequal values of data3")
			}
			if !isEqualTime(sep3, refDate, false) {
				return fmt.Errorf("expected SEP3: %v and got: %v", refDate, sep3)
			}
			if data4 != expectedData4 {
				return errors.New("unequal values of data4")
			}
			if sep4 != expectedSep4 {
				return fmt.Errorf("expected SEP4: %s and got: %s", expectedSep4, sep4)
			}
			index++
		}
		return nil
	}

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
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
