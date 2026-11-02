package types

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestCreateNewType(t *testing.T) {
	var v1 *Blob
	var v2 *Clob
	var v3 *Vector
	var v4 *int
	for _, dst := range []interface{}{&v1, &v2, &v3, &v4} {
		rv := reflect.ValueOf(dst).Elem()
		if err := createNewType(rv, rv.Type()); err != nil {
			t.Fatalf("%T: %v", dst, err)
		}
		if rv.IsNil() {
			t.Errorf("%T: expected a new value and got nil", dst)
		}
	}
	var notPtr int
	rv := reflect.ValueOf(&notPtr).Elem()
	if err := createNewType(rv, rv.Type()); err == nil {
		t.Error("expected an error for a non-pointer destination")
	}
}

func TestCopy(t *testing.T) {
	var v1 string
	err := Copy(&v1, "foo")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v1)

	var v2 Blob
	err = Copy(&v2, []byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v2)
}

func TestCopyStringFromOthers(t *testing.T) {
	var v_string string
	err := Copy(&v_string, 100)
	if err != nil {
		t.Fatal(err)
	}
	if v_string != "100" {
		t.Fatal("error copy int to string")
	}
	err = Copy(&v_string, int64(100))
	if err != nil {
		t.Fatal(err)
	}
	if v_string != "100" {
		t.Fatal("error copy int64 to string")
	}
	err = Copy(&v_string, float64(1.1))
	if err != nil {
		t.Fatal(err)
	}
	if v_string != "1.100000" {
		t.Fatal("error copy float64 to string")
	}
	err = Copy(&v_string, true)
	if err != nil {
		t.Fatal(err)
	}
	if v_string != "true" {
		t.Fatal("error copy true to string")
	}
}

func TestRCopyString(t *testing.T) {
	var intVar int
	if err := RCopy(reflect.ValueOf(&intVar).Elem(), "15"); err != nil {
		t.Fatal(err)
	}
	if intVar != 15 {
		t.Errorf("expected 15 and got %d", intVar)
	}
	var uint8Var uint8
	if err := RCopy(reflect.ValueOf(&uint8Var).Elem(), "17"); err != nil {
		t.Fatal(err)
	}
	if uint8Var != 17 {
		t.Errorf("expected 17 and got %d", uint8Var)
	}
	var int32Var sql.NullInt32
	if err := RCopy(reflect.ValueOf(&int32Var).Elem(), "18"); err != nil {
		t.Fatal(err)
	}
	if !int32Var.Valid || int32Var.Int32 != 18 {
		t.Errorf("expected {18 true} and got %v", int32Var)
	}
}

func TestRCopyPointerNullString(t *testing.T) {
	var testString *sql.NullString
	if err := RCopy(reflect.ValueOf(&testString).Elem(), "this is a test"); err != nil {
		t.Fatal(err)
	}
	if testString == nil || !testString.Valid || testString.String != "this is a test" {
		t.Errorf("expected {this is a test true} and got %v", testString)
	}
}

func TestRCopyStructFields(t *testing.T) {
	type test1 struct {
		Id   int
		Name string
		Date sql.NullTime
	}
	test := test1{}
	sField := reflect.Indirect(reflect.ValueOf(&test))
	if err := RCopy(sField.Field(0), int64(15)); err != nil {
		t.Fatal(err)
	}
	if err := RCopy(sField.Field(1), "this is a test"); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Truncate(time.Second)
	if err := RCopy(sField.Field(2), now.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if test.Id != 15 || test.Name != "this is a test" || !test.Date.Valid || !test.Date.Time.Equal(now) {
		t.Errorf("unexpected result %+v", test)
	}
}

func TestRCopyNil(t *testing.T) {
	var x int = 10
	var xx float64 = 3.3
	var xxx string = "test"
	for _, v := range []reflect.Value{reflect.ValueOf(&x).Elem(), reflect.ValueOf(&xx).Elem(), reflect.ValueOf(&xxx).Elem()} {
		if err := RCopy(v, nil); err != nil {
			t.Fatal(err)
		}
	}
	if x != 0 {
		t.Error("expected 0 get ", x)
	}
	if xx != 0 {
		t.Error("expected 0 get: ", xx)
	}
	if xxx != "" {
		t.Error("expected empty get: ", xxx)
	}
}
