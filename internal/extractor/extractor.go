package extractor

import (
	"fmt"
	"strings"
)

// Extract dispatches to the language-specific extractor based on cfg.Language.
// ignoreTests is the CLI flag value; cfg.IncludeTests overrides it when true.
func Extract(dir string, cfg Config, ignoreTestsFlag bool) (*GoResult, error) {
	ignoreTests := ignoreTestsFlag
	if cfg.IncludeTests {
		ignoreTests = false
	}

	var result *GoResult
	var err error

	switch cfg.Language {
	case "go":
		result, err = ExtractGoPackages(dir, ignoreTests)
	case "php":
		result, err = ExtractPHPPackages(dir, ignoreTests)
	case "typescript":
		return nil, fmt.Errorf("typescript extractor is planned for v0.2")
	default:
		return nil, fmt.Errorf("unsupported language %q", cfg.Language)
	}
	if err != nil {
		return nil, err
	}

	if cfg.NamespaceStrip != "" {
		stripPrefix(result, cfg.NamespaceStrip)
	}

	return result, nil
}

// stripPrefix removes a common namespace prefix from all concept names and deps.
func stripPrefix(result *GoResult, prefix string) {
	prefix = strings.TrimRight(prefix, "/") + "/"
	rename := func(s string) string {
		return strings.TrimPrefix(s, prefix)
	}
	for _, pkg := range result.Packages {
		pkg.ConceptName = rename(pkg.ConceptName)
		for i, dep := range pkg.Deps {
			pkg.Deps[i] = rename(dep)
		}
	}
}
