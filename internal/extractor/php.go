package extractor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	rePHPNamespace = regexp.MustCompile(`(?m)^namespace\s+([\w\\]+)\s*;`)
	rePHPUse       = regexp.MustCompile(`(?m)^use\s+([\w\\]+)(?:\s+as\s+\w+)?\s*;`)
	rePHPClassDecl = regexp.MustCompile(`(?mi)(?:abstract\s+|final\s+|readonly\s+)*(?:class|interface|trait|enum)\s+(\w+)` +
		`(?:\s+extends\s+([\w\\]+))?` +
		`(?:\s+implements\s+([\w\\,\s]+))?`)
)

// ExtractPHPPackages extracts class-level concepts from PHP source files.
// Each class/interface/trait becomes one concept; use+extends+implements become deps.
func ExtractPHPPackages(rootDir string, ignoreTests bool) (*GoResult, error) {
	classMap := make(map[string]*Package)
	var order []string

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || (name != "." && strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			if ignoreTests && isPHPTestDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".php") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(src)

		ns := extractPHPNamespace(content)

		// Build alias → FQN map from use statements
		aliases := extractPHPUseAliases(content)

		// Process each class/interface/trait declaration in the file
		for _, m := range rePHPClassDecl.FindAllStringSubmatch(content, -1) {
			className := m[1]
			conceptFQN := phpFQN(ns, className)

			pkg, exists := classMap[conceptFQN]
			if !exists {
				pkg = &Package{ConceptName: conceptFQN}
				classMap[conceptFQN] = pkg
				order = append(order, conceptFQN)
			}

			// Deps from use statements
			for _, fqn := range aliases {
				dep := phpToConceptName(fqn)
				if dep != conceptFQN && !contains(pkg.Deps, dep) {
					pkg.Deps = append(pkg.Deps, dep)
				}
			}

			// Dep from extends (same-namespace resolution)
			if m[2] != "" {
				dep := phpResolve(m[2], ns, aliases)
				if dep != "" && dep != conceptFQN && !contains(pkg.Deps, dep) {
					pkg.Deps = append(pkg.Deps, dep)
				}
			}

			// Deps from implements (comma-separated)
			if m[3] != "" {
				for _, iface := range strings.Split(m[3], ",") {
					iface = strings.TrimSpace(iface)
					if iface == "" {
						continue
					}
					dep := phpResolve(iface, ns, aliases)
					if dep != "" && dep != conceptFQN && !contains(pkg.Deps, dep) {
						pkg.Deps = append(pkg.Deps, dep)
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Filter deps: keep only concepts defined in this extraction
	known := make(map[string]bool, len(order))
	for _, name := range order {
		known[name] = true
	}

	result := &GoResult{}
	for _, name := range order {
		pkg := classMap[name]
		var deps []string
		for _, dep := range pkg.Deps {
			if known[dep] {
				deps = append(deps, dep)
			}
		}
		pkg.Deps = deps
		result.Packages = append(result.Packages, pkg)
	}
	return result, nil
}

// phpResolve resolves a class name (possibly unqualified) to a concept name.
// Checks use-aliases first, then falls back to current namespace.
func phpResolve(name, ns string, aliases map[string]string) string {
	// Already fully qualified (starts with \) or contains \
	if strings.HasPrefix(name, "\\") {
		return phpToConceptName(strings.TrimPrefix(name, "\\"))
	}
	// Simple name — check aliases first
	base := strings.SplitN(name, "\\", 2)[0]
	if fqn, ok := aliases[base]; ok {
		if strings.Contains(name, "\\") {
			// e.g. alias\Sub
			suffix := strings.SplitN(name, "\\", 2)[1]
			return phpToConceptName(fqn + "\\" + suffix)
		}
		return phpToConceptName(fqn)
	}
	// Assume same namespace
	if ns != "" {
		return phpToConceptName(ns + "\\" + name)
	}
	return phpToConceptName(name)
}

func extractPHPNamespace(content string) string {
	if m := rePHPNamespace.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	return ""
}

func extractPHPUseAliases(content string) map[string]string {
	aliases := make(map[string]string)
	for _, m := range rePHPUse.FindAllStringSubmatch(content, -1) {
		fqn := m[1]
		parts := strings.Split(fqn, "\\")
		alias := parts[len(parts)-1]
		aliases[alias] = fqn
	}
	return aliases
}

// phpFQN builds the fully-qualified class name from namespace + class.
func phpFQN(ns, className string) string {
	if ns == "" {
		return phpToConceptName(className)
	}
	return phpToConceptName(ns + "\\" + className)
}

// phpToConceptName replaces PHP namespace separator \ with DSL-safe /.
func phpToConceptName(fqn string) string {
	return strings.ReplaceAll(strings.TrimPrefix(fqn, "\\"), "\\", "/")
}

func isPHPTestDir(name string) bool {
	lower := strings.ToLower(name)
	return lower == "tests" || lower == "test" || lower == "spec" || lower == "specs"
}
