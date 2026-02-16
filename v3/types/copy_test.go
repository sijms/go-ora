package types

import (
	"reflect"
	"testing"
)

func TestCreateNewType(t *testing.T) {
	var v1 Blob
	rv1 := reflect.ValueOf(&v1).Elem()

	err := createNewType(rv1, reflect.TypeOf(&v1).Elem())
	if err != nil {
		t.Fatal(err)
	}
	//addSomeData(v1)
	t.Log(v1)

	var v2 Clob
	rv2 := reflect.ValueOf(&v2).Elem()
	err = createNewType(rv2, reflect.TypeOf(&v2).Elem())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v2)

	var v3 Vector
	rv3 := reflect.ValueOf(&v3).Elem()
	err = createNewType(rv3, reflect.TypeOf(&v3).Elem())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v3)

	var v4 *int
	rv4 := reflect.ValueOf(&v4).Elem()
	err = createNewType(rv4, reflect.TypeOf(&v4).Elem())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*v4)
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
