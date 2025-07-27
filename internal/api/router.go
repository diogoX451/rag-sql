package api

import (
	"encoding/json"
	"log"
	"net/http"
	"rag-sql/internal/db/contextbuilder"
	"rag-sql/internal/db/dbschema"
	"rag-sql/internal/db/exec"
	"rag-sql/internal/llm"
	"regexp"
	"strings"
)

type RouterDeps struct {
	SchemaService *dbschema.Service
	Builder       *contextbuilder.Builder
	Executor      *exec.Executor
	LLM           *llm.Client
}

func NewRouter(schemaService *dbschema.Service, builder *contextbuilder.Builder, executor *exec.Executor, llmClient *llm.Client) http.Handler {
	mux := http.NewServeMux()
	deps := &RouterDeps{schemaService, builder, executor, llmClient}

	mux.HandleFunc("/api/ask", deps.handleAsk)
	mux.HandleFunc("/api/schema", deps.handleSchema)

	return mux
}

type askResponse struct {
	SQL  string      `json:"sql"`
	Data interface{} `json:"data"`
}

func (r *RouterDeps) handleAsk(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}

	schema, err := r.SchemaService.GetCreateTableStatements()
	if err != nil {
		http.Error(w, "erro ao extrair schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prompt := r.Builder.BuildPrompt(schema, q, nil, "")

	println("Prompt para LLM:", prompt)

	sql, err := r.LLM.GenerateSQL(prompt)
	if err != nil {
		http.Error(w, "erro ao gerar SQL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, execErr := r.Executor.Execute(sql)
	if execErr == nil {
		respondJSON(w, askResponse{SQL: sql, Data: data})
		return
	}

	log.Printf("Erro ao executar SQL: %v", execErr)
	suggestion := analyzeSQLError(execErr.Error())
	promptRetry := r.Builder.BuildPrompt(schema, q, nil, suggestion)

	sqlRetry, err := r.LLM.GenerateSQL(promptRetry)
	if err != nil {
		respondJSON(w, askResponse{SQL: sql, Data: "Erro ao gerar SQL na segunda tentativa: " + err.Error()}, http.StatusInternalServerError)
		return
	}

	dataRetry, execErr := r.Executor.Execute(sqlRetry)
	if execErr != nil {
		respondJSON(w, askResponse{SQL: sqlRetry, Data: "Erro ao executar SQL na segunda tentativa: " + execErr.Error()}, http.StatusInternalServerError)
		return
	}

	respondJSON(w, askResponse{SQL: sqlRetry, Data: dataRetry})
}

func (r *RouterDeps) handleSchema(w http.ResponseWriter, req *http.Request) {
	schema, err := r.SchemaService.GetCreateTableStatements()
	if err != nil {
		http.Error(w, "erro ao extrair schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(schema))
}

func analyzeSQLError(err string) string {
	err = strings.ToLower(err)

	if strings.Contains(err, "function sum(") && strings.Contains(err, "does not exist") {
		return "Erro: a função de agregação SUM foi usada em uma coluna do tipo texto. Use CAST(coluna AS NUMERIC) ou coluna::numeric."
	}

	if strings.Contains(err, "operator does not exist") {
		re := regexp.MustCompile(`operator does not exist: ([a-z0-9_]+) [^:]+ ([a-z0-9_]+)`)
		matches := re.FindStringSubmatch(err)
		if len(matches) == 3 {
			return "Erro: operador não existe entre tipos " + matches[1] + " e " + matches[2] + ". Considere fazer cast explícito usando ::tipo ou CAST()."
		}
		return "Erro de operador entre tipos. Considere fazer cast com ::tipo."
	}

	if strings.Contains(err, "invalid input syntax") && strings.Contains(err, "for type") {
		return "Erro de sintaxe de tipo. Pode ser necessário converter texto para número ou data usando ::tipo."
	}

	return err
}

func respondJSON(w http.ResponseWriter, payload interface{}, opt ...interface{}) {
	if len(opt) > 0 {
		if status, ok := opt[0].(int); ok {
			w.WriteHeader(status)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
