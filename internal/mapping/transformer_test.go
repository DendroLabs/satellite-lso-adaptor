package mapping

import (
	"context"
	"testing"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/satellite"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

func newTestTransformer() *Transformer {
	tc := telesat.NewClient("http://mock", "")
	sc := satellite.NewClient("http://mock:8090")
	return NewTransformer(tc, sc)
}

func TestGetProductCatalog(t *testing.T) {
	tr := newTestTransformer()
	catalog, err := tr.GetProductCatalog(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(catalog) != 3 {
		t.Fatalf("expected 3 offerings, got %d", len(catalog))
	}

	// Verify E-Line offering
	eline := catalog[0]
	if eline.ID != "telesat-ls-eline" {
		t.Errorf("expected ID telesat-ls-eline, got %s", eline.ID)
	}
	if eline.ServiceLevelSpec.PerformanceTier != "PT5" {
		t.Errorf("expected PT5 performance tier, got %s", eline.ServiceLevelSpec.PerformanceTier)
	}
	if eline.Status != "active" {
		t.Errorf("expected active status, got %s", eline.Status)
	}

	// Verify all offerings have PT5
	for _, offering := range catalog {
		if offering.ServiceLevelSpec.PerformanceTier != "PT5" {
			t.Errorf("offering %s should have PT5, got %s", offering.ID, offering.ServiceLevelSpec.PerformanceTier)
		}
	}
}

func TestSubscriptionToMEFProduct(t *testing.T) {
	tr := newTestTransformer()
	sub := telesat.Subscription{
		ID:          "sub-test",
		Status:      "active",
		ServiceType: "e-line",
		BandwidthMbps: 500,
		EndpointA: telesat.ServicePoint{
			Latitude: 45.42, Longitude: -75.70,
			TerminalID: "term-a",
		},
		EndpointZ: telesat.ServicePoint{
			Latitude: 48.86, Longitude: 2.35,
			TerminalID: "term-z",
		},
		CreatedAt: time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	product := tr.SubscriptionToMEFProduct(sub)

	if product.ID != "mef-sub-test" {
		t.Errorf("expected ID mef-sub-test, got %s", product.ID)
	}
	if product.Status != "active" {
		t.Errorf("expected active status, got %s", product.Status)
	}
	if len(product.Site) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(product.Site))
	}
	if product.Site[0].Role != "UNI-A" {
		t.Errorf("expected UNI-A role, got %s", product.Site[0].Role)
	}
	if product.Site[1].Role != "UNI-Z" {
		t.Errorf("expected UNI-Z role, got %s", product.Site[1].Role)
	}
}

func TestSubscriptionToMEFProduct_StatusMapping(t *testing.T) {
	tr := newTestTransformer()
	tests := []struct {
		telesatStatus string
		mefStatus     string
	}{
		{"active", "active"},
		{"provisioning", "pendingActive"},
		{"suspended", "suspended"},
		{"terminated", "terminated"},
		{"unknown-status", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.telesatStatus, func(t *testing.T) {
			sub := telesat.Subscription{ID: "test", Status: tt.telesatStatus}
			product := tr.SubscriptionToMEFProduct(sub)
			if product.Status != tt.mefStatus {
				t.Errorf("status %s -> got %s, want %s", tt.telesatStatus, product.Status, tt.mefStatus)
			}
		})
	}
}

func TestGetInventory(t *testing.T) {
	tr := newTestTransformer()
	products, err := tr.GetInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product")
	}
	for _, p := range products {
		if p.ID == "" {
			t.Error("product ID should not be empty")
		}
		if p.Status == "" {
			t.Error("product status should not be empty")
		}
	}
}

func TestListTerminals(t *testing.T) {
	tr := newTestTransformer()
	terminals, err := tr.ListTerminals(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(terminals) != 4 {
		t.Fatalf("expected 4 terminals, got %d", len(terminals))
	}
	for _, term := range terminals {
		if term.MaxThroughput <= 0 {
			t.Errorf("terminal %s should have positive throughput", term.ID)
		}
	}
}

func TestListLandingStations(t *testing.T) {
	tr := newTestTransformer()
	stations, err := tr.ListLandingStations(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stations) != 4 {
		t.Fatalf("expected 4 landing stations, got %d", len(stations))
	}

	gatineauFound := false
	for _, s := range stations {
		if s.ID == "ls-gatineau" {
			gatineauFound = true
			if s.Status != "operational" {
				t.Errorf("Gatineau should be operational, got %s", s.Status)
			}
		}
	}
	if !gatineauFound {
		t.Error("Gatineau landing station not found")
	}
}
