package nettw

import "testing"

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
