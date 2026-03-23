package mapping

import (
	"context"
	"testing"
)

func TestCreateQuote_InstantSync(t *testing.T) {
	tr := newTestTransformer()
	req := MEFQuoteCreate{
		BuyerRequestedQuoteLevel: "firm",
		InstantSyncQuote:         true,
		QuoteItem: []MEFQuoteItemCreate{
			{
				ID:     "1",
				Action: "add",
				Product: QuoteProduct{
					ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
					Place: []QuotePlace{
						{Role: "UNI-A", Latitude: 45.42, Longitude: -75.70},
						{Role: "UNI-Z", Latitude: 48.86, Longitude: 2.35},
					},
					ProductConfig: &ProductConfig{
						Type:          "AccessElineOvc",
						BandwidthMbps: 500,
					},
				},
				RequestedQuoteItemTerm: &MEFItemTerm{
					Name:     "36-month",
					Duration: Duration{Amount: 36, Units: "months"},
				},
			},
		},
	}

	quote, err := tr.CreateQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quote.State != "approved.orderable" {
		t.Errorf("expected state approved.orderable, got %s", quote.State)
	}
	if !quote.InstantSyncQuote {
		t.Error("expected InstantSyncQuote to be true")
	}
	if quote.ValidFor == nil {
		t.Error("expected ValidFor to be set for instant quote")
	}
	if len(quote.QuoteItem) != 1 {
		t.Fatalf("expected 1 quote item, got %d", len(quote.QuoteItem))
	}

	item := quote.QuoteItem[0]
	if item.State != "approved.orderable" {
		t.Errorf("expected item state approved.orderable, got %s", item.State)
	}
	if len(item.QuoteItemPrice) != 2 {
		t.Fatalf("expected 2 prices (MRC + NRC), got %d", len(item.QuoteItemPrice))
	}

	// Verify MRC
	mrc := item.QuoteItemPrice[0]
	if mrc.PriceType != "recurring" {
		t.Errorf("expected recurring price type, got %s", mrc.PriceType)
	}
	if mrc.Price.DutyFreeAmount.Value <= 0 {
		t.Error("MRC should be positive")
	}
	if mrc.Price.DutyFreeAmount.Unit != "USD" {
		t.Errorf("expected USD, got %s", mrc.Price.DutyFreeAmount.Unit)
	}

	// Verify NRC
	nrc := item.QuoteItemPrice[1]
	if nrc.PriceType != "nonRecurring" {
		t.Errorf("expected nonRecurring price type, got %s", nrc.PriceType)
	}
	if nrc.Price.DutyFreeAmount.Value <= 0 {
		t.Error("NRC should be positive")
	}
}

func TestCreateQuote_AsyncQuote(t *testing.T) {
	tr := newTestTransformer()
	req := MEFQuoteCreate{
		BuyerRequestedQuoteLevel: "budgetary",
		InstantSyncQuote:         false,
		QuoteItem: []MEFQuoteItemCreate{
			{
				ID:     "1",
				Action: "add",
				Product: QuoteProduct{
					ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eaccess"},
				},
			},
		},
	}

	quote, err := tr.CreateQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quote.State != "inProgress" {
		t.Errorf("expected state inProgress for async quote, got %s", quote.State)
	}
	if quote.ExpectedQuoteCompletionDate == nil {
		t.Error("expected ExpectedQuoteCompletionDate for async quote")
	}
}

func TestCreateQuote_InvalidAction(t *testing.T) {
	tr := newTestTransformer()
	req := MEFQuoteCreate{
		BuyerRequestedQuoteLevel: "firm",
		InstantSyncQuote:         true,
		QuoteItem: []MEFQuoteItemCreate{
			{
				ID:     "1",
				Action: "invalid",
				Product: QuoteProduct{
					ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
				},
			},
		},
	}

	quote, err := tr.CreateQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error (should return rejected quote, not error): %v", err)
	}
	if quote.State != "rejected" {
		t.Errorf("expected rejected state for invalid action, got %s", quote.State)
	}
}

func TestCreateQuote_PricingScales(t *testing.T) {
	tr := newTestTransformer()

	// 500 Mbps at 36 months should be cheaper per Mbps than 500 Mbps at 12 months
	quote36, _ := tr.CreateQuote(context.Background(), MEFQuoteCreate{
		BuyerRequestedQuoteLevel: "firm",
		InstantSyncQuote:         true,
		QuoteItem: []MEFQuoteItemCreate{{
			ID: "1", Action: "add",
			Product: QuoteProduct{
				ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
				ProductConfig:      &ProductConfig{BandwidthMbps: 500},
			},
			RequestedQuoteItemTerm: &MEFItemTerm{Duration: Duration{Amount: 36, Units: "months"}},
		}},
	})

	quote12, _ := tr.CreateQuote(context.Background(), MEFQuoteCreate{
		BuyerRequestedQuoteLevel: "firm",
		InstantSyncQuote:         true,
		QuoteItem: []MEFQuoteItemCreate{{
			ID: "1", Action: "add",
			Product: QuoteProduct{
				ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
				ProductConfig:      &ProductConfig{BandwidthMbps: 500},
			},
			RequestedQuoteItemTerm: &MEFItemTerm{Duration: Duration{Amount: 12, Units: "months"}},
		}},
	})

	mrc36 := quote36.QuoteItem[0].QuoteItemPrice[0].Price.DutyFreeAmount.Value
	mrc12 := quote12.QuoteItem[0].QuoteItemPrice[0].Price.DutyFreeAmount.Value

	if mrc36 >= mrc12 {
		t.Errorf("36-month MRC ($%.0f) should be less than 12-month ($%.0f)", mrc36, mrc12)
	}
}

func TestListQuotes(t *testing.T) {
	tr := newTestTransformer()
	quotes, err := tr.ListQuotes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quotes) == 0 {
		t.Fatal("expected at least one quote")
	}
	for _, q := range quotes {
		if q.ID == "" {
			t.Error("quote ID should not be empty")
		}
		if q.State == "" {
			t.Error("quote state should not be empty")
		}
	}
}

func TestGetQuote(t *testing.T) {
	tr := newTestTransformer()

	// Get existing quote
	quote, err := tr.GetQuote(context.Background(), "quote-tq-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.ID != "quote-tq-001" {
		t.Errorf("expected ID quote-tq-001, got %s", quote.ID)
	}
	if len(quote.QuoteItem) == 0 {
		t.Error("expected at least one quote item")
	}
	if len(quote.QuoteItem[0].QuoteItemPrice) != 2 {
		t.Errorf("expected 2 prices, got %d", len(quote.QuoteItem[0].QuoteItemPrice))
	}

	// Non-existent quote
	_, err = tr.GetQuote(context.Background(), "quote-nonexistent")
	if err == nil {
		t.Error("expected error for non-existent quote")
	}
}

func TestMapOfferingToServiceType(t *testing.T) {
	tests := []struct {
		offeringID string
		want       string
	}{
		{"telesat-ls-eline", "e-line"},
		{"telesat-ls-eaccess", "e-access"},
		{"telesat-ls-vno-pool", "vno-pool"},
		{"unknown", "e-line"},
	}

	for _, tt := range tests {
		t.Run(tt.offeringID, func(t *testing.T) {
			got := mapOfferingToServiceType(&ProductOfferingRef{ID: tt.offeringID})
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}

	// nil ref
	if got := mapOfferingToServiceType(nil); got != "e-line" {
		t.Errorf("nil ref should default to e-line, got %s", got)
	}
}
