package nettw

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"net"
	"strconv"
)

type ParsePortArgs struct {
	IgnoreInvalidPort bool

	NewPortFrom int
	NewPortTo   int
	MaxTries    int
	Seed        string
}

type ParsePortOption func(*ParsePortArgs)

type Port struct {
	Str string
	Int int
}

func ParsePortOrPickAnother(port string, options ...ParsePortOption) (Port, error) {
	args := defaultParsePortArgs()
	for _, option := range options {
		option(&args)
	}

	return parsePortOrPickAnother(port, args)
}

// ParsePortOrPickAnotherWithArgs is kept for compatibility.
// Prefer ParsePortOrPickAnother with functional options for new code.
func ParsePortOrPickAnotherWithArgs(port string, args ParsePortArgs) (Port, error) {
	return parsePortOrPickAnother(port, args)
}

func WithIgnoreInvalidPort(ignore bool) ParsePortOption {
	return func(args *ParsePortArgs) {
		args.IgnoreInvalidPort = ignore
	}
}

func WithPortRange(from, to int) ParsePortOption {
	return func(args *ParsePortArgs) {
		args.NewPortFrom = from
		args.NewPortTo = to
	}
}

func WithMaxTries(maxTries int) ParsePortOption {
	return func(args *ParsePortArgs) {
		args.MaxTries = maxTries
	}
}

func WithSeed(seed string) ParsePortOption {
	return func(args *ParsePortArgs) {
		args.Seed = seed
	}
}

func defaultParsePortArgs() ParsePortArgs {
	return ParsePortArgs{
		IgnoreInvalidPort: false,

		NewPortFrom: 10000,
		NewPortTo:   20000,
		MaxTries:    100,
	}
}

func parsePortOrPickAnother(port string, args ParsePortArgs) (Port, error) {
	pport, err := strconv.Atoi(port)
	if err != nil && !args.IgnoreInvalidPort {
		return Port{}, err
	}
	if err == nil && isPortAvailable(pport) {
		return Port{
			Str: strconv.Itoa(pport),
			Int: pport,
		}, nil
	}

	return findAvailablePort(args)
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	ln.Close()
	return true
}

func findAvailablePort(args ParsePortArgs) (Port, error) {
	minPort, maxPort := args.NewPortFrom, args.NewPortTo
	if maxPort < minPort {
		return Port{}, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	portsCount := maxPort - minPort + 1

	if args.Seed != "" {
		return findAvailablePortFromSeed(args, portsCount)
	}

	tried := make(map[int]struct{})

	for attemptsLeft := args.MaxTries; attemptsLeft > 0; attemptsLeft-- {
		if len(tried) >= portsCount {
			return Port{}, fmt.Errorf("failed to find available port after %d attempts", args.MaxTries)
		}

		currentPort := minPort + rand.IntN(portsCount)

		if !isPortAvailable(currentPort) {
			tried[currentPort] = struct{}{}
			continue
		}

		return Port{
			Str: strconv.Itoa(currentPort),
			Int: currentPort,
		}, nil
	}

	return Port{}, fmt.Errorf("failed to find available port after %d attempts", args.MaxTries)
}

func findAvailablePortFromSeed(args ParsePortArgs, portsCount int) (Port, error) {
	start := seedOffset(args.Seed, portsCount)
	maxAttempts := min(args.MaxTries, portsCount)

	for attempt := range maxAttempts {
		currentPort := args.NewPortFrom + ((start + attempt) % portsCount)

		if !isPortAvailable(currentPort) {
			continue
		}

		return Port{
			Str: strconv.Itoa(currentPort),
			Int: currentPort,
		}, nil
	}

	return Port{}, fmt.Errorf("failed to find available port after %d attempts", args.MaxTries)
}

func seedOffset(seed string, portsCount int) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(seed))

	return int(hash.Sum64() % uint64(portsCount))
}
