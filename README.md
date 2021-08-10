# go-ora
## Pure go oracle client
### note:
    - Use version 2 you will need to import github.com/sijms/go-ora/v2
    - V2 is more preferred for oracle servers 10.2 and above
### version 2.1.23
* now support auto-login oracle wallet (non-local)
* **note**:
to use wallet you need to specify directory path for wallet the directory
  should contain cwallet.sso file "the file that will be used"
```bigquery
sqlQuery := "oracle://user@127.0.0.1:1522/service"
sqlQuery += "?TRACE FILE=trace.log"
sqlQuery += "&wallet=path_to_wallet_directory"
conn, err := sql.open("oracle", sqlQuery)
```
###### server:port/service ---> should be supplied when using wallet
###### user ---> is optional when omitted the reader will return first matched dsn
###### password ---> should be empty as it will be supplied from wallet
### version 2.1.22
* now support data packet integrity check using MD5, SHA1,
 SHA256, SHA384, SHA512
* key is exchanged between server and client using
  Diffie Hellman method
* **note**:
to enable data integrity check add the following line to sqlnet.ora of the server
```bigquery
# possible values ([accepted | rejected | requested | required])
SQLNET.CRYPTO_CHECKSUM_SERVER = required
# possible values ([MD5 | SHA1 | SHA256 | SHA384 | SHA512])
SQLNET.CRYPTO_CHECKSUM_TYPES_SERVER = SHA512
```
### version 2.1.21
* now support data packet encryption using AES. 
* key is exchanged between server and client using
  Diffie Hellman method
* note:
to enable AES encryption add the following line to sqlnet.ora 
  of the server
```bigquery
# possible values ([accepted | rejected | requested | required])
SQLNET.ENCRYPTION_SERVER = required
# possible values for AES (AES256 | AES192 | AES128)
SQLNET.ENCRYPTION_TYPES_SERVER = AES256
```

### version 2.1.20
* add new type **go_ora.NVarChar**
now you can pass string parameter in 2 way:
##### &nbsp; &nbsp; 1- varchar string:

```
_, err := conn.Exec(inputSql, "7586")
```
##### &nbsp; &nbsp;2- nvarchar string:
```
_, err := conn.Exec(inputSql, go_ora.NVarChar("7586"))
```

### version 2.1.19
* support more charsets (0x33D, 0x33E, 0x33F, 0x340, 0x352, 0x353, 0x354)

### version 2.0-beta
* update client version to 317
* update ttc version to: 9
* use 4 byte packet length instead of 2 bytes
* use advanced negotiation
* use big clear chunks
* use more verifier type in authentication object

# Usage:
## there are 2 way to use the client
### A. Using sql/database interface
#### 1- importing:
    import (
      "database/sql"
      _ "github.com/sijms/go-ora/v2"
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
    conn, err := go_ora.NewConnection("oracle://user:pass@server/service_name")
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
 
 ## Server's URL options
The complete syntax of connection url is: 

    oracle://user:pass@server/service_name[?OPTION1=VALUE1[&OPTIONn=VALUEn]...]

Check possible options in `connection_string.go` 

### TRACE FILE 
This option enables logging driver activity and packet content into a file.

    oracle://user:pass@server/service_name?TRACE FILE=trace.log

The log file is created into the current directory.


This produce this kind of log:
```
2020-11-22T07:51:42.8137: Open :(DESCRIPTION=(ADDRESS=(PROTOCOL=tcp)(HOST=192.168.10.10)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=xe)(CID=(PROGRAM=C:\Users\Me\bin\hello_ora.exe)(HOST=workstation)(USER=Me))))
2020-11-22T07:51:42.8147: Connect
2020-11-22T07:51:42.8256: 
Write packet:
00000000  00 3a 00 00 01 00 00 00  01 38 01 2c 0c 01 ff ff  |.:.......8.,....|
00000010  ff ff 4f 98 00 00 00 01  00 ea 00 3a 00 00 00 00  |..O........:....|
00000020  04 04 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  00 00 00 00 00 00 00 00  00 00                    |..........|

...

2020-11-22T07:51:42.8705: Query:
SELECT * FROM v$version
2020-11-22T07:51:42.8705: 
Write packet:
00000000  00 55 00 00 06 00 00 00  00 00 03 5e 00 02 81 21  |.U.........^...!|
00000010  00 01 01 17 01 01 0d 00  00 00 01 19 01 01 00 00  |................|
00000020  00 00 00 00 00 00 00 00  00 01 00 00 00 00 00 53  |...............S|
00000030  45 4c 45 43 54 20 2a 20  46 52 4f 4d 20 76 24 76  |ELECT * FROM v$v|
00000040  65 72 73 69 6f 6e 01 01  00 00 00 00 00 00 01 01  |ersion..........|
00000050  00 00 00 00 00                                    |.....|
2020-11-22T07:51:42.9094: 
Read packet:
00000000  01 a7 00 00 06 00 00 00  00 00 10 17 3f d5 ec 21  |............?..!|
00000010  d5 37 e0 67 cc 0f eb 03  cc c5 d1 d8 78 78 0b 15  |.7.g........xx..|
00000020  0c 21 20 01 50 01 01 51  01 80 00 00 01 50 00 00  |.! .P..Q.....P..|
00000030  00 00 02 03 69 01 01 50  01 06 01 06 06 42 41 4e  |....i..P.....BAN|
00000040  4e 45 52 00 00 00 00 01  07 07 78 78 0b 16 07 34  |NER.......xx...4|
00000050  2b 00 02 1f e8 01 0a 01  0a 00 06 22 01 01 00 01  |+.........."....|
00000060  19 00 00 00 07 49 4f 72  61 63 6c 65 20 44 61 74  |.....IOracle Dat|
00000070  61 62 61 73 65 20 31 31  67 20 45 78 70 72 65 73  |abase 11g Expres|
00000080  73 20 45 64 69 74 69 6f  6e 20 52 65 6c 65 61 73  |s Edition Releas|
00000090  65 20 31 31 2e 32 2e 30  2e 32 2e 30 20 2d 20 36  |e 11.2.0.2.0 - 6|
000000a0  34 62 69 74 20 50 72 6f  64 75 63 74 69 6f 6e 07  |4bit Production.|
000000b0  26 50 4c 2f 53 51 4c 20  52 65 6c 65 61 73 65 20  |&PL/SQL Release |
000000c0  31 31 2e 32 2e 30 2e 32  2e 30 20 2d 20 50 72 6f  |11.2.0.2.0 - Pro|
000000d0  64 75 63 74 69 6f 6e 15  01 01 01 07 1a 43 4f 52  |duction......COR|
000000e0  45 09 31 31 2e 32 2e 30  2e 32 2e 30 09 50 72 6f  |E.11.2.0.2.0.Pro|
000000f0  64 75 63 74 69 6f 6e 15  01 01 01 07 2e 54 4e 53  |duction......TNS|
00000100  20 66 6f 72 20 4c 69 6e  75 78 3a 20 56 65 72 73  | for Linux: Vers|
00000110  69 6f 6e 20 31 31 2e 32  2e 30 2e 32 2e 30 20 2d  |ion 11.2.0.2.0 -|
00000120  20 50 72 6f 64 75 63 74  69 6f 6e 15 01 01 01 07  | Production.....|
00000130  26 4e 4c 53 52 54 4c 20  56 65 72 73 69 6f 6e 20  |&NLSRTL Version |
00000140  31 31 2e 32 2e 30 2e 32  2e 30 20 2d 20 50 72 6f  |11.2.0.2.0 - Pro|
00000150  64 75 63 74 69 6f 6e 08  01 06 03 14 97 b7 00 01  |duction.........|
00000160  01 01 02 00 00 00 00 00  04 01 05 01 07 01 05 02  |................|
00000170  05 7b 00 00 01 01 00 03  00 01 20 00 00 00 00 00  |.{........ .....|
00000180  00 00 00 00 00 00 00 01  01 00 00 00 00 19 4f 52  |..............OR|
00000190  41 2d 30 31 34 30 33 3a  20 6e 6f 20 64 61 74 61  |A-01403: no data|
000001a0  20 66 6f 75 6e 64 0a                              | found.|
2020-11-22T07:51:42.9104: Summary: RetCode:1403, Error Message:"ORA-01403: no data found\n"
2020-11-22T07:51:42.9104: Row 0
2020-11-22T07:51:42.9104:   BANNER              : Oracle Database 11g Express Edition Release 11.2.0.2.0 - 64bit Production
2020-11-22T07:51:42.9104: Row 1
2020-11-22T07:51:42.9104:   BANNER              : PL/SQL Release 11.2.0.2.0 - Production
2020-11-22T07:51:42.9104: Row 2
2020-11-22T07:51:42.9104:   BANNER              : CORE	11.2.0.2.0	Production
2020-11-22T07:51:42.9104: Row 3
2020-11-22T07:51:42.9104:   BANNER              : TNS for Linux: Version 11.2.0.2.0 - Production
2020-11-22T07:51:42.9104: Row 4
2020-11-22T07:51:42.9104:   BANNER              : NLSRTL Version 11.2.0.2.0 - Production
2020-11-22T07:51:42.9114: 
```
### PREFETCH_ROWS
Default value is 25 increase this value to higher level will significantly
speed up the query
## RefCursor
to use RefCursor follow these steps:
* create the connection object and open
* create NewStmt from connection
* pass RefCursorParam
* cast parameter to go_ora.RefCursor
* call cursor.Query()
* reterive records use for loop 
#### code:
```buildoutcfg
conn, err := go_ora.NewConnection(url)
// check error

err = conn.Open()
// check error

defer conn.Close()

cmdText := `BEGIN    
    proc_1(:1); 
end;`
stmt := go_ora.NewStmt(cmdText, conn)
stmt.AddRefCursorParam("1")
defer stmt.Close()

_, err = stmt.Exec(nil)
//check errors

if cursor, ok := stmt.Pars[0].Value.(go_ora.RefCursor); ok {
    defer cursor.Close()
    rows, err := cursor.Query()
    // check for error
    
    var (
        var_1 int64
        var_2 string
    )
    values := make([]driver.Value, 2)
    for {
        err = rows.Next(values)
        // check for error and if == io.EOF break
        
        if var_1, ok = values[0].(int64); !ok {
            // error
        }
        if var_2, ok = values[1].(string); !ok {
            // error
        }
        fmt.Println(var_1, var_2)
    }
}
```


