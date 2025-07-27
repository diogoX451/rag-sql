package main

import (
	"context"
	"fmt"
	"log"
	"rag-sql/internal/config"
	"rag-sql/internal/db"
	"rag-sql/internal/db/dbschema"
	"rag-sql/internal/graph"
	"rag-sql/internal/llm"
	"rag-sql/internal/tools"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar config:", err)
	}

	dbConn := db.Connect(cfg.DB)
	schemaService := dbschema.NewService(dbConn)

	tables := schemaService.ExtractTableNames()

	llmClient := llm.New("llama3", "http://localhost:11434")

	neo, err := graph.NewGraph(cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		log.Fatal("Erro ao conectar no Neo4j:", err)
	}
	defer neo.Close(ctx)

	for _, table := range tables {
		fmt.Printf("\n🔍 Gerando aliases para: %s\n", table)

		collumns, _ := schemaService.GetColumnsTable(table)
		aliases, raw, err := tools.GenerateAliasesFromLLM(ctx, llmClient, table, collumns)
		if err != nil {
			fmt.Printf("❌ Erro para %s: %v\n", table, err)
			fmt.Println("📤 Resposta bruta:")
			fmt.Println(raw)
			continue
		}

		fmt.Println("📄 Definição gerada:")
		fmt.Println(raw)

		err = neo.AddAliasesToEntity(ctx, table, aliases)
		if err != nil {
			fmt.Printf("❌ Erro ao salvar %s no grafo: %v\n", table, err)
			continue
		}

		fmt.Printf("✅ %s => %v\n", table, aliases)
	}

}

func ExtractTableNames(schema string) []string {
	re := regexp.MustCompile(`(?i)create table "?(\w+)"?`)
	matches := re.FindAllStringSubmatch(schema, -1)

	var names []string
	seen := map[string]bool{}
	for _, m := range matches {
		name := strings.ToLower(m[1])
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}
