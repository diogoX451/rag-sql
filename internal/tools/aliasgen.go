package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"rag-sql/internal/llm"
	"strings"
)

type AliasResult struct {
	Table   string   `json:"table"`
	Aliases []string `json:"aliases"`
}

func GenerateAliasesFromLLM(ctx context.Context, model *llm.Client, tableName string, columnNames []string) ([]string, string, error) {

	context := `Você está atuando como um assistente que trabalha com o banco de dados da empresa **Produzindo Certo**, especializada em análises socioambientais e assistência técnica para produtores rurais no Brasil.

Sua tarefa é gerar sinônimos e expressões comuns que usuários da empresa poderiam utilizar para se referir a uma tabela específica. Use o nome da tabela e suas colunas para inferir o propósito da tabela e sugerir nomes alternativos coerentes com o contexto da empresa.

As variações devem incluir:
- Sinônimos diretos e traduções naturais (ex: "analyst" → "analista")
- Funções ou papéis que representem o uso da tabela (ex: "analyst" → "consultor socioambiental")
- Termos informais ou jargões internos (ex: "farms" → "propriedades", "unidades produtoras")
- Expressões compostas ou nomes descritivos usados por usuários (ex: "certificates" → "documentos de regularização")

Não repita apenas o nome da tabela com mínima variação (como plural ou underscore removido).

Responda exclusivamente em um JSON válido.
`

	prompt := fmt.Sprintf(`%s

Considere a tabela "%s" no banco de dados. As colunas dessa tabela são: %s.

Gere até 5 sinônimos ou expressões comuns em **português do Brasil** que usuários da empresa poderiam usar para se referir a essa tabela.

Retorne **exclusivamente** no seguinte formato JSON válido:

{
  "table": "%s",
  "aliases": ["...", "..."]
}
`, context, tableName, strings.Join(columnNames, ", "), tableName)

	resp, err := model.Generate(ctx, prompt)
	if err != nil {
		return nil, "", err
	}

	cleaned := strings.ReplaceAll(resp, "`", "")
	cleaned = strings.ReplaceAll(cleaned, "“", `"`)
	cleaned = strings.ReplaceAll(cleaned, "”", `"`)
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, cleaned, fmt.Errorf("nenhum JSON encontrado na resposta:\n%s", cleaned)
	}
	jsonStr := cleaned[start : end+1]

	var result AliasResult
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, jsonStr, fmt.Errorf("erro ao interpretar JSON: %w\nJSON bruto:\n%s", err, jsonStr)
	}

	if len(result.Aliases) > 5 {
		result.Aliases = result.Aliases[:5]
	}

	return result.Aliases, jsonStr, nil
}
