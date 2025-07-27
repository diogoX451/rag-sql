package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	reqBody, _ := json.Marshal(generateRequest{
		Model:  c.Model,
		Prompt: prompt,
		System: "Você é um assistente especialista em auxiliar nas perguntas sobre contexto de banco de dados.",
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
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res generateResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("erro ao parsear resposta do LLM: %w\n%s", err, string(body))
	}

	return res.Response, nil
}
