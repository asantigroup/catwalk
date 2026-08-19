// Package main provides a command-line tool to fetch models from the Asanti
// LiteLLM gateway and generate a configuration file for the provider.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/catwalk/pkg/catwalk"
)

// Model represents a model from the OpenAI-compatible models API.
type Model struct {
	ID string `json:"id"`
}

// ModelsResponse is the response structure for the models API.
type ModelsResponse struct {
	Data []Model `json:"data"`
}

func fetchAsantiModels(apiEndpoint, apiKey string) (*ModelsResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiEndpoint+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "Crush-Client/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading models response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var mr ModelsResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return nil, fmt.Errorf("decoding models response: %w", err)
	}

	return &mr, nil
}

func modelDisplayName(id string) string {
	words := strings.Fields(strings.ReplaceAll(id, "-", " "))
	for i, w := range words {
		upper := strings.ToUpper(w)
		switch upper {
		case "GLM", "GPT", "AI", "V2", "V2.5", "OMNI", "K3", "HY3", "PRO":
			words[i] = upper
		default:
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// This is used to generate the asanti.json config file.
func main() {
	apiKey := os.Getenv("ASANTI_API_KEY")
	if apiKey == "" {
		log.Fatal("ASANTI_API_KEY environment variable is required")
	}

	asantiProvider := catwalk.Provider{
		Name:                "Asanti",
		ID:                  "asanti",
		APIKey:              "$ASANTI_API_KEY",
		APIEndpoint:         "https://llm.asanti.dev/v1",
		Type:                catwalk.TypeOpenAICompat,
		DefaultLargeModelID: "zen-deepseek-v4-pro",
		DefaultSmallModelID: "zen-deepseek-v4-flash",
	}

	modelsResp, err := fetchAsantiModels(asantiProvider.APIEndpoint, apiKey)
	if err != nil {
		log.Fatal("Error fetching Asanti models:", err)
	}

	for _, model := range modelsResp.Data {
		// Skip ids with a provider prefix separator, they duplicate aliases.
		if strings.Contains(model.ID, "/") {
			continue
		}

		var defaultMaxTokens int64 = 32768
		canReason := true
		if strings.Contains(model.ID, "flash") || strings.Contains(model.ID, "-air") {
			defaultMaxTokens = 8192
			canReason = false
		}

		m := catwalk.Model{
			ID:               model.ID,
			Name:             modelDisplayName(model.ID),
			ContextWindow:    262144,
			DefaultMaxTokens: defaultMaxTokens,
			CanReason:        canReason,
			SupportsImages:   false,
		}

		asantiProvider.Models = append(asantiProvider.Models, m)
	}

	slices.SortFunc(asantiProvider.Models, func(a catwalk.Model, b catwalk.Model) int {
		return strings.Compare(a.ID, b.ID)
	})

	data, err := json.MarshalIndent(asantiProvider, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling Asanti provider:", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile("internal/providers/configs/asanti.json", data, 0o600); err != nil {
		log.Fatal("Error writing Asanti provider config:", err)
	}

	fmt.Printf("Generated asanti.json with %d models\n", len(asantiProvider.Models))
}
