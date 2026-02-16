package TestIssues

import (
	"github.com/sijms/go-ora/dbms"
	"testing"
)

func TestDBMS_OUTPUT(t *testing.T) {

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
	output, err := dbms.NewOutput(db, 0x7FFF)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = output.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	err = execCmd(db, `BEGIN
DBMS_OUTPUT.PUT_LINE('this is a test');
END;`)
	if err != nil {
		t.Error(err)
		return
	}
	line, err := output.GetOutput()
	if err != nil {
		t.Error(err)
		return
	}
	if line != "this is a test\n" {
		t.Errorf("expected: %s and got: %s", "this is a test", line)
	}

	err = execCmd(db, `BEGIN
DBMS_OUTPUT.PUT_LINE('this is a test2');
END;`)
	if err != nil {
		t.Error(err)
		return
	}
	line, err = output.GetOutput()
	if err != nil {
		t.Error(err)
		return
	}
	if line != "this is a test2\n" {
		t.Errorf("expected: %s and got: %s", "this is a test2", line)
	}
}
