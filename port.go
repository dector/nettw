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

type Port struct {
	Str string
	Int int
}

func ParsePortOrPickAnother(port string) (Port, error) {
	return ParsePortOrPickAnotherWithArgs(port, ParsePortArgs{
		IgnoreInvalidPort: false,

		NewPortFrom: 10000,
		NewPortTo:   20000,
		MaxTries:    100,
	})
}

func ParsePortOrPickAnotherWithArgs(port string, args ParsePortArgs) (Port, error) {
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
