package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func execStmt(conn *sql.DB, sqlText, helpString string) error {
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println(helpString, time.Now().Sub(t))
	return nil
}
func insertData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("insert into GOORA_TEMP_DEPT values (10,'ACCOUNTING','NEW YORK')")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_DEPT values (20,'RESEARCH','DALLAS')")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_DEPT values (30,'SALES','CHICAGO')")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_DEPT values (40,'OPERATIONS','BOSTON')")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7369,'SMITH','CLERK',7902,to_date('17-12-1980','dd-mm-yyyy'),800,null,20)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7499,'ALLEN','SALESMAN',7698,to_date('20-2-1981','dd-mm-yyyy'),1600,300,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7521,'WARD','SALESMAN',7698,to_date('22-2-1981','dd-mm-yyyy'),1250,500,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7566,'JONES','MANAGER',7839,to_date('2-4-1981','dd-mm-yyyy'),2975,null,20)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7654,'MARTIN','SALESMAN',7698,to_date('28-9-1981','dd-mm-yyyy'),1250,1400,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7698,'BLAKE','MANAGER',7839,to_date('1-5-1981','dd-mm-yyyy'),2850,null,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7782,'CLARK','MANAGER',7839,to_date('9-6-1981','dd-mm-yyyy'),2450,null,10)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7788,'SCOTT','ANALYST',7566,to_date('13-JUL-87','dd-mm-rr')-85,3000,null,20)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7839,'KING','PRESIDENT',null,to_date('17-11-1981','dd-mm-yyyy'),5000,null,10)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7844,'TURNER','SALESMAN',7698,to_date('8-9-1981','dd-mm-yyyy'),1500,0,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7876,'ADAMS','CLERK',7788,to_date('13-JUL-87', 'dd-mm-rr')-51,1100,null,20)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7900,'JAMES','CLERK',7698,to_date('3-12-1981','dd-mm-yyyy'),950,null,30)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7902,'FORD','ANALYST',7566,to_date('3-12-1981','dd-mm-yyyy'),3000,null,20)")
	if err != nil {
		return err
	}
	_, err = conn.Exec("insert into GOORA_TEMP_EMP values (7934,'MILLER','CLERK',7782,to_date('23-1-1982','dd-mm-yyyy'),1300,null,10)")
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}

func usage() {
	fmt.Println()
	fmt.Println("using_with")
	fmt.Println("  a code that use with clause.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  using_with -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  using_with -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}

func main() {
	var (
		server string
	)

	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", server)
	if err != nil {
		fmt.Println("Can't open the driver: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close driver: ", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection: ", err)
		return
	}

	sqlText := `create table GOORA_TEMP_DEPT (
  deptno number(2) constraint pk_dept primary key,
  dname varchar2(14),
  loc varchar2(13)
) `
	err = execStmt(conn, sqlText, "Finish create table GOORA_TEMP_DEPT: ")
	if err != nil {
		fmt.Println("Can't create table GOORA_TEMP_DEPT", err)
		return
	}
	defer func() {
		err = execStmt(conn, "drop table GOORA_TEMP_DEPT purge", "Finish drop table GOORA_TEMP_DEPT: ")
		if err != nil {
			fmt.Println("Can't drop table GOORA_TEMP_DEPT", err)
		}
	}()
	sqlText = `create table GOORA_TEMP_EMP (
  empno number(4) constraint pk_emp primary key,
  ename varchar2(10),
  job varchar2(9),
  mgr number(4),
  hiredate date,
  sal number(7,2),
  comm number(7,2),
  deptno number(2) constraint fk_deptno references GOORA_TEMP_DEPT
)`
	err = execStmt(conn, sqlText, "Finish create table GOORA_TEMP_EMP: ")
	if err != nil {
		fmt.Println("Can't create table GOORA_TEMP_EMP", err)
		return
	}
	defer func() {
		err = execStmt(conn, "drop table GOORA_TEMP_EMP purge", "Finish drop table GOORA_TEMP_EMP: ")
		if err != nil {
			fmt.Println("Can't drop table GOORA_TEMP_EMP", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data ", err)
		return
	}

	sqlText = `with dept_count as (
  select deptno, count(*) as dept_count
  from   GOORA_TEMP_EMP
  group by deptno)
select e.ename as employee_name,
       dc.dept_count as emp_dept_count
from   GOORA_TEMP_EMP e,
       dept_count dc
where  e.deptno = dc.deptno`
	t := time.Now()
	rows, err := conn.Query(sqlText)
	if err != nil {
		fmt.Println("Can't query rows", err)
		return
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close dataset: ", err)
		}
	}()
	var (
		name  string
		count int64
	)
	for rows.Next() {
		err = rows.Scan(&name, &count)
		if err != nil {
			fmt.Println("Can't scan rows", err)
		}
		fmt.Println("Name: ", name, "\tCount: ", count)
	}
	if rows.Err() != nil {
		fmt.Println("Can't scan rows", err)
		return
	}
	fmt.Println("Finish run with clause: ", time.Now().Sub(t))

}
