package dag

import (
	"fmt"
)

type Node struct {
	ID string
}

type Edge struct {
	From      string
	To        string
	Condition map[string]any
}

type Graph struct {
	Nodes    map[string]Node
	OutEdges map[string][]Edge
	InEdges  map[string][]Edge
}

// Constructs a graph and validates it for missing nodes and cycles
func New(nodes []Node, edges []Edge) (*Graph, error) {
	g := &Graph{
		Nodes:    make(map[string]Node),
		OutEdges: make(map[string][]Edge),
		InEdges:  make(map[string][]Edge),
	}

	for _, n := range nodes {
		g.Nodes[n.ID] = n
	}

	for _, e := range edges {
		if _, ok := g.Nodes[e.From]; !ok {
			return nil, fmt.Errorf("edge references unknown parent node: %s", e.From)
		}
		if _, ok := g.Nodes[e.To]; !ok {
			return nil, fmt.Errorf("edge references unknown child node: %s", e.To)
		}
		g.OutEdges[e.From] = append(g.OutEdges[e.From], e)

		g.InEdges[e.To] = append(g.InEdges[e.To], e)
	}
	if hasCycle(g) {
		return nil, fmt.Errorf("workflow graph contains a cycle")
	}
	return g, nil
}

// Cycle detection using Kahn's algo
func hasCycle(g *Graph) bool {
	inDegree := make(map[string]int)
	for id := range g.Nodes {
		inDegree[id] = len(g.InEdges[id])
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	visitedCount := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visitedCount++
		for _, edge := range g.OutEdges[curr] {
			inDegree[edge.To]--
			if inDegree[edge.To] == 0 {
				queue = append(queue, edge.To)
			}
		}
	}
	// if we can't visit all the nodes, a cycle exists
	return visitedCount != len(g.Nodes)
}

// Constructs and returns a 2D array of node IDs grouped by execution order
// Nodes in the same inner array can be safely executed in parallel
func (g *Graph) Waves() [][]string {
	inDegree := make(map[string]int)
	for id := range g.Nodes {
		inDegree[id] = len(g.InEdges[id])
	}

	var waves [][]string
	var currwave []string

	for id, deg := range inDegree {
		if deg == 0 {
			currwave = append(currwave, id)
		}
	}

	for len(currwave) > 0 {
		waves = append(waves, currwave)
		var nextwave []string
		for _, nodeID := range currwave {
			for _, edge := range g.OutEdges[nodeID] {
				inDegree[edge.To]--
				if inDegree[edge.To] == 0 {
					nextwave = append(nextwave, edge.To)
				}
			}
		}
		currwave = nextwave
	}
	return waves
}

// Returns all incoming edges for a give nodeID
func (g *Graph) Parents(nodeID string) []Edge {
	return g.InEdges[nodeID]
}
