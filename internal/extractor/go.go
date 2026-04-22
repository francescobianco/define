package extractor

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	ConceptName string
	Deps        []string
}

type GoResult struct {
	Packages []*Package
}

// ExtractGoPackages walks rootDir, groups .go files by directory,
// and returns one concept per package with internal import deps.
// It performs two passes: first to resolve dir→conceptName, then to map deps.
func ExtractGoPackages(rootDir string, ignoreTests bool) (*GoResult, error) {
	module, err := readGoModule(filepath.Join(rootDir, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	fset := token.NewFileSet()

	type rawPkg struct {
		conceptName string
		rawImports  []string // internal import paths, resolved after first pass
	}

	rawMap := make(map[string]*rawPkg) // relDir → rawPkg
	var dirOrder []string              // relDir declaration order

	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			// Skip vendor, hidden dirs, and _-prefixed dirs (ignored by Go build)
			if name == "vendor" || (name != "." && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_"))) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if ignoreTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}

		relDir, _ := filepath.Rel(rootDir, filepath.Dir(path))
		relDir = filepath.ToSlash(relDir)

		pkg, exists := rawMap[relDir]
		if !exists {
			pkg = &rawPkg{conceptName: dirToConceptName(relDir, f.Name.Name)}
			rawMap[relDir] = pkg
			dirOrder = append(dirOrder, relDir)
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if isInternalImport(importPath, module) && !contains(pkg.rawImports, importPath) {
				pkg.rawImports = append(pkg.rawImports, importPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build importPath → conceptName lookup using completed rawMap.
	importToConceptName := make(map[string]string)
	for relDir, pkg := range rawMap {
		var importPath string
		if relDir == "." || relDir == "" {
			importPath = module
		} else {
			importPath = module + "/" + relDir
		}
		importToConceptName[importPath] = pkg.conceptName
	}

	// Root package first, then rest in declaration order.
	sortedDirs := make([]string, 0, len(dirOrder))
	for _, d := range dirOrder {
		if d == "." || d == "" {
			sortedDirs = append([]string{d}, sortedDirs...)
		} else {
			sortedDirs = append(sortedDirs, d)
		}
	}

	result := &GoResult{}
	for _, relDir := range sortedDirs {
		pkg := rawMap[relDir]
		p := &Package{ConceptName: pkg.conceptName}
		for _, imp := range pkg.rawImports {
			if dep, ok := importToConceptName[imp]; ok && dep != pkg.conceptName {
				if !contains(p.Deps, dep) {
					p.Deps = append(p.Deps, dep)
				}
			}
		}
		result.Packages = append(result.Packages, p)
	}
	return result, nil
}

// dirToConceptName maps a relative directory to a DSL concept name.
// Root package uses the Go package name (e.g. "main", "gin", "chi").
// Every other directory uses its full relative path — always unique.
func dirToConceptName(relDir, goPkgName string) string {
	if relDir == "." || relDir == "" {
		return goPkgName
	}
	return relDir
}

func isInternalImport(importPath, module string) bool {
	return importPath == module || strings.HasPrefix(importPath, module+"/")
}

func readGoModule(gomodPath string) (string, error) {
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in %s", gomodPath)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
