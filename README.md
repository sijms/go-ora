# go-ora
## Pure go oracle client
### note:
    - Use version 2 you will need to import github.com/sijms/go-ora/v2
    - V2 is more preferred for oracle servers 10.2 and above
    - I always update the driver fixing issues and add new features so
      always ensure that you get latest release
    - See examples for more help
### version 2.4.28: Binary Double And Float Fix
- Now you can read binary double and float without error issue#217
- You can avoid calling cgo function `user.Current()` if you define environmental variable $USER
### version 2.4.20: Query To Struct
- you can query to struct that contain basic types (int, float, string, datetime)
or any types that implement sql.Scanner interface
- see query to struct example for more information
### version 2.4.18: Add support for proxy user
if you need to connect with proxy user pass following connection
string
```golang
oracle://proxy_user:proxy_password@host:port/service?proxy client name=schema_owner
```
### version 2.4.8: JDBC connect string
* Add new function go_ora.BuildJDBC
```golang
    // program will extract server, ports and protocol and build
    // connection table
    connStr := `(DESCRIPTION=
    (ADDRESS_LIST=
    	(LOAD_BALANCE=OFF)
        (FAILOVER=ON)
    	(address=(PROTOCOL=tcps)(host=localhost)(PORT=2484))
    	(address=(protocol=tcp)(host=localhost)(port=1521))
    )
    (CONNECT_DATA=
    	(SERVICE_NAME=service)
        (SERVER=DEDICATED)
    )
    (SOURCE_ROUTE=yes)
    )`
    // use urlOption to set other options like:
    // TRACE FILE = for debug
    // note SSL automatically set from connStr (address=...
    // SSL Verify = need to cancel certifiate verification
    // wallet path
    databaseUrl := go_ora.BuildJDBC(user, password, connStr, urlOptions)
    conn, err := sql.Open("oracle", databaseUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
    err = conn.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
```
### version 2.4.5: Support BFile
* connect as sys and create directory object that refer to physical directory
* `grant read,write on directory 'dirName' to user`
* put text file in the directory with name = fileName
```golang
// create and open connection before use BFile
conn, err := go_ora.NewConnection(connStr)
// check for error
err = conn.Open()
// check for error
defer conn.Close()

// Create BFile object
file, err := go_ora.BFile(conn, dirName, fileName)
// check for error

// before use BFile it must be opened
err = file.Open()
// check for error
defer file.Close()

// does the file exist
exists, err := file.Exists()
// check for error

if exists {
    length, err := file.GetLength()
    // check for error
    
    // read all data
    data, err := file.Read()
    
    // read at position 2
    data, err = file.ReadFromPos(2)
    
    // read 5 bytes count start at position 2
    data, err = file.ReadBytesFromPos(2, 5)
```
* you can pass BFile object as input parameter or receive it from query or output parameters
for more detail see example bfile
### version 2.4.4: Support for unix socket IPC
you can use this option if server and client on same linux machine
by specify the following url option
```golang
urlOptions := map[string]string{
	// change the value according to your machine
	"unix socket": "/usr/tmp/.oracle/sEXTPROC1"
}
```
### version 2.4.3: Input Parameter CLOB and BLOB Accept Large Data Size
you can pass input CLOB and BLOB with any data size up to
[data type limit](https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/datatype-limits.html#GUID-963C79C9-9303-49FE-8F2D-C8AAF04D3095)
### version 2.4.1: Add support for connection time out + context read and write
you can determine connection overall lifetime through url options
```golang
// set connection time for 3 second
urlOptions := map[string]string {
    "CONNECTION TIMEOUT": "3"
}
databaseUrl := go_ora.BuildUrl(server, port, service, user, password, urlOptions)
```
see context example for more help about using context
### version 2.4.0: Add support for Arrays
* add support for oracle associative array as input and output parameter type
* add BulkInsert function which dramatically improve performance (> 10x) during insert
* add support for nullable type in DataSet.Scan function
* Bug fixes
* examples (bulk_insert and arrays) contain explanation of use of this 2 major features
```golang
// sqlText: sql text with parameters
// rowNum: number of rows to insert
// columns: each column contain array of driver.Value size of column should
//          equal to rowNum
func (conn *Connection) BulkInsert(sqlText string, rowNum int, columns ...[]driver.Value) (*QueryResult, error) 
```
### version 2.3.5: Add support for OS Auth (Windows) With Password Hash
now you can pass password hash of the user instead of real password
#### source of hash:
* windows registry
* create the hash by md4(unicode(password))
passing hash through url option as follow
```golang
urlOptions := map[string]string {
	"OS HASH": "yourpasswordhash"
	// or
	"OS PassHash": "yourpasswordhash"
	// or
	"OS Password Hash": "yourpasswordhash"
}
```
#### note:
you can use NTSAuthInterface
```golang
type YourCustomNTSManager struct {
	NTSAuthDefault
}
func (nts *NTSAuthHash) ProcessChallenge(chaMsgData []byte, user, password string) ([]byte, error) {
    // password = get (extract) password hash from Windows registry
	return ntlmssp.ProcessChallengeWithHash(chaMsgData, user, password)
}
// now you can pass empty user and password to the driver
```
### version 2.3.3: Add support for OS Auth (Windows)
you can see windows_os_auth example for more detail
* NTS packets are supplied from the following github package:
  [go-ntlmssp](https://github.com/Azure/go-ntlmssp)
* empty username or password will suppose OS Auth by default
* `AUTH TYPE: "OS"` optional
* `OS USER` optional if omit the client will use logon user
* `OS PASS` is obligatory to make OS Auth using NTS
* `DOMAIN` optional for windows domain
* `AUTH SERV: "NTS"` optional as NTS is automatically added if the client running on Windows machine
* `DBA PRIVILEGE: "SYSDBA"` optional if you need a SYSDBA access
```golang
urlOptions := map[string]string{
    // automatically set if you pass an empty oracle user or password
    // otherwise you need to set it
    "AUTH TYPE": "OS",
    // operating system user if empty the driver will use logon user name
    "OS USER": user,
    // operating system password needed for os logon
     "OS PASS": password,
    // Windows system domain name
    "DOMAIN": domain,
    // NTS is the required for Windows os authentication
    // when you run the program from Windows machine it will be added automatically
    // otherwise you need to specify it
    "AUTH SERV": "NTS",
    // uncomment this option for debugging
    "TRACE FILE": "trace.log",
}
databaseUrl := go_ora.BuildUrl(server, port, service, "", "", urlOptions)
```
#### note (Remote OS Auth):
* you can make OS Auth **on the same machine** (Windows Server) 
or **different machine** (Windows Server) and (Other Client) and in this situation you need to pass 
`AUTH SERV: "NTS"` as url parameter
#### note (advanced users):
* You can use custom NTS auth manager by implementing the following interface
```Golang
type NTSAuthInterface interface {
	NewNegotiateMessage(domain, machine string) ([]byte, error)
	ProcessChallenge(chaMsgData []byte, user, password string) ([]byte, error)
}
```
* set newNTS auth manager before open the connection
```golang
go_ora.SetNTSAuth(newNTSManager)
```
* advantage of custom manager: you may not need to provide OS Password. for example using
.NET or Windows API code as original driver
```cs
// CustomStream will take data from NegotiateStream and give it to the driver
// through NewNegotiateMessage
// Then take data form the driver (Challenge Message) to NegotiateStream
// And return back Authentication msg to the driver through ProcessChallenge
// as you see here CredentialCache.DefaultNetworkCredentials will take auth data
// (username and password) from logon user
new NegotiateStream(new YourCustomStream(), true).AuthenticateAsClient(CredentialCache.DefaultNetworkCredentials, "", ProtectionLevel.None, TokenImpersonationLevel.Identification);
```

### version 2.3.1: Fix issue related to use ipv6
now you can define url that contain ipv6
```go
url := go_ora.BuildUrl("::1", 1521, "service", "user", "password", nil)
url = "oracle://user:password@[::1]:1521/service"
```
### version 2.3.0: Add support for Nullable types
* support for nullable type in output parameters
* add more nullable type NullTimeStamp and NullNVarChar
### version 2.2.25: Add support for User Defined Type (UDT) as input and output parameter
* see example udt_pars for more help
### version 2.2.23: User Defined Type (UDT) as input parameters
* Add support for UDT as input parameter
* Add go_ora.Out struct with Size member to set output parameter size
### version 2.2.22: Lob for output parameters
* Add new types for output parameter which is `go_ora.Clob` and `go_ora.Blob`
used for receiving Clob and Blob from output parameters **_see clob example for 
more details_**
* Fix some issue related to reading output parameters
* Fix issue related to reading user defined type UDT
### version 2.2.19: improve lob reading with high prefetch rows value
* Now Prefetch rows value is **_automatically calculated (when left with its default value = 25)_** according to column
size 
* Reading lob is retarded until all record has been read this fix error happen
when you try to read lob with large PREFETCH_ROWS value
### version 2.2.9: add support for connect to multiple servers
define multiple server in 2 way
* in url string options
```golang
// using url options
databaseURL := "oracle://user:pass@server1/service?server=server2&server=server3"
/* now the driver will try connection as follow
1- server1
2- server2
3- server3
*/
```
* using BuildUrl function
```golang
urlOptions := map[string] string {
    "TRACE FILE": "trace.log",
    "SERVER": "server2, server3",
    "PREFETCH_ROWS": "500",
    //"SSL": "enable",
    //"SSL Verify": "false",
}
databaseURL := go_ora.BuildUrl(server1, 1521, "service", "user", "pass", urlOptions)
```

### version 2.2.8: add OracleError class 
OracleError carry error message from the server
### version 2.2.7: Add support for user defined types
* this feature is now tested against these oracle versions 10.2, 12.2, 19.3.
* RegisterType function need extra parameter owner (oracle user who create the type).
### version 2.2.6 (pre-release - experimental): Add support for user defined types
to use make the following (oracle 12c)
* define custom type in the oracle
```golang
create or replace TYPE TEST_TYPE1 IS OBJECT 
( 
    TEST_ID NUMBER(6, 0),
    TEST_NAME VARCHAR2(10)
)
```
* define struct in go with tag
```golang
type test1 struct {
    // note use int64 not int
    // all tagged fields should be exported 
    // tag name:field_name --> case insensitive
    Id int64       `oracle:"name:test_id"`
    Name string    `oracle:"name:test_name"`
}
```
* connect to database
```golang
databaseURL := go_ora.BuildUrl("localhost", 1521, "service", "user", "pass", nil)
conn, err := sql.Open("oracle", databaseURL)
// check for err
err = conn.Ping()
// check for err
defer func() {
    err := conn.Close()
    // check for err
}()
```
* register type
```golang
if drv, ok := conn.Driver().(*go_ora.OracleDriver); ok {
    err = drv.Conn.RegisterType("owner", "TEST_TYPE1", test1{})
    // check for err
}
```
* select and display data
```golang
rows, err := conn.Query("SELECT test_type1(10, 'test') from dual")
// check for err
var test test1
for rows.Next() {
    err = rows.Scan(&test)
    // check for err
    fmt.Println(test)
}
```
### version 2.2.5
* add function go_ora.BuildUrl to escape special characters 
### version 2.2.4
* add support for tcps. you can enable tcps through the following url options
* this [link](https://oracle-base.com/articles/misc/configure-tcpip-with-ssl-and-tls-for-database-connections) explain how to enable tcps in your server
```golang
wallet=wallet_dir // wallet should contain server and client certificates
SSL=true          // true or enabled
SSL Verify=false  // to bypass certificate verification
```
### version 2.1.23
* now support auto-login oracle wallet (non-local)
* **note**:
to use wallet you need to specify directory path for wallet the directory
  should contain cwallet.sso file "the file that will be used"
```golang
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

```golang
_, err := conn.Exec(inputSql, "7586")
```
##### &nbsp; &nbsp;2- nvarchar string:
```golang
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
#### note: See examples for using RefCursor with sql package
to use RefCursor follow these steps:
* create the connection object and open
* create NewStmt from connection
* pass RefCursorParam
* cast parameter to go_ora.RefCursor
* call cursor.Query()
* reterive records use for loop 
#### code:
```Golang
urlOptions := map[string] string {
	"trace file": "trace.log" ,
}
databaseURL := go_ora.BuildUrl(server, port, service, user, password, urlOptions)
conn, err := sql.Open("oracle", databaseURL)
// check error

err = conn.Ping()
// check error

defer conn.Close()

cmdText := `BEGIN    
    proc_1(:1); 
end;`
var cursor go_ora.RefCursor
_, err = conn.Exec(cmdText, sql.Out{Dest: &cursor})
//check errors

defer cursor.Close()
rows, err := cursor.Query()
// check for error

var (
    var_1 int64
    var_2 string
)
for rows.Next_() {
    err = rows.Scan(&var_1, &var_2)
    // check for error
	fmt.Println(var_1, var_2)
}
```


