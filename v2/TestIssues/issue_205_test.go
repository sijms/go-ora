package TestIssues

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestIssue205(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_205(
	ID NUMBER(10) NOT NULL,
		BIG_TEXT LONG,
		PRIMARY KEY (ID)
	) NOCOMPRESS`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_205 PURGE`)
	}
	insert := func(db *sql.DB) error {
		type ttb_205 struct {
			Id   int            `db:"ID"`
			Name sql.NullString `db:"NAME"`
		}
		data := make([]ttb_205, 100)
		for index := range data {
			data[index].Id = index + 1
			if (index+1)%2 == 0 {
				data[index].Name.Valid = false
			} else {
				data[index].Name.String = strings.Repeat("a", 0x5000)
				data[index].Name.Valid = true
			}
		}
		_, err := db.Exec("INSERT ALL INTO TTB_205(BIG_TEXT, ID) VALUES(:NAME, :ID) SELECT * FROM DUAL", data)
		if err != nil {
			return err
		}
		return nil
	}
	sqlQuery := func(db *sql.DB) error {
		rows, err := db.Query("SELECT BIG_TEXT, ID FROM TTB_205 WHERE ID < 3")
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
		var name sql.NullString
		for rows.Next() {
			err = rows.Scan(&name, &id)
			if err != nil {
				return err
			}
			if id == 1 {
				expected := strings.Repeat("a", 0x5000)
				if name.String != expected {
					return errors.New("long data is not correct")
				}
			} else {
				if name.Valid {
					return errors.New("long should be null")
				}
			}
		}
		return rows.Err()
	}
	outputQuery := func(db *sql.DB, id int) error {
		temp := struct {
			ID   int            `db:"ID,,,output"`
			Name sql.NullString `db:"NAME,,100000,output"`
		}{}
		_, err := db.Exec("BEGIN SELECT ID, BIG_TEXT INTO :ID, :NAME FROM TTB_205 WHERE ID = :IID; END;", &temp,
			sql.Named("IID", id))
		if err != nil {
			return err
		}
		if temp.ID%2 == 0 {
			if temp.Name.Valid {
				return errors.New("long should be null")
			}
		} else {
			expected := strings.Repeat("a", 0x5000)
			if temp.Name.String != expected {
				return errors.New("long data is not correct")
			}
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
	err = sqlQuery(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = outputQuery(db, 1)
	if err != nil {
		t.Error(err)
		return
	}
	err = outputQuery(db, 2)
	if err != nil {
		t.Error(err)
		return
	}
}
