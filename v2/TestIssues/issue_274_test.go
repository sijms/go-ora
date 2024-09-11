package TestIssues

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIssue274(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_274(
    ID NUMBER(10) NOT NULL,
    NAME VARCHAR(200),
    VAL NUMBER(10, 2),
    LDATE DATE,
    PRIMARY KEY (ID)
    )`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "DROP TABLE TTB_274 PURGE")
	}
	dbLock := func(db *sql.DB) error {
		execCtx, execCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer execCancel()
		_, err := db.ExecContext(execCtx, "begin DBMS_LOCK.sleep(5); end;")
		return err
	}
	insert := func(db *sql.DB, rowNum int) error {
		type TTB_274 struct {
			ID   int       `db:"ID"`
			Name string    `db:"NAME"`
			Val  float64   `db:"VAL"`
			Date time.Time `db:"LDATE"`
		}
		interval := 1.1
		data := make([]TTB_274, rowNum)
		for index := range data {
			data[index].ID = index + 1
			data[index].Name = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			data[index].Val = float64(index) + interval
			data[index].Date = time.Now()
		}
		_, err := db.Exec("INSERT INTO TTB_274(ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)", data)
		if err != nil {
			return err
		}
		return nil
	}
	query := func(db *sql.DB, rowNum int) error {
		execCtx, execCancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer execCancel()
		rows, err := db.QueryContext(execCtx, `SELECT ID, NAME, VAL, LDATE FROM TTB_274 WHERE ID < :1 ORDER BY ID`, rowNum)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		var id int
		var name string
		var val float64
		var date time.Time
		for rows.Next() {
			err = rows.Scan(&id, &name, &val, &date)
			if err != nil {
				return err
			}
		}
		return rows.Err()
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
	err = dbLock(db)
	if err != nil {
		t.Log(err)
	}
	err = insert(db, 100000)
	if err != nil {
		t.Error(err)
		return
	}
	err = query(db, 10000)
	if err != nil {
		t.Error(err)
		return
	}
}
