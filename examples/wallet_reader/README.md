# Wallet Reader Example

This example demonstrates how to load Oracle Wallets using the `io.Reader` interface instead of file paths.

## Why Use io.Reader?

Using `io.Reader` provides flexibility to load wallets from various sources:

- **Embedded Files**: Bundle wallets in your binary using `go:embed`
- **Remote Sources**: Load from HTTP, S3, Google Cloud Storage, etc.
- **Secrets Managers**: Retrieve from HashiCorp Vault, AWS Secrets Manager, etc.
- **In-Memory**: Use wallet data stored in memory without touching disk
- **Encrypted Storage**: Decrypt wallets on-the-fly from encrypted sources

## Concurrency Safe

This implementation uses `OracleConnector` with `sql.OpenDB()`, which means:

- Each connector maintains its own wallet configuration
- Multiple goroutines can use different wallets without interference
- No global state is modified

## Basic Usage

```go
import (
    "database/sql"
    "os"

    go_ora "github.com/sijms/go-ora/v2"
)

func main() {
    // Create a new connector - isolated from other connections
    connector := go_ora.NewConnector(
        "oracle://user:pass@server:1521/service?SSL=enable",
    )

    // Open wallet file (could be any io.Reader)
    walletFile, _ := os.Open("/path/to/cwallet.sso")
    defer walletFile.Close()

    // Load wallet from reader - data is read immediately
    connector.WithWallet(walletFile)

    // Use sql.OpenDB with the connector
    db := sql.OpenDB(connector)
    defer db.Close()
}
```

## Running the Example

```bash
go run main.go "oracle://user:pass@server:1521/service?SSL=enable" "/path/to/cwallet.sso"
```

## Embedded Wallet Example

See `embedded_example.go` for how to embed a wallet file directly in your binary:

```go
import "embed"

//go:embed wallets/cwallet.sso
var walletFS embed.FS

func ConnectWithEmbeddedWallet() (*sql.DB, error) {
    connector := go_ora.NewConnector("oracle://user:pass@server/service?SSL=enable")

    walletFile, _ := walletFS.Open("wallets/cwallet.sso")
    defer walletFile.Close()

    connector.WithWallet(walletFile)

    return sql.OpenDB(connector), nil
}
```

## Multiple Wallets in Concurrent Goroutines

Each connector is independent, so you can safely use different wallets:

```go
func connectWithWallet(connStr string, walletPath string) (*sql.DB, error) {
    connector := go_ora.NewConnector(connStr)

    walletFile, err := os.Open(walletPath)
    if err != nil {
        return nil, err
    }
    defer walletFile.Close()

    if err := connector.WithWallet(walletFile); err != nil {
        return nil, err
    }

    return sql.OpenDB(connector), nil
}

// Safe to call from multiple goroutines with different wallets
go connectWithWallet("oracle://user1:pass@server1/svc", "/path/to/wallet1/cwallet.sso")
go connectWithWallet("oracle://user2:pass@server2/svc", "/path/to/wallet2/cwallet.sso")
```

## Advanced Examples

### Loading from HTTP

```go
resp, _ := http.Get("https://secrets.example.com/wallet/cwallet.sso")
defer resp.Body.Close()

connector := go_ora.NewConnector("oracle://user:pass@server/service?SSL=enable")
connector.WithWallet(resp.Body)
db := sql.OpenDB(connector)
```

### Loading from In-Memory Data

```go
import "bytes"

walletData := []byte{...} // Your wallet data
reader := bytes.NewReader(walletData)

connector := go_ora.NewConnector("oracle://user:pass@server/service?SSL=enable")
connector.WithWallet(reader)
db := sql.OpenDB(connector)
```

### Direct API Usage

You can also use the configuration API directly:

```go
import "github.com/sijms/go-ora/v2/configurations"

walletFile, _ := os.Open("cwallet.sso")
defer walletFile.Close()

wallet, _ := configurations.NewWalletFromReader(walletFile)

// Use wallet directly with your configuration
```

## Notes

- This feature supports **cwallet.sso** files only
- The wallet data is read immediately when `WithWallet()` is called
- The reader can be closed after `WithWallet()` returns
- All existing wallet functionality (credential extraction, certificate handling) works the same
- Backward compatibility is maintained - file path-based loading still works via URL parameters
