package nettw

import (
	"strconv"
	"testing"
)

func TestSeedOffsetIsDeterministic(t *testing.T) {
	const portsCount = 1000

	first := seedOffset("/full/path/to/file.txt", portsCount)
	second := seedOffset("/full/path/to/file.txt", portsCount)

	if first != second {
		t.Fatalf("expected same offset for same seed, got %d and %d", first, second)
	}
}

func TestSeedOffsetStaysInRange(t *testing.T) {
	const portsCount = 1000
	offset := seedOffset("/full/path/to/file.txt", portsCount)

	if offset < 0 || offset >= portsCount {
		t.Fatalf("expected offset in [0, %d), got %d", portsCount, offset)
	}
}

func TestPortSuffixKeepsRepoBlockAcrossSubdirectories(t *testing.T) {
	args := defaultParsePortArgs()
	args.Seed = "/repo"
	args.PortSuffixSeed = "apps/web"
	args.PortSuffixDigits = 1

	webBase, webSuffix, suffixCount, err := portSuffixCandidate(args)
	if err != nil {
		t.Fatal(err)
	}

	args.PortSuffixSeed = "apps/api"
	apiBase, apiSuffix, apiSuffixCount, err := portSuffixCandidate(args)
	if err != nil {
		t.Fatal(err)
	}

	if webBase != apiBase {
		t.Fatalf("expected same repo block, got %d and %d", webBase, apiBase)
	}
	if suffixCount != 10 || apiSuffixCount != suffixCount {
		t.Fatalf("expected 10 suffixes, got %d and %d", suffixCount, apiSuffixCount)
	}
	if webSuffix < 0 || webSuffix >= suffixCount || apiSuffix < 0 || apiSuffix >= suffixCount {
		t.Fatalf("suffixes must be within [0, %d), got %d and %d", suffixCount, webSuffix, apiSuffix)
	}
	if webBase%10 != 0 {
		t.Fatalf("expected block base to align to 10, got %d", webBase)
	}
}

func TestPortSuffixSupportsMultipleWidths(t *testing.T) {
	for _, digits := range []int{1, 2, 3} {
		t.Run(strconv.Itoa(digits), func(t *testing.T) {
			args := defaultParsePortArgs()
			args.Seed = "/repo"
			args.PortSuffixSeed = "apps/web"
			args.PortSuffixDigits = digits

			base, suffix, suffixCount, err := portSuffixCandidate(args)
			if err != nil {
				t.Fatal(err)
			}
			if suffixCount != pow10(digits) {
				t.Fatalf("expected %d suffixes, got %d", pow10(digits), suffixCount)
			}
			if base%suffixCount != 0 {
				t.Fatalf("expected base %d to align to block size %d", base, suffixCount)
			}
			if suffix < 0 || suffix >= suffixCount {
				t.Fatalf("expected suffix in [0, %d), got %d", suffixCount, suffix)
			}
		})
	}
}

func TestPortSuffixRejectsInvalidWidthAndIncompleteBlocks(t *testing.T) {
	if _, err := ParsePortOrPickAnother(
		"",
		WithIgnoreInvalidPort(true),
		WithPortSuffix("app", 0),
	); err == nil {
		t.Fatal("expected invalid suffix width error")
	}

	args := defaultParsePortArgs()
	args.Seed = "/repo"
	args.PortSuffixSeed = "app"
	args.PortSuffixDigits = 1
	args.NewPortFrom = 10001
	args.NewPortTo = 10009
	if _, _, _, err := portSuffixCandidate(args); err == nil {
		t.Fatal("expected incomplete port block error")
	}
}

func pow10(digits int) int {
	value := 1
	for range digits {
		value *= 10
	}
	return value
}
