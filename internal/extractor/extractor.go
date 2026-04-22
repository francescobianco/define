package extractor

import "fmt"

// Extract dispatches to the language-specific extractor based on cfg.Language.
func Extract(dir string, cfg Config, ignoreTests bool) (*GoResult, error) {
	switch cfg.Language {
	case "go":
		return ExtractGoPackages(dir, ignoreTests)
	case "php", "typescript":
		return nil, fmt.Errorf("language %q extractor is planned for v0.2", cfg.Language)
	default:
		return nil, fmt.Errorf("unsupported language %q", cfg.Language)
	}
}
