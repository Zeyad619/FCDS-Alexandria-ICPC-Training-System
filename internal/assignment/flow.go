package assignment

// A minimal Edmonds-Karp max-flow implementation.
//
// Deliberately simple (BFS-based Ford-Fulkerson) rather than Dinic's
// or push-relabel: community sizes here are small (tens to low
// hundreds of people), so the simpler algorithm is fast enough and
// much easier to read and verify than a more advanced one.

type flowEdge struct {
	to   int
	cap  int
	flow int
}

type flowGraph struct {
	edges [][]int // edges[u] = indices into `all` of edges leaving u
	all   []flowEdge
}

func newFlowGraph(n int) *flowGraph {
	return &flowGraph{edges: make([][]int, n)}
}

// addEdge adds a directed edge u->v with the given capacity, plus its
// zero-capacity reverse edge (standard residual-graph setup). Edges
// are always added in (forward, reverse) pairs, so for any edge index
// ei, ei^1 is always its paired reverse edge.
func (g *flowGraph) addEdge(u, v, cap int) {
	g.edges[u] = append(g.edges[u], len(g.all))
	g.all = append(g.all, flowEdge{to: v, cap: cap})
	g.edges[v] = append(g.edges[v], len(g.all))
	g.all = append(g.all, flowEdge{to: u, cap: 0})
}

// bfsAugmentingPath finds a shortest augmenting path from s to t using
// BFS over residual capacity (cap-flow > 0). It returns the edge
// indices along the path in order from s to t, or nil if t is
// unreachable.
func (g *flowGraph) bfsAugmentingPath(s, t int) []int {
	parentEdge := make([]int, len(g.edges))
	visited := make([]bool, len(g.edges))
	for i := range parentEdge {
		parentEdge[i] = -1
	}
	visited[s] = true
	queue := []int{s}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		if u == t {
			break
		}
		for _, ei := range g.edges[u] {
			e := g.all[ei]
			if !visited[e.to] && e.cap-e.flow > 0 {
				visited[e.to] = true
				parentEdge[e.to] = ei
				queue = append(queue, e.to)
			}
		}
	}
	if !visited[t] {
		return nil
	}
	var path []int
	cur := t
	for cur != s {
		ei := parentEdge[cur]
		path = append([]int{ei}, path...)
		cur = g.all[ei^1].to // the edge's origin node
	}
	return path
}

// maxFlow repeatedly finds augmenting paths and pushes flow along them
// until none remain, returning the total flow pushed from s to t.
func (g *flowGraph) maxFlow(s, t int) int {
	total := 0
	for {
		path := g.bfsAugmentingPath(s, t)
		if path == nil {
			break
		}
		bottleneck := 1 << 30
		for _, ei := range path {
			if room := g.all[ei].cap - g.all[ei].flow; room < bottleneck {
				bottleneck = room
			}
		}
		for _, ei := range path {
			g.all[ei].flow += bottleneck
			g.all[ei^1].flow -= bottleneck
		}
		total += bottleneck
	}
	return total
}
