package d_graph

import (
	"fmt"

	"crawler/persistence"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gorm.io/gorm"
)

type pageNode struct {
	id int64
}

/**
 * ID returns the Gonum node ID.
 *
 * @return unique node identifier
 */
func (n pageNode) ID() int64 {
	return n.id
}

/*
 * ComputePageRank calculates PageRank values for all pages in a dataset
 * using a Gonum directed graph and stores the results back into the database.
 *
 * @param db         database connection
 * @param dataset    dataset identifier
 * @param damping    damping factor, usually 0.85
 * @param iterations number of PageRank iterations
 * @return error if computation fails
 */
func ComputePageRank(db *gorm.DB, dataset string, damping float64, iterations int) error {
	if dataset == "" {
		return fmt.Errorf("dataset cannot be empty")
	}

	if damping <= 0 || damping >= 1 {
		return fmt.Errorf("damping factor must be between 0 and 1")
	}

	if iterations <= 0 {
		return fmt.Errorf("iterations must be greater than 0")
	}

	pages, err := loadPagesByDataset(db, dataset)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages found for dataset %s", dataset)
	}

	pageSet := make(map[uint]bool, len(pages))
	for _, p := range pages {
		pageSet[p.ID] = true
	}

	links, err := loadLinks(db)
	if err != nil {
		return err
	}

	g := buildDirectedGraph(pages, links, pageSet)
	ranks := runPageRank(g, damping, iterations)

	for pageID, rank := range ranks {
		if err := persistence.UpdatePageRank(db, uint(pageID), rank); err != nil {
			return err
		}
	}

	return nil
}

/*
 * loadPagesByDataset retrieves all pages belonging to a dataset.
 *
 * @param db      database connection
 * @param dataset dataset identifier
 * @return slice of pages in the dataset
 */
func loadPagesByDataset(db *gorm.DB, dataset string) ([]persistence.Page, error) {
	var pages []persistence.Page

	result := db.Where("dataset = ?", dataset).Find(&pages)
	if result.Error != nil {
		return nil, result.Error
	}

	return pages, nil
}

/*
 * loadLinks retrieves all links from the database.
 *
 * @param db database connection
 * @return slice of stored links
 */
func loadLinks(db *gorm.DB) ([]persistence.Link, error) {
	var links []persistence.Link

	result := db.Find(&links)
	if result.Error != nil {
		return nil, result.Error
	}

	return links, nil
}

/*
 * buildDirectedGraph creates a Gonum directed graph for the given dataset pages and links.
 *
 * @param pages   pages in the dataset
 * @param links   all stored links
 * @param pageSet set of valid page IDs for the dataset
 * @return directed graph containing only dataset nodes and edges
 */
func buildDirectedGraph(
	pages []persistence.Page,
	links []persistence.Link,
	pageSet map[uint]bool,
) *simple.DirectedGraph {
	g := simple.NewDirectedGraph()

	for _, p := range pages {
		g.AddNode(pageNode{id: int64(p.ID)})
	}

	for _, l := range links {
		if !pageSet[l.FromPageID] || !pageSet[l.ToPageID] {
			continue
		}

		if l.FromPageID == l.ToPageID {
			continue
		}

		from := g.Node(int64(l.FromPageID))
		to := g.Node(int64(l.ToPageID))

		if from == nil || to == nil {
			continue
		}

		g.SetEdge(g.NewEdge(from, to))
	}

	return g
}

/*
 * runPageRank executes iterative PageRank over a Gonum directed graph.
 *
 * @param g          directed graph
 * @param damping    damping factor
 * @param iterations number of iterations
 * @return map of node ID to PageRank score
 */
func runPageRank(g *simple.DirectedGraph, damping float64, iterations int) map[int64]float64 {
	nodes := graph.NodesOf(g.Nodes())
	n := len(nodes)

	ranks := make(map[int64]float64, n)
	if n == 0 {
		return ranks
	}

	initial := 1.0 / float64(n)
	for _, node := range nodes {
		ranks[node.ID()] = initial
	}

	for range iterations {
		next := make(map[int64]float64, n)
		base := (1.0 - damping) / float64(n)

		for _, node := range nodes {
			next[node.ID()] = base
		}

		// leaf nodes: nodes with no outgoing edges
		leaf := 0.0
		for _, node := range nodes {
			if g.From(node.ID()).Len() == 0 {
				leaf += ranks[node.ID()]
			}
		}

		danglingContribution := damping * leaf / float64(n)
		for _, node := range nodes {
			next[node.ID()] += danglingContribution
		}

		// Add incoming-link
		for _, node := range nodes {
			incoming := graph.NodesOf(g.To(node.ID()))

			for _, src := range incoming {
				outDegree := g.From(src.ID()).Len()
				if outDegree == 0 {
					continue
				}

				next[node.ID()] += damping * (ranks[src.ID()] / float64(outDegree))
			}
		}

		ranks = next
	}

	return ranks
}