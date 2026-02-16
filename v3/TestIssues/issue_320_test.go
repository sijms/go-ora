package TestIssues

import (
	"context"
	"database/sql"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
)

func TestIssue320(t *testing.T) {
	type Mat struct {
		Id       sql.NullString
		Response go_ora.Clob
	}
	var MatCol = func(colname string, mat *Mat) interface{} {
		switch colname {
		case "ID":
			return &mat.Id
		case "RESPONSE":
			return &mat.Response
		default:
			return new(string)
		}
	}
	var create = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_320(
    ID          varchar2(100),
    RESPONSE    CLOB
	)`,
			`CREATE OR REPLACE FUNCTION TP_320 RETURN SYS_REFCURSOR AS
    L_CURSOR SYS_REFCURSOR;
BEGIN
	OPEN L_CURSOR FOR SELECT ID, RESPONSE FROM TTB_320;
	return L_CURSOR;
END TP_320;`,
			`INSERT INTO TTB_320(ID, RESPONSE) VALUES('1', 'THIS IS A TEST')`,
		)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db,
			"DROP FUNCTION TP_320",
			"DROP TABLE TTB_320 PURGE",
		)
	}
	var query = func(db *sql.DB) error {
		var cursor go_ora.RefCursor
		_, err := db.Exec(`BEGIN :1 := TP_320(); END;`, sql.Out{Dest: &cursor})
		if err != nil {
			return err
		}
		rows, err := go_ora.WrapRefCursor(context.Background(), db, &cursor)
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		columns, err := rows.Columns()
		if err != nil {
			return err
		}
		for rows.Next() {
			mat := Mat{}
			values := make([]interface{}, len(columns))
			for i, v := range columns {
				values[i] = MatCol(v, &mat)
			}
			err = rows.Scan(values...)
			if err != nil {
				return err
			}
			t.Log(values)
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
	err = create(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = drop(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
