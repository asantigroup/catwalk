// Package main provides a command-line tool to generate the Asanti
// provider configuration file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
		APIEndpoint:         "https://agent.asanti.dev/",
		Type:                catwalk.TypeOpenAICompat,
		DefaultLargeModelID: "asanti-coder",
		DefaultSmallModelID: "asanti-flash",
	}

	// The model catalog is discovered by clients at runtime, so the
	// generated config carries no static model list. Fetch the gateway
	// catalog anyway to verify the default models are still served.
	modelsResp, err := fetchAsantiModels(asantiProvider.APIEndpoint, apiKey)
	if err != nil {
		log.Fatal("Error fetching Asanti models:", err)
	}

	available := make(map[string]bool, len(modelsResp.Data))
	for _, model := range modelsResp.Data {
		available[model.ID] = true
	}
	for _, id := range []string{asantiProvider.DefaultLargeModelID, asantiProvider.DefaultSmallModelID} {
		if !available[id] {
			log.Fatalf("Default model %q is not served by the Asanti gateway", id)
		}
	}

	data, err := json.MarshalIndent(asantiProvider, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling Asanti provider:", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile("internal/providers/configs/asanti.json", data, 0o600); err != nil {
		log.Fatal("Error writing Asanti provider config:", err)
	}

	fmt.Println("Generated asanti.json (catalog auto-discovered by clients)")
}
