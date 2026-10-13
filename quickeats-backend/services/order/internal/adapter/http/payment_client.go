package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quickeats/services/order/internal/port"
)

type PaymentHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPaymentClient(baseURL string) *PaymentHTTPClient {
	return &PaymentHTTPClient{baseURL: baseURL, httpClient: &http.Client{}}
}

func (c *PaymentHTTPClient) CreatePaymentIntent(ctx context.Context, req port.CreatePaymentIntentRequest) (*port.PaymentIntentResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/payments/intent", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("payment intent: %w", err) }
	defer resp.Body.Close()

	if resp.StatusCode != 201 { return nil, fmt.Errorf("payment intent failed: status %d", resp.StatusCode) }

	var result struct{ Data port.PaymentIntentResponse `json:"data"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return &result.Data, nil
}

func (c *PaymentHTTPClient) Refund(ctx context.Context, paymentID, orderID string, amount int64, reason string) error {
	body, _ := json.Marshal(map[string]interface{}{"orderId": orderID, "amount": amount, "reason": reason})
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/internal/payments/%s/refund", c.baseURL, paymentID), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return err }
	resp.Body.Close()
	return nil
}
