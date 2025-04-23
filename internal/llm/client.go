package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/tldr-llm-telegram-bot/internal/config"
	"google.golang.org/api/option"
)

type LLMClient interface {
	AnalyzePrompt(nrApp *newrelic.Application, prompt string) (response string, err error)
}

type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
	Language   string
}

func (o *OllamaClient) AnalyzePrompt(nrApp *newrelic.Application, prompt string) (response string, err error) {
	txn := nrApp.StartTransaction("OllamaClient.AnalyzePrompt")
	defer txn.End()

	txn.AddAttribute("model", o.ModelName)
	txn.AddAttribute("language", o.Language)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  o.ModelName,
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", o.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	segment := txn.StartSegment("OllamaAPI.Request")
	resp, err := o.HTTPClient.Do(req)
	segment.End()

	if err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("received non-200 response: %s", resp.Status)
		txn.NoticeError(err)
		return "", err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		err = fmt.Errorf("response field not found in the API response")
		txn.NoticeError(err)
		return "", err
	}

	txn.AddAttribute("response_length", len(response))
	return response, nil
}

type GeminiClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
	APIkey     string
	Language   string
}

func (g *GeminiClient) AnalyzePrompt(nrApp *newrelic.Application, prompt string) (response string, err error) {
	txn := nrApp.StartTransaction("GeminiClient.AnalyzePrompt")
	defer txn.End()

	txn.AddAttribute("model", g.ModelName)
	txn.AddAttribute("language", g.Language)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(g.APIkey))
	if err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(g.ModelName)

	model.SetTemperature(0.9)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(4096)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	segment := txn.StartSegment("GeminiAPI.GenerateContent")
	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	segment.End()

	if err != nil {
		txn.NoticeError(err)
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		err := fmt.Errorf("no response generated")
		txn.NoticeError(err)
		return "", err
	}

	textResponse, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		err := fmt.Errorf("unexpected response format")
		txn.NoticeError(err)
		return "", err
	}

	response = string(textResponse)
	txn.AddAttribute("response_length", len(response))
	return response, nil
}

func NewLLMClient(config *config.Config) (LLMClient, error) {
	provider := config.LLMProvider

	switch provider {
	case "ollama":
		return &OllamaClient{
			BaseURL:    config.OllamaApiUrl,
			HTTPClient: &http.Client{},
			ModelName:  config.OllamaModel,
			Language:   config.Language,
		}, nil
	case "gemini":
		return &GeminiClient{
			BaseURL:    config.GeminiApiUrl,
			HTTPClient: &http.Client{},
			ModelName:  config.GeminiModel,
			APIkey:     config.GeminiApiKey,
			Language:   config.Language,
		}, nil
	default:
		return nil, errors.New("unsupported LLM provider")
	}
}
