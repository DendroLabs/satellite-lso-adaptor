package mapping

import (
	"context"
	"fmt"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/satellite"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// Transformer translates between MEF LSO Sonata API models and Telesat service models.
type Transformer struct {
	telesat   *telesat.Client
	satellite *satellite.Client
}

func NewTransformer(tc *telesat.Client, sc *satellite.Client) *Transformer {
	return &Transformer{telesat: tc, satellite: sc}
}

// --- MEF LSO Sonata Product Catalog Models ---

// MEFProductOffering represents an MEF-standard product offering.
// Maps Telesat services to MEF Access E-Line / E-Access product schemas.
type MEFProductOffering struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Status            string            `json:"status"` // active, retired
	ProductSpecRef    ProductSpecRef    `json:"productSpecification"`
	ServiceLevelSpec  ServiceLevelSpec  `json:"serviceLevelSpecification"`
	Place             []GeographicArea  `json:"place,omitempty"`
}

type ProductSpecRef struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Name string `json:"name"` // AccessElineSpec, EAccessSpec
}

type ServiceLevelSpec struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	PerformanceTier   string  `json:"performanceTier"` // PT5 for satellite
	Availability      float64 `json:"availabilityPct"`
	MaxLatencyMs      float64 `json:"maxOneWayDelayMs"`
	MaxJitterMs       float64 `json:"maxDelayVariationMs"`
	MaxPacketLossPct  float64 `json:"maxFrameLossRatioPct"`
}

type GeographicArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // region, global
}

// --- MEF Product Inventory Models ---

// MEFProduct represents a provisioned service in MEF inventory format.
type MEFProduct struct {
	ID                string          `json:"id"`
	Href              string          `json:"href"`
	Status            string          `json:"status"` // active, suspended, terminated
	ProductOffering   ProductSpecRef  `json:"productOffering"`
	ProductSpec       ProductSpecRef  `json:"productSpecification"`
	StartDate         time.Time       `json:"startDate"`
	Site              []MEFSite       `json:"relatedSite,omitempty"`
	ProductPrice      []ProductPrice  `json:"productPrice,omitempty"`
}

type MEFSite struct {
	ID   string  `json:"id"`
	Role string  `json:"role"` // UNI-A, UNI-Z
	Lat  float64 `json:"latitude"`
	Lon  float64 `json:"longitude"`
}

type ProductPrice struct {
	PriceType string  `json:"priceType"` // recurring, nonRecurring
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
}

// --- Transformation Methods ---

// GetProductCatalog returns Telesat Lightspeed services expressed as MEF product offerings.
func (t *Transformer) GetProductCatalog(ctx context.Context) ([]MEFProductOffering, error) {
	return []MEFProductOffering{
		{
			ID:          "telesat-ls-eline",
			Name:        "Telesat Lightspeed Access E-Line",
			Description: "Point-to-point Carrier Ethernet over Telesat Lightspeed LEO constellation. MEF 3.0 certified, sub-100ms latency, 99.9% availability SLA.",
			Status:      "active",
			ProductSpecRef: ProductSpecRef{
				ID:   "mef-access-eline-106",
				Href: "/productSpecification/mef-access-eline-106",
				Name: "AccessElineSpec",
			},
			ServiceLevelSpec: ServiceLevelSpec{
				ID:               "telesat-pt5-standard",
				Name:             "Lightspeed PT5 Standard",
				PerformanceTier:  "PT5",
				Availability:     99.9,
				MaxLatencyMs:     50.0,
				MaxJitterMs:      10.0,
				MaxPacketLossPct: 0.1,
			},
			Place: []GeographicArea{
				{ID: "global", Name: "Global Coverage", Type: "global"},
			},
		},
		{
			ID:          "telesat-ls-eaccess",
			Name:        "Telesat Lightspeed E-Access",
			Description: "Satellite access segment providing Carrier Ethernet connectivity to remote locations via Telesat Lightspeed LEO constellation.",
			Status:      "active",
			ProductSpecRef: ProductSpecRef{
				ID:   "mef-eaccess-51",
				Href: "/productSpecification/mef-eaccess-51",
				Name: "EAccessSpec",
			},
			ServiceLevelSpec: ServiceLevelSpec{
				ID:               "telesat-pt5-remote",
				Name:             "Lightspeed PT5 Remote Access",
				PerformanceTier:  "PT5",
				Availability:     99.5,
				MaxLatencyMs:     60.0,
				MaxJitterMs:      15.0,
				MaxPacketLossPct: 0.2,
			},
			Place: []GeographicArea{
				{ID: "global", Name: "Global Coverage", Type: "global"},
			},
		},
		{
			ID:          "telesat-ls-vno-pool",
			Name:        "Telesat Lightspeed VNO Capacity Pool",
			Description: "Dedicated capacity allocation on Telesat Lightspeed with self-service management portal and APIs. Supports dynamic bandwidth allocation across multiple sites.",
			Status:      "active",
			ProductSpecRef: ProductSpecRef{
				ID:   "telesat-vno-pool-spec",
				Href: "/productSpecification/telesat-vno-pool-spec",
				Name: "VNOPoolSpec",
			},
			ServiceLevelSpec: ServiceLevelSpec{
				ID:               "telesat-pt5-vno",
				Name:             "Lightspeed PT5 VNO",
				PerformanceTier:  "PT5",
				Availability:     99.9,
				MaxLatencyMs:     50.0,
				MaxJitterMs:      10.0,
				MaxPacketLossPct: 0.1,
			},
			Place: []GeographicArea{
				{ID: "global", Name: "Global Coverage", Type: "global"},
			},
		},
	}, nil
}

// SubscriptionToMEFProduct converts a Telesat subscription to an MEF Product inventory item.
func (t *Transformer) SubscriptionToMEFProduct(sub telesat.Subscription) MEFProduct {
	specName := "AccessElineSpec"
	specID := "mef-access-eline-106"
	if sub.ServiceType == "e-access" {
		specName = "EAccessSpec"
		specID = "mef-eaccess-51"
	}

	return MEFProduct{
		ID:     fmt.Sprintf("mef-%s", sub.ID),
		Href:   fmt.Sprintf("/productInventory/v4/product/mef-%s", sub.ID),
		Status: mapTelesatStatusToMEF(sub.Status),
		ProductOffering: ProductSpecRef{
			ID:   fmt.Sprintf("telesat-ls-%s", sub.ServiceType),
			Href: fmt.Sprintf("/productOffering/telesat-ls-%s", sub.ServiceType),
			Name: fmt.Sprintf("Telesat Lightspeed %s", specName),
		},
		ProductSpec: ProductSpecRef{
			ID:   specID,
			Name: specName,
		},
		StartDate: sub.CreatedAt,
		Site: []MEFSite{
			{ID: sub.EndpointA.TerminalID, Role: "UNI-A", Lat: sub.EndpointA.Latitude, Lon: sub.EndpointA.Longitude},
			{ID: sub.EndpointZ.TerminalID, Role: "UNI-Z", Lat: sub.EndpointZ.Latitude, Lon: sub.EndpointZ.Longitude},
		},
	}
}

// GetInventory returns all active Telesat services as MEF Product inventory items.
func (t *Transformer) GetInventory(ctx context.Context) ([]MEFProduct, error) {
	subs, err := t.telesat.ListSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching subscriptions: %w", err)
	}

	products := make([]MEFProduct, 0, len(subs))
	for _, sub := range subs {
		products = append(products, t.SubscriptionToMEFProduct(sub))
	}
	return products, nil
}

// ListTerminals returns approved Telesat Lightspeed terminals.
func (t *Transformer) ListTerminals(ctx context.Context) ([]telesat.Terminal, error) {
	return t.telesat.ListTerminals(ctx)
}

// ListLandingStations returns known Telesat ground stations.
func (t *Transformer) ListLandingStations(ctx context.Context) ([]telesat.LandingStation, error) {
	return t.telesat.ListLandingStations(ctx)
}

func mapTelesatStatusToMEF(status string) string {
	switch status {
	case "active":
		return "active"
	case "provisioning":
		return "pendingActive"
	case "suspended":
		return "suspended"
	case "terminated":
		return "terminated"
	default:
		return "unknown"
	}
}
