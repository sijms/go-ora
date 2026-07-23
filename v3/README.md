# go-ora v3

Pure Go Oracle database driver for [database/sql](https://pkg.go.dev/database/sql) - a major rewrite with new abstractions, new types, and improved performance.

## What's New in v3

### 1. Architecture Abstractions

v3 introduces a modular architecture with clear separation of concerns:

#### Network & Session (`network/`)
- Full abstraction of the Oracle TTC (Two-Task Common) protocol
- Packet-based communication layer (`ConnectPacket`, `DataPacket`, `MarkerPacket`, etc.)
- `MemorySession` for in-memory buffer operations (parameter encoding, AQ message marshaling)
- TLS/SSL negotiation and support
- Connection break/cancel with OOB (out-of-band) support
- Redirect handling

#### Connection (`Connection`)
- `driver.Connector` interface support via `OracleConnector`
- Pluggable `Dialer`, `TLSConfig`, `Kerberos`, and `Wallet` configuration
- Automatic reconnection on bad connections
- Session parameter management at driver level

#### Types (`types/`)
- Complete type system with encoding/decoding abstraction
- Each Oracle type has its own Go implementation
- Unified `SetValue` / `Value` / `Scan` / `CopyTo` interface
- LOB streaming support via `LobStreamer` interface

#### Parameter Encoding/Decoding (`parameter_coder/`)
- Pluggable `OracleParameterCoder` interface for encoding and decoding
- Three-way type mapping: Oracle type ID → Go `reflect.Type` → SQL type name
- Custom type registration via `AddParameterCoder`
- Each parameter type (string, number, date, vector, json, bool, LOB, etc.) has its own coder implementation

### 2. Fast Login, Token Login & Cookie Login

v3 implements Oracle's **Fast Authentication** mechanism (TTC version 24):

- **Fast Login**: When the server supports it (`FastAuthEnabled`), the driver uses a reduced negotiation round-trip, skipping full TCP and data type negotiation on subsequent connections.
- **Cookie Login**: Server negotiation results are cached in an in-memory `ConnectionCookie` store. On reconnection, cached data (charset, capabilities, version) is sent directly to the server, avoiding repeated negotiation.
- **Token Login**: Support for token-based authentication via `TokenFile` and `TokenPrivateKeyFile` configuration options.

```go
// Token-based connection
db, err := sql.Open("oracle", "oracle://user@host:1521/service?TOKEN_FILE=token.enc&TOKEN_PRIVATE_KEY_FILE=key.pem")
```

### 3. Client Version 24

v3 upgrades the TTC protocol version to **24**, enabling:

- Big CLR chunks (`ClrChunkSize = 0x7FFF`)
- Fast Session Affinity Protocol (FSAP) capability
- Extended data type negotiation
- Improved session property handling

### 4. New Oracle Type Support

#### VECTOR (Oracle 23ai)
Full support for the `VECTOR` data type with multiple element formats:

```go
import "github.com/sijms/go-ora/v3/types"

// Create vectors from Go slices
v1, _ := types.CreateVector([]uint8{10, 20, 30})       // INT8
v2, _ := types.CreateVector([]float32{-10.1, -20.2})    // FLOAT32
v3, _ := types.CreateVector([]float64{10.1, 20.2, 30.3}) // FLOAT64

// Scan from database
var vec types.Vector
row.Scan(&vec)

// Copy to typed slices
var data []float32
vec.CopyTo(&data)
```

Supported formats: `INT8` (uint8), `FLOAT32`, `FLOAT64` - both dense and sparse vectors.

#### JSON (Oracle 21c+)
Native JSON type support with pluggable coders:

```go
import "github.com/sijms/go-ora/v3/types"

var js types.Json
js.SetValue(`{"key": "value"}`)

// Copy to Go types
var s string
js.CopyTo(&s)

var m map[string]interface{}
js.CopyTo(&m)
```

Supports `oson` (Oracle Binary JSON) encoding via the `types/oson` package.

#### BOOLEAN (Oracle 23c+)
Native Oracle BOOLEAN type support:

```go
import "github.com/sijms/go-ora/v3/types"

input := types.Bool{}
input.SetValue(true)

// Use in PL/SQL calls
db.Exec("BEGIN my_proc(:1, :2); END;", input, go_ora.Out{Dest: &message})
```

### 5. Advanced Queuing (AQ)

Full Oracle Advanced Queuing support via the `aq` package:

```go
import "github.com/sijms/go-ora/v3/aq"

// Create a queue
queue, err := aq.CreateQueue(db, "my_queue", aq.RAW, "")

// Enqueue a message
msg, _ := queue.NewMessage([]byte("hello"))
queue.Enqueue(msg)

// Dequeue a message
deqOpts := &aq.DequeueOptions{
    Consumer: "my_consumer",
    Mode:     aq.DequeueModeBrowse,
    Wait:     5,  // seconds
}
msg, err = queue.Dequeue(deqOpts)
```

Supported message types: `RAW`, `JSON`, `UDT`, `XML`.

Features:
- Single and batch enqueue/dequeue
- Persistent and buffered delivery modes
- Visibility modes (on-commit, immediate)
- Dequeue modes (browse, locked, remove)
- Navigation modes (first, next, transactional)
- Message expiration and delay
- Correlation-based filtering

### 6. User-Defined Types (UDT)

Enhanced UDT registration with nested type support:

```go
// Register a type with its array counterpart
go_ora.RegisterType(db, "MY_OBJECT", "MY_ARRAY", MyStruct{})

// Register with explicit owner
go_ora.RegisterTypeWithOwner(db, "SCHEMA", "MY_OBJECT", "MY_ARRAY", MyStruct{})
```

Supports nested objects, collections (VARRAY, TABLE OF), and automatic struct mapping via `udt` tags.

### 7. Session Parameters

Runtime session parameter management without reconnecting:

```go
// Set session parameters
go_ora.AddSessionParam(db, "cursor_sharing", "force")
go_ora.AddSessionParam(db, "nls_language", "arabic")

// Parameters persist across connections in the pool
go_ora.DelSessionParam(db, "nls_language")
```

### 8. Custom Type Coders

Register custom encoders/decoders for new types:

```go
go_ora.AddParameterCoder(db, reflect.TypeOf(MyType{}), MY_ORACLE_TYPE_ID, &MyCoder{})
```

## Installation

```bash
go get github.com/sijms/go-ora/v3
```

## Quick Start

```go
import (
    "database/sql"
    _ "github.com/sijms/go-ora/v3"
)

func main() {
    db, err := sql.Open("oracle", "oracle://user:pass@host:1521/service")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        panic(err)
    }
}
```

## Connection String Options

| Option | Description |
|--------|-------------|
| `SERVER` | Database server hostname or IP |
| `PORT` | Database port (default 1521) |
| `SERVICE` | Oracle service name |
| `USER` | Database username |
| `PASSWORD` | Database password |
| `SSL` | Enable SSL/TLS connection |
| `SSL VERIFY` | Verify server certificate |
| `WALLET` | Path to Oracle wallet |
| `AUTH TYPE` | Authentication type (KERBEROS, etc.) |
| `FAST LOGIN` | Enable fast login optimization |
| `TOKEN FILE` | Path to authentication token file |
| `TOKEN PRIVATE KEY FILE` | Path to token private key |
| `TRACE DIR` | Directory for trace files |
| `CONNECT TIMEOUT` | Connection timeout duration |
| `LOB READ` | LOB read mode: `AUTO` or `IMPLICIT` (driver reads LOB automatically, default) or `NO` or `EXPLICIT` (manual LOB read by application) |

## Package Structure

```
go-ora/v3/
├── advanced_nego/     # Advanced authentication negotiation (NTS, Kerberos)
├── aq/                # Advanced Queuing (enqueue/dequeue)
├── configurations/    # Connection configuration parsing
├── converters/        # String and data converters
├── lazy_init/         # Lazy initialization utilities
├── network/           # TTC protocol, packets, session management
│   └── security/      # Security-related network utilities
├── parameter_coder/   # Parameter encoding/decoding implementations
├── trace/             # Trace and logging
├── types/             # Oracle type implementations
│   └── oson/          # Oracle Binary JSON (OSON) coder
├── utils/             # General utilities
├── connection.go      # Connection implementation
├── driver.go          # Driver registration and type coder maps
├── command.go         # Statement execution
├── parameter.go       # Parameter handling
├── lob.go             # LOB streaming
├── udt.go             # User-Defined Type support
├── transaction.go     # Transaction support
└── bulk_copy.go       # Bulk copy operations
```

## Migration from v2

1. Update import path: `github.com/sijms/go-ora/v2` → `github.com/sijms/go-ora/v3`
2. New types (`Vector`, `Json`, `Bool`) are in `github.com/sijms/go-ora/v3/types`
3. AQ API changed from `go_ora/dbms.NewAQ` to `aq.CreateQueue`
4. UDT registration uses `go_ora.RegisterType` with struct-based type mapping
5. Session parameters managed via `go_ora.AddSessionParam` / `go_ora.DelSessionParam`
6. Connection string format remains compatible
