// issue 554
package TestIssues

import (
	"database/sql"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
)

func TestLargeUDTArray(t *testing.T) {
	var createTypes = func(db *sql.DB) error {
		return execCmd(db, `
		create or replace type stringsType as object(
			STRING1 varchar2(60),
			STRING2 varchar2(300)
		)`, `create or replace type stringsTypeCol as table of stringsType`)
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db, "DROP TYPE stringsTypeCol", "DROP TYPE stringsType")
	}
	type StringsType struct {
		String1 string `udt:"STRING1"`
		String2 string `udt:"STRING2"`
	}
	var inputPars = func(db *sql.DB, input []StringsType) (int, error) {
		var length int
		_, err := db.Exec(`
	DECLARE
		inp stringsTypeCol;
	BEGIN
		inp := :1;
		:2 := inp.count;
	END;`, input, go_ora.Out{Dest: &length})
		return length, err
	}
	var outputPars = func(db *sql.DB, length int) ([]StringsType, error) {
		var output []StringsType
		_, err := db.Exec(`
	declare
		outp stringsTypeCol;
		ext number;
	begin 
		ext := :1;
		outp := stringsTypeCol();
		outp.extend(ext);
		for n in 1..ext
		loop
			outp(n) := stringsType('string1','string2');
		end loop;

		:2 := outp;
	end;
	`, length, go_ora.Out{Dest: &output})
		return output, err
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
	err = createTypes(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTypes(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = go_ora.RegisterType(db, "stringsType", "stringsTypeCol", StringsType{})
	if err != nil {
		t.Error(err)
		return
	}
	size := 0x200
	output, err := inputPars(db, make([]StringsType, size))
	if err != nil {
		t.Error(err)
		return
	}
	if output != size {
		t.Errorf("expected size: %d and got: %d", size, output)
		return
	}
	outputArray, err := outputPars(db, size)
	if err != nil {
		t.Error(err)
		return
	}
	if len(outputArray) != size {
		t.Errorf("expected size: %d and got: %d", size, len(outputArray))
	}

}
