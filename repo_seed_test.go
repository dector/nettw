package nettw

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRepoAwareSeedNormalizesPaths(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	root := filepath.Join(cwd, "repo")

	var expected ParsePortArgs
	for _, tc := range []struct {
		name string
		root string
		path string
	}{
		{"relative path", root, filepath.Join("apps", "web")},
		{"absolute path", root, filepath.Join(root, "apps", "web")},
		{"relative repo", "repo", filepath.Join("apps", "web")},
		{"cleaned paths", filepath.Join(root, "apps", ".."), "apps/./api/../web"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := defaultParsePortArgs()
			WithRepoAwareSeed(tc.root, tc.path, 2)(&args)
			if args.optionErr != nil {
				t.Fatal(args.optionErr)
			}
			if args.Seed != filepath.ToSlash(root) || args.PortSuffixSeed != "apps/web" {
				t.Fatalf("unexpected seeds: %q, %q", args.Seed, args.PortSuffixSeed)
			}
			if args.PortSuffixDigits != 2 || !args.UsePortSuffix {
				t.Fatal("expected two-digit suffix selection to be enabled")
			}
			if expected.Seed == "" {
				expected = args
			}
			base, suffix, count, err := portSuffixCandidate(args)
			if err != nil {
				t.Fatal(err)
			}
			wantBase, wantSuffix, wantCount, err := portSuffixCandidate(expected)
			if err != nil {
				t.Fatal(err)
			}
			if base != wantBase || suffix != wantSuffix || count != wantCount {
				t.Fatal("equivalent paths must select the same port candidate")
			}
		})
	}
}

func TestRepoAwareSeedRootPath(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"", ".", root} {
		args := defaultParsePortArgs()
		WithRepoAwareSeed(root, path, 1)(&args)
		if args.optionErr != nil {
			t.Fatal(args.optionErr)
		}
		if args.PortSuffixSeed != "." {
			t.Fatalf("expected root suffix seed, got %q", args.PortSuffixSeed)
		}
	}
}

func TestRepoAwareSeedRejectsOutsidePaths(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	for _, path := range []string{
		"..",
		filepath.Join("..", "other"),
		filepath.Join(parent, "other"),
		filepath.Join(parent, "repo-other", "app"),
	} {
		_, err := ParsePortOrPickAnother("8080", WithRepoAwareSeed(root, path, 1))
		if err == nil || !strings.Contains(err.Error(), "outside repo") {
			t.Fatalf("expected outside-repo error for %q, got %v", path, err)
		}
	}
}

func TestRepoAwareSeedRejectsInvalidDigits(t *testing.T) {
	root := t.TempDir()
	for _, digits := range []int{-1, 0, 5} {
		_, err := ParsePortOrPickAnother("8080", WithRepoAwareSeed(root, "app", digits))
		if err == nil || !strings.Contains(err.Error(), "invalid port suffix digits") {
			t.Fatalf("expected invalid-width error for %d, got %v", digits, err)
		}
	}
}

func TestRepoAwareSeedSharesBlockAcrossSubdirectories(t *testing.T) {
	root := t.TempDir()
	var wantBase int
	for _, path := range []string{"apps/web", "apps/api", "docs"} {
		args := defaultParsePortArgs()
		WithRepoAwareSeed(root, path, 3)(&args)
		if args.optionErr != nil {
			t.Fatal(args.optionErr)
		}
		base, _, _, err := portSuffixCandidate(args)
		if err != nil {
			t.Fatal(err)
		}
		if wantBase == 0 {
			wantBase = base
		}
		if base != wantBase {
			t.Fatalf("expected shared repo block %d, got %d", wantBase, base)
		}
	}
}
