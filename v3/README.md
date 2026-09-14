# go-ora v3

A pure Go Oracle database driver for [`database/sql`](https://pkg.go.dev/database/sql).

v3 is a major rewrite focused on extensibility, modern Oracle features, and clean architecture. It uses Go interfaces throughout to make adding new types, nested objects, and array types straightforward.

## Features

- **Pure Go** -- no C dependencies, no Oracle client installation required
- **`database/sql` compatible** -- works with any `database/sql` pool, retry logic, and health checking
- **Oracle 23ai types** -- VECTOR, BOOLEAN, JSON
- **Advanced Queuing** -- native AQ support (RAW, JSON, UDT, XML)
- **Fast Authentication** -- reduced round-trips with cookie-based caching and token login
- **TTC v24** -- latest protocol version with FSAP capability
- **User-Defined Types** -- nested objects, collections, and struct mapping
- **Extensible type system** -- plug in custom encoders/decoders for any Oracle type
- **TLS/SSL** -- full support with wallet, mutual TLS, and Kerberos

## Installation

```bash
go get github.com/sijms/go-ora/v3
```

**Requires Go 1.24+**

## Quick Start

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/sijms/go-ora/v3"
)

func main() {
    db, err := sql.Open("oracle", "oracle://user:pass@host:1521/service")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to Oracle")
}
```

## Connection Options

```text
oracle://USER:PASSWORD@HOST:PORT/SERVICE_NAME?PARAM1=value&PARAM2=value
```

| Parameter | Description | Default |
|-----------|-------------|---------|
| `USER` | Database username | -- |
| `PASSWORD` | Database password | -- |
| `SERVER` | Hostname or IP | -- |
| `PORT` | Listener port | `1521` |
| `SERVICE` | Oracle service name | -- |
| `SSL` | Enable TLS | `false` |
| `SSL VERIFY` | Verify server certificate | `true` |
| `WALLET` | Path to Oracle wallet | -- |
| `AUTH TYPE` | Authentication type (`KERBEROS`) | -- |
| `FAST LOGIN` | Enable fast login optimization | `false` |
| `TOKEN FILE` | Path to authentication token file | -- |
| `TOKEN PRIVATE KEY FILE` | Path to token private key | -- |
| `LOB READ` | `AUTO`/`IMPLICIT` (default) or `NO`/`EXPLICIT` | `AUTO` |
| `TRACE DIR` | Directory for trace files | -- |
| `CONNECT TIMEOUT` | Connection timeout | -- |

## New Types

### VECTOR (Oracle 23ai)

```go
import "github.com/sijms/go-ora/v3/types"

// Create from Go slices
v1, _ := types.CreateVector([]uint8{10, 20, 30})          // INT8
v2, _ := types.CreateVector([]float32{-10.1, -20.2})       // FLOAT32
v3, _ := types.CreateVector([]float64{10.1, 20.2, 30.3})   // FLOAT64

// Scan from database
var vec types.Vector
row.Scan(&vec)

// Copy to typed slices
var data []float32
vec.CopyTo(&data)
```

Formats: `INT8` (`[]uint8`), `FLOAT32`, `FLOAT64` -- dense and sparse.

### JSON (Oracle 21c+)

```go
var js types.Json
js.SetValue(`{"key": "value"}`)

var s string
js.CopyTo(&s)

var m map[string]interface{}
js.CopyTo(&m)
```

Oracle Binary JSON (OSON) encoding is supported via the `types/oson` package.

### BOOLEAN (Oracle 23c+)

```go
input := types.Bool{}
input.SetValue(true)

db.Exec("BEGIN my_proc(:1, :2); END;", input, go_ora.Out{Dest: &message})
```

## Advanced Queuing

```go
import "github.com/sijms/go-ora/v3/aq"

// Create a queue
queue, err := aq.CreateQueue(db, "my_queue", aq.RAW, "")

// Enqueue
msg, _ := queue.NewMessage([]byte("hello"))
queue.Enqueue(msg)

// Dequeue
msg, err = queue.Dequeue(&aq.DequeueOptions{
    Consumer: "my_consumer",
    Mode:     aq.DequeueModeBrowse,
    Wait:     5,
})
```

Message types: `RAW`, `JSON`, `UDT`, `XML`

Features: batch enqueue/dequeue, persistent and buffered delivery, visibility modes, navigation modes, message expiration, correlation filtering.

## User-Defined Types

```go
// Register with array counterpart
go_ora.RegisterType(db, "MY_OBJECT", "MY_ARRAY", MyStruct{})

// Register with explicit owner
go_ora.RegisterTypeWithOwner(db, "SCHEMA", "MY_OBJECT", "MY_ARRAY", MyStruct{})
```

Supports nested objects, collections (VARRAY, TABLE OF), and struct mapping via `udt` tags.

## Session Parameters

```go
go_ora.AddSessionParam(db, "cursor_sharing", "force")
go_ora.AddSessionParam(db, "nls_language", "arabic")
go_ora.DelSessionParam(db, "nls_language")
```

Parameters persist across connections in the pool.

## Custom Type Coders

```go
go_ora.AddParameterCoder(db, reflect.TypeOf(MyType{}), MY_ORACLE_TYPE_ID, &MyCoder{})
```

Implement the `OracleParameterCoder` interface to add support for any Oracle type.

## Architecture

```
go-ora/v3/
├── advanced_nego/     # NTS, Kerberos authentication
├── aq/                # Advanced Queuing
├── configurations/    # Connection string parsing
├── converters/        # String and data converters
├── network/           # TTC protocol, packets, session
│   └── security/      # Network security utilities
├── parameter_coder/   # Type encoding/decoding
├── trace/             # Logging and tracing
├── types/             # Oracle type implementations
│   └── oson/          # Oracle Binary JSON (OSON)
├── utils/             # General utilities
├── connection.go      # Connection implementation
├── driver.go          # Driver registration
├── command.go         # Statement execution
├── parameter.go       # Parameter handling
├── lob.go             # LOB streaming
├── udt.go             # User-Defined Types
├── transaction.go     # Transactions
└── bulk_copy.go       # Bulk copy
```

### Design Principles

- **Interface-based type system** -- each Oracle type has its own `OracleParameterCoder` implementation, making it trivial to add new types
- **Three-way type mapping** -- Oracle type ID ↔ Go `reflect.Type` ↔ SQL type name
- **Pluggable coders** -- register custom encoders/decoders without modifying driver internals
- **MemorySession** -- lightweight in-memory buffer for parameter encoding and AQ message marshaling
- **No driver-level failover** -- relies on `database/sql` connection pooling and retry logic

## Migration from v2

| v2 | v3 |
|----|-----|
| `github.com/sijms/go-ora/v2` | `github.com/sijms/go-ora/v3` |
| Types in driver package | Types in `github.com/sijms/go-ora/v3/types` |
| `go_ora/dbms.NewAQ` | `aq.CreateQueue` |
| Manual UDT setup | `go_ora.RegisterType` with struct tags |
| URL options for session params | `go_ora.AddSessionParam` / `DelSessionParam` |

Connection string format is compatible.
