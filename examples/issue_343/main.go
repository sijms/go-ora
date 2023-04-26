package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
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

func merge(conn *sql.DB) error {
	t := time.Now()
	sqlText := `MERGE INTO testshort t1 USING(select :ID ID from dual) tmp ON (tmp.ID=t1.ID) 
    WHEN MATCHED THEN UPDATE SET TM=:TM,SN=:SN,CUS=:CUS, AID=:AID,TR=:TR,PID=:PID,CODE=:CODE,TTNO=:TTNO,UPDATETIME=:UPDATETIME WHERE t1.ID=:ID AND t1.UPDATETIME<=:UPDATETIME 
    WHEN NOT MATCHED THEN INSERT (ID,TM,SN,CUS,AID,TR,PID,CODE,TTNO,UPDATETIME) VALUES (:ID,:TM,:SN,:CUS,:AID,:TR,:PID,:CODE,:TTNO,:UPDATETIME)`
	length := 500
	id := make([]int, length)
	tm := make([]sql.NullString, length)
	sn := make([]string, length)
	cus := make([]string, length)
	aid := make([]sql.NullString, length)
	tr := make([]string, length)
	pid := make([]string, length)
	code := make([]string, length)
	ttno := make([]sql.NullString, length)
	updateTime := make([]go_ora.TimeStamp, length)
	for x := 0; x < length; x++ {
		id[x] = x + 1
		if x > 0 && x%10 == 0 {
			tm[x] = sql.NullString{String: "", Valid: false}
			aid[x] = sql.NullString{String: "", Valid: false}
			ttno[x] = sql.NullString{String: "", Valid: false}
		} else {
			tm[x] = sql.NullString{String: "tm text", Valid: true}
			aid[x] = sql.NullString{String: "aid text", Valid: true}
			ttno[x] = sql.NullString{String: "ttno text", Valid: false}
		}
		sn[x] = "sn text"
		cus[x] = "cus text"
		tr[x] = "tr text"
		pid[x] = "pid text"
		code[x] = "code text"
		updateTime[x] = go_ora.TimeStamp(time.Now())
	}
	_, err := conn.Exec(sqlText, sql.Named("ID", id),
		sql.Named("TM", tm),
		sql.Named("SN", sn),
		sql.Named("CUS", cus),
		sql.Named("AID", aid),
		sql.Named("TR", tr),
		sql.Named("PID", pid),
		sql.Named("CODE", code),
		sql.Named("TTNO", ttno),
		sql.Named("UPDATETIME", updateTime))

	if err != nil {
		return err
	}
	fmt.Println("finish merge: ", time.Now().Sub(t))
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
	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("can't drop table: ", err)
	//	}
	//}()
	err = merge(conn)
	if err != nil {
		fmt.Println("can't merge: ", err)
		return
	}
}
