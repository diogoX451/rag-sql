package schemautil

import (
	"regexp"
	"strings"
)

type TableRelation struct {
	Table       string
	ForeignKeys []string
}

type SchemaGraph struct {
	Relations map[string]TableRelation
}

func BuildSchemaGraph(schema string) *SchemaGraph {
	graph := &SchemaGraph{Relations: make(map[string]TableRelation)}
	tableDefs := strings.Split(schema, "\n\n")

	reTableName := regexp.MustCompile(`(?i)create table (\w+)`)
	reForeignKey := regexp.MustCompile(`(?i)foreign key.*?references (\w+)`)

	for _, def := range tableDefs {
		lines := strings.Split(def, "\n")
		if len(lines) == 0 {
			continue
		}

		tableNameMatch := reTableName.FindStringSubmatch(lines[0])
		if len(tableNameMatch) < 2 {
			continue
		}
		tableName := tableNameMatch[1]
		foreignKeys := []string{}

		for _, line := range lines {
			if matches := reForeignKey.FindStringSubmatch(line); len(matches) == 2 {
				foreignKeys = append(foreignKeys, matches[1])
			}
		}

		graph.Relations[tableName] = TableRelation{
			Table:       tableName,
			ForeignKeys: foreignKeys,
		}
	}

	return graph
}
