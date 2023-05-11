package main

import (
	"context"
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"math/rand"
	"os"
	"sync"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TESTSHORT 
   (	"ID" NUMBER(20,0) NOT NULL ENABLE, 
	"TM" VARCHAR2(30), 
	"SN" VARCHAR2(25) NOT NULL ENABLE, 
	"CUS" VARCHAR2(20) NOT NULL ENABLE, 
	"AID" VARCHAR2(20), 
	"TR" VARCHAR2(8) NOT NULL ENABLE, 
	"PID" VARCHAR2(20), 
	"CODE" VARCHAR2(20) NOT NULL ENABLE, 
	"TTNO" VARCHAR2(20), 
	"UPDATETIME" TIMESTAMP (6) DEFAULT systimestamp, 
	"OUT_TEST" VARCHAR2(200), 
	 PRIMARY KEY ("ID")
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
	_, err := conn.Exec("drop table TESTSHORT purge")
	if err != nil {
		return err
	}
	fmt.Println("finish drop table: ", time.Now().Sub(t))
	return nil
}

func truncateTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("truncate table TESTSHORT")
	if err != nil {
		return err
	}
	fmt.Println("finish truncate table: ", time.Now().Sub(t))
	return nil
}
func merge(db *sql.DB) error {
	t := time.Now()
	sqlText := `MERGE INTO TESTSHORT t1 USING(select :ID ID from dual) tmp ON (tmp.ID=t1.ID) 
    WHEN MATCHED THEN UPDATE SET TM=:TM,SN=:SN,CUS=:CUS, AID=:AID,TR=:TR,PID=:PID,CODE=:CODE,TTNO=:TTNO,UPDATETIME=:UPDATETIME,OUT_TEST=:OUT_TEST WHERE t1.ID=:ID AND t1.UPDATETIME<=:UPDATETIME 
    WHEN NOT MATCHED THEN INSERT (ID,TM,SN,CUS,AID,TR,PID,CODE,TTNO,UPDATETIME,OUT_TEST) VALUES (:ID,:TM,:SN,:CUS,:AID,:TR,:PID,:CODE,:TTNO,:UPDATETIME,:OUT_TEST)`
	length := 500
	var bindall []interface{}
	testjson := []byte(`{"ag":"a3077220"}`)
	rand.Seed(time.Now().UnixMilli())
	for _, colname := range []string{"ID", "TM", "SN", "CUS", "AID", "TR", "PID", "CODE", "TTNO", "UPDATETIME", "OUT_TEST"} {
		var bind []interface{}
		for x := 0; x < length; x++ {
			switch colname {
			case "ID":
				bind = append(bind, rand.Intn(500)+1)
			case "UPDATETIME":
				bind = append(bind, go_ora.TimeStamp(time.Now()))
			case "TM":
				bind = append(bind, string(testjson))
			default:
				bind = append(bind, "text")
			}

		}
		bindall = append(bindall, sql.Named(colname, bind))
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	conn, err := db.Conn(ctxTimeout)
	if err != nil {
		return err
	}
	defer func(conn *sql.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("can't close conn: ", err)
		}
		fmt.Println("connection closed")
	}(conn)
	//tx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})
	//if err != nil {
	//	return err
	//}
	//_, err = tx.ExecContext(ctxTimeout, sqlText, bindall...)
	_, err = conn.ExecContext(ctxTimeout, sqlText, bindall...)
	if err != nil {
		//if err2 := tx.Rollback(); err2 != nil {
		//	fmt.Println("error in rollback: ", err2)
		//}
		return err
	}
	fmt.Println("finish merge: ", time.Now().Sub(t))
	//return tx.Commit()
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
	//conn.SetMaxIdleConns(10)
	//conn.SetMaxOpenConns(10)
	//err = createTable(conn)
	//if err != nil {
	//	fmt.Println("can't create table: ", err)
	//	return
	//}
	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("can't drop table: ", err)
	//	}
	//}()
	err = truncateTable(conn)
	if err != nil {
		fmt.Println("can't truncate table: ", err)
		return
	}
	st := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(seq int, wgin *sync.WaitGroup) {
			fmt.Printf("%d: start to merge\n", seq)
			defer wgin.Done()
			err = merge(conn)
			if err != nil {
				fmt.Println("can't merge: ", err)
				return
			}
		}(i, &wg)
	}
	wg.Wait()
	et := time.Now()
	fmt.Println("finished total:", et.Sub(st))
	// time.Sleep(time.Second * 10)
}
