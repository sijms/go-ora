# go-ora
# Pure go oracle client

## note:
###### the original oracle drivers are very complex and contain many features which are difficult to add them at one time
###### your feedbacks are very important for this project to proceed
```
    - To use version 2 you should import github.com/sijms/go-ora/v2
    - V2 is more preferred for oracle servers 10.2 and above
    - I always update the driver fixing issues and add new features so
      always ensure that you get latest release
    - See examples for more help
```

# Sponsors
<p>
  <a href="https://jb.gg/OpenSourceSupport" rel="noopener sponsored" target="_blank"><img height="128" width="128" src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png?_gl=1*txv9x8*_ga*MzY1MjAzNDI2LjE3MDAzMDc5NTg.*_ga_9J976DJZ68*MTcwNTE3NzM4My4zLjEuMTcwNTE3Nzg1MC40OS4wLjA.&_ga=2.97733338.412104364.1705177384-365203426.1700307958" alt="JetBrains" title="Essential tools for software developers and teams" loading="lazy" /></a>
</p>

# How To Use
## Connect to Database
* ### Simple Connection
  this connection require server name or IP, port, service name, username and password
  * using database/sql
  ```golang
  port := 1521
  connStr := go_ora.BuildUrl("server", port, "service_name", "username", "password", nil)
  conn, err := sql.Open("oracle", connStr)
  // check for error
  err = conn.Ping()
  // check for error
  ```
  * using package directly
  ```golang
  port := 1521
  connStr := go_ora.BuildUrl("server", port, "service_name", "username", "password", nil)
  conn, err := go_ora.NewConnection(connStr)
  // check for error
  err = conn.Open()
  // check for error
  ```
* ### Connect using SID
here we should pass urlOptions
note that service name is empty
```golang
port := 1521
urlOptions := map[string]string {
  "SID": "SID_VALUE",
}
connStr := go_ora.BuildUrl("server", port, "", "username", "password", urlOptions)
conn, err := sql.Open("oracle", connStr)
// check for error
```
* ### Connect using JDBC string
either pass a urlOption `connStr` with JDBC string
server, port and service name will be collected from JDBC string
```golang
urlOptions := map[string]string {
  "connStr": "JDBC string",
}
connStr := go_ora.BuildUrl("", 0, "", "username", "password", urlOptions)
conn, err := sql.Open("oracle", connStr)
// check for error
```
or use `go_ora.BuildJDBC`
```golang
urlOptions := map[string] string {
	// other options
}
connStr := go_ora.BuildJDBC("username", "password", "JDBC string", urlOptions)
conn, err := sql.Open("oracle", connStr)
// check for error
```
* ### SSL Connection
to use ssl connection you should pass required url options.
```golang
port := 2484
urlOptions := map[string] string {
	"ssl": "true", // or enable
	"ssl verify": "false", // stop ssl certificate verification
	"wallet": "path to folder that contains oracle wallet",
}
connStr := go_ora.BuildUrl("server", port, "service_name", "username", "password", urlOptions)
```

* ### OS Auth (for windows)
connect to oracle using OS user instead of oracle user
username and password parameters passed empty to `BuildUrl`
see [examples/windows_os_auth](https://github.com/sijms/go-ora/blob/master/examples/windows_os_auth/main.go) for more help
```golang
urlOptions := map[string]string {
    // optional as it will be automatically set 
	// if you pass an empty oracle user or password
    "AUTH TYPE": "OS",
    // operating system user if empty the driver will use logon user name
    "OS USER": user,
    // operating system password needed for os logon
    "OS PASS": password,
    // Windows system domain name
    "DOMAIN": domain,
	// optional as it will be automatically set 
	// when you define AUTH TYPE=OS in windows
    "AUTH SERV": "NTS",
}
port := 1521
connStr := go_ora.BuildUrl("server", port, "service_name", "", "", urlOptions)
```

* ### Proxy Auth
  if you need to connect with proxy user pass following connection string
  ```
  oracle://proxy_user:proxy_password@host:port/service?proxy client name=schema_owner
  ```

* ### Client Auth
  you should have server and client certificate store in wallets + working TCPS communication
  > create oracle user as follows:
  ```sql
  CREATE USER "SSLCLIENT" IDENTIFIED EXTERNALLY AS 'CN=ORCLCLIENT';
  ```
  > configure sqlnet.ora in the server to use client authentication

  ```sql
  SQLNET.AUTHENTICATION_SERVICES=(TCPS,NTS)
  SSL_CLIENT_AUTHENTICATION=TRUE
  ```
  > connect
  ```golang
  urlOptions := map[string]string {
  "TRACE FILE": "trace.log",
  "AUTH TYPE":  "TCPS",
  "SSL": "enable",
  "SSL VERIFY": "FALSE",
  "WALLET": "PATH TO WALLET"
  }
  connStr := go_ora.BuildUrl("server", 2484, "service", "", "", urlOptions)
  ```

* ### KERBEROS5 Auth
  > note that kerberos need an intact dns system and 3 separate machines to test it
* kerberos server you can use this link to install [on ubuntu](https://ubuntu.com/server/docs/service-kerberos)
* oracle server you can configure it from this [link](https://docs.oracle.com/cd/E11882_01/network.112/e40393/asokerb.htm#ASOAG9636)
* client which contain our gocode using package [gokrb5](https://github.com/jcmturner/gokrb5)
* Complete code found in [examples/kerberos](https://github.com/sijms/go-ora/blob/master/examples/kerberos/main.go)
  ```golang
  urlOptions := map[string]string{
      "AUTH TYPE":  "KERBEROS",
  }
  // note empty password
  connStr := go_ora.BuildUrl("server", 1521, "service", "krb_user", "", urlOptions)
  
  type KerberosAuth struct{}
  func (kerb KerberosAuth) Authenticate(server, service string) ([]byte, error) {
      // see implementation in examples/kerberos
  }
  advanced_nego.SetKerberosAuth(&KerberosAuth{})
  ```
before run the code you should run command `kinit user`
## Other Connection Options

<details>

* ### Define more servers to Connect
```golang
urlOptions := map[string]string {
	"server": "server2,server3",
}
connStr := go_ora.BuildUrl("server1", 1251, "service", "username", "password", urlOptions)
/* now the driver will try to connect as follows
1- server1
2- server2
3- server3
*/
```
* ### Client Encryption
this option give the client control weather to use encryption or not
```golang
urlOptions := map[string]string {
	// values can be "required", "accepted", "requested", and rejected"
	"encryption": "required",
}
```
* ### Client Data Integrity
this option give the client control weather to user data integrity or not
```golang
urlOptions := map[string]string {
    // values can be "required", "accepted", "requested", and rejected"
    "data integrity": "rejected",
}
```
* ### Using Unix Socket
you can use this option if server and client on same linux machine by specify the following url option
```golang
urlOptions := map[string]string{
	// change the value according to your machine 
	"unix socket": "/usr/tmp/.oracle/sEXTPROC1",
}
```
* ### Using Timeout
  * activate global timeout value (default=120 sec) to protect against block read/write if no timeout context specified
  * timeout value should be numeric string which represent number of seconds that should pass before operation finish or canceled by the driver
  * to disable this option pass 0 value start from v2.7.15
```golang
urlOptions := map[string]string {
	"TIMEOUT": "60",
}
```
* ### Using Proxy user
```golang
urlOptions := map[string]string {
	"proxy client name": "schema_owner",
}
connStr := go_ora.BuildUrl("server", 1521, "service", "proxy_user", "proxy_password", urlOptions)
```
* ### Define DBA Privilege
  * define dba privilege of the connection
  * default value is `NONE`
  * using user `sys` automatically set its value to `SYSDBA`
```golang
urlOptions := map[string]string {
	"dba privilege" : "sysdba", // other values "SYSOPER"
}
```
* ### Define Lob Fetching Mode
  * this option define how lob data will be loaded
  * default value is `pre` means lob data is send online with other values
  * other value is `post` means lob data will be loaded after finish loading other value through a separate network call
```golang
urlOptions := map[string]string {
	"lob fetch": "post",
}
```
* ### Define Client Charset
  * this option will allow controlling string encoding and decoding at client level
  * so using this option you can define a charset for the client that is different from the server
  * client charset will work in the following situation
    * encoding sql text
    * decoding varchar column
    * encoding and decoding varchar parameters
    * encoding and decoding CLOB
  * nvarchar, nclob and server messages are excluded from client charset
```golang
urlOptions := map[string]string {
    // you can use also 
    //"charset": "UTF8",
    "client charset": "UTF8",
}
```
* ### Define Client Territory and Language
  * this will control the language of the server messages
```golang
urlOptions := map[string]string {
    "language": "PORTUGUESE",
    "territory": "BRAZILIAN",
}
```
* ### Loging
this option used for logging driver work and network data for debugging purpose
```golang
urlOptions := map[string]string {
	"trace file": "trace.log",
}
```

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

* ### CID
use this option if you want to pass your own CID started from v2.7.15

> default CID
```golang
FulCid := "(CID=(PROGRAM=" + op.ProgramPath + ")(HOST=" + op.HostName + ")(USER=" + op.UserName + "))" 
```


</details>

## Execute SQL
* ### simple query
execute a query follows standards that defined in go package database/sql
you have `Query` used for query rows and `Exec` used for DML/DDL and PL/SQL
> Exec example
```golang
// note no semicolon at the end
_, err := conn.Exec(`CREATE TABLE TABLE1(
ID number(10),
NAME varchar2(50),
DAT DATE
)`)
// check for errors
```
> query example:
```golang
rows, err := conn.Query("SELECT ID, NAME, DAT FROM TABLE1")
// check for errors
defer rows.Close()
var (
	id int64
	name sql.NullString
	date sql.NullTime
)
for rows.Next() {
	err = rows.Scan(&id, &name, &date)
	// check for errors
}
```
> PL/SQL
```golang
// note semicolon at the end
_, err := conn.Exec("begin DBMS_LOCK.sleep(7); end;")
// check for errors
```
complete example found in [examples/crud](https://github.com/sijms/go-ora/blob/master/examples/crud/main.go)
* ### input parameters
  * #### parameters in oracle should start with `:` for example `:pr1`
passing input parameters as defined by database/sql package.
> parameter type
>   * int64 / float64 and their equivalent
>   * string
>   * time.Time
>   * any type that support Valuer interface
>   * NVarChar
>   * TimeStamp
>   * TimeStampTZ
>   * sql.Null* and go_ora.Null* for all the above
>   * Clob, NClob and Blob
* ### output parameters
  * passing parameter to Exec to return a value.
  * output parameter should be passed as pointers.
  * all output parameter should be passed inside `go_ora.Out` or `sql.Out` structures
  * output parameters like strings should be passed in `go_ora.Out` to define max size.
```golang
var(
	id int64
	name sql.NullString
	date sql.NullTime
)
_, err := conn.Exec("SELECT ID, NAME, DAT INTO :pr1, :pr2, :pr3 FROM TABLE1 WHERE ID=:pr4",
	sql.Out{Dest: &id},
	go_ora.Out{Dest: &name, Size: 100},
	go_ora.Out{Dest: &date},
	1)
```
* ### Lob Types
  * Blob, Clob and NClob
  * Clob use database charset and NClob use database ncharset for string encoding and decoding
  * complete code is found in [examples/clob](https://github.com/sijms/go-ora/blob/master/examples/clob/main.go)
> input parameters
```golang
var1 := go_ora.Clob{String: "large string value"}
var2 := go_ora.Blob{Data: []byte("large []byte value")}
_, err := conn.Exec("insert into tb values(:1, :2)", var1, var2)
```
> output parameters
```golang
var {
	var1 go_ora.NClob
	var2 go_ora.Blob
}
// set size according to size of your data
_, err := conn.Exec("BEGIN SELECT col1, col2 into :1, :2 FROM tb; END;",
	go_ora.Out{Dest: &var1, size: 100000},
	go_ora.Out{Dest: &var2, size: 300000})
```

* ### BFile (v2.7.23)
  * BFile require oracle directory object name + file name
  * create new BFile object by calling `CreateBFile`, `CreateBFile2` or `CreateNullBFile`
  * `CreateBFile` create BFile using `*sql.DB` while `CreateBFile2` take `*go_ora.Connection`
  * to create null BFile call `CreateNullBFile` or create object with `valid=false`
  * you can read BFile value from database either by Query or Exec (output par) using *BFile object
  * You can use object functions after creation or read from database, but you should test for null first
  * BFile functions:
    * Open: will open the file (file object returned from database unopened)
    * Exists: check if the file exists
    * GetLength: return file length
    * Read: read entire file content
    * ReadFromPos: start read from specific position till end of the file
    * ReadBytesFromPos: read portion of the file
    * Close: close file
    * IsOpen: check if the file opened
    * GetDirName: return directory object name
    * GetFileName: return file name
  * complete code for BFile found in [examples/bfile](https://github.com/sijms/go-ora/blob/master/examples/bfile/main.go)

* ### Named Parameters
  * to use named parameters just wrap all you parameters inside `sql.Named`
  * if one of the parameters doesn't have name driver will switch to positional mode
  * parameter named `:pr1` in sql should be passed as `sql.Named("pr1", 1)`
  * Named parameter is useful if you have one value passed in sql multiple times.
  * order is not important
  * complete code for named parameters found in [examples/named_pars](https://github.com/sijms/go-ora/blob/master/examples/named_pars/main.go)

* ### structures with tag
you can pass a structure parameter to sql in one of the following situation
- structure that implement Valuer interface
- oracle user defined type UDT
- struct with tag `db`
> data in `db` tag can be recognized by its position or as key=value
> ```golang
> type TEMP_TABLE struct {
>   // tag by position: db:"name,type,size,direction"
>   Id int  `db: "ID,number"`
> 
>   Name string `db:"type=varchar,name=NAME"`
> }  
> ```

> **struct with tag uses named parameters** so you should pass at least the name of the parameter to use this feature.
>
> Type is important in some situations
> for example if you have field with type time.Time and you want to pass timestamp
> to database so put `type=timestamp`

> type can be one of the following
> ```
> number      mapped to golang types integer, float, string, bool
> varchar     mapped to any golang types that can converted to string
> nvarchar    mapped to any golang types that can converted to string
> date        mapped to golang types time.Time and string
> timestamp   mapped to golang types time.Time and string
> timestamptz mapped to golang types time.Time and string
> raw         mapped to golang types []byte and string
> blob        mapped to golang types []byte and string
> clob        mapped to any golang types that can converted to string
> nclob       mapped to any golang types that can converted to string
> ```

> size and direction are required if the fields mapped to an output parameter

complete code can be found in [examples/struct_par](https://github.com/sijms/go-ora/blob/master/examples/struct_par/main.go)

* ### Arrays
> passing array as a parameter is useful in the following situations
> * Multiple insert/merge
> * Associative Array. You can find complete code in [examples/array](https://github.com/sijms/go-ora/blob/master/examples/arrays/main.go)
> * UDT array. You can find complete code in [examples/udt_array](https://github.com/sijms/go-ora/blob/master/examples/udt_array/main.go)

> Bulk insert/merge will be activated when you 
> passing all parameters as arrays of same size.
> 
> you can also pass an array of tagged structure to do same thing.
> complete code for bulk-insert/merge can be found in [examples/merge](https://github.com/sijms/go-ora/blob/master/examples/merge/main.go)

* ### UDT
* Created inside oracle using `create type`
* `UDT` mapped to golang struct type.
* To use UDT you should create struct with `udt` tag then call `go_ora.RegisterType(...)`
* complete code is found in [examples/UDT](https://github.com/sijms/go-ora/blob/master/examples/UDT/main.go)

* ### RefCursor
> as an output parameter
> ```golang
>   var cursor go_ora.RefCursor
>   _, err = conn.Exec(`BEGIN PROC1(:1, :2); END;`, 1, sql.Out{Dest: &cursor})
> ```
> you can use `go_ora.WrapRefCursor(...)` to convert `*RefCursor` into `*sql.Rows` started from v2.7.17

> complete code for RefCursor as output parameter found in [examples/refcursor](https://github.com/sijms/go-ora/blob/master/examples/refcursor/main.go)

> Map RefCursor to sql.Rows
```golang
// TEMP_FUNC_316 is sql function that return RefCursor
sqlText := `SELECT TEMP_FUNC_316(10) from dual`

// use Query and don't use QueryRow
rows, err := conn.Query(sqlText)
if err != nil {
	return err
}

// closing the parent rows will automatically close cursor
defer rows.Close()

for rows.Next() {
    var cursor sql.Rows
	err = rows.Scan(&cursor)
	if err != nil {
		return err
	}
	var (
        id   int64
        name string
        val  float64
        date time.Time
    )
	
    // reading operation should be inside rows.Next
    for cursor.Next() {
        err = cursor.Scan(&id, &name, &val, &date)
        if err != nil {
            return err
        }
        fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
    }
}
```
complete code for mapping refcursor to sql.Rows is found in [example/refcursor_to_rows](https://github.com/sijms/go-ora/blob/master/examples/refcursor_to_rows/main.go)

* ### Connect to multiple database
  * note that `sql.Open(...)` will use default driver so it will be suitable for one database projects.
  * to use multiple database you should create a separate driver for each one (don't use default driver) 
```golang
  // Get a driver-specific connector.   
  connector, err := go_ora.NewConnector(connStr)
  if err != nil {
    log.Fatal(err)
  }

  // Get a database handle.
  db = sql.OpenDB(connector)
```

* ### Use Custom String encode/decode
  * if your database charset is not supported you can create a custom object that implement IStringConverter interface and pass it to the driver as follows
```golang
  db, err := sql.Open("oracle", connStr)
  if err != nil {
	  // error handling
  }
  
  // call SetStringConverter before use db object
  // charset and nCharset are custom object that implement 
  // IStringConverter interface
  // if you pass nil for any of them then the driver will use 
  // default StringConverter
  go_ora.SetStringConverter(db, charset, nCharset)
  
  // rest of your code
```

* ### Session Parameters
  * you can update session parameter after connection as follow
  ```golang
  db, err := sql.Open("oracle", connStr)
  if err != nil {
    // error handling
  }
  // pass database, key, value
  err = go_ora.AddSessionParameter(db, "nls_language", "english")
  if err != nil {
    // error handling
  }
  ```
[//]: # (### Go and Oracle type mapping + special types)

[//]: # ()
[//]: # (### Supported DBMS features)
### releases
<details>

### version 2.8.6
* add support for nested user define type (UDT) array. 
* add testing file (TestIssue/nested_udt_array_test.go) with 2 level nesting
* fix issue related to date with time zone which occur with some oracle servers
* correct reading of oracle date with local time zone as output col/par.
* more testing is done for oracle date/time types now you can pass time.Time{}
as input/output for oracle date/time types except:
  * associative array which require strict oracle type
* add testing file (TestIssue/time_test.go)

### version 2.8.5
* add support for nested user defined types (UDT)
* add test file for nested UDT
* fix issue related to passing time with timezone as default input parameter for DATE, TIMESTAMP, TIMESTAMP with timezone
now user should define which type will be used according to oracle data type

| go type | oracle type |
|------ | -------|
|time.Time | DATE|
|go_ora.TimeStamp | TIMESTAMP|
|go_ora.TimeStampTZ | TIMESTAMP WITH TIME ZONE|

* fix issue related to returning clause

### version 2.8.4
* fix regression occur with custom types that support valuer and scanner interfaces
* fix regression occur with struct par that contain pointer (output)
* fix issue related to struct pars contain LOBs
* add DelSessionParam to remove session parameters when it is not needed
* add messageSize for Dequeue function in dbms.AQ 
* add tests for module, features and issues
### version 2.8.2
* now most of charsets are supported. still some of them are not fully tested.
* fix issue related to nested pointers
### version 2.8.0
* use buffered input during network read to avoid network data loss (occur with slow connections).
* fix charset mapping for charset id 846.
* add support for charset id 840
* re-code timeouts and connection break to get clean non-panic exit when context is timout and operation is cancelled
### version 2.7.25
* Add feature that enable the driver to read oracle 23c wallet
* introduce passing time.Time{} as input parameter for
DATE, TIMESTAMP and TIMESTAMP WITH TIMEZONE data types
note still output parameters need these types
* other bug fixes

### version 2.7.23
* Update BFile code to support null value and use *sql.DB
* Fix issue: BFile not working with lob pre-fetch mode
* Improve Bulk insert by removing un-necessary calls for `EncodeValue` so Bulk insert now can support Lob objects
* Fix issue related to refcursor

### version 2.7.20
* fix time not in timezone issue specially with oracle 19c
* add function to set value for session parameters that will be applied for subsequent connections
* fix issue #461
* bug fixes and improvements

### version 2.7.18
* Add 2 function `go_ora.NewDriver` and `go_ora.NewConnector`
* Add new function `go_ora.SetStringConverter` which accept custom converter for unsupported charset and nCharset
* `go_ora.SetStringConverter` accept `*sql.DB` as first parameter and IStringConveter interface object for charset and nCharset (you can pass nil to use default converter)
* Add support for charset ID 846

### version 2.7.17
* add `WrapRefCursor` which converts `*RefCursor` into `*sql.Rows`
* code:
```golang
// conn is *sql.DB
// cursor comming from output parameter
rows, err := go_ora.WrapRefCursor(context.Background(), conn, cursor)
```

### version 2.7.11
* add support for DBMS_OUTPUT
```golang
import (
  "database/sql"
  db_out "github.com/sijms/go-ora/dbms_output"
  _ "github.com/sijms/go-ora/v2"
  "os"
)

// create new output
// conn is *sql.DB
// bufferSize between 2000 and 32767
output, err := db_out.NewOutput(conn, 0x7FFF)

// close before exit
defer func() {
  err = output.Close()
}()

// put some line 
err = exec_simple_stmt(conn, `BEGIN
DBMS_OUTPUT.PUT_LINE('this is a test');
END;`)

// get data as string
line, err := output.GetOutput()

// or print it to io.StringWriter
err = output.Print(os.Stdout)
```
* complete code found in examples/dbms_output/main.go
### version 2.7.7:
* add support for CLOB/BLOB in UDT
* add support for UDT array as output parameters
* add function `go-ora.RegisterType(...)` so you can use it with database/sql package
* add `arrayTypeName` (input for array type can be empty) to `RegisterType(...)` to support UDT array
* `examples/udt_array` contain complete code that explain how to use udt array
* parameter encode/decode is recoded from the start
* fix uint64 truncation
* fix some other issue
### version 2.7.4:
* activate global timeout value to protect against block read/write
  if no timeout context specified
* default value for timeout is 120 second you can change by
  passing one of the following ["TIMEOUT", "CONNECT TIMEOUT", "CONNECTION TIMEOUT"]
* other feature/issues:
  * fix passing empty `[]byte{}` will produce error
  * fix passing empty array as a parameter will produce error
  * return first binding error when the driver return `ora-24381: error in DML array`
### version 2.7.3: Use database/sql fail over
* use database/sql fail over by returning driver.ErrBadConn
  when connection interrupted
* other features:
  * add support for RC4 encryption
### version 2.7.2: Use golang structure as an oracle (output) parameters
all rules used for input will be required for output plus:
* structure should be passed as a pointer
* tag direction is required to be output or inout. size is used with
  some types like strings
* each field will be translated to a parameter as follows
```
number      mapped to sql.NullFloat64
varchar     mapped to sql.NullString
nvarchar    mapped to sql.NullNVarchar
date        mapped to sql.NullTime
timestamp   mapped to NullTimeStamp
timestamptz mapped to NullTimeStampTZ
raw         mapped to []byte
clob        mapped to Clob
nclob       mapped to NClob
blob        mapped to Blob
```
all fields that support driver.Valuer interface will be passed as it is
* data assigned back to structure fields after exec finish when a null
  value read then field value will set to `reflect.Zero`
* `examples/struct_pars/main.go` contain full example for reading and
  writing struct pars
### version 2.7.1: Use golang structure as an oracle (input) parameters
* by define structure tag `db` now you can pass information to sql.Exec
* data in `db` tag can be recognized by its position or as key=value
```golang
type TEMP_TABLE struct {
	// tag by position: db:"name,type,size,direction"
	Id    int      `db:"ID,number"`
	// tag as key=value: db:"size=size,name=name,dir=directiontype=type"
	Name  string   `db:"type=varchar,name=NAME"`
}
```
* you should pass at least the name of the parameter to use this feature
* input parameters need only name and type. if you omit type driver will
  use field value directly as input parameter. type is used to make
  some flexibility
  example use time.Time field and pass type=timestamp in this
  case timestamp will be used instead of default value for time.Time which is date
* type can be one of the following:
```
number      mapped to golang types integer, float, string, bool
varchar     mapped to any golang types that can converted to string
nvarchar    mapped to any golang types that can converted to string
date        mapped to golang types time.Time and string
timestamp   mapped to golang types time.Time and string
timestamptz mapped to golang types time.Time and string
raw         mapped to golang types []byte and string
blob        mapped to golang types []byte and string
clob        mapped to any golang types that can converted to string
nclob       mapped to any golang types that can converted to string
```
* other features:
  * tag for user defined type UDT changed from `oracle` to `udt`
  * add 2 url options give the client control weather to use encryption, data integrity or not
  ```golang
  urlOptions := map[string]string {
    // values can be "required", "accepted", "requested", and rejected"
    "encryption": "required",
    "data integrity": "rejected",
  }
  ```
  * fix issue #350
### version 2.6.17: Implement Bulk(Insert/Merge) in ExecContext
* now you can make bulk (insert/merge) with sql driver Exec as follows:
  * declare sql text with Insert or Merge
  * pass all parameter as array
  * number of rows inserted will equal to the least array size
* Named parameter is also supported
* full code is present in examples/merge
### version 2.6.16: Map RefCursor to sql.Rows
* mapping RefCursor to sql.Rows will work with select/scan.
```golang
// TEMP_FUNC_316 is sql function that return RefCursor
sqlText := `SELECT TEMP_FUNC_316(10) from dual`

// use Query and don't use QueryRow
rows, err := conn.Query(sqlText)
if err != nil {
	return err
}

// closing the parent rows will automatically close cursor
defer rows.Close()

for rows.Next() {
    var cursor sql.Rows
	err = rows.Scan(&cursor)
	if err != nil {
		return err
	}
	var (
        id   int64
        name string
        val  float64
        date time.Time
    )
	
    // reading operation should be inside rows.Next
    for cursor.Next() {
        err = cursor.Scan(&id, &name, &val, &date)
        if err != nil {
            return err
        }
        fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
    }
}
```
* complete code is present in `examples/refcursor_to_rows/main.go`
### version 2.6.14: Add Support for Named Parameters
* to switch on named parameter mode simply pass all
  your parameter to `Query` or `Exec` as `sql.Named("name", Value)`
* if one of the parameter doesn't contain **name** the driver automatically switch to
  positional mode
* parameter name in sql will be for example `:pr1`
  and its value will be `sql.Named("pr1", 1)`
* when using named parameters the order of the parameters is not important as
  the driver will re-arrange the parameter according to declaration in
  sql text
* See `examples/named_pars/main.go` for example code
### version 2.6.12: Add Client Charset option
* this option will allow controlling string encoding and decoding at client level
* so using this option you can define a charset for the client that is different from the server
* client charset will work in the following situation
  * encoding sql text
  * decoding varchar column
  * encoding and decoding varchar parameters
  * encoding and decoding CLOB
* nvarchar, nclob and server messages are excluded from client charset
* code
```golang
urlOptions := map[string]string {
	// you can use also 
	//"charset": "UTF8",
	"client charset": "UTF8",
	"trace file": "trace.log",
}
connStr := go_ora.BuildUrl("server", 1521, "service", "", "", urlOptions)
```
### version 2.6.9: Re-Code Failover
* now failover start when receive the following error:
  * io.EOF
  * syscall.EPIPE
* failover added for the following
  * Query
  * Fetch
  * Exec
  * Ping
  * Commit
  * Rollback
  * RefCursor Query
* In all situation Failover will try to reconnect before returning error except in case of Query failover will reconnect + requery
### version 2.6.8: Fix return long data type with lob prefetch option:
* now you can return up to 0x3FFFFFFF of data from long coumn type
* examples/long insert 0x3FFF bytes of data into long column and query it again
* for large data size better use `lob fetch=post`
### version 2.6.5: Add New Url Options (Language and Territory)
* this will control the language of the server messages
```golang
urlOptions := map[string]string {
"language": "PORTUGUESE",
"territory": "BRAZILIAN",
}
url := go_ora.BuildUrl(server, port, service, user, password, urlOptions)
```
### version 2.6.4: Add Support for TimeStamp with timezone
* now you can use TimeStampTZ as input/output parameters to manage timestamp with timezone
* see code in examples/timestamp_tz
### version 2.6.2: Add Support for Lob Prefetch
* now you can control how you need to get lob data
  * **pre-fetch (default)** = lob data is sent from the server before send lob locator
  * **post-fetch** = lob data is sent from the server after send lob locator (need network call)
* you can do this using url options
```golang
urlOptions := map[string]string {
  "TRACE FILE": "trace.log",
  "LOB FETCH": "PRE", // other value "POST"
}
connStr := go_ora.BuildUrl("server", 1521, "service", "", "", urlOptions)
```
### version 2.5.33: Add Support for Client Authentication
* you should have server and client certificate store in wallets + working TCPS communication
* create oracle user as follows:
```sql
CREATE USER "SSLCLIENT" IDENTIFIED EXTERNALLY AS 'CN=ORCLCLIENT';
```
* configure sqlnet.ora in the server to use client authentication
```sql
SQLNET.AUTHENTICATION_SERVICES=(TCPS,NTS)
SSL_CLIENT_AUTHENTICATION=TRUE
```
* now connect
```golang
urlOptions := map[string]string {
  "TRACE FILE": "trace.log",
  "AUTH TYPE":  "TCPS",
  "SSL": "TRUE",
  "SSL VERIFY": "FALSE",
  "WALLET": "PATH TO WALLET"
}
connStr := go_ora.BuildUrl("server", 2484, "service", "", "", urlOptions)
```
### version 2.5.31: Add BulkCopy using DirectPath (experimental)
* it is a way to insert large amount of rows in table or view
* this feature use oracle [direct path](https://docs.oracle.com/database/121/ODPNT/featBulkCopy.htm#ODPNT212)
* this feature still not implemented for the following types:
  * LONG
  * CLOB
  * BLOB
* for more help about using this feature return to bulk_copy example
### version 2.5.19: Add Support for Kerberos5 Authentication
* note that kerberos need an intact dns system
* to test kerberos you need 3 machine
  * kerberos server you can use this link to install [i use ubuntu because easy steps](https://ubuntu.com/server/docs/service-kerberos)
  * oracle server you can configure it from this [link](https://docs.oracle.com/cd/E11882_01/network.112/e40393/asokerb.htm#ASOAG9636)
  * client which contain our gocode using package [gokrb5](https://github.com/jcmturner/gokrb5)
* there is an example code for kerberos, but you need to call `kinit user` before using the example
```golang
urlOptions := map[string]string{
    "TRACE FILE": "trace.log",
    "AUTH TYPE":  "KERBEROS",
}
// note empty password
connStr := go_ora.BuildUrl("server", 1521, "service", "krb_user", "", urlOptions)

type KerberosAuth struct{}
func (kerb KerberosAuth) Authenticate(server, service string) ([]byte, error) {
    // see implementation in examples/kerberos
}
advanced_nego.SetKerberosAuth(&KerberosAuth{})
```
### version 2.5.16: Add Support for cwallet.sso created with -auto_login_local option
* note that this type of oracle wallets only work on the machine where they were created
### version 2.5.14: Failover and wallet update
* Exec will return error after connection restore
* add new field _**WALLET PASSWORD**_ to read ewallet.p12 file
### version 2.5.13: Add Support For Failover (Experimental)
* to use failover pass it into connection string as follow
```golang
urlOptions := map[string]string{
	"FAILOVER": "5",
	"TRACE FILE": "trace.log",
}
databaseUrl := go_ora.BuildUrl(server, port, service, user, password, urlOptions)
```
* FAILOVER value is integer indicate how many times the driver will try to reconnect after lose connection default value = 0
* failover will activated when stmt receive io.EOF error during read or write
* FAILOVER work in 3 places:
  * Query when fail the driver will reconnect and re-query up to failover number.
  * Exec when fail the driver will reconnect up to failover times then return the error to avoid unintended re-execution.
  * Fetch when fail the driver will reconnect up to failover times then return the error (whatever failover success or fail)

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
</details>
