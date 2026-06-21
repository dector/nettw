package nettw

import (
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
)

type ParsePortArgs struct {
	IgnoreInvalidPort bool

	NewPortFrom int
	NewPortTo   int
	MaxTries    int
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
	tried := make(map[int]struct{})

	minPort, maxPort := args.NewPortFrom, args.NewPortTo
	if maxPort < minPort {
		return Port{}, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	portsRange := maxPort - minPort

	for attemptsLeft := args.MaxTries; attemptsLeft > 0; attemptsLeft-- {
		if len(tried) >= portsRange {
			return Port{}, fmt.Errorf("failed to find available port after %d attempts", args.MaxTries)
		}

		randomPort := func() int {
			return minPort + rand.IntN(portsRange) + 1
		}
		currentPort := randomPort()

		if !isPortAvailable(currentPort) {
			tried[currentPort] = struct{}{}
			continue
		}

		return Port{
			Str: fmt.Sprintf("%d", currentPort),
			Int: currentPort,
		}, nil
	}

	return Port{}, fmt.Errorf("failed to find available port after %d attempts", args.MaxTries)
}
