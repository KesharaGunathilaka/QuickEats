package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CatalogClient calls the catalog service's internal (service-to-service) API.
type CatalogClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCatalogClient(baseURL string) *CatalogClient {
	return &CatalogClient{baseURL: baseURL, httpClient: &http.Client{Timeout: 5 * time.Second}}
}

func (c *CatalogClient) CreateShop(ctx context.Context, orgID, ownerID, name, description string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"organisationId": orgID, "ownerId": ownerID, "name": name, "description": description,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/shops", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("catalog create shop: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("catalog create shop failed: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Data.ID == "" {
		return "", fmt.Errorf("catalog create shop: bad response")
	}
	return result.Data.ID, nil
}
