package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"strings"
	"testing"
)

func TestVarray(t *testing.T) {
	var createTypes = func(db *sql.DB) error {
		return execCmd(db, `create type StringArray as VARRAY(10) of varchar2(20) not null`)
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db, `drop type StringArray`)
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
	err = go_ora.RegisterType(db, "varchar2", "StringArray", nil)
	if err != nil {
		t.Error(err)
		return
	}
	var output []string
	_, err = db.Exec(`
DECLARE
	l_array StringArray := StringArray();
BEGIN
	for x in 1..10 loop
		l_array.extend;
		l_array(x) := 'string_' || x;
	end loop;
	:1 := l_array;
END;`, go_ora.Object{Name: "StringArray", Value: &output})
	if err != nil {
		t.Error(err)
		return
	}
	var expected string
	for index, item := range output {
		expected = fmt.Sprintf("string_%d", index+1)
		if !strings.EqualFold(item, expected) {
			t.Errorf("expected: %s and got: %s", expected, item)
			return
		}
	}
}
