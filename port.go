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

	PortSuffixSeed   string
	PortSuffixDigits int
	UsePortSuffix    bool

	SeedSegments []SeedSegment

	optionErr error
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

// WithPortSuffix uses seed to choose the trailing digits of a fallback port.
// WithSeed chooses the shared leading part. digits must be between 1 and 4.
func WithPortSuffix(seed string, digits int) ParsePortOption {
	return func(args *ParsePortArgs) {
		args.PortSuffixSeed = seed
		args.PortSuffixDigits = digits
		args.UsePortSuffix = true
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
	if args.optionErr != nil {
		return Port{}, args.optionErr
	}

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

	if args.SeedSegments != nil {
		return findAvailablePortWithSegments(args)
	}

	if args.UsePortSuffix || args.PortSuffixDigits != 0 {
		return findAvailablePortWithSuffix(args)
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

func findAvailablePortWithSuffix(args ParsePortArgs) (Port, error) {
	basePort, startSuffix, suffixCount, err := portSuffixCandidate(args)
	if err != nil {
		return Port{}, err
	}

	return findAvailablePortInBlock(basePort, startSuffix, suffixCount, args.MaxTries)
}

func findAvailablePortInBlock(basePort, startSuffix, suffixCount, maxTries int) (Port, error) {
	maxAttempts := min(maxTries, suffixCount)
	for attempt := range maxAttempts {
		currentPort := basePort + ((startSuffix + attempt) % suffixCount)
		if !isPortAvailable(currentPort) {
			continue
		}

		return Port{
			Str: strconv.Itoa(currentPort),
			Int: currentPort,
		}, nil
	}

	return Port{}, fmt.Errorf("failed to find available port after %d attempts", maxAttempts)
}

// portSuffixCandidate returns a seed-selected block and a suffix within it.
func portSuffixCandidate(args ParsePortArgs) (basePort, startSuffix, suffixCount int, err error) {
	digits := args.PortSuffixDigits
	if digits < 1 || digits > 4 {
		return 0, 0, 0, fmt.Errorf("invalid port suffix digits: %d (must be between 1 and 4)", digits)
	}

	minPort, maxPort := args.NewPortFrom, args.NewPortTo
	if minPort < 1 || maxPort > 65535 || maxPort < minPort {
		return 0, 0, 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	suffixCount = 1
	for range digits {
		suffixCount *= 10
	}

	firstBlock := ((minPort + suffixCount - 1) / suffixCount) * suffixCount
	lastBlock := ((maxPort + 1) / suffixCount * suffixCount) - suffixCount
	if lastBlock < firstBlock {
		return 0, 0, 0, fmt.Errorf("port range %d-%d has no complete block for %d trailing digits", minPort, maxPort, digits)
	}

	blockCount := (lastBlock-firstBlock)/suffixCount + 1
	var blockOffset int
	if args.Seed != "" {
		blockOffset = seedOffset(args.Seed, blockCount)
	} else {
		blockOffset = rand.IntN(blockCount)
	}

	basePort = firstBlock + blockOffset*suffixCount
	startSuffix = seedOffset(args.PortSuffixSeed, suffixCount)
	return basePort, startSuffix, suffixCount, nil
}

func seedOffset(seed string, portsCount int) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(seed))

	return int(hash.Sum64() % uint64(portsCount))
}
