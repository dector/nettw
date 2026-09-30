package nettw

import (
	"fmt"
	"slices"
)

// SeedSegment selects part of a decimal port number using a deterministic seed.
// The first segment must have Digits == 0 and selects the remaining leading part.
// Subsequent segments each specify between 1 and 4 digits, totaling at most 4.
type SeedSegment struct {
	Seed   string
	Digits int
}

// WithSeedSegments groups fallback ports by independently seeded decimal segments.
// At least two segments are required. Retries vary only the final segment, keeping
// the preceding segments fixed. The leading block must fit inside the port range.
// This option takes precedence over WithSeed and WithPortSuffix.
// Invalid segments cause ParsePortOrPickAnother to return an error, even when an
// explicitly requested port is available. Seeds are opaque strings, not paths.
func WithSeedSegments(segments ...SeedSegment) ParsePortOption {
	segments = slices.Clone(segments)
	return func(args *ParsePortArgs) {
		if args.optionErr != nil {
			return
		}
		if err := validateSeedSegments(segments); err != nil {
			args.optionErr = err
			return
		}
		args.SeedSegments = slices.Clone(segments)
	}
}

func validateSeedSegments(segments []SeedSegment) error {
	if len(segments) < 2 {
		return fmt.Errorf("seed segments require at least two segments")
	}
	if segments[0].Digits != 0 {
		return fmt.Errorf("first seed segment must have zero digits to select the leading part")
	}

	total := 0
	for i, segment := range segments[1:] {
		if segment.Digits < 1 || segment.Digits > 4 {
			return fmt.Errorf("invalid digits for seed segment %d: %d (must be between 1 and 4)", i+1, segment.Digits)
		}
		total += segment.Digits
		if total > 4 {
			return fmt.Errorf("seed segments may use at most 4 trailing digits in total")
		}
	}
	return nil
}

func findAvailablePortWithSegments(args ParsePortArgs) (Port, error) {
	base, start, count, err := portSegmentsCandidate(args)
	if err != nil {
		return Port{}, err
	}
	return findAvailablePortInBlock(base, start, count, args.MaxTries)
}

func portSegmentsCandidate(args ParsePortArgs) (base, start, count int, err error) {
	segments := args.SeedSegments
	if err := validateSeedSegments(segments); err != nil {
		return 0, 0, 0, err
	}
	minPort, maxPort := args.NewPortFrom, args.NewPortTo
	if minPort < 1 || maxPort > 65535 || maxPort < minPort {
		return 0, 0, 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	blockSize := 1
	for _, segment := range segments[1:] {
		blockSize *= decimalSize(segment.Digits)
	}
	firstBlock := ((minPort + blockSize - 1) / blockSize) * blockSize
	lastBlock := ((maxPort + 1) / blockSize * blockSize) - blockSize
	if lastBlock < firstBlock {
		return 0, 0, 0, fmt.Errorf("port range %d-%d has no complete block for seed segments", minPort, maxPort)
	}
	blockCount := (lastBlock-firstBlock)/blockSize + 1
	base = firstBlock + seedOffset(segments[0].Seed, blockCount)*blockSize

	remaining := blockSize
	for _, segment := range segments[1 : len(segments)-1] {
		size := decimalSize(segment.Digits)
		remaining /= size
		base += seedOffset(segment.Seed, size) * remaining
	}
	last := segments[len(segments)-1]
	count = decimalSize(last.Digits)
	start = seedOffset(last.Seed, count)
	return base, start, count, nil
}

func decimalSize(digits int) int {
	count := 1
	for range digits {
		count *= 10
	}
	return count
}
