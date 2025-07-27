package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jGraph struct {
	Driver neo4j.DriverWithContext
}

func NewGraph(uri, username, password string) (*Neo4jGraph, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Neo4j: %w", err)
	}
	return &Neo4jGraph{Driver: driver}, nil
}

func (g *Neo4jGraph) Close(ctx context.Context) {
	_ = g.Driver.Close(ctx)
}
