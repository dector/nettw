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

### Keep ports for one repo visually grouped

Use `WithSeed()` for the repo and `WithPortSuffix()` for the subdirectory. The repo seed selects a shared port block; the subdirectory seed selects the requested number of trailing digits:

```go
port, err := nettw.ParsePortOrPickAnother(
    "",
    nettw.WithIgnoreInvalidPort(true),
    nettw.WithSeed("/work/my-repo"),
    nettw.WithPortSuffix("apps/web", 1),
)
```

For example, a one-digit suffix keeps ports in a block like `12340`–`12349`. Use `2` or `3` to vary the last two or three digits. Each repo-relative subdirectory can use a different suffix seed, while the repo seed keeps the leading part shared. If a candidate is busy, nettw tries other suffixes in that same block.

The selected block must fit completely inside the configured port range. `WithPortSuffix()` supports 1–4 trailing digits.

## License

This project is licensed under the MIT License.
