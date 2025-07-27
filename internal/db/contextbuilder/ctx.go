package contextbuilder

import (
	"context"
	"fmt"
	"rag-sql/internal/graph"
	"strings"
)

type Builder struct {
	graph *graph.Neo4jGraph
}

func New(g *graph.Neo4jGraph) *Builder {
	return &Builder{graph: g}
}

func (b *Builder) BuildPrompt(schema string, question string, logic []string, lastError string) string {
	var sb strings.Builder

	tablesToInclude := b.selectRelevantTables(schema, question)

	sb.WriteString("## ESQUEMA DO BANCO DE DADOS:\n")
	sb.WriteString(tablesToInclude)
	sb.WriteString("\n\n")

	if len(logic) > 0 {
		sb.WriteString("## LÓGICA DE NEGÓCIO:\n")
		for _, rule := range logic {
			sb.WriteString("- " + rule + "\n")
		}
		sb.WriteString("\n")
	}

	if lastError != "" {
		sb.WriteString("## ERRO NA CONSULTA ANTERIOR:\n")
		sb.WriteString(lastError + "\n\n")
		sb.WriteString("A consulta anterior falhou. Corrija o SQL levando isso em consideração.\n\n")
	}

	sb.WriteString("## PERGUNTA:\n")
	sb.WriteString(question)
	sb.WriteString("\nSQL:")

	return sb.String()
}

func (b *Builder) selectRelevantTables(schema, question string) string {
	tables := strings.Split(schema, "\n\n")
	var baseTables []string
	qLower := strings.ToLower(question)

	for _, table := range tables {
		tableLower := strings.ToLower(table)
		if strings.Contains(tableLower, "create table") {
			tableName := extractTableName(tableLower)
			if strings.Contains(qLower, tableName) || containsAnyColumn(qLower, tableLower) {
				baseTables = append(baseTables, tableName)
			}
		}
	}

	graphTables := b.findTablesByGraph(qLower)
	baseTables = append(baseTables, graphTables...)
	baseTables = uniqueStrings(baseTables)

	ctx := context.Background()
	expandedTables, err := b.expandTablesFromGraph(ctx, baseTables, 1)
	if err != nil {
		expandedTables = baseTables
	}

	var relevantDefs []string
	for _, table := range tables {
		tableName := extractTableName(strings.ToLower(table))
		if contains(expandedTables, tableName) {
			relevantDefs = append(relevantDefs, table)
		}
	}

	if len(relevantDefs) == 0 {
		return schema
	}
	return strings.Join(relevantDefs, "\n\n")
}

func (b *Builder) expandTablesFromGraph(ctx context.Context, baseTables []string, depth int) ([]string, error) {
	var expanded []string
	seen := map[string]bool{}

	for _, base := range baseTables {
		related, err := b.graph.FindRelatedEntities(ctx, base, depth)
		if err != nil {
			return nil, err
		}
		for _, r := range related {
			if !seen[r] {
				expanded = append(expanded, r)
				seen[r] = true
			}
		}
	}

	return expanded, nil
}

func extractTableName(tableDef string) string {
	start := strings.Index(tableDef, "create table") + len("create table")
	end := strings.Index(tableDef[start:], "(")
	if start < 0 || end < 0 {
		return ""
	}
	return strings.TrimSpace(tableDef[start : start+end])
}

func containsAnyColumn(q string, tableDef string) bool {
	colsStart := strings.Index(tableDef, "(")
	colsEnd := strings.LastIndex(tableDef, ")")
	if colsStart == -1 || colsEnd == -1 || colsEnd <= colsStart {
		return false
	}
	cols := tableDef[colsStart+1 : colsEnd]
	lines := strings.Split(cols, "\n")
	for _, line := range lines {
		col := strings.TrimSpace(strings.Split(line, " ")[0])
		if col != "" && strings.Contains(q, col) {
			return true
		}
	}
	return false
}

func (b *Builder) findTablesByGraph(question string) []string {
	ctx := context.Background()

	tables, err := b.graph.FindEntitiesByAlias(ctx, question)
	if err != nil {
		fmt.Printf("Erro ao buscar aliases no grafo: %v\n", err)
		return nil
	}

	println("Tabelas encontradas no grafo:", strings.Join(tables, ", "))

	if len(tables) == 0 {
		best := graph.FindBestFuzzyMatch(question)
		if best != "" {
			tables = append(tables, best)
		}
	}

	return tables
}

func uniqueStrings(input []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, s := range input {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	return result
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
