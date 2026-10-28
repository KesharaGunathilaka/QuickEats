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

func (c *CatalogClient) CreateShop(ctx context.Context, orgIDs []string, ownerID, name, description string) (string, error) {
	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	err := c.do(ctx, "POST", "/internal/shops", map[string]interface{}{
		"organisationIds": orgIDs, "ownerId": ownerID, "name": name, "description": description,
	}, http.StatusCreated, &result)
	if err != nil {
		return "", fmt.Errorf("catalog create shop: %w", err)
	}
	if result.Data.ID == "" {
		return "", fmt.Errorf("catalog create shop: bad response")
	}
	return result.Data.ID, nil
}

func (c *CatalogClient) ShopOrganisations(ctx context.Context, shopID string) ([]string, error) {
	var result struct {
		Data struct {
			OrganisationIDs []string `json:"organisationIds"`
		} `json:"data"`
	}
	if err := c.do(ctx, "GET", "/internal/shops/"+shopID+"/organisations", nil, http.StatusOK, &result); err != nil {
		return nil, fmt.Errorf("catalog shop organisations: %w", err)
	}
	return result.Data.OrganisationIDs, nil
}

func (c *CatalogClient) SetShopOrganisations(ctx context.Context, shopID string, orgIDs []string) error {
	err := c.do(ctx, "PUT", "/internal/shops/"+shopID+"/organisations",
		map[string]interface{}{"organisationIds": orgIDs}, http.StatusOK, nil)
	if err != nil {
		return fmt.Errorf("catalog set shop organisations: %w", err)
	}
	return nil
}

func (c *CatalogClient) do(ctx context.Context, method, path string, body interface{}, wantStatus int, out interface{}) error {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("bad response: %w", err)
		}
	}
	return nil
}
