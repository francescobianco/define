package report

import (
	"fmt"
	"io"

	"github.com/francescobianco/define/internal/graph"
)

func Print(w io.Writer, root string, g *graph.Graph, closure graph.ClosureResult, dead graph.DeadReport, reachable map[string]bool) {
	if closure.Closed {
		fmt.Fprintf(w, "%s is closed\n", root)
	} else {
		fmt.Fprintf(w, "%s is NOT closed\n", root)
		fmt.Fprintf(w, "\nErrors:\n")
		for _, e := range closure.Errors {
			fmt.Fprintf(w, "  - %s\n", e)
		}
	}

	invokedSet := g.AllInvoked()
	invokedCount := 0
	for name := range invokedSet {
		if _, ok := g.Concepts[name]; ok {
			invokedCount++
		}
	}

	fmt.Fprintf(w, "\nCoverage:\n")
	fmt.Fprintf(w, "  - declared:  %d\n", len(g.Concepts))
	fmt.Fprintf(w, "  - reachable: %d\n", len(reachable))
	fmt.Fprintf(w, "  - invoked:   %d\n", invokedCount)

	hasDeadConcepts := len(dead.NeverReferenced) > 0 || len(dead.NeverInvoked) > 0 || len(dead.Unreachable) > 0

	if !hasDeadConcepts {
		fmt.Fprintf(w, "\nNo dead concepts found.\n")
		return
	}

	fmt.Fprintf(w, "\nDead concepts:\n")
	if len(dead.NeverReferenced) > 0 {
		fmt.Fprintf(w, "  - never referenced:\n")
		for _, name := range dead.NeverReferenced {
			fmt.Fprintf(w, "      %s\n", name)
		}
	}
	if len(dead.NeverInvoked) > 0 {
		fmt.Fprintf(w, "  - never invoked:\n")
		for _, name := range dead.NeverInvoked {
			fmt.Fprintf(w, "      %s\n", name)
		}
	}
	if len(dead.Unreachable) > 0 {
		fmt.Fprintf(w, "  - unreachable:\n")
		for _, name := range dead.Unreachable {
			fmt.Fprintf(w, "      %s\n", name)
		}
	}
}