package telesat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the Telesat Lightspeed VNO / service APIs.
// In production this hits Telesat's real API; for development we use mock responses
// based on publicly documented service parameters.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListPools returns available VNO capacity pools.
func (c *Client) ListPools(ctx context.Context) ([]VNOPool, error) {
	// TODO: Replace with real Telesat API call when API access is available.
	// For now, return mock data based on publicly documented service parameters.
	return mockPools(), nil
}

// GetPool returns a specific VNO pool by ID.
func (c *Client) GetPool(ctx context.Context, id string) (*VNOPool, error) {
	pools := mockPools()
	for _, p := range pools {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("pool %s not found", id)
}

// ListSubscriptions returns active subscriptions.
func (c *Client) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	return mockSubscriptions(), nil
}

// GetSubscription returns a specific subscription by ID.
func (c *Client) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	subs := mockSubscriptions()
	for _, s := range subs {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("subscription %s not found", id)
}

// ListTerminals returns approved user terminals.
func (c *Client) ListTerminals(ctx context.Context) ([]Terminal, error) {
	return approvedTerminals(), nil
}

// ListLandingStations returns known Telesat ground stations.
func (c *Client) ListLandingStations(ctx context.Context) ([]LandingStation, error) {
	return knownLandingStations(), nil
}

// ListQuoteRequests returns pricing requests.
func (c *Client) ListQuoteRequests(ctx context.Context) ([]QuoteRequest, error) {
	return mockQuoteRequests(), nil
}

// GetQuoteRequest returns a specific quote request by ID.
func (c *Client) GetQuoteRequest(ctx context.Context, id string) (*QuoteRequest, error) {
	quotes := mockQuoteRequests()
	for _, q := range quotes {
		if q.ID == id {
			return &q, nil
		}
	}
	return nil, fmt.Errorf("quote request %s not found", id)
}

// CreateQuoteRequest submits a pricing request to Telesat.
func (c *Client) CreateQuoteRequest(ctx context.Context, req QuoteRequest) (*QuoteRequest, error) {
	// In production, this would POST to Telesat's pricing API.
	// For now, simulate an instant price response.
	req.Status = "priced"
	req.MonthlyPriceUSD = estimateMonthlyPrice(req.BandwidthMbps, req.TermMonths)
	req.SetupPriceUSD = estimateSetupPrice(req.BandwidthMbps)
	req.ValidUntil = time.Now().Add(90 * 24 * time.Hour)
	req.CreatedAt = time.Now()
	return &req, nil
}

// ListOrderRequests returns service provisioning orders.
func (c *Client) ListOrderRequests(ctx context.Context) ([]OrderRequest, error) {
	return mockOrderRequests(), nil
}

// GetOrderRequest returns a specific order by ID.
func (c *Client) GetOrderRequest(ctx context.Context, id string) (*OrderRequest, error) {
	orders := mockOrderRequests()
	for _, o := range orders {
		if o.ID == id {
			return &o, nil
		}
	}
	return nil, fmt.Errorf("order request %s not found", id)
}

// CreateOrderRequest submits a service provisioning order to Telesat.
func (c *Client) CreateOrderRequest(ctx context.Context, req OrderRequest) (*OrderRequest, error) {
	// In production, this would POST to Telesat's provisioning API.
	req.Status = "acknowledged"
	req.CreatedAt = time.Now()
	return &req, nil
}

// ListTroubleTickets returns trouble tickets.
func (c *Client) ListTroubleTickets(ctx context.Context) ([]TroubleTicket, error) {
	return mockTroubleTickets(), nil
}

// GetTroubleTicket returns a specific trouble ticket by ID.
func (c *Client) GetTroubleTicket(ctx context.Context, id string) (*TroubleTicket, error) {
	tickets := mockTroubleTickets()
	for _, t := range tickets {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("trouble ticket %s not found", id)
}

// CreateTroubleTicket submits a new trouble ticket.
func (c *Client) CreateTroubleTicket(ctx context.Context, ticket TroubleTicket) (*TroubleTicket, error) {
	ticket.Status = "acknowledged"
	ticket.CreatedAt = time.Now().UTC()
	ticket.UpdatedAt = ticket.CreatedAt
	return &ticket, nil
}

// UpdateTroubleTicket updates a trouble ticket's status or resolution.
func (c *Client) UpdateTroubleTicket(ctx context.Context, id string, status, resolutionNote string) (*TroubleTicket, error) {
	ticket, err := c.GetTroubleTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	ticket.Status = status
	ticket.UpdatedAt = now
	if resolutionNote != "" {
		ticket.ResolutionNote = resolutionNote
	}
	if status == "resolved" || status == "closed" {
		ticket.ResolvedAt = &now
	}
	return ticket, nil
}

// ListPerformanceReports returns performance reports, optionally filtered by subscription.
func (c *Client) ListPerformanceReports(ctx context.Context, subscriptionID string) ([]PerformanceReport, error) {
	reports := mockPerformanceReports()
	if subscriptionID == "" {
		return reports, nil
	}
	var filtered []PerformanceReport
	for _, r := range reports {
		if r.SubscriptionID == subscriptionID {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// GetPerformanceReport returns a specific performance report by ID.
func (c *Client) GetPerformanceReport(ctx context.Context, id string) (*PerformanceReport, error) {
	reports := mockPerformanceReports()
	for _, r := range reports {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("performance report %s not found", id)
}

// estimateMonthlyPrice generates a realistic price based on bandwidth and term.
// Based on publicly available LEO satellite service pricing benchmarks.
func estimateMonthlyPrice(bandwidthMbps, termMonths int) float64 {
	basePerMbps := 9.0 // $/Mbps/month base rate
	if termMonths >= 36 {
		basePerMbps = 7.0
	} else if termMonths >= 12 {
		basePerMbps = 8.0
	}
	return float64(bandwidthMbps) * basePerMbps
}

// estimateSetupPrice generates a realistic NRC based on bandwidth.
func estimateSetupPrice(bandwidthMbps int) float64 {
	base := 10000.0
	if bandwidthMbps > 500 {
		base = 15000.0
	}
	if bandwidthMbps > 2000 {
		base = 25000.0
	}
	return base
}

// doRequest executes an authenticated request against the Telesat API.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
	}
	return resp, nil
}

// decodeResponse reads a JSON response body into the target.
func decodeResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
