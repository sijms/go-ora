# go-ora

Pure Go Oracle database driver for [`database/sql`](https://pkg.go.dev/database/sql).

[![JetBrains logo.](https://resources.jetbrains.com/storage/products/company/brand/logos/jetbrains.svg)](https://jb.gg/OpenSource)

## Install

```bash
go get github.com/sijms/go-ora/v2
```

Requires Oracle server 10.2+. See [v3](v3/) for the latest version with Oracle 23ai support.

## Quick Start

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/sijms/go-ora/v2"
)

func main() {
    connStr := "oracle://user:pass@server:1521/service"
    db, err := sql.Open("oracle", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    var version string
    err = db.QueryRow("SELECT * FROM v$version").Scan(&version)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(version)
}
```

## Connection Options

Build connection strings with `go_ora.BuildUrl` or `go_ora.BuildJDBC`:

```go
// Basic
connStr := go_ora.BuildUrl("server", 1521, "service", "user", "pass", nil)

// With options
connStr := go_ora.BuildUrl("server", 1521, "service", "user", "pass", map[string]string{
    "SSL":         "true",
    "SSL VERIFY":  "false",
    "WALLET":      "/path/to/wallet",
    "TIMEOUT":     "60",
    "TRACE FILE":  "trace.log",
})

// JDBC string
connStr := go_ora.BuildJDBC("user", "pass", "JDBC_STRING", nil)
```

| Option | Description | Default |
|--------|-------------|---------|
| `TIMEOUT` | Socket read/write timeout (seconds, 0 = disabled) | `15` |
| `CONNECTION TIMEOUT` | Connection timeout (seconds, 0 = disabled) | `60` |
| `FAILOVER` | Reconnect attempts on connection loss | `0` |
| `SSL` | Enable TLS/SSL | `false` |
| `SSL VERIFY` | Verify server certificate | `true` |
| `WALLET` | Path to Oracle wallet directory | |
| `AUTH TYPE` | `OS`, `KERBEROS`, or `TCPS` | |
| `DBA PRIVILEGE` | `SYSDBA` or `SYSOPER` | `NONE` |
| `LOB FETCH` | `pre`/`inline` (default) or `post`/`stream` | `pre` |
| `client charset` | Override client-side character set | |
| `language` / `territory` | Server message language | |
| `TRACE FILE` | Enable packet logging | |
| `proxy client name` | Proxy user schema | |

## Usage

### Query

```go
rows, err := db.Query("SELECT id, name, created_at FROM users")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var (
        id    int64
        name  string
        createdAt sql.NullTime
    )
    if err := rows.Scan(&id, &name, &createdAt); err != nil {
        log.Fatal(err)
    }
    fmt.Println(id, name, createdAt)
}
```

### Exec (DDL/DML)

```go
_, err := db.Exec(`CREATE TABLE users (
    id NUMBER(10),
    name VARCHAR2(50),
    created_at DATE
)`)
```

### PL/SQL

```go
_, err := db.Exec("BEGIN DBMS_LOCK.sleep(5); END;")
```

### Parameters

```go
// Input
_, err := db.Exec("INSERT INTO users (id, name) VALUES (:1, :2)", 1, "Alice")

// Named parameters
_, err := db.Exec("INSERT INTO users (id, name) VALUES (:id, :name)", sql.Named("id", 1), sql.Named("name", "Alice"))

// Output
var name string
_, err = db.Exec("BEGIN SELECT name INTO :1 FROM users WHERE id=:2; END;",
    go_ora.Out{Dest: &name, Size: 100}, 1)
```

### LOB Types

```go
// Insert
clob := go_ora.Clob{String: "large text value"}
blob := go_ora.Blob{Data: []byte("binary data")}
_, err := db.Exec("INSERT INTO docs (text_col, blob_col) VALUES (:1, :2)", clob, blob)

// Output
var text go_ora.NClob
_, err = db.Exec("BEGIN SELECT text_col INTO :1 FROM docs WHERE id=:2; END;",
    go_ora.Out{Dest: &text, Size: 100000}, 1)
```

### Struct Parameters

```go
type User struct {
    Id   int64  `db:"ID,number"`
    Name string `db:"type=varchar,name=NAME"`
}

// Input
_, err := db.Exec("INSERT INTO users (id, name) VALUES (:ID, :NAME)", User{Id: 1, Name: "Alice"})

// Output (requires direction + size)
_, err = db.Exec("BEGIN SELECT id, name INTO :ID, :NAME FROM users WHERE id=:1; END;",
    go_ora.Out{Dest: &User{Id: 1}, Size: 100}, 1)
```

### User-Defined Types (UDT)

```sql
CREATE OR REPLACE TYPE address_type AS OBJECT (
    street VARCHAR2(100),
    city   VARCHAR2(50)
);
```

```go
type Address struct {
    Street string `udt:"STREET"`
    City   string `udt:"CITY"`
}

// Register before use
if drv, ok := db.Driver().(*go_ora.OracleDriver); ok {
    err := drv.Conn.RegisterType("SCHEMA", "ADDRESS_TYPE", Address{})
}

// Use in queries
var addr Address
rows, err := db.Query("SELECT address_type('123 Main', 'NYC') FROM dual")
rows.Scan(&addr)
```

### Vector (Oracle 23ai)

```go
v, err := go_ora.NewVector([]float32{0.1, 0.2, 0.3})
_, err = db.Exec("INSERT INTO embeddings (id, vec) VALUES (:1, :2)", 1, v)

// Query
var vec go_ora.Vector
rows.Scan(&vec)
data := vec.Data.([]float32) // cast to slice
```

### Ref Cursor

```go
rows, err := db.Query("BEGIN :1 := my_func(:2); END;",
    go_ora.Out{Dest: &cursor}, 1)
```

### Multiple Databases

```go
connector, err := go_ora.NewConnector(connStr)
db := sql.OpenDB(connector)
```

### Custom Configuration

```go
config, err := go_ora.ParseConfig(connStr)
config.RegisterDial(func(ctx context.Context, network, address string) (net.Conn, error) {
    // custom dialer
})
go_ora.RegisterConnConfig(config)
db, err := sql.Open("oracle", "")
```

### Session Parameters

```go
err := go_ora.AddSessionParameter(db, "nls_language", "english")
err = go_ora.DelSessionParameter(db, "nls_language")
```

### Custom String Converter

```go
go_ora.SetStringConverter(db, charset, nCharset) // implement IStringConverter
```

## Authentication

| Method | Options |
|--------|---------|
| Password | Default (user/pass in connection string) |
| OS Auth (Windows) | `AUTH TYPE=OS`, `OS USER`, `OS PASS`, `DOMAIN` |
| Kerberos5 | `AUTH TYPE=KERBEROS` + [gokrb5](https://github.com/jcmturner/gokrb5) |
| Client Cert | `AUTH TYPE=TCPS` + `SSL=TRUE` + `WALLET` |
| Wallet | `wallet=/path/to/cwallet.sso` |
| Proxy | `proxy client name=schema_owner` |

## Encryption

Data encryption (AES) and integrity checking (SHA) are negotiated via Diffie-Hellman key exchange. Configure on the server side in `sqlnet.ora`:

```ini
SQLNET.ENCRYPTION_SERVER = required
SQLNET.ENCRYPTION_TYPES_SERVER = AES256
SQLNET.CRYPTO_CHECKSUM_SERVER = required
SQLNET.CRYPTO_CHECKSUM_TYPES_SERVER = SHA512
```

Control client-side behavior:

```go
urlOptions := map[string]string{
    "encryption":     "required",  // accepted/rejected/required
    "data integrity": "required",
}
```

## Features

- **Pure Go** -- no C dependencies or Oracle client required
- **`database/sql` compatible** -- works with standard Go DB pools
- **Oracle types** -- DATE, TIMESTAMP, TIMESTAMP TZ, RAW, LONG, CLOB/BLOB/NCLOB, BFILE
- **Advanced types** -- Vector (23ai), JSON (23c), BOOLEAN (23c), UDT, nested UDT
- **Bulk operations** -- insert/merge/update/delete with array parameters
- **TLS/SSL** -- wallet, mutual TLS, Kerberos5, client certificates
- **Encryption** -- AES data encryption, SHA data integrity
- **Failover** -- automatic reconnection on connection loss
- **DBMS OUTPUT** -- `dbms_output.NewOutput(conn, bufferSize)`
- **Multiple result sets** -- query returning multiple cursors
- **BulkCopy** -- Direct Path insert for large datasets

## Release History

<details>
<summary>Click to expand full changelog</summary>

### v2.8.24
* Per-session Kerberos auth settings
* Bulk update & delete support
* Bug fixes

### v2.8.19
* Long input support (up to 1GB for LONG/LOB columns)
* `RegisterDial` for custom connection dialers

### v2.9.0
* Multiple result sets
* Connect timeout option
* RefCursor fixes

### v2.8.12
* `ParseConfig` / `RegisterConnConfig` API
* LONG and JSON fixes

### v2.8.8
* Async connection break
* LOB prefetch returns BLOB as longRAW (up to 1GB)

### v2.8.7
* Regular type array support
* Nested null object/array support
* `go_ora.Object` wrapper for UDT parameters

### v2.8.6
* Nested UDT array support
* Date/time fixes across Oracle versions

### v2.8.5
* Nested UDT support
* TIME input parameter type resolution

### v2.8.4
* Valuer/Scanner regressions fixed
* Struct LOB parameters
* `DelSessionParam`

### v2.8.2 / v2.8.0
* Most character sets supported
* Buffered network read for slow connections

### v2.7.25
* Oracle 23c wallet support
* `time.Time{}` as input for DATE/TIMESTAMP

### v2.7.23
* BFile null value support
* Bulk insert with LOB objects

### v2.7.20
* Timezone fixes for Oracle 19c
* Session parameter functions

### v2.7.18
* `NewDriver` / `NewConnector`
* `SetStringConverter` for unsupported charsets

### v2.7.17
* `WrapRefCursor` converts RefCursor to `*sql.Rows`

### v2.7.11
* DBMS_OUTPUT support

### v2.7.7
* CLOB/BLOB in UDT
* UDT array output parameters

### v2.7.4
* Global timeout protection
* Empty array/byte fixes

### v2.7.3
* `database/sql` failover via `driver.ErrBadConn`
* RC4 encryption

### v2.7.1 / v2.7.2
* Struct tags for input/output parameters
* Encryption/data integrity URL options

### v2.6.17
* Bulk insert/merge in ExecContext

### v2.6.16
* RefCursor mapped to `*sql.Rows`

### v2.6.14
* Named parameters (`sql.Named`)

### v2.6.12
* Client charset option

### v2.6.9
* Failover re-coded (Query reconnect + re-query)

### v2.6.2
* Lob prefetch control (`pre`/`post`)

### v2.5.33
* Client certificate authentication

### v2.5.31
* BulkCopy using DirectPath (experimental)

### v2.5.19
* Kerberos5 authentication

### v2.5.13
* Failover support (reconnect on io.EOF)

### v2.4.28
* Binary Double/Float fix

### v2.4.20
* Query to struct

### v2.4.18
* Proxy user support

### v2.4.8
* JDBC connect string support

### v2.4.5
* BFile support

### v2.4.4
* Unix socket IPC

### v2.4.3
* Large CLOB/BLOB input

### v2.4.1
* Connection timeout + context

### v2.4.0
* Array parameters (associative array)
* BulkInsert

### v2.3.5
* OS Auth with password hash

### v2.3.3
* OS Auth (Windows NTS)

### v2.3.1
* IPv6 support

### v2.3.0
* Nullable types

### v2.2.25 - v2.2.6
* UDT support progression (input, output, arrays, nested)

### v2.2.5
* `BuildUrl` for special characters

### v2.2.4
* TCPS support

### v2.1.23
* Auto-login wallet

### v2.1.22
* Data integrity check (SHA)

### v2.1.21
* AES encryption

### v2.1.20
* NVarChar type

### v2.0-beta
* TTC v9, 4-byte packet length, advanced negotiation

</details>

## Examples

See [examples/](https://github.com/sijms/go-ora/tree/master/examples) for complete code samples covering CRUD, LOB, UDT, arrays, BulkCopy, RefCursor, and more.
