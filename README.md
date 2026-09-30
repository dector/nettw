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

### Multiple seed segments

`WithSeedSegments()` groups ports using independently seeded decimal segments:

```go
port, err := nettw.ParsePortOrPickAnother(
    "",
    nettw.WithIgnoreInvalidPort(true),
    nettw.WithSeedSegments(
        nettw.SeedSegment{Seed: "my-project"},
        nettw.SeedSegment{Seed: "backend", Digits: 1},
        nettw.SeedSegment{Seed: "api", Digits: 1},
    ),
)
```

For a five-digit port, this produces a pattern like `PPPST`: the project selects `PPP`, the service group selects `S`, and the server selects `T`. Changing a later seed preserves the preceding segments. Busy-port retries change only the final segment, wrapping within that segment; an exhausted group returns an error rather than changing its prefix.

The first segment must have `Digits: 0` (the default) and uses the remaining leading digits. At least two segments are required. Later segments each use 1–4 digits, with at most 4 trailing digits in total. The complete leading block must fit inside the configured port range. More trailing digits leave fewer distinct leading groups.

This option takes precedence over `WithSeed()` and `WithPortSuffix()`. Invalid segments return an error even if an explicitly requested port is available. Seeds are opaque strings: paths are not normalized or resolved. Seeds may collide, so distinct seeds do not guarantee unique ports.

`WithRepoAwareSeed()` has been removed. To migrate its two-part grouping, pass the normalized root and relative path as two segments:

```go
nettw.WithSeedSegments(
    nettw.SeedSegment{Seed: normalizedRoot},
    nettw.SeedSegment{Seed: relativePath, Digits: 2},
)
```

## License

This project is licensed under the MIT License.
