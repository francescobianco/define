package graph

// DeadReport groups the three categories of dead concepts.
type DeadReport struct {
	NeverReferenced []string // never appears in any "with" clause
	NeverInvoked    []string // never appears in any body invocation
	Unreachable     []string // not reachable from root via deps or invocations
}

// FindDead computes dead concept categories relative to root.
// The root itself is excluded from all categories.
func FindDead(g *Graph, root string) DeadReport {
	referenced := g.AllReferenced()
	invoked := g.AllInvoked()
	reachable := Reachable(g, root)

	var report DeadReport
	for _, name := range g.Order {
		if name == root {
			continue
		}
		if !referenced[name] {
			report.NeverReferenced = append(report.NeverReferenced, name)
		}
		if !invoked[name] {
			report.NeverInvoked = append(report.NeverInvoked, name)
		}
		if !reachable[name] {
			report.Unreachable = append(report.Unreachable, name)
		}
	}
	return report
}