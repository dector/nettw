# nettw

Small network-oriented Go library.

## Installation

```bash
go get github.com/dector/nettw
```

## Usage

### Get available port (preferebly user-specified)

Use `ParsePortOrPickAnother()`:

```go
package main

import (
    "fmt"
    "github.com/dector/nettw"
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

### Getting a Port with Custom Arguments

Use `ParsePortOrPickAnotherWithArgs()`:

```go
package main

import (
    "fmt"
    "github.com/dector/nettw"
)

func main() {
    port, err := nettw.ParsePortOrPickAnotherWithArgs(
        "8080",
        nettw.ParsePortArgs{
            NewPortFrom:       3000,
            NewPortTo:         4000,
            MaxTries:          50,
        }
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Using port: %d\n", port.Int)
}
```

## License

This project is licensed under the MIT License.
