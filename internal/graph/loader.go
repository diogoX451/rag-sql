package graph

import (
	"context"
	"fmt"
	"rag-sql/internal/db/schemautil"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (g *Neo4jGraph) LoadEntityTypes(ctx context.Context, entities []EntityType) error {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Write)})
	defer session.Close(ctx)

	for _, et := range entities {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			// Cria o nó do tipo
			_, err := tx.Run(ctx,
				`MERGE (e:Entity {name: $name}) SET e.aliases = $aliases`,
				map[string]any{
					"name":    et.Name,
					"aliases": et.Aliases,
				})
			if err != nil {
				return nil, err
			}
			// Cria relações de compatibilidade
			for _, comp := range et.CompatibleWith {
				_, err := tx.Run(ctx,
					`MATCH (a:Entity {name: $a}), (b:Entity {name: $b})
					MERGE (a)-[:COMPATIBLE_WITH]->(b)`,
					map[string]any{"a": et.Name, "b": comp})
				if err != nil {
					return nil, err
				}
			}
			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("falha ao criar entidade %s: %w", et.Name, err)
		}
	}
	return nil
}

func (g *Neo4jGraph) LoadSchemaGraph(ctx context.Context, graphSchema *schemautil.SchemaGraph) error {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Write)})
	defer session.Close(ctx)

	for tableName, relation := range graphSchema.Relations {
		table := tableName
		fks := relation.ForeignKeys

		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			_, err := tx.Run(ctx, `MERGE (:Entity {name: $name})`, map[string]any{"name": table})
			if err != nil {
				return nil, err
			}

			for _, fk := range fks {
				_, err := tx.Run(ctx, `
					MATCH (a:Entity {name: $from}), (b:Entity {name: $to})
					MERGE (a)-[:REFERENCES]->(b)
				`, map[string]any{"from": table, "to": fk})
				if err != nil {
					return nil, fmt.Errorf("erro ao criar relação de %s para %s: %w", table, fk, err)
				}
			}
			return nil, nil
		})

		if err != nil {
			return fmt.Errorf("falha ao processar entidade %s: %w", tableName, err)
		}
	}

	return nil
}

func (g *Neo4jGraph) AddAliasesToEntity(ctx context.Context, tableName string, aliases []string) error {
	session := g.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessMode(neo4j.Write)})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (e:Entity {name: $name}) SET e.aliases = $aliases`,
			map[string]any{"name": tableName, "aliases": aliases})
		return nil, err
	})

	return err
}
