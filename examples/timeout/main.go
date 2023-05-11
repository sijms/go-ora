package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

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
	t := time.Now()
	defer func() {
		fmt.Println("finish: ", time.Now().Sub(t))
	}()
	execCtx, execCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer execCancel()
	_, err = conn.ExecContext(execCtx, "begin DBMS_LOCK.sleep(7); end;")
	if err != nil {
		fmt.Println(err)
		return
	}
}
