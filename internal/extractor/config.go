package extractor

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Language string   `yaml:"language"`
	Sources  []string `yaml:"sources"`
	Ignore   []string `yaml:"ignore"`
}

// LoadConfig reads the profile file. If cfgFile is empty it looks for
// define.yml inside dir, then falls back to auto-detection from the source tree.
func LoadConfig(dir, cfgFile string) (Config, error) {
	if cfgFile == "" {
		cfgFile = filepath.Join(dir, "define.yml")
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		// No config file: auto-detect language from dir
		return autoDetect(dir)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", cfgFile, err)
	}
	return cfg, nil
}

func autoDetect(dir string) (Config, error) {
	checks := []struct {
		glob string
		lang string
	}{
		{"*.go", "go"},
		{"go.mod", "go"},
		{"package.json", "typescript"},
		{"composer.json", "php"},
	}
	for _, c := range checks {
		matches, _ := filepath.Glob(filepath.Join(dir, c.glob))
		if len(matches) > 0 {
			return Config{Language: c.lang}, nil
		}
	}
	return Config{}, fmt.Errorf("cannot detect language in %s — use --config", dir)
}
