package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/quickeats/services/order/internal/port"
)

type PaymentHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPaymentClient(baseURL string) *PaymentHTTPClient {
	return &PaymentHTTPClient{baseURL: baseURL, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *PaymentHTTPClient) CreatePaymentIntent(ctx context.Context, req port.CreatePaymentIntentRequest) (*port.PaymentIntentResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/payments/intent", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("payment intent: %w", err) }
	defer resp.Body.Close()

	// A declined card comes back as 422 with the reason in error.details.reason.
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var result struct {
			Error struct {
				Details struct{ Reason string `json:"reason"` } `json:"details"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		reason := result.Error.Details.Reason
		if reason == "" { reason = port.DeclineOther }
		return nil, &port.PaymentDeclinedError{Reason: reason}
	}
	if resp.StatusCode != http.StatusCreated { return nil, fmt.Errorf("payment intent failed: status %d", resp.StatusCode) }

	var result struct{ Data port.PaymentIntentResponse `json:"data"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("payment intent: bad response: %w", err)
	}
	return &result.Data, nil
}

func (c *PaymentHTTPClient) RefundOrder(ctx context.Context, orderID, customerID string) error {
	body, _ := json.Marshal(map[string]string{"orderId": orderID, "customerId": customerID})
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/payments/refund", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return fmt.Errorf("payment refund: %w", err) }
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return fmt.Errorf("payment refund failed: status %d", resp.StatusCode) }
	return nil
}

func (c *PaymentHTTPClient) CheckCard(ctx context.Context, customerID string, amount int64) (port.CardCheck, error) {
	u := fmt.Sprintf("%s/internal/cards/%s/check?amount=%d", c.baseURL, url.PathEscape(customerID), amount)
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return port.CardCheck{}, fmt.Errorf("card check: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return port.CardCheck{}, fmt.Errorf("card check failed: status %d", resp.StatusCode) }

	var result struct{ Data port.CardCheck `json:"data"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return port.CardCheck{}, fmt.Errorf("card check: bad response: %w", err)
	}
	return result.Data, nil
}
