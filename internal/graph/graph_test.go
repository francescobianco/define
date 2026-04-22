package graph_test

import (
	"testing"

	"github.com/francescobianco/define/internal/dsl"
	"github.com/francescobianco/define/internal/graph"
)

const exampleDef = `
define Example1 with Example2, Routes
define Example2 with load, test
define load
define test
define Routes with /routes/a
define /routes/a {
    load;
    Example2 test;
}
define Debug
define /routes/old
`

func buildGraph(t *testing.T, src string) *graph.Graph {
	t.Helper()
	file, err := dsl.Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return graph.Build(file)
}

func TestClosure_Closed(t *testing.T) {
	g := buildGraph(t, exampleDef)
	result := graph.CheckClosure(g, "Example1")
	if !result.Closed {
		t.Errorf("expected closed, errors: %v", result.Errors)
	}
}

func TestClosure_UndefinedDep(t *testing.T) {
	g := buildGraph(t, `define A with B`)
	result := graph.CheckClosure(g, "A")
	if result.Closed {
		t.Error("expected not closed for undefined dep B")
	}
}

func TestClosure_Cycle(t *testing.T) {
	g := buildGraph(t, `
define A with B
define B with A
`)
	result := graph.CheckClosure(g, "A")
	if result.Closed {
		t.Error("expected not closed due to cycle")
	}
}

func TestReachable(t *testing.T) {
	g := buildGraph(t, exampleDef)
	reachable := graph.Reachable(g, "Example1")

	expected := []string{"Example1", "Example2", "Routes", "load", "test", "/routes/a"}
	for _, name := range expected {
		if !reachable[name] {
			t.Errorf("%q should be reachable", name)
		}
	}

	if reachable["Debug"] {
		t.Error("Debug should not be reachable")
	}
	if reachable["/routes/old"] {
		t.Error("/routes/old should not be reachable")
	}
}

func TestDeadConcepts(t *testing.T) {
	g := buildGraph(t, exampleDef)
	dead := graph.FindDead(g, "Example1")

	unreachableSet := make(map[string]bool)
	for _, n := range dead.Unreachable {
		unreachableSet[n] = true
	}
	if !unreachableSet["Debug"] {
		t.Error("Debug should be unreachable")
	}
	if !unreachableSet["/routes/old"] {
		t.Error("/routes/old should be unreachable")
	}
}