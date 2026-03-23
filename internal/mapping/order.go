package mapping

import (
	"context"
	"fmt"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// MEF LSO Sonata Product Order Management models (MEF 123.1 / Mplify 123.1)

// MEFProductOrder represents a complete product order in MEF LSO Sonata format.
type MEFProductOrder struct {
	ID                        string                `json:"id"`
	Href                      string                `json:"href"`
	State                     string                `json:"state"`
	OrderDate                 time.Time             `json:"orderDate"`
	ExternalID                string                `json:"externalId,omitempty"`
	ProjectID                 string                `json:"projectId,omitempty"`
	CompletionDate            *time.Time            `json:"completionDate,omitempty"`
	CancellationDate          *time.Time            `json:"cancellationDate,omitempty"`
	CancellationReason        string                `json:"cancellationReason,omitempty"`
	ProductOrderItem          []MEFProductOrderItem `json:"productOrderItem"`
	StateChange               []StateChange         `json:"stateChange"`
	RelatedContactInformation []ContactInfo         `json:"relatedContactInformation,omitempty"`
	Note                      []Note                `json:"note,omitempty"`
}

// MEFProductOrderFind is the abbreviated order returned in list responses.
type MEFProductOrderFind struct {
	ID             string     `json:"id"`
	State          string     `json:"state"`
	OrderDate      time.Time  `json:"orderDate"`
	ExternalID     string     `json:"externalId,omitempty"`
	ProjectID      string     `json:"projectId,omitempty"`
	CompletionDate *time.Time `json:"completionDate,omitempty"`
}

// MEFProductOrderCreate is the request body for creating a product order.
type MEFProductOrderCreate struct {
	ExternalID                string                      `json:"externalId,omitempty"`
	ProjectID                 string                      `json:"projectId,omitempty"`
	ProductOrderItem          []MEFProductOrderItemCreate `json:"productOrderItem"`
	RelatedContactInformation []ContactInfo               `json:"relatedContactInformation,omitempty"`
	Note                      []Note                      `json:"note,omitempty"`
}

// MEFProductOrderItem represents an item within a product order.
type MEFProductOrderItem struct {
	ID                               string           `json:"id"`
	State                            string           `json:"state"`
	Action                           string           `json:"action"` // add, modify, delete
	Product                          OrderProduct     `json:"product"`
	RequestedCompletionDate          time.Time        `json:"requestedCompletionDate"`
	ExpectedCompletionDate           *time.Time       `json:"expectedCompletionDate,omitempty"`
	CompletionDate                   *time.Time       `json:"completionDate,omitempty"`
	QuoteItem                        *QuoteItemRef    `json:"quoteItem,omitempty"`
	ProductOfferingQualificationItem *POQItemRef      `json:"productOfferingQualificationItem,omitempty"`
	ItemTerm                         *MEFItemTerm     `json:"itemTerm,omitempty"`
	StateChange                      []StateChange    `json:"stateChange,omitempty"`
	Note                             []Note           `json:"note,omitempty"`
}

// MEFProductOrderItemCreate is the request body for an order item.
type MEFProductOrderItemCreate struct {
	ID                               string        `json:"id"`
	Action                           string        `json:"action"` // add, modify, delete
	Product                          OrderProduct  `json:"product"`
	RequestedCompletionDate          time.Time     `json:"requestedCompletionDate"`
	QuoteItem                        *QuoteItemRef `json:"quoteItem,omitempty"`
	ProductOfferingQualificationItem *POQItemRef   `json:"productOfferingQualificationItem,omitempty"`
	RequestedItemTerm                *MEFItemTerm  `json:"requestedItemTerm,omitempty"`
	EndCustomerName                  string        `json:"endCustomerName,omitempty"`
	ExpediteIndicator                bool          `json:"expediteIndicator,omitempty"`
	Note                             []Note        `json:"note,omitempty"`
}

// OrderProduct identifies the product being ordered.
type OrderProduct struct {
	ID                 string              `json:"id,omitempty"` // existing product ID for modify/delete
	ProductOfferingRef *ProductOfferingRef `json:"productOffering,omitempty"`
	ProductConfig      *ProductConfig      `json:"productConfiguration,omitempty"`
	Place              []QuotePlace        `json:"place,omitempty"`
}

// QuoteItemRef references a prior quote item.
type QuoteItemRef struct {
	QuoteID   string `json:"quoteId"`
	ID        string `json:"id"`
	QuoteHref string `json:"quoteHref,omitempty"`
}

// --- Order Transformation Methods ---

// CreateOrder processes an MEF Product Order and submits to Telesat for provisioning.
func (t *Transformer) CreateOrder(ctx context.Context, req MEFProductOrderCreate) (*MEFProductOrder, error) {
	now := time.Now().UTC()
	orderID := fmt.Sprintf("order-%d", now.UnixMilli())

	order := &MEFProductOrder{
		ID:                        orderID,
		Href:                      fmt.Sprintf("/mef/v4/productOrder/%s", orderID),
		OrderDate:                 now,
		ExternalID:                req.ExternalID,
		ProjectID:                 req.ProjectID,
		RelatedContactInformation: req.RelatedContactInformation,
		Note:                      req.Note,
	}

	allAcknowledged := true
	for _, itemReq := range req.ProductOrderItem {
		item, err := t.processOrderItem(ctx, itemReq)
		if err != nil {
			allAcknowledged = false
			item.State = "rejected"
			item.StateChange = []StateChange{
				{State: "acknowledged", ChangeDate: now},
				{State: "rejected", ChangeDate: now, ChangeReason: err.Error()},
			}
		}
		order.ProductOrderItem = append(order.ProductOrderItem, item)
	}

	if !allAcknowledged {
		order.State = "rejected"
		order.StateChange = []StateChange{
			{State: "acknowledged", ChangeDate: now},
			{State: "rejected", ChangeDate: now, ChangeReason: "one or more order items failed validation"},
		}
	} else {
		order.State = "inProgress"
		order.StateChange = []StateChange{
			{State: "acknowledged", ChangeDate: now},
			{State: "inProgress", ChangeDate: now},
		}
	}

	return order, nil
}

// processOrderItem submits a single order item to Telesat for provisioning.
func (t *Transformer) processOrderItem(ctx context.Context, req MEFProductOrderItemCreate) (MEFProductOrderItem, error) {
	now := time.Now().UTC()

	item := MEFProductOrderItem{
		ID:                               req.ID,
		Action:                           req.Action,
		Product:                          req.Product,
		RequestedCompletionDate:          req.RequestedCompletionDate,
		QuoteItem:                        req.QuoteItem,
		ProductOfferingQualificationItem: req.ProductOfferingQualificationItem,
		ItemTerm:                         req.RequestedItemTerm,
		Note:                             req.Note,
	}

	if req.Action != "add" && req.Action != "modify" && req.Action != "delete" {
		return item, fmt.Errorf("unsupported action: %s", req.Action)
	}

	// Extract service parameters
	bandwidthMbps := 100
	if req.Product.ProductConfig != nil && req.Product.ProductConfig.BandwidthMbps > 0 {
		bandwidthMbps = req.Product.ProductConfig.BandwidthMbps
	}

	// Build Telesat order request
	telesatOrder := telesat.OrderRequest{
		ID:            fmt.Sprintf("to-%s", req.ID),
		ServiceType:   mapOfferingToServiceType(req.Product.ProductOfferingRef),
		BandwidthMbps: bandwidthMbps,
		RequestedDate: req.RequestedCompletionDate,
	}
	if req.QuoteItem != nil {
		telesatOrder.QuoteID = req.QuoteItem.QuoteID
	}
	if len(req.Product.Place) >= 2 {
		telesatOrder.EndpointA = telesat.ServicePoint{Latitude: req.Product.Place[0].Latitude, Longitude: req.Product.Place[0].Longitude}
		telesatOrder.EndpointZ = telesat.ServicePoint{Latitude: req.Product.Place[1].Latitude, Longitude: req.Product.Place[1].Longitude}
	}

	submitted, err := t.telesat.CreateOrderRequest(ctx, telesatOrder)
	if err != nil {
		return item, fmt.Errorf("failed to submit order to Telesat: %w", err)
	}

	item.State = mapTelesatOrderItemStateToMEF(submitted.Status)
	expectedCompletion := now.Add(15 * 24 * time.Hour)
	item.ExpectedCompletionDate = &expectedCompletion
	item.StateChange = []StateChange{
		{State: "acknowledged", ChangeDate: now},
		{State: item.State, ChangeDate: now},
	}

	return item, nil
}

// ListOrders returns all orders in MEF format.
func (t *Transformer) ListOrders(ctx context.Context) ([]MEFProductOrderFind, error) {
	telesatOrders, err := t.telesat.ListOrderRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching orders: %w", err)
	}

	orders := make([]MEFProductOrderFind, 0, len(telesatOrders))
	for _, to := range telesatOrders {
		orders = append(orders, MEFProductOrderFind{
			ID:             fmt.Sprintf("order-%s", to.ID),
			State:          mapTelesatOrderStateToMEF(to.Status),
			OrderDate:      to.CreatedAt,
			CompletionDate: to.CompletionDate,
		})
	}
	return orders, nil
}

// GetOrder returns a single order in full MEF format.
func (t *Transformer) GetOrder(ctx context.Context, id string) (*MEFProductOrder, error) {
	telesatID := id
	if len(id) > 6 && id[:6] == "order-" {
		telesatID = id[6:]
	}

	to, err := t.telesat.GetOrderRequest(ctx, telesatID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return t.telesatOrderToMEF(to), nil
}

// telesatOrderToMEF converts a Telesat order to full MEF format.
func (t *Transformer) telesatOrderToMEF(to *telesat.OrderRequest) *MEFProductOrder {
	mefState := mapTelesatOrderStateToMEF(to.Status)
	orderID := fmt.Sprintf("order-%s", to.ID)

	order := &MEFProductOrder{
		ID:             orderID,
		Href:           fmt.Sprintf("/mef/v4/productOrder/%s", orderID),
		State:          mefState,
		OrderDate:      to.CreatedAt,
		CompletionDate: to.CompletionDate,
		StateChange: []StateChange{
			{State: "acknowledged", ChangeDate: to.CreatedAt},
			{State: mefState, ChangeDate: to.CreatedAt},
		},
		ProductOrderItem: []MEFProductOrderItem{
			{
				ID:     "1",
				State:  mapTelesatOrderItemStateToMEF(to.Status),
				Action: "add",
				Product: OrderProduct{
					ProductOfferingRef: &ProductOfferingRef{
						ID: fmt.Sprintf("telesat-ls-%s", to.ServiceType),
					},
					Place: []QuotePlace{
						{Role: "UNI-A", Latitude: to.EndpointA.Latitude, Longitude: to.EndpointA.Longitude},
						{Role: "UNI-Z", Latitude: to.EndpointZ.Latitude, Longitude: to.EndpointZ.Longitude},
					},
					ProductConfig: &ProductConfig{
						Type:          "urn:mef:lso:spec:sonata:AccessElineOvc:v5.0.0:order",
						BandwidthMbps: to.BandwidthMbps,
					},
				},
				RequestedCompletionDate: to.RequestedDate,
				CompletionDate:          to.CompletionDate,
				QuoteItem: func() *QuoteItemRef {
					if to.QuoteID != "" {
						return &QuoteItemRef{QuoteID: fmt.Sprintf("quote-%s", to.QuoteID), ID: "1"}
					}
					return nil
				}(),
				StateChange: []StateChange{
					{State: "acknowledged", ChangeDate: to.CreatedAt},
					{State: mapTelesatOrderItemStateToMEF(to.Status), ChangeDate: to.CreatedAt},
				},
			},
		},
	}

	return order
}

func mapTelesatOrderStateToMEF(status string) string {
	switch status {
	case "submitted", "acknowledged":
		return "acknowledged"
	case "provisioning":
		return "inProgress"
	case "active":
		return "completed"
	case "failed":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return "acknowledged"
	}
}

func mapTelesatOrderItemStateToMEF(status string) string {
	switch status {
	case "submitted", "acknowledged":
		return "acknowledged"
	case "provisioning":
		return "inProgress"
	case "active":
		return "completed"
	case "failed":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return "acknowledged"
	}
}
