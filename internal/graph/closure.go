package graph

import "fmt"

type ClosureResult struct {
	Closed bool
	Errors []string
}

// CheckClosure performs DFS from root, verifying:
//   - every referenced/invoked symbol is defined
//   - there are no dependency cycles
func CheckClosure(g *Graph, root string) ClosureResult {
	result := ClosureResult{Closed: true}

	if _, ok := g.Concepts[root]; !ok {
		result.Closed = false
		result.Errors = append(result.Errors, fmt.Sprintf("root concept %q is not defined", root))
		return result
	}

	const (
		white = 0 // not visited
		gray  = 1 // in current path (cycle detection)
		black = 2 // fully visited
	)
	color := make(map[string]int)
	path := []string{}

	var dfs func(name string)
	dfs = func(name string) {
		switch color[name] {
		case black:
			return
		case gray:
			// find where cycle starts in path
			result.Closed = false
			cycleStart := -1
			for i, n := range path {
				if n == name {
					cycleStart = i
					break
				}
			}
			cycle := append(path[cycleStart:], name)
			result.Errors = append(result.Errors, fmt.Sprintf("cycle detected: %v", cycle))
			return
		}

		c, ok := g.Concepts[name]
		if !ok {
			result.Closed = false
			result.Errors = append(result.Errors, fmt.Sprintf("%q is referenced but not defined", name))
			return
		}

		color[name] = gray
		path = append(path, name)

		for _, dep := range c.Deps {
			dfs(dep)
		}
		for _, inv := range c.Invocations {
			for _, target := range inv {
				dfs(target)
			}
		}

		path = path[:len(path)-1]
		color[name] = black
	}

	dfs(root)
	return result
}