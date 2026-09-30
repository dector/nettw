package nettw

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WithRepoAwareSeed combines WithSeed and WithPortSuffix using normalized paths.
// repoRoot is made absolute relative to the working directory. path is either
// relative to repoRoot or absolute within it. An empty path denotes the repo root.
// digits must be between 1 and 4. Invalid arguments cause ParsePortOrPickAnother
// to return an error, even when an explicitly requested port is available.
// Paths are cleaned lexically; symlinks are not resolved and paths need not exist.
func WithRepoAwareSeed(repoRoot, path string, digits int) ParsePortOption {
	return func(args *ParsePortArgs) {
		if args.optionErr != nil {
			return
		}
		if digits < 1 || digits > 4 {
			args.optionErr = fmt.Errorf("invalid port suffix digits: %d (must be between 1 and 4)", digits)
			return
		}

		repoSeed, suffixSeed, err := repoAwareSeeds(repoRoot, path)
		if err != nil {
			args.optionErr = err
			return
		}

		WithSeed(repoSeed)(args)
		WithPortSuffix(suffixSeed, digits)(args)
	}
}

func repoAwareSeeds(repoRoot, path string) (repoSeed, suffixSeed string, err error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve repo root: %w", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return "", "", fmt.Errorf("resolve path relative to repo: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q is outside repo %q", path, root)
	}

	return filepath.ToSlash(root), filepath.ToSlash(relative), nil
}
