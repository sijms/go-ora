package TestIssues

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

func TestLongInput(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_557 (
			ID NUMBER(10), 
			DATA_BLOB BLOB, 
			DATA_CLOB CLOB, 
			DATA_NCLOB NCLOB, 
			LDATE DATE
		) NOCOMPRESS`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "drop table TTB_557 purge")
	}

	raw := func(db *sql.DB, data []byte) error {
		_, err := db.Exec(`INSERT INTO TTB_557(DATA_BLOB, ID, LDATE) VALUES(:1, :2, :3)`, data, 1, time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_BLOB, LDATE) VALUES(:1, :2, :3)`, 2, data, time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_BLOB) VALUES(:1, :2, :3)`, 3, time.Now(), data)
		if err != nil {
			return err
		}
		rows, err := db.Query(`SELECT DATA_BLOB FROM TTB_557`)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error("can't close rows: ", err)
			}
		}()
		var output []byte
		for rows.Next() {
			err = rows.Scan(&output)
			if err != nil {
				return err
			}
			if len(output) != len(data) {
				return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
			}
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		err = execCmd(db, "DELETE FROM TTB_557")
		if err != nil {
			return err
		}
		t.Logf("finish insert %d raw at (begin, middle and end) of sql\n", len(data))
		return nil
	}

	varchar := func(db *sql.DB, data string) error {
		_, err := db.Exec(`INSERT INTO TTB_557(DATA_CLOB, ID, LDATE) VALUES(:1, :2, :3)`, data, 1, time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_CLOB, LDATE) VALUES(:1, :2, :3)`, 2, data, time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_CLOB) VALUES(:1, :2, :3)`, 3, time.Now(), data)
		if err != nil {
			return err
		}
		rows, err := db.Query(`SELECT DATA_CLOB FROM TTB_557`)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error("can't close rows: ", err)
			}
		}()
		var output string
		for rows.Next() {
			err = rows.Scan(&output)
			if err != nil {
				return err
			}
			if len(output) != len(data) {
				return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
			}
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		err = execCmd(db, "DELETE FROM TTB_557")
		if err != nil {
			return err
		}
		t.Logf("finish insert %d varchar string at (begin, middle and end) of sql", len(data))
		return nil
	}

	nvarchar := func(db *sql.DB, data string) error {
		_, err := db.Exec(`INSERT INTO TTB_557(DATA_NCLOB, ID, LDATE) VALUES(:1, :2, :3)`, go_ora.NVarChar(data), 1, time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, DATA_NCLOB, LDATE) VALUES(:1, :2, :3)`, 2, go_ora.NVarChar(data), time.Now())
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO TTB_557(ID, LDATE, DATA_NCLOB) VALUES(:1, :2, :3)`, 3, time.Now(), go_ora.NVarChar(data))
		if err != nil {
			return err
		}
		rows, err := db.Query(`SELECT DATA_NCLOB FROM TTB_557`)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		var output string
		for rows.Next() {
			err = rows.Scan(&output)
			if err != nil {
				return err
			}
			if len(output) != len(data) {
				return fmt.Errorf("expected data size: %d and got: %d", len(data), len(output))
			}
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		err = execCmd(db, "DELETE FROM TTB_557")
		if err != nil {
			return err
		}
		t.Logf("finish insert %d nvarchar string at (begin, middle and end) of sql", len(data))
		return nil
	}
	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err := db.Close()
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
	// insert short string
	err = varchar(db, strings.Repeat("*", 10000))
	if err != nil {
		t.Error("varchar short insert: ", err)
		return
	}
	// insert long string
	err = varchar(db, strings.Repeat("*", 40000))
	if err != nil {
		t.Error("varchar long insert: ", err)
		return
	}

	// insert short NVARCHAR
	err = nvarchar(db, strings.Repeat("早上好", 2000))
	if err != nil {
		t.Error("nvarchar short insert: ", err)
		return
	}
	// insert long NVARCHAR
	err = nvarchar(db, strings.Repeat("早上好", 40000))
	if err != nil {
		t.Error("nvarchar long insert: ", err)
		return
	}

	// insert short RAW
	err = raw(db, bytes.Repeat([]byte{3}, 10000))
	if err != nil {
		t.Error("raw short insert: ", err)
		return
	}

	// insert long RAW
	err = raw(db, bytes.Repeat([]byte{3}, 40000))
	if err != nil {
		t.Error("raw long insert: ", err)
		return
	}
}
