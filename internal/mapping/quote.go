package mapping

import (
	"context"
	"fmt"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// MEF LSO Sonata Quote Management models (MEF 115.1 / Mplify 115.1)

// MEFQuote represents a complete quote in MEF LSO Sonata format.
type MEFQuote struct {
	ID                             string           `json:"id"`
	Href                           string           `json:"href"`
	State                          string           `json:"state"`
	QuoteDate                      time.Time        `json:"quoteDate"`
	QuoteLevel                     string           `json:"quoteLevel"` // budgetary, firm, firmSubjectToFeasibilityCheck
	Description                    string           `json:"description,omitempty"`
	ExternalID                     string           `json:"externalId,omitempty"`
	InstantSyncQuote               bool             `json:"instantSyncQuote"`
	RequestedQuoteCompletionDate   *time.Time       `json:"requestedQuoteCompletionDate,omitempty"`
	ExpectedQuoteCompletionDate    *time.Time       `json:"expectedQuoteCompletionDate,omitempty"`
	EffectiveQuoteCompletionDate   *time.Time       `json:"effectiveQuoteCompletionDate,omitempty"`
	ValidFor                       *TimePeriod      `json:"validFor,omitempty"`
	QuoteItem                      []MEFQuoteItem   `json:"quoteItem"`
	StateChange                    []StateChange    `json:"stateChange"`
	RelatedContactInformation      []ContactInfo    `json:"relatedContactInformation,omitempty"`
	Note                           []Note           `json:"note,omitempty"`
}

// MEFQuoteFind is the abbreviated quote returned in list responses.
type MEFQuoteFind struct {
	ID                             string     `json:"id"`
	State                          string     `json:"state"`
	QuoteDate                      time.Time  `json:"quoteDate"`
	QuoteLevel                     string     `json:"quoteLevel,omitempty"`
	ExternalID                     string     `json:"externalId,omitempty"`
	RequestedQuoteCompletionDate   *time.Time `json:"requestedQuoteCompletionDate,omitempty"`
	ExpectedQuoteCompletionDate    *time.Time `json:"expectedQuoteCompletionDate,omitempty"`
	EffectiveQuoteCompletionDate   *time.Time `json:"effectiveQuoteCompletionDate,omitempty"`
}

// MEFQuoteCreate is the request body for creating a quote.
type MEFQuoteCreate struct {
	BuyerRequestedQuoteLevel     string                `json:"buyerRequestedQuoteLevel"` // budgetary, firm
	InstantSyncQuote             bool                  `json:"instantSyncQuote"`
	Description                  string                `json:"description,omitempty"`
	ExternalID                   string                `json:"externalId,omitempty"`
	ProjectID                    string                `json:"projectId,omitempty"`
	RequestedQuoteCompletionDate *time.Time            `json:"requestedQuoteCompletionDate,omitempty"`
	QuoteItem                    []MEFQuoteItemCreate  `json:"quoteItem"`
	RelatedContactInformation    []ContactInfo         `json:"relatedContactInformation,omitempty"`
	Note                         []Note                `json:"note,omitempty"`
}

// MEFQuoteItem represents an item within a quote.
type MEFQuoteItem struct {
	ID                                 string          `json:"id"`
	State                              string          `json:"state"`
	Action                             string          `json:"action"` // add, modify, delete
	Product                            QuoteProduct    `json:"product"`
	QuoteItemPrice                     []QuotePrice    `json:"quoteItemPrice,omitempty"`
	QuoteItemTerm                      *MEFItemTerm    `json:"quoteItemTerm,omitempty"`
	QuoteItemInstallationInterval      *Duration       `json:"quoteItemInstallationInterval,omitempty"`
	SubjectToFeasibilityCheck          bool            `json:"subjectToFeasibilityCheck"`
	ProductOfferingQualificationItem   *POQItemRef     `json:"productOfferingQualificationItem,omitempty"`
	StateChange                        []StateChange   `json:"stateChange,omitempty"`
	Note                               []Note          `json:"note,omitempty"`
}

// MEFQuoteItemCreate is the request body for a quote item.
type MEFQuoteItemCreate struct {
	ID                                 string       `json:"id"`
	Action                             string       `json:"action"` // add, modify, delete
	Product                            QuoteProduct `json:"product"`
	ProductOfferingQualificationItem   *POQItemRef  `json:"productOfferingQualificationItem,omitempty"`
	RequestedQuoteItemTerm             *MEFItemTerm `json:"requestedQuoteItemTerm,omitempty"`
	RequestedQuoteItemInstallationInterval *Duration `json:"requestedQuoteItemInstallationInterval,omitempty"`
	Note                               []Note       `json:"note,omitempty"`
}

// QuoteProduct identifies the product being quoted.
type QuoteProduct struct {
	ProductOfferingRef *ProductOfferingRef `json:"productOffering,omitempty"`
	Place              []QuotePlace        `json:"place,omitempty"`
	ProductConfig      *ProductConfig      `json:"productConfiguration,omitempty"`
}

// ProductOfferingRef references a product offering from the catalog.
type ProductOfferingRef struct {
	ID   string `json:"id"`
	Href string `json:"href,omitempty"`
}

// ProductConfig holds satellite-specific product configuration.
type ProductConfig struct {
	Type          string  `json:"@type"`
	BandwidthMbps int    `json:"bandwidthMbps,omitempty"`
	CIR           int    `json:"cirMbps,omitempty"`
	TerminalType  string `json:"terminalType,omitempty"`
}

// QuotePlace represents a geographic location in a quote.
type QuotePlace struct {
	Role      string  `json:"role"` // UNI-A, UNI-Z
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// QuotePrice represents pricing for a quote item.
type QuotePrice struct {
	Name                 string   `json:"name"`
	PriceType            string   `json:"priceType"` // recurring, nonRecurring
	RecurringChargePeriod string  `json:"recurringChargePeriod,omitempty"`
	Price                Price    `json:"price"`
	Description          string   `json:"description,omitempty"`
}

// Price represents a monetary amount.
type Price struct {
	DutyFreeAmount Money   `json:"dutyFreeAmount"`
	TaxRate        float64 `json:"taxRate,omitempty"`
}

// Money represents a currency value.
type Money struct {
	Unit  string  `json:"unit"` // ISO 4217 currency code
	Value float64 `json:"value"`
}

// POQItemRef references a prior Product Offering Qualification item.
type POQItemRef struct {
	ProductOfferingQualificationID   string `json:"productOfferingQualificationId"`
	ID                               string `json:"id"`
	ProductOfferingQualificationHref string `json:"productOfferingQualificationHref,omitempty"`
}

// MEFItemTerm defines the contract term.
type MEFItemTerm struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	Duration      Duration `json:"duration"`
	EndOfTermAction string `json:"endOfTermAction,omitempty"` // roll, autoDisconnect, autoRenew
}

// Duration represents a time duration with units.
type Duration struct {
	Amount int    `json:"amount"`
	Units  string `json:"units"` // calendarDays, calendarHours, months, years
}

// TimePeriod represents a validity period.
type TimePeriod struct {
	StartDateTime time.Time `json:"startDateTime"`
	EndDateTime   time.Time `json:"endDateTime"`
}

// StateChange records a state transition.
type StateChange struct {
	State        string    `json:"state"`
	ChangeDate   time.Time `json:"changeDate"`
	ChangeReason string    `json:"changeReason,omitempty"`
}

// ContactInfo represents a contact associated with a quote or order.
type ContactInfo struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	Number       string `json:"number,omitempty"`
	Role         string `json:"role"` // buyerContactInformation, sellerContact
}

// Note represents a free-text note.
type Note struct {
	ID     string    `json:"id"`
	Author string    `json:"author"`
	Date   time.Time `json:"date"`
	Source string    `json:"source"` // buyer, seller
	Text   string    `json:"text"`
}

// --- Quote Transformation Methods ---

// CreateQuote processes an MEF Quote request and returns pricing from Telesat.
func (t *Transformer) CreateQuote(ctx context.Context, req MEFQuoteCreate) (*MEFQuote, error) {
	now := time.Now().UTC()
	quoteID := fmt.Sprintf("quote-%d", now.UnixMilli())

	quote := &MEFQuote{
		ID:               quoteID,
		Href:             fmt.Sprintf("/mef/v4/quote/%s", quoteID),
		QuoteDate:        now,
		InstantSyncQuote: req.InstantSyncQuote,
		ExternalID:       req.ExternalID,
		Description:      req.Description,
		RelatedContactInformation: req.RelatedContactInformation,
		Note:             req.Note,
	}

	// Process each quote item
	for _, itemReq := range req.QuoteItem {
		item, err := t.processQuoteItem(ctx, itemReq, req.BuyerRequestedQuoteLevel)
		if err != nil {
			quote.State = "rejected"
			quote.QuoteLevel = req.BuyerRequestedQuoteLevel
			quote.StateChange = []StateChange{
				{State: "acknowledged", ChangeDate: now},
				{State: "rejected", ChangeDate: now, ChangeReason: err.Error()},
			}
			quote.QuoteItem = []MEFQuoteItem{item}
			return quote, nil
		}
		quote.QuoteItem = append(quote.QuoteItem, item)
	}

	// For instant sync quotes, return pricing immediately
	if req.InstantSyncQuote {
		quote.State = "approved.orderable"
		quote.QuoteLevel = mapBuyerToSellerQuoteLevel(req.BuyerRequestedQuoteLevel)
		validUntil := now.Add(90 * 24 * time.Hour)
		quote.ValidFor = &TimePeriod{StartDateTime: now, EndDateTime: validUntil}
		completionDate := now
		quote.EffectiveQuoteCompletionDate = &completionDate
		quote.StateChange = []StateChange{
			{State: "acknowledged", ChangeDate: now},
			{State: "inProgress", ChangeDate: now},
			{State: "answered", ChangeDate: now},
			{State: "approved.orderable", ChangeDate: now},
		}
	} else {
		quote.State = "inProgress"
		quote.QuoteLevel = req.BuyerRequestedQuoteLevel
		expected := now.Add(2 * 24 * time.Hour)
		quote.ExpectedQuoteCompletionDate = &expected
		quote.StateChange = []StateChange{
			{State: "acknowledged", ChangeDate: now},
			{State: "inProgress", ChangeDate: now},
		}
	}

	return quote, nil
}

// processQuoteItem prices a single quote item by calling Telesat's pricing.
func (t *Transformer) processQuoteItem(ctx context.Context, req MEFQuoteItemCreate, quoteLevel string) (MEFQuoteItem, error) {
	item := MEFQuoteItem{
		ID:     req.ID,
		Action: req.Action,
		Product: req.Product,
		ProductOfferingQualificationItem: req.ProductOfferingQualificationItem,
		Note: req.Note,
	}

	if req.Action != "add" && req.Action != "modify" && req.Action != "delete" {
		item.State = "rejected"
		return item, fmt.Errorf("unsupported action: %s", req.Action)
	}

	// Extract service parameters from the product configuration
	bandwidthMbps := 100
	termMonths := 12
	if req.Product.ProductConfig != nil {
		if req.Product.ProductConfig.BandwidthMbps > 0 {
			bandwidthMbps = req.Product.ProductConfig.BandwidthMbps
		}
	}
	if req.RequestedQuoteItemTerm != nil && req.RequestedQuoteItemTerm.Duration.Units == "months" {
		termMonths = req.RequestedQuoteItemTerm.Duration.Amount
	}

	// Get pricing from Telesat
	telesatQuote := telesat.QuoteRequest{
		ID:            fmt.Sprintf("tq-%s", req.ID),
		ServiceType:   mapOfferingToServiceType(req.Product.ProductOfferingRef),
		BandwidthMbps: bandwidthMbps,
		TermMonths:    termMonths,
	}
	if len(req.Product.Place) >= 2 {
		telesatQuote.EndpointA = telesat.ServicePoint{Latitude: req.Product.Place[0].Latitude, Longitude: req.Product.Place[0].Longitude}
		telesatQuote.EndpointZ = telesat.ServicePoint{Latitude: req.Product.Place[1].Latitude, Longitude: req.Product.Place[1].Longitude}
	}

	priced, err := t.telesat.CreateQuoteRequest(ctx, telesatQuote)
	if err != nil {
		item.State = "unableToProvide"
		return item, err
	}

	// Map Telesat pricing to MEF quote prices
	item.State = "approved.orderable"
	item.SubjectToFeasibilityCheck = false
	item.QuoteItemPrice = []QuotePrice{
		{
			Name:      "Monthly Recurring Charge",
			PriceType: "recurring",
			RecurringChargePeriod: "month",
			Price: Price{
				DutyFreeAmount: Money{Unit: "USD", Value: priced.MonthlyPriceUSD},
			},
		},
		{
			Name:      "Non-Recurring Setup Charge",
			PriceType: "nonRecurring",
			Price: Price{
				DutyFreeAmount: Money{Unit: "USD", Value: priced.SetupPriceUSD},
			},
		},
	}

	if req.RequestedQuoteItemTerm != nil {
		item.QuoteItemTerm = req.RequestedQuoteItemTerm
	} else {
		item.QuoteItemTerm = &MEFItemTerm{
			Name:     "Standard Term",
			Duration: Duration{Amount: termMonths, Units: "months"},
			EndOfTermAction: "autoRenew",
		}
	}

	item.QuoteItemInstallationInterval = &Duration{Amount: 15, Units: "businessDays"}

	return item, nil
}

// ListQuotes returns all quotes in MEF format.
func (t *Transformer) ListQuotes(ctx context.Context) ([]MEFQuoteFind, error) {
	telesatQuotes, err := t.telesat.ListQuoteRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching quotes: %w", err)
	}

	quotes := make([]MEFQuoteFind, 0, len(telesatQuotes))
	for _, tq := range telesatQuotes {
		quotes = append(quotes, MEFQuoteFind{
			ID:         fmt.Sprintf("quote-%s", tq.ID),
			State:      mapTelesatQuoteStateToMEF(tq.Status),
			QuoteDate:  tq.CreatedAt,
			QuoteLevel: "firm",
		})
	}
	return quotes, nil
}

// GetQuote returns a single quote in full MEF format.
func (t *Transformer) GetQuote(ctx context.Context, id string) (*MEFQuote, error) {
	// Try to find in mock data by stripping the "quote-" prefix
	telesatID := id
	if len(id) > 6 && id[:6] == "quote-" {
		telesatID = id[6:]
	}

	tq, err := t.telesat.GetQuoteRequest(ctx, telesatID)
	if err != nil {
		return nil, fmt.Errorf("quote not found: %w", err)
	}

	return t.telesatQuoteToMEF(tq), nil
}

// telesatQuoteToMEF converts a Telesat quote to full MEF format.
func (t *Transformer) telesatQuoteToMEF(tq *telesat.QuoteRequest) *MEFQuote {
	mefState := mapTelesatQuoteStateToMEF(tq.Status)
	quoteID := fmt.Sprintf("quote-%s", tq.ID)

	quote := &MEFQuote{
		ID:               quoteID,
		Href:             fmt.Sprintf("/mef/v4/quote/%s", quoteID),
		State:            mefState,
		QuoteDate:        tq.CreatedAt,
		QuoteLevel:       "firm",
		InstantSyncQuote: true,
		ValidFor:         &TimePeriod{StartDateTime: tq.CreatedAt, EndDateTime: tq.ValidUntil},
		StateChange: []StateChange{
			{State: "acknowledged", ChangeDate: tq.CreatedAt},
			{State: mefState, ChangeDate: tq.CreatedAt},
		},
		QuoteItem: []MEFQuoteItem{
			{
				ID:     "1",
				State:  mefState,
				Action: "add",
				Product: QuoteProduct{
					ProductOfferingRef: &ProductOfferingRef{
						ID: fmt.Sprintf("telesat-ls-%s", tq.ServiceType),
					},
					Place: []QuotePlace{
						{Role: "UNI-A", Latitude: tq.EndpointA.Latitude, Longitude: tq.EndpointA.Longitude},
						{Role: "UNI-Z", Latitude: tq.EndpointZ.Latitude, Longitude: tq.EndpointZ.Longitude},
					},
					ProductConfig: &ProductConfig{
						Type:          "urn:mef:lso:spec:sonata:AccessElineOvc:v5.0.0:quote",
						BandwidthMbps: tq.BandwidthMbps,
					},
				},
				QuoteItemPrice: []QuotePrice{
					{
						Name:      "Monthly Recurring Charge",
						PriceType: "recurring",
						RecurringChargePeriod: "month",
						Price:     Price{DutyFreeAmount: Money{Unit: "USD", Value: tq.MonthlyPriceUSD}},
					},
					{
						Name:      "Non-Recurring Setup Charge",
						PriceType: "nonRecurring",
						Price:     Price{DutyFreeAmount: Money{Unit: "USD", Value: tq.SetupPriceUSD}},
					},
				},
				QuoteItemTerm: &MEFItemTerm{
					Name:     "Contract Term",
					Duration: Duration{Amount: tq.TermMonths, Units: "months"},
					EndOfTermAction: "autoRenew",
				},
				QuoteItemInstallationInterval: &Duration{Amount: 15, Units: "businessDays"},
			},
		},
	}

	return quote
}

func mapTelesatQuoteStateToMEF(status string) string {
	switch status {
	case "pending":
		return "inProgress"
	case "priced":
		return "approved.orderable"
	case "rejected":
		return "rejected"
	case "expired":
		return "expired"
	default:
		return "inProgress"
	}
}

func mapBuyerToSellerQuoteLevel(level string) string {
	switch level {
	case "firm":
		return "firm"
	case "budgetary":
		return "budgetary"
	default:
		return "firmSubjectToFeasibilityCheck"
	}
}

func mapOfferingToServiceType(ref *ProductOfferingRef) string {
	if ref == nil {
		return "e-line"
	}
	switch ref.ID {
	case "telesat-ls-eaccess":
		return "e-access"
	case "telesat-ls-vno-pool":
		return "vno-pool"
	default:
		return "e-line"
	}
}
