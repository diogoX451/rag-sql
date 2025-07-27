package graph

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (g *Neo4jGraph) IsCompatible(ctx context.Context, a, b string) (bool, error) {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Read)})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (a:Entity {name: $a})-[:COMPATIBLE_WITH]->(b:Entity {name: $b}) RETURN a`,
			map[string]any{"a": a, "b": b})
		if err != nil {
			return false, err
		}
		if res.Next(ctx) {
			return true, nil
		}
		return false, nil
	})
	return result.(bool), err
}

func (g *Neo4jGraph) FindRelatedEntities(ctx context.Context, name string, depth int) ([]string, error) {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Read)})
	defer session.Close(ctx)

	query := `
		MATCH (start:Entity {name: $name})-[:REFERENCES*1..$depth]-(related:Entity)
		RETURN DISTINCT related.name AS name
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, query, map[string]any{
			"name":  name,
			"depth": depth,
		})
		if err != nil {
			return nil, err
		}

		var related []string
		for res.Next(ctx) {
			if val, ok := res.Record().Get("name"); ok {
				if str, ok := val.(string); ok {
					related = append(related, str)
				}
			}
		}

		return related, res.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

func (g *Neo4jGraph) FindEntitiesByAlias(ctx context.Context, question string) ([]string, error) {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Read)})
	defer session.Close(ctx)

	query := `
		MATCH (e:Entity)
		WHERE ANY(alias IN e.aliases WHERE toLower($q) CONTAINS toLower(alias))
		RETURN DISTINCT e.name AS name
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, query, map[string]any{"q": question})
		if err != nil {
			return nil, err
		}

		var matches []string
		for res.Next(ctx) {
			if val, ok := res.Record().Get("name"); ok {
				if nameStr, ok := val.(string); ok {
					matches = append(matches, nameStr)
				}
			}
		}
		return matches, res.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}
