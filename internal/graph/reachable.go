package graph

// Reachable returns the set of symbols reachable from root via deps and invocations.
func Reachable(g *Graph, root string) map[string]bool {
	visited := make(map[string]bool)

	var dfs func(name string)
	dfs = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		c, ok := g.Concepts[name]
		if !ok {
			return
		}
		for _, dep := range c.Deps {
			dfs(dep)
		}
		for _, inv := range c.Invocations {
			for _, target := range inv {
				dfs(target)
			}
		}
	}

	dfs(root)
	return visited
}