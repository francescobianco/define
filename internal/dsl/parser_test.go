package dsl_test

import (
	"testing"

	"github.com/francescobianco/define/internal/dsl"
)

func TestParseSimple(t *testing.T) {
	src := `define load
define test`

	file, err := dsl.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(file.Definitions))
	}
	if file.Definitions[0].Name != "load" {
		t.Errorf("expected name=load, got %q", file.Definitions[0].Name)
	}
}

func TestParseWithDeps(t *testing.T) {
	src := `define Example1 with Example2, Routes`

	file, err := dsl.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := file.Definitions[0]
	if def.Name != "Example1" {
		t.Errorf("expected name=Example1, got %q", def.Name)
	}
	if len(def.Deps) != 2 || def.Deps[0] != "Example2" || def.Deps[1] != "Routes" {
		t.Errorf("unexpected deps: %v", def.Deps)
	}
}

func TestParseBody(t *testing.T) {
	src := `define /routes/a {
    load;
    Example2 test;
}`

	file, err := dsl.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := file.Definitions[0]
	if len(def.Invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(def.Invocations))
	}
	if len(def.Invocations[0].Targets) != 1 || def.Invocations[0].Targets[0] != "load" {
		t.Errorf("unexpected first invocation: %v", def.Invocations[0])
	}
	if len(def.Invocations[1].Targets) != 2 {
		t.Errorf("unexpected second invocation: %v", def.Invocations[1])
	}
}

func TestDuplicateDefinition(t *testing.T) {
	src := `define foo
define foo`

	_, err := dsl.Parse(src)
	if err == nil {
		t.Fatal("expected error for duplicate definition")
	}
}

func TestCommentSkipped(t *testing.T) {
	src := `# this is a comment
define foo`

	file, err := dsl.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Definitions) != 1 || file.Definitions[0].Name != "foo" {
		t.Errorf("unexpected definitions: %v", file.Definitions)
	}
}