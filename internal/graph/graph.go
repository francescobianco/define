package graph

import "github.com/francescobianco/define/internal/dsl"

type Concept struct {
	Name        string
	Deps        []string
	Invocations [][]string // each entry is the targets of one invocation statement
	Line        int
}

type Graph struct {
	Concepts map[string]*Concept
	Order    []string // declaration order
}

func Build(file *dsl.File) *Graph {
	g := &Graph{
		Concepts: make(map[string]*Concept),
	}
	for _, def := range file.Definitions {
		c := &Concept{
			Name: def.Name,
			Deps: def.Deps,
			Line: def.Line,
		}
		for _, inv := range def.Invocations {
			c.Invocations = append(c.Invocations, inv.Targets)
		}
		g.Concepts[def.Name] = c
		g.Order = append(g.Order, def.Name)
	}
	return g
}

// AllInvoked returns the set of all symbols ever invoked across all bodies.
func (g *Graph) AllInvoked() map[string]bool {
	invoked := make(map[string]bool)
	for _, c := range g.Concepts {
		for _, inv := range c.Invocations {
			for _, t := range inv {
				invoked[t] = true
			}
		}
	}
	return invoked
}

// AllReferenced returns the set of all symbols that appear in any "with" clause.
func (g *Graph) AllReferenced() map[string]bool {
	referenced := make(map[string]bool)
	for _, c := range g.Concepts {
		for _, dep := range c.Deps {
			referenced[dep] = true
		}
	}
	return referenced
}