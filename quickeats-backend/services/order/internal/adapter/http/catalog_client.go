package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quickeats/services/order/internal/port"
)

type CatalogHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCatalogClient(baseURL string) *CatalogHTTPClient {
	return &CatalogHTTPClient{baseURL: baseURL, httpClient: &http.Client{}}
}

func (c *CatalogHTTPClient) ReserveStock(ctx context.Context, req port.ReserveStockRequest) (*port.ReserveStockResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/stock/reserve", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("catalog reserve: %w", err) }
	defer resp.Body.Close()

	if resp.StatusCode != 200 { return nil, fmt.Errorf("catalog reserve failed: status %d", resp.StatusCode) }

	var result struct{ Data port.ReserveStockResponse `json:"data"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return &result.Data, nil
}

func (c *CatalogHTTPClient) ReleaseStock(ctx context.Context, orderID string, reservationIDs []string) error {
	body, _ := json.Marshal(map[string]interface{}{"orderId": orderID, "reservationIds": reservationIDs})
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/stock/release", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return err }
	resp.Body.Close()
	return nil
}
