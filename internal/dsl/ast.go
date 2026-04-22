package dsl

// Invocation is a semicolon-terminated statement in a definition body.
// Targets holds each whitespace-separated token, e.g. "Example2 test;" → ["Example2", "test"].
type Invocation struct {
	Targets []string
}

type Definition struct {
	Name        string
	Deps        []string
	Invocations []Invocation
	Line        int
}

type File struct {
	Definitions []*Definition
}