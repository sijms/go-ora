package go_ora

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestGetValue(t *testing.T) {
	var val interface{}
	var err error
	// passing nil
	val, err = getValue(nil)
	if err != nil {
		t.Error(err)
		return
	}
	if val != nil {
		t.Errorf("passing nil expected nil and get %v", val)
	}
	// passing primitive value
	val, err = getValue(20)
	if err != nil {
		t.Error(err)
		return
	}
	if val != 20 {
		t.Errorf("passing nil expected 20 and get %v", val)
	}
	// passing sql.Null* value
	val, err = getValue(sql.NullString{String: "this is a test", Valid: true})
	if err != nil {
		t.Error(err)
		return
	}
	if val != "this is a test" {
		t.Errorf("passing nil expected 'this is a test' and get %v", val)
	}
	// passing primitive pointer
	temp := 30
	val, err = getValue(&temp)
	if err != nil {
		t.Error(err)
		return
	}
	if val != 30 {
		t.Errorf("passing pointer int64 value 30 and get %v", val)
	}
	// passing sqlNull pointer
	val, err = getValue(&sql.NullFloat64{Float64: 5.78, Valid: true})
	if err != nil {
		t.Error(err)
		return
	}
	if val != 5.78 {
		t.Errorf("passing pointer to sql.NullFloat64 with value 5.78 and get %v", val)
	}
}

func TestGetInt(t *testing.T) {
	checkGetInt := func(testedValue interface{}, expectedValue int64) error {
		val, err := getInt(testedValue)
		if err != nil {
			return err
		}
		if val != expectedValue {
			return fmt.Errorf("passing %v to getInt and recieve %v", testedValue, val)
		}
		return nil
	}
	// passing nil
	err := checkGetInt(nil, 0)
	if err != nil {
		t.Error(err)
	}
	// passing float
	err = checkGetInt(3.35, 3)
	if err != nil {
		t.Error(err)
	}
	// passing float32
	err = checkGetInt(float32(5.78), 5)
	if err != nil {
		t.Error(err)
	}
	//passing int
	err = checkGetInt(5, 5)
	if err != nil {
		t.Error(err)
	}
	// passing int64
	err = checkGetInt(int64(7), 7)
	if err != nil {
		t.Error(err)
	}
	// passing *float
	tempFloat := float32(5.78)
	err = checkGetInt(&tempFloat, 5)
	if err != nil {
		t.Error(err)
	}
	// passing *int
	tempInt := 9
	err = checkGetInt(&tempInt, 9)
	if err != nil {
		t.Error(err)
	}
	// passing NullFlat
	err = checkGetInt(sql.NullFloat64{5.78, true}, 5)
	if err != nil {
		t.Error(err)
	}
	// passing NullInt
	err = checkGetInt(sql.NullInt64{2, true}, 2)
	if err != nil {
		t.Error(err)
	}
	// passing *NullFloat
	err = checkGetInt(&sql.NullFloat64{8.44, true}, 8)
	if err != nil {
		t.Error(err)
	}
	// passing *NullInt
	err = checkGetInt(&sql.NullInt64{10, true}, 10)
	if err != nil {
		t.Error(err)
	}
	// passing string
	err = checkGetInt("10", 10)
	if err != nil {
		t.Error(err)
	}
	// passing *string
	tempString := "11"
	err = checkGetInt(&tempString, 11)
	if err != nil {
		t.Error(err)
	}

	// passing NullString
	err = checkGetInt(sql.NullString{"12", true}, 12)
	if err != nil {
		t.Error(err)
	}
	// passing *NullString
	err = checkGetInt(&sql.NullString{"13", true}, 13)
	if err != nil {
		t.Error(err)
	}
}
func TestGetFloat(t *testing.T) {
	checkGetFloat := func(testedValue interface{}, expectedValue float64) error {
		val, err := getFloat(testedValue)
		if err != nil {
			return err
		}
		if val != expectedValue {
			return fmt.Errorf("passing %v to getFloat and recieve %v", testedValue, val)
		}
		return nil
	}
	// passing nil
	err := checkGetFloat(nil, 0)
	if err != nil {
		t.Error(err)
	}
	// passing float
	err = checkGetFloat(3.35, 3.35)
	if err != nil {
		t.Error(err)
	}
	// passing float32
	err = checkGetFloat(float32(5.78), 5.78)
	if err != nil {
		t.Error(err)
	}
	//passing int
	err = checkGetFloat(5, 5)
	if err != nil {
		t.Error(err)
	}
	// passing int64
	err = checkGetFloat(int64(7), 7)
	if err != nil {
		t.Error(err)
	}
	// passing *float
	tempFloat := float32(5.78)
	err = checkGetFloat(&tempFloat, 5.78)
	if err != nil {
		t.Error(err)
	}
	// passing *int
	tempInt := 9
	err = checkGetFloat(&tempInt, 9)
	if err != nil {
		t.Error(err)
	}
	// passing NullFlat
	err = checkGetFloat(sql.NullFloat64{5.43, true}, 5.43)
	if err != nil {
		t.Error(err)
	}
	// passing NullInt
	err = checkGetFloat(sql.NullInt64{2, true}, 2)
	if err != nil {
		t.Error(err)
	}
	// passing *NullFloat
	err = checkGetFloat(&sql.NullFloat64{8.44, true}, 8.44)
	if err != nil {
		t.Error(err)
	}
	// passing *NullInt
	err = checkGetFloat(&sql.NullInt64{10, true}, 10)
	if err != nil {
		t.Error(err)
	}
	// passing string
	err = checkGetFloat("10.8", 10.8)
	if err != nil {
		t.Error(err)
	}
	// passing *string
	tempString := "11.70"
	err = checkGetFloat(&tempString, 11.7)
	if err != nil {
		t.Error(err)
	}

	// passing NullString
	err = checkGetFloat(sql.NullString{"12.8", true}, 12.8)
	if err != nil {
		t.Error(err)
	}
	// passing *NullString
	err = checkGetFloat(&sql.NullString{"13.2", true}, 13.2)
	if err != nil {
		t.Error(err)
	}
}

func TestSetString(t *testing.T) {
	var intVar int
	err := setString(reflect.ValueOf(&intVar).Elem(), "15")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(intVar)
	var uint8Var uint8
	err = setString(reflect.ValueOf(&uint8Var).Elem(), "17")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(uint8Var)
	var int32Var sql.NullInt32
	err = setString(reflect.ValueOf(&int32Var).Elem(), "18")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(int32Var)
}
func TestSetFieldValue(t *testing.T) {
	var testString *sql.NullString
	err := setFieldValue(reflect.ValueOf(&testString).Elem(), nil, "this is a test")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(*testString)
	type test1 struct {
		Id   int
		Name string
		Date sql.NullTime
	}
	var test = test1{}
	sField := reflect.Indirect(reflect.ValueOf(&test))
	err = setFieldValue(sField.Field(0), nil, int64(15))
	if err != nil {
		t.Error(err)
		return
	}
	err = setFieldValue(sField.Field(1), nil, "this is a test")
	if err != nil {
		t.Error(err)
		return
	}
	err = setFieldValue(sField.Field(2), nil, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(test)
}

func TestSetArray(t *testing.T) {
	var array []int
	pars := []ParameterInfo{}
	pars = append(pars, ParameterInfo{oPrimValue: int64(5)})
	pars = append(pars, ParameterInfo{oPrimValue: int64(6)})
	pars = append(pars, ParameterInfo{oPrimValue: int64(7)})
	err := setArray(reflect.ValueOf(&array), pars)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(array)
}

func TestSetNull(t *testing.T) {
	var x int = 10
	var xx float64 = 3.3
	var xxx string = "test"
	var rx = reflect.ValueOf(&x).Elem()
	var rxx = reflect.ValueOf(&xx).Elem()
	var rxxx = reflect.ValueOf(&xxx).Elem()
	//var xx = reflect.ValueOf(float64(3.3))
	//var xxx = reflect.ValueOf("test")
	setNull(rx)
	setNull(rxx)
	setNull(rxxx)
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
