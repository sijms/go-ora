package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_343(
	ID	number(10)	NOT NULL,
	NAME		VARCHAR2(500),
	NAME2       NVARCHAR2(500),
	VAL			number(10,2),
	VAL2        number(10),
	LDATE   		date,
	LDATE2          TIMESTAMP,
	LDATE3          TIMESTAMP WITH TIME ZONE,
	PRIMARY KEY(ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_343 purge")
	if err != nil {
		return err
	}
	fmt.Println("finish drop table: ", time.Now().Sub(t))
	return nil
}

func query(conn *sql.DB) error {
	t := time.Now()
	rows, err := conn.Query(`SELECT ID, NAME, NAME2, VAL, VAL2, LDATE, LDATE2, LDATE3 FROM TEMP_TABLE_343`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var (
		id                 int64
		name, name2        string
		val                sql.NullFloat64
		val2               int
		date, date2, date3 time.Time
	)
	for rows.Next() {
		err = rows.Scan(&id, &name, &name2, &val, &val2, &date, &date2, &date3)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id, "\tName: ", name, "\tName2: ", name2, "\tVal: ", val, "\tVal2: ", val2)
		fmt.Println("Date: ", date, "\tDate2: ", date2, "\tDate3: ", date3)
	}
	fmt.Println("finish query: ", time.Now().Sub(t))
	return nil
}
func query2(conn *sql.DB) error {
	t := time.Now()
	temp := struct {
		Id    int       `db:"ID,,,output"`
		Name  string    `db:"NAME,,500,output"`
		Name2 string    `db:"NAME2,nvarchar,500,output"`
		Val   float32   `db:"VAL,,,output"`
		Val2  int       `db:"VAL2,,,output"`
		Date  time.Time `db:"LDATE,,,output"`
		Date2 time.Time `db:"LDATE2,timestamp,,output"`
		Date3 time.Time `db:"LDATE3,timestamptz,,output"`
	}{}

	sqlText := `BEGIN
SELECT ID, NAME, NAME2, VAL, VAL2, LDATE, LDATE2, LDATE3 INTO :ID,:NAME,:NAME2,:VAL,:VAL2,:LDATE,:LDATE2,:LDATE3
    FROM TEMP_TABLE_343;
end;`
	//sqlText := `BEGIN SELECT ID INTO :ID FROM TEMP_TABLE_343; END;`
	//sqlText := `BEGIN SELECT 2 into :ID FROM DUAL; END;`
	_, err := conn.Exec(sqlText, &temp)
	if err != nil {
		return err
	}
	fmt.Println(temp)
	fmt.Println("finish query: ", time.Now().Sub(t))
	return nil
}
func insert(conn *sql.DB) error {
	t := time.Now()
	sqlText := `INSERT INTO TEMP_TABLE_343(ID, NAME, NAME2, VAL, VAL2, LDATE, LDATE2, LDATE3) VALUES(:ID, :NAME, :NAME2, :VAL, :VAL2, :LDATE, :LDATE2, :LDATE3)`
	temp := struct {
		Id    int             `db:"ID"`
		Name  string          `db:"NAME"`
		Name2 string          `db:"NAME2,nvarchar"`
		Val   *sql.NullString `db:"VAL,number"`
		Val2  int             `db:"VAL2,number"`
		Date  time.Time       `db:"LDATE"`
		Date2 time.Time       `db:"LDATE2,timestamp"`
		Date3 string          `db:"LDATE3,timestamptz"`
	}{1, "TEXT", "我想试的 执行存储过程返回的", &sql.NullString{String: "1.1", Valid: true}, 10, time.Now(), time.Now(),
		"2023-04-01T12:10:44+03:00"}
	_, err := conn.Exec(sqlText, temp)
	if err != nil {
		return err
	}
	fmt.Println("finish insert: ", time.Now().Sub(t))
	return nil
}

func bulkInsert(conn *sql.DB) error {
	t := time.Now()
	sqlText := `INSERT INTO TEMP_TABLE_343(ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)`
	type TABLE_343 struct {
		Id   int       `db:"ID"`
		Name string    `db:"NAME"`
		Val  string    `db:"VAL,number"`
		Date time.Time `db:"LDATE"`
	}
	length := 500
	tableArray := make([]TABLE_343, length)
	for x := 0; x < length; x++ {
		tableArray[x].Id = x + 1
		tableArray[x].Name = strings.Repeat("*", x+1)
		tableArray[x].Val = strconv.FormatFloat(float64(length)/float64(x+1), 'f', 2, 32)
		tableArray[x].Date = time.Now()
	}
	_, err := conn.Exec(sqlText, tableArray)
	if err != nil {
		return err
	}
	fmt.Println("finish insert: ", time.Now().Sub(t))
	return nil
}

func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
			return
		}
	}()
	err = createTable(conn)
	if err != nil {
		fmt.Println("can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()
	err = insert(conn)
	if err != nil {
		fmt.Println("can't insert: ", err)
		return
	}
	//err = bulkInsert(conn)
	//if err != nil {
	//	fmt.Println("can't bulk insert: ", err)
	//	return
	//}
	err = query(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	err = query2(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
}
