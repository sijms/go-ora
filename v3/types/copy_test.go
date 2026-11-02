package types

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestCopyRejectsNumericOverflow(t *testing.T) {
	var i8 int8
	var u8 uint8
	var i int
	var f32 float32
	cases := []struct {
		name string
		dest any
		src  any
	}{
		{"int64 300 into int8", &i8, int64(300)},
		{"int64 -1 into uint8", &u8, int64(-1)},
		{"uint64 max into int", &i, uint64(1<<64 - 1)},
		{"float64 1e30 into int", &i, float64(1e30)},
		{"float64 -1.5 into uint8", &u8, float64(-1.5)},
		{"float64 1e300 into float32", &f32, float64(1e300)},
	}
	for _, tc := range cases {
		if err := Copy(tc.dest, tc.src); err == nil {
			t.Errorf("%s: got %v, want an overflow error", tc.name, reflect.ValueOf(tc.dest).Elem())
		}
	}
}

func TestCopyKeepsValuesThatFit(t *testing.T) {
	var i8 int8
	var u8 uint8
	var i int
	var f32 float32
	if err := Copy(&i8, int64(-128)); err != nil || i8 != -128 {
		t.Errorf("int64 -128 into int8: got %d, %v", i8, err)
	}
	if err := Copy(&u8, int64(255)); err != nil || u8 != 255 {
		t.Errorf("int64 255 into uint8: got %d, %v", u8, err)
	}
	// truncating a fraction matches v2's getInt
	if err := Copy(&i, float64(1.9)); err != nil || i != 1 {
		t.Errorf("float64 1.9 into int: got %d, %v", i, err)
	}
	if err := Copy(&f32, float64(1.5)); err != nil || f32 != 1.5 {
		t.Errorf("float64 1.5 into float32: got %v, %v", f32, err)
	}
}

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
