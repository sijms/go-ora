// issue 578
// can't read or write UDT when contain string > 251
package TestIssues

import (
	"database/sql"
	"errors"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
	"time"
)

func TestUDTString(t *testing.T) {
	type Data struct {
		DateCreated time.Time `udt:"DATE_CREATED" json:"dateCreated"`
		Qty         int       `udt:"QTY" json:"qty"`
		FirstName   string    `udt:"FIRST_NAME" json:"firstName"`
	}
	var create = func(db *sql.DB) error {
		return execCmd(db, `
	create table TAB_EMPLOYEE_K(
		firstname varchar2(4000),
		DATE_CREATED DATE,
		qty NUMBER
	)`, `create or replace type OBJ_TYPE_DATA as object(
		DATE_CREATED DATE,
		QTY NUMBER,
		FIRST_NAME varchar2(4000)
	)`, `create or replace PROCEDURE USP_SAVE_TEST(
	 P_DATE IN OBJ_TYPE_DATA,
	 p_code out number,
	 p_msg out varchar2
	) AS
	BEGIN
	  INSERT INTO TAB_EMPLOYEE_K(firstname, DATE_CREATED, qty)
	  VALUES (p_date.FIRST_NAME, sysdate, length(p_date.FIRST_NAME));
		p_code := 200;
		p_msg := 'Success';
	exception when others then
		p_code := 500;
		p_msg := 'Error';
	END;`, `CREATE OR REPLACE PROCEDURE USP_LOAD_TEST
	(
	L_DATA OUT OBJ_TYPE_DATA,
	L_QTY NUMBER
	) AS
	  P_DATA OBJ_TYPE_DATA;
	BEGIN
	  P_DATA := OBJ_TYPE_DATA(NULL, NULL, NULL);
	  SELECT FIRSTNAME, DATE_CREATED, QTY
		INTO P_DATA.FIRST_NAME, P_DATA.DATE_CREATED, P_DATA.QTY
		FROM TAB_EMPLOYEE_K
		WHERE QTY = L_QTY;
	  L_DATA := P_DATA;
	END USP_LOAD_TEST;`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db, `DROP TYPE OBJ_TYPE_DATA`,
			`DROP PROCEDURE USP_LOAD_TEST`,
			`DROP PROCEDURE USP_SAVE_TEST`,
			`drop table TAB_EMPLOYEE_K`)
	}

	var callSave = func(db *sql.DB, data Data) error {
		var code int
		var message string
		_, err := db.Exec(`BEGIN USP_SAVE_TEST(:1, :2, :3); END;`, data,
			go_ora.Out{Dest: &code}, go_ora.Out{Dest: &message, Size: 2000})
		if err != nil {
			return err
		}
		if code != 200 || message != "Success" {
			return errors.New("Error occur inside function save")
		}
		return nil
	}
	var callLoad = func(db *sql.DB, qty int) error {
		var obj Data
		_, err := db.Exec(`BEGIN USP_LOAD_TEST(:1, :2); END;`, go_ora.Object{
			Owner: "LAB",
			Name:  "OBJ_TYPE_DATA",
			Value: &obj,
		}, qty)
		if err != nil {
			return err
		}
		if len(obj.FirstName) != qty {
			return fmt.Errorf("expected string length: %d and got %d", qty, len(obj.FirstName))
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
	err = go_ora.RegisterType(db, "OBJ_TYPE_DATA", "", Data{})
	if err != nil {
		t.Error(err)
		return
	}
	// call save with string < 251
	err = callSave(db, Data{time.Now(), 241, strings.Repeat("a", 241)})
	if err != nil {
		t.Error(err)
		return
	}
	// call save with string > 251
	err = callSave(db, Data{time.Now(), 261, strings.Repeat("a", 261)})
	if err != nil {
		t.Error(err)
		return
	}
	// call load for string < 251
	err = callLoad(db, 241)
	if err != nil {
		t.Error(err)
		return
	}
	err = callLoad(db, 261)
	if err != nil {
		t.Error(err)
		return
	}
}
