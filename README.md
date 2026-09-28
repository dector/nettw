# nettw

> [!NOTE]
> This package moved from `github.com/dector/nettw` to `dector.space/go/nettw`.
> Use the new import path.

Small network-oriented Go library.

## Installation

```bash
go get dector.space/go/nettw
```

## Usage

### Get available port (preferebly user-specified)

Use `ParsePortOrPickAnother()`:

```go
package main

import (
    "fmt"
    "dector.space/go/nettw"
)

func main() {
    port, err := nettw.ParsePortOrPickAnother("8080")
    if err != nil {
        panic(err)
    }

    // If port 8080 free - we will get it here.
    // If not - random free port will be returned.
    fmt.Printf("Using port: %d\n", port.Int)
}
```

### Getting a Port with Custom Options

Use `ParsePortOrPickAnother()` with functional options:

```go
package main

import (
    "fmt"
    "dector.space/go/nettw"
)

func main() {
    port, err := nettw.ParsePortOrPickAnother(
        "8080",
        nettw.WithPortRange(3000, 4000),
        nettw.WithMaxTries(50),
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Using port: %d\n", port.Int)
}
```

### Getting a Deterministic Random Port

Use `WithSeed()` to prefer the same fallback port for the same seed and port range:

```go
port, err := nettw.ParsePortOrPickAnother(
    "",
    nettw.WithIgnoreInvalidPort(true),
    nettw.WithPortRange(3000, 4000),
    nettw.WithSeed("/full/path/to/file.txt"),
)
```

If the seeded port is unavailable, nettw checks the next ports in a deterministic order.

## License

This project is licensed under the MIT License.
