# go-ora
## Pure go oracle client
important notes:
 * the client is tested against oracle 10G, 11G and 12G and working properly
 * supported parameter types is integer, double, strings and time.Time
 * named parameter not supported to define parameter just put ':' + parameter_name in sql statment
 * integeration with sql/database is done using simple form

# Usage:
## there are 2 way to use the client
### A. Using sql/database interface
#### 1- importing:
    import (
      "database/sql"
      "fmt"
      _ "go-ora"
      "time"
    )
      
#### 2- create the connection
    conn, err := sql.Open("oracle", "oracle://user:pass@server/service_name")
    // check for error
    defer conn.Close()
   
#### 3- create statment
    stmt, err := conn.Prepare("SELECT col_1, col_2, col_3 FROM table WHERE col_1 = :1 or col_2 = :2")
    // check for error
    defer stmt.CLose()
   
#### 4- query
    // suppose we have 2 params one time.Time and other is double
    rows, err := stmt.Query(time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC), 9.2)
    // check for error
    defer rows.Close()
   
#### 5- extract data using next
     for rows.Next() {
        // define vars
        err = rows.Scan(/*vars here */)
        // check for error
     }
    
#### 6- use exec instead of query for update and insert stmt
    // i make change in parameter no 4 to explain that you can use string in parameter name instead of numbers
    stmt, err := conn.Prepare("UPDATE table SET col_1=:1, col_2=:2 WHERE col_3 = :3 or col_4 = :col_4_par")
    // check for error
    defer stmt.Close()
    result, err := stmt.Exec(/*pars value*/)
    // check for error
    fmt.Println(result.RowsAffected())

#### 7- using transaction:
    // after step 2 "Create Connection"
    tx, err := conn.Begin()
    // check for error
    stmt, err := tx.Prepare("sql text")
    // check for error
    // continue as above
    tx.Commit()
    // or
    tx.Rollback()
    // note: any stmt created from conn will not be committed or rolled back
     
### B. direct use of the package
  the benefit here is that you can use pl/sql and output parameters
#### 1- import go_ora "go-ora"
#### 2- create connection
    conn, err := go_ora.NewConnection("oracle://user:pass@dbname/service_name")
    // check for error
    err = conn.Open()
    // check for error
    defer conn.Close()
#### 2- create stmt
    stmt := go_ora.NewStmt("sql or pl/sql text", conn)
    defer stmt.Close()
#### 3- add parameters
    stmt.AddParam("name", value, size, go_ora.Input /* or go_ora.Output*/)
    // note that size is need when you define string output parameters
#### 4- exec or query as above and pass nil for parameters
#### 5- after that you can read the output parameters using Pars variable of stmt structure
 
