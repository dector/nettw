package nettw

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestSeedSegmentsMatchExistingTwoSeedSelection(t *testing.T) {
	for _, digits := range []int{1, 2, 3, 4} {
		t.Run(fmt.Sprint(digits), func(t *testing.T) {
			args := defaultParsePortArgs()
			WithSeed("project")(&args)
			WithPortSuffix("web", digits)(&args)
			wantBase, wantStart, wantCount, err := portSuffixCandidate(args)
			if err != nil {
				t.Fatal(err)
			}
			WithSeedSegments(SeedSegment{Seed: "project"}, SeedSegment{Seed: "web", Digits: digits})(&args)
			base, start, count, err := portSegmentsCandidate(args)
			if err != nil {
				t.Fatal(err)
			}
			if base != wantBase || start != wantStart || count != wantCount {
				t.Fatalf("got %d/%d/%d, want %d/%d/%d", base, start, count, wantBase, wantStart, wantCount)
			}
		})
	}
}

func TestSeedSegmentsSelectIndependentDecimalParts(t *testing.T) {
	args := defaultParsePortArgs()
	args.SeedSegments = []SeedSegment{
		{Seed: "project"},
		{Seed: "backend", Digits: 1},
		{Seed: "api", Digits: 2},
	}
	base, start, count, err := portSegmentsCandidate(args)
	if err != nil {
		t.Fatal(err)
	}
	if base%100 != 0 || (base/100)%10 != seedOffset("backend", 10) {
		t.Fatalf("incorrect middle segment in base %d", base)
	}
	if start != seedOffset("api", 100) || count != 100 {
		t.Fatalf("incorrect final segment: %d/%d", start, count)
	}
	if base < args.NewPortFrom || base+count-1 > args.NewPortTo {
		t.Fatal("selected block is outside the port range")
	}

	args.SeedSegments[2].Seed = "worker"
	otherBase, _, _, err := portSegmentsCandidate(args)
	if err != nil {
		t.Fatal(err)
	}
	if otherBase != base {
		t.Fatal("changing the final seed must preserve the preceding segments")
	}
	args.SeedSegments[1].Seed = "frontend"
	otherBase, _, _, err = portSegmentsCandidate(args)
	if err != nil {
		t.Fatal(err)
	}
	if otherBase/1000 != base/1000 {
		t.Fatal("changing the middle seed must preserve the leading segment")
	}
}

func TestSeedSegmentsRejectInvalidConfigurations(t *testing.T) {
	for _, segments := range [][]SeedSegment{
		nil,
		{{Seed: "project"}},
		{{Digits: 1}, {Digits: 1}},
		{{}, {}},
		{{}, {Digits: -1}},
		{{}, {Digits: 5}},
		{{}, {Digits: 3}, {Digits: 2}},
	} {
		_, err := ParsePortOrPickAnother("8080", WithSeedSegments(segments...))
		if err == nil {
			t.Fatalf("expected invalid segment error for %+v", segments)
		}
	}
}

func TestSeedSegmentsValidateRange(t *testing.T) {
	for _, ports := range [][2]int{{0, 10000}, {10000, 65536}, {20000, 10000}, {10001, 10099}} {
		args := defaultParsePortArgs()
		args.NewPortFrom, args.NewPortTo = ports[0], ports[1]
		args.SeedSegments = []SeedSegment{{Seed: "project"}, {Seed: "web", Digits: 2}}
		if _, _, _, err := portSegmentsCandidate(args); err == nil {
			t.Fatalf("expected invalid or incomplete range error for %v", ports)
		}
	}
}

func TestSeedSegmentsCopiesInput(t *testing.T) {
	segments := []SeedSegment{{Seed: "project"}, {Seed: "web", Digits: 1}}
	option := WithSeedSegments(segments...)
	segments[0].Seed = "mutated"
	first := defaultParsePortArgs()
	option(&first)
	if first.SeedSegments[0].Seed != "project" {
		t.Fatal("option must snapshot its input")
	}
	first.SeedSegments[0].Seed = "mutated again"
	second := defaultParsePortArgs()
	option(&second)
	if second.SeedSegments[0].Seed != "project" {
		t.Fatal("reusing the option must produce an independent segment slice")
	}
}

func TestSeedSegmentsEmptySeedsAreDeterministic(t *testing.T) {
	args := defaultParsePortArgs()
	args.SeedSegments = []SeedSegment{{}, {Digits: 1}, {Digits: 1}, {Digits: 1}, {Digits: 1}}
	base, start, count, err := portSegmentsCandidate(args)
	if err != nil {
		t.Fatal(err)
	}
	for range 10 {
		otherBase, otherStart, otherCount, err := portSegmentsCandidate(args)
		if err != nil {
			t.Fatal(err)
		}
		if base != otherBase || start != otherStart || count != otherCount {
			t.Fatal("empty seeds must still produce deterministic segments")
		}
	}
}

func TestPortBlockRetriesWrapWithoutChangingPrefix(t *testing.T) {
	// Reserve an entire decimal block to test exhaustion and wrapping reliably.
	var listeners []net.Listener
	base := 40000
	for ; base < 60000; base += 10 {
		listeners = nil
		for suffix := range 10 {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", base+suffix))
			if err != nil {
				break
			}
			listeners = append(listeners, ln)
		}
		if len(listeners) == 10 {
			break
		}
		for _, ln := range listeners {
			ln.Close()
		}
	}
	if len(listeners) != 10 {
		t.Fatal("unable to reserve a port block")
	}
	t.Cleanup(func() {
		for _, ln := range listeners {
			ln.Close()
		}
	})

	if _, err := findAvailablePortInBlock(base, 9, 10, 100); err == nil || !strings.Contains(err.Error(), "10 attempts") {
		t.Fatalf("expected exhaustion after 10 attempts, got %v", err)
	}
	listeners[0].Close()
	if _, err := findAvailablePortInBlock(base, 9, 10, 1); err == nil {
		t.Fatal("expected max tries to be respected")
	}
	port, err := findAvailablePortInBlock(base, 9, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	if port.Int != base || port.Str != fmt.Sprint(base) {
		t.Fatalf("expected wrapped port %d, got %+v", base, port)
	}
}
