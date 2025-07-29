package main

import (
	"context"
	"log"
	"net/http"

	"rag-sql/internal/api"
	"rag-sql/internal/config"
	"rag-sql/internal/db"
	"rag-sql/internal/db/contextbuilder"
	"rag-sql/internal/db/dbschema"
	"rag-sql/internal/db/exec"
	"rag-sql/internal/db/schemautil"
	"rag-sql/internal/graph"
	"rag-sql/internal/llm"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar configuração: %v", err)
	}

	dbConn := db.Connect(cfg.DB)

	neoGraph, err := graph.NewGraph(cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		log.Fatalf("erro ao conectar ao Neo4j: %v", err)
	}
	defer neoGraph.Close(context.Background())

	schemaService := dbschema.NewService(dbConn)
	llmClient := llm.New("natural-sql-q4-k-s", "http://localhost:11434")
	builder := contextbuilder.New(neoGraph)

	executor := exec.New(dbConn)

	schemaStr, err := schemaService.GetAllAsString()
	if err != nil {
		log.Fatalf("Erro ao obter schema como string: %v", err)
	}

	schemaGraph := schemautil.BuildSchemaGraph(schemaStr)

	clientLlm := llm.New("llama3.1:8b", "http://localhost:11434")

	err = neoGraph.LoadSchemaGraph(context.Background(), schemaGraph, clientLlm)
	if err != nil {
		log.Fatalf("Erro ao carregar schema no grafo: %v", err)
	}

	router := api.NewRouter(schemaService, builder, executor, llmClient)

	log.Println("🚀 API rodando em http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
