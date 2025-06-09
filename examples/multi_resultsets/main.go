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
	err := execCmd(db, `create table TTB_573_1(
    ID number(10),
    NAME varchar2(100)
    )`, `create table TTB_573_2(
    ID number(10),
    CUST_ID number(10),
    NAME varchar2(100)
    )`)
	if err != nil {
		return err
	}
	fmt.Println("Finish create tables: ", time.Since(t))
	return nil
}

func dropTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, "drop table TTB_573_1 purge", "drop table TTB_573_2 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop tables: ", time.Since(t))
	return nil
}

func insert(db *sql.DB) error {
	t := time.Now()
	_, err := db.Exec("INSERT INTO TTB_573_1(ID, NAME) VALUES (:1, :2)", []int{1, 2}, []string{"Tom", "Julia"})
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO TTB_573_2(ID, CUST_ID, NAME) VALUES(:1, :2, :3)",
		[]int{1000, 2000}, []int{1, 2}, []string{"BOOKS", "FURNITURE"})
	if err != nil {
		return err
	}
	fmt.Println("Finish insert: ", time.Since(t))
	return nil
}

func query(db *sql.DB) error {
	t := time.Now()
	rows, err := db.Query(`declare
    cust_cur sys_refcursor;
    sales_cur sys_refcursor;
begin
    open cust_cur for SELECT ID, NAME FROM TTB_573_1;
    dbms_sql.return_result(cust_cur);

    open sales_cur for SELECT ID, CUST_ID, NAME FROM TTB_573_2;
    dbms_sql.return_result(sales_cur);
end;`)
	if err != nil {
		return nil
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	type Customer struct {
		id   int
		name string
	}
	type Product struct {
		id     int
		custID int
		name   string
	}
	customer := Customer{}
	product := Product{}

	// get first result set
	fmt.Println("Customers: ")
	for rows.Next() {
		err = rows.Scan(&customer.id, &customer.name)
		if err != nil {
			return err
		}
		fmt.Println(customer)
	}
	fmt.Println("Products: ")
	if rows.NextResultSet() {
		for rows.Next() {
			err = rows.Scan(&product.id, &product.custID, &product.name)
			if err != nil {
				return err
			}
			fmt.Println(product)
		}
	}
	fmt.Println("Finish query: ", time.Since(t))
	return nil
}
func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't connect: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()
	err = db.Ping()
	if err != nil {
		fmt.Println("can't ping database: ", err)
		return
	}
	err = createTable(db)
	if err != nil {
		fmt.Println("can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()
	err = insert(db)
	if err != nil {
		fmt.Println("can't insert table: ", err)
		return
	}
	err = query(db)
	if err != nil {
		fmt.Println("can't query table: ", err)
		return
	}

}
