package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func execCmd(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			if len(stmts) > 1 {
				return fmt.Errorf("error: %v in execuation of stmt: %s", err, stmt)
			} else {
				return err
			}
		}
	}
	return nil
}

func createTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, `create table TTB_562(id number not null, data xmltype)`)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table: ", time.Since(t))
	return nil
}

func dropTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, `drop table TTB_562 purge`)
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Since(t))
	return nil
}

func insert(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec(`insert into TTB_562(id, data) values(:1, xmltype(:2))`, 1,
		"<xml><test>this is a test</test></xml>")
	if err != nil {
		//if ora_err, ok := err.(*go_ora_network.OracleError); ok {
		//	fmt.Println(ora_err.ErrPos())
		//	fmt.Println(ora_err.ErrMsg)
		//	fmt.Println(ora_err.ErrCode)
		//}
		return err
	}
	fmt.Println("Finish insert table: ", time.Since(t))
	return nil
}

func xmlGen(db *sql.DB) error {
	t := time.Now()
	var data string
	err := db.QueryRow("SELECT sys_xmlgen('this is a test') DATA FROM DUAL").Scan(&data)
	if err != nil {
		return err
	}
	fmt.Println("Finish xmlGen: ", time.Since(t))
	fmt.Println(data)
	return nil
}
func queryStringValue(db *sql.DB) error {
	t := time.Now()
	var data string
	err := db.QueryRow("SELECT T.DATA.extract('/xml/test/text()').getStringVal() DATA FROM TTB_562 T").Scan(&data)
	if err != nil {
		return err
	}
	fmt.Println("Finish queryStringValue: ", time.Since(t))
	fmt.Println(data)
	return nil
}

func extractValue(db *sql.DB) error {
	t := time.Now()
	var data string
	err := db.QueryRow("SELECT extractValue(DATA, '/xml/test') as DATA FROM TTB_562").Scan(&data)
	if err != nil {
		return err
	}
	fmt.Println("Finish extractValue: ", time.Since(t))
	fmt.Println(data)
	return nil
}

func updateXML(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec("UPDATE TTB_562 SET DATA = updateXML(data, :1, :2)",
		"/xml/test/text()", "this is a test 2")
	if err != nil {
		return err
	}
	fmt.Println("Finish updateXML: ", time.Since(t))
	return nil
}

func updateNodeXML(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec("UPDATE TTB_562 SET DATA = updateXML(data, :1, xmltype.createXML(:2))",
		"/xml/test", "<test>this is a test 3</test>")
	if err != nil {
		return err
	}
	fmt.Println("Finish updateNodeXML: ", time.Since(t))
	return nil
}
func query(db *sql.DB) error {
	t := time.Now()
	var id int
	var data string
	err := db.QueryRow("SELECT ID, DATA FROM TTB_562").Scan(&id, &data)
	if err != nil {
		return err
	}
	fmt.Println("Finish query: ", time.Since(t))
	fmt.Println(id, data)
	return nil
}
func main() {
	db, err := sql.Open("oracle", os.Getenv("LOCAL_DSN"))
	if err != nil {
		fmt.Println("can't open database:", err)
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Println("can't close database:", err)
		}
	}()
	err = createTable(db)
	if err != nil {
		fmt.Println("can't create table:", err)
		return
	}
	defer func() {
		err := dropTable(db)
		if err != nil {
			fmt.Println("can't drop table:", err)
		}
	}()
	err = insert(db)
	if err != nil {
		fmt.Println("can't insert table:", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query table:", err)
		return
	}
	err = queryStringValue(db)
	if err != nil {
		fmt.Println("can't query value:", err)
		return
	}
	err = updateXML(db)
	if err != nil {
		fmt.Println("can't update xml:", err)
		return
	}
	err = extractValue(db)
	if err != nil {
		fmt.Println("can't extract value:", err)
		return
	}
	err = updateNodeXML(db)
	if err != nil {
		fmt.Println("can't update node xml:", err)
		return
	}
	err = extractValue(db)
	if err != nil {
		fmt.Println("can't extract value:", err)
		return
	}
	err = xmlGen(db)
	if err != nil {
		fmt.Println("can't xmlGen:", err)
		return
	}
}
