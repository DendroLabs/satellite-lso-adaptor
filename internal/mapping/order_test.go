package mapping

import (
	"context"
	"testing"
	"time"
)

func TestCreateOrder(t *testing.T) {
	tr := newTestTransformer()
	req := MEFProductOrderCreate{
		ProductOrderItem: []MEFProductOrderItemCreate{
			{
				ID:     "1",
				Action: "add",
				Product: OrderProduct{
					ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
					Place: []QuotePlace{
						{Role: "UNI-A", Latitude: 45.42, Longitude: -75.70},
						{Role: "UNI-Z", Latitude: 48.86, Longitude: 2.35},
					},
					ProductConfig: &ProductConfig{BandwidthMbps: 500},
				},
				RequestedCompletionDate: time.Date(2028, 5, 1, 0, 0, 0, 0, time.UTC),
				QuoteItem: &QuoteItemRef{QuoteID: "quote-tq-001", ID: "1"},
			},
		},
	}

	order, err := tr.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.State != "inProgress" {
		t.Errorf("expected state inProgress, got %s", order.State)
	}
	if len(order.ProductOrderItem) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.ProductOrderItem))
	}

	item := order.ProductOrderItem[0]
	if item.State != "acknowledged" {
		t.Errorf("expected item state acknowledged, got %s", item.State)
	}
	if item.ExpectedCompletionDate == nil {
		t.Error("expected ExpectedCompletionDate to be set")
	}
	if item.QuoteItem == nil || item.QuoteItem.QuoteID != "quote-tq-001" {
		t.Error("expected quote reference to be preserved")
	}
}

func TestCreateOrder_InvalidAction(t *testing.T) {
	tr := newTestTransformer()
	req := MEFProductOrderCreate{
		ProductOrderItem: []MEFProductOrderItemCreate{
			{
				ID:     "1",
				Action: "invalid",
				Product: OrderProduct{
					ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"},
				},
			},
		},
	}

	order, err := tr.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error (should return rejected order): %v", err)
	}
	if order.State != "rejected" {
		t.Errorf("expected rejected state, got %s", order.State)
	}
}

func TestCreateOrder_MultipleItems(t *testing.T) {
	tr := newTestTransformer()
	req := MEFProductOrderCreate{
		ProductOrderItem: []MEFProductOrderItemCreate{
			{
				ID: "1", Action: "add",
				Product: OrderProduct{ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eline"}},
				RequestedCompletionDate: time.Now().Add(30 * 24 * time.Hour),
			},
			{
				ID: "2", Action: "add",
				Product: OrderProduct{ProductOfferingRef: &ProductOfferingRef{ID: "telesat-ls-eaccess"}},
				RequestedCompletionDate: time.Now().Add(30 * 24 * time.Hour),
			},
		},
	}

	order, err := tr.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order.ProductOrderItem) != 2 {
		t.Fatalf("expected 2 order items, got %d", len(order.ProductOrderItem))
	}
}

func TestListOrders(t *testing.T) {
	tr := newTestTransformer()
	orders, err := tr.ListOrders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) == 0 {
		t.Fatal("expected at least one order")
	}
	for _, o := range orders {
		if o.ID == "" {
			t.Error("order ID should not be empty")
		}
	}
}

func TestGetOrder(t *testing.T) {
	tr := newTestTransformer()

	order, err := tr.GetOrder(context.Background(), "order-to-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.State != "completed" {
		t.Errorf("expected completed state for active order, got %s", order.State)
	}

	// Non-existent
	_, err = tr.GetOrder(context.Background(), "order-nonexistent")
	if err == nil {
		t.Error("expected error for non-existent order")
	}
}

func TestOrderStateMapping(t *testing.T) {
	tests := []struct {
		telesat string
		mef     string
	}{
		{"submitted", "acknowledged"},
		{"acknowledged", "acknowledged"},
		{"provisioning", "inProgress"},
		{"active", "completed"},
		{"failed", "failed"},
		{"cancelled", "cancelled"},
		{"unknown", "acknowledged"},
	}

	for _, tt := range tests {
		t.Run(tt.telesat, func(t *testing.T) {
			got := mapTelesatOrderStateToMEF(tt.telesat)
			if got != tt.mef {
				t.Errorf("got %s, want %s", got, tt.mef)
			}
		})
	}
}
