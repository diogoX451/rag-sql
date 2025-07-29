package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Client struct {
	Model  string
	Host   string
	client *http.Client
}

func New(model, host string) *Client {
	return &Client{
		Model:  model,
		Host:   host,
		client: &http.Client{},
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	System string `json:"system"`
}

type generateResponse struct {
	Response string `json:"response"`
}

type Option func(*Client)

func (c *Client) GenerateSQL(prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.Host)

	reqBody, _ := json.Marshal(generateRequest{
		Model:  c.Model,
		Prompt: prompt,
		System: "Você é um assistente especialista em SQL. Gere apenas a query SQL correta para responder à pergunta. Não explique. Não use linguagem natural. Apenas retorne a query SQL pura.",
		Stream: false,
	})

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("erro ao chamar LLM: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res generateResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("erro ao parsear resposta do LLM: %w\n%s", err, string(body))
	}

	return res.Response, nil
}

func WithModel(model string) Option {
	return func(c *Client) {
		c.Model = model
	}
}

func (c *Client) Generate(ctx context.Context, prompt string, opts ...Option) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.Host)

	println("Enviando prompt para LLM:", url)

	reqBody, _ := json.Marshal(generateRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao chamar LLM: %w", err)
	}

	println("Status da resposta:", resp.StatusCode)
	println("Headers da resposta:", resp.Header.Values("Content-Type"))
	body, _ := io.ReadAll(resp.Body)
	println("Corpo da resposta:", string(body))
	defer resp.Body.Close()

	var res generateResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("erro ao parsear resposta do LLM: %w\n%s", err, string(body))
	}

	return res.Response, nil
}

func (c *Client) InferRelationship(ctx context.Context, fromEntity, toEntity string) (string, error) {
	prompt := fmt.Sprintf(`
		Aja como um Arquiteto de Dados especialista em modelagem de grafos para Neo4j.
		Sua tarefa é definir um tipo de relação semântica entre duas entidades de um banco de dados.
		A relação deve ser um verbo ou uma frase verbal curta em maiúsculas, usando o padrão SNAKE_CASE.
		Analise a ação que a entidade de origem ('%s') exerce sobre a entidade de destino ('%s').

		Exemplos Estratégicos:
		1. Origem: 'checklists', Destino: 'farms' -> <relationship>AVALIOU</relationship>
		2. Origem: 'users', Destino: 'addresses' -> <relationship>RESIDE_EM</relationship>
		3. Origem: 'orders', Destino: 'customers' -> <relationship>FEITO_POR</relationship>
		4. Origem: 'order_items', Destino: 'orders' -> <relationship>PERTENCE_A</relationship>
		5. Origem: 'employees', Destino: 'departments' -> <relationship>TRABALHA_EM</relationship>
		6. Origem: 'farm_files', Destino: 'farms' -> <relationship>ANEXADO_A</relationship>

		Agora, analise o seguinte caso:
		Entidade de Origem: '%s'
		Entidade de Destino: '%s'

		Forneça sua resposta final, e APENAS a resposta, dentro de tags XML <relationship></relationship>. Não adicione nenhuma outra explicação ou texto.
`, fromEntity, toEntity, fromEntity, toEntity)

	responseText, err := c.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	println("Resposta da LLM:", responseText)

	return parseXMLRelationship(responseText)
}

func parseXMLRelationship(xmlString string) (string, error) {
	r := regexp.MustCompile(`<relationship>(.*?)<\/relationship>`)
	matches := r.FindStringSubmatch(xmlString)

	if len(matches) < 2 {

		cleanedFallback := strings.TrimSpace(xmlString)
		cleanedFallback = strings.ToUpper(cleanedFallback)
		reg := regexp.MustCompile("[^A-Z_]")
		cleanedFallback = reg.ReplaceAllString(cleanedFallback, "")
		if cleanedFallback == "" {
			return "REFERENCES", fmt.Errorf("a resposta da LLM não continha a tag <relationship> e estava vazia: %s", xmlString)
		}
		return cleanedFallback, nil
	}

	relationship := strings.TrimSpace(matches[1])
	if relationship == "" {
		return "REFERENCES", fmt.Errorf("a tag <relationship> estava vazia na resposta da LLM")
	}
	return relationship, nil
}
