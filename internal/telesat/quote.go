package telesat

import "time"

// QuoteRequest represents a Telesat-side pricing request for a Lightspeed service.
type QuoteRequest struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"` // pending, priced, rejected, expired
	ServiceType     string       `json:"serviceType"`
	BandwidthMbps   int          `json:"bandwidthMbps"`
	TermMonths      int          `json:"termMonths"`
	EndpointA       ServicePoint `json:"endpointA"`
	EndpointZ       ServicePoint `json:"endpointZ"`
	MonthlyPriceUSD float64      `json:"monthlyPriceUsd"`
	SetupPriceUSD   float64      `json:"setupPriceUsd"`
	ValidUntil      time.Time    `json:"validUntil"`
	CreatedAt       time.Time    `json:"createdAt"`
}

// OrderRequest represents a Telesat-side service provisioning order.
type OrderRequest struct {
	ID             string       `json:"id"`
	Status         string       `json:"status"` // submitted, acknowledged, provisioning, active, failed, cancelled
	QuoteID        string       `json:"quoteId,omitempty"`
	ServiceType    string       `json:"serviceType"`
	BandwidthMbps  int          `json:"bandwidthMbps"`
	EndpointA      ServicePoint `json:"endpointA"`
	EndpointZ      ServicePoint `json:"endpointZ"`
	RequestedDate  time.Time    `json:"requestedDate"`
	CompletionDate *time.Time   `json:"completionDate,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
}
