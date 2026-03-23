package mapping

import (
	"context"
	"fmt"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// MEF LSO Sonata Trouble Ticket Management models (MEF 124.1)

// MEFTroubleTicket represents a trouble ticket in MEF LSO Sonata format.
type MEFTroubleTicket struct {
	ID                        string              `json:"id"`
	Href                      string              `json:"href"`
	Status                    string              `json:"status"` // acknowledged, inProgress, pending, resolved, closed, cancelled
	Severity                  string              `json:"severity"` // minor, moderate, significant, extensive
	TicketType                string              `json:"ticketType"`
	CreationDate              time.Time           `json:"creationDate"`
	LastUpdate                time.Time           `json:"lastUpdate"`
	ResolutionDate            *time.Time          `json:"resolutionDate,omitempty"`
	Description               string              `json:"description"`
	ExternalID                string              `json:"externalId,omitempty"`
	Priority                  int                 `json:"priority"` // 0 (highest) to 4 (lowest)
	RelatedEntity             []RelatedEntity     `json:"relatedEntity,omitempty"`
	Attachment                []Attachment        `json:"attachment,omitempty"`
	Note                      []Note              `json:"note,omitempty"`
	StatusChange              []StatusChange      `json:"statusChange,omitempty"`
	RelatedContactInformation []ContactInfo       `json:"relatedContactInformation,omitempty"`
	TroubleTicketRelationship []TicketRelationship `json:"troubleTicketRelationship,omitempty"`
}

// MEFTroubleTicketFind is the abbreviated ticket returned in list responses.
type MEFTroubleTicketFind struct {
	ID             string    `json:"id"`
	Href           string    `json:"href"`
	Status         string    `json:"status"`
	Severity       string    `json:"severity"`
	TicketType     string    `json:"ticketType"`
	CreationDate   time.Time `json:"creationDate"`
	LastUpdate     time.Time `json:"lastUpdate"`
	ResolutionDate *time.Time `json:"resolutionDate,omitempty"`
	Description    string    `json:"description"`
	Priority       int       `json:"priority"`
	ExternalID     string    `json:"externalId,omitempty"`
}

// MEFTroubleTicketCreate is the request body for creating a trouble ticket.
type MEFTroubleTicketCreate struct {
	Description               string           `json:"description"`
	ExternalID                string           `json:"externalId,omitempty"`
	Severity                  string           `json:"severity"`
	TicketType                string           `json:"ticketType"`
	Priority                  int              `json:"priority"`
	RelatedEntity             []RelatedEntity  `json:"relatedEntity,omitempty"`
	Attachment                []Attachment     `json:"attachment,omitempty"`
	Note                      []Note           `json:"note,omitempty"`
	RelatedContactInformation []ContactInfo    `json:"relatedContactInformation,omitempty"`
}

// MEFTroubleTicketUpdate is the request body for patching a trouble ticket.
type MEFTroubleTicketUpdate struct {
	Status         *string      `json:"status,omitempty"`
	Severity       *string      `json:"severity,omitempty"`
	Priority       *int         `json:"priority,omitempty"`
	ExternalID     *string      `json:"externalId,omitempty"`
	Note           []Note       `json:"note,omitempty"`
	Attachment     []Attachment `json:"attachment,omitempty"`
	ResolutionNote *string      `json:"resolutionNote,omitempty"`
}

// RelatedEntity links a ticket to a product/service.
type RelatedEntity struct {
	ID   string `json:"id"`
	Href string `json:"href,omitempty"`
	Role string `json:"role"` // affectedProduct, affectedService, affectedResource
	Name string `json:"name,omitempty"`
	Type string `json:"@referredType,omitempty"`
}

// Attachment represents a file attachment on a ticket.
type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType"`
	URL         string `json:"url"`
	Size        int    `json:"size,omitempty"`
}

// StatusChange records a ticket status transition.
type StatusChange struct {
	Status     string    `json:"status"`
	ChangeDate time.Time `json:"changeDate"`
	Reason     string    `json:"changeReason,omitempty"`
}

// TicketRelationship links related trouble tickets.
type TicketRelationship struct {
	ID               string `json:"id"`
	Href             string `json:"href,omitempty"`
	RelationshipType string `json:"relationshipType"` // parent, child, duplicate, relatedTo
}

// --- Trouble Ticket Transformation Methods ---

// CreateTicket processes an MEF Trouble Ticket creation request.
func (t *Transformer) CreateTicket(ctx context.Context, req MEFTroubleTicketCreate) (*MEFTroubleTicket, error) {
	now := time.Now().UTC()
	ticketID := fmt.Sprintf("tt-%d", now.UnixMilli())

	// Map MEF severity to Telesat severity
	telesatSeverity := mapMEFSeverityToTelesat(req.Severity)

	// Map ticket type
	telesatType := mapMEFTicketTypeToTelesat(req.TicketType)

	// Extract subscription ID from related entities
	subscriptionID := ""
	for _, entity := range req.RelatedEntity {
		if entity.Role == "affectedProduct" || entity.Role == "affectedService" {
			subscriptionID = entity.ID
			break
		}
	}

	telesatTicket := telesat.TroubleTicket{
		ID:             ticketID,
		Severity:       telesatSeverity,
		Type:           telesatType,
		SubscriptionID: subscriptionID,
		Summary:        req.Description,
		Description:    req.Description,
	}

	created, err := t.telesat.CreateTroubleTicket(ctx, telesatTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to create trouble ticket: %w", err)
	}

	ticket := t.telesatTicketToMEF(created)
	ticket.ExternalID = req.ExternalID
	ticket.Priority = req.Priority
	ticket.RelatedEntity = req.RelatedEntity
	ticket.Attachment = req.Attachment
	ticket.Note = req.Note
	ticket.RelatedContactInformation = req.RelatedContactInformation
	ticket.StatusChange = []StatusChange{
		{Status: "acknowledged", ChangeDate: now},
	}

	return ticket, nil
}

// ListTickets returns all trouble tickets in MEF format.
func (t *Transformer) ListTickets(ctx context.Context) ([]MEFTroubleTicketFind, error) {
	telesatTickets, err := t.telesat.ListTroubleTickets(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching trouble tickets: %w", err)
	}

	tickets := make([]MEFTroubleTicketFind, 0, len(telesatTickets))
	for _, tt := range telesatTickets {
		tickets = append(tickets, MEFTroubleTicketFind{
			ID:             fmt.Sprintf("mef-tt-%s", tt.ID),
			Href:           fmt.Sprintf("/mef/v4/troubleTicket/mef-tt-%s", tt.ID),
			Status:         mapTelesatTicketStatusToMEF(tt.Status),
			Severity:       mapTelesatSeverityToMEF(tt.Severity),
			TicketType:     mapTelesatTicketTypeToMEF(tt.Type),
			CreationDate:   tt.CreatedAt,
			LastUpdate:     tt.UpdatedAt,
			ResolutionDate: tt.ResolvedAt,
			Description:    tt.Summary,
			Priority:       severityToPriority(tt.Severity),
		})
	}
	return tickets, nil
}

// GetTicket returns a single trouble ticket in full MEF format.
func (t *Transformer) GetTicket(ctx context.Context, id string) (*MEFTroubleTicket, error) {
	telesatID := id
	if len(id) > 7 && id[:7] == "mef-tt-" {
		telesatID = id[7:]
	}

	tt, err := t.telesat.GetTroubleTicket(ctx, telesatID)
	if err != nil {
		return nil, fmt.Errorf("trouble ticket not found: %w", err)
	}

	ticket := t.telesatTicketToMEF(tt)

	// Add related entity for the affected subscription
	if tt.SubscriptionID != "" {
		ticket.RelatedEntity = []RelatedEntity{
			{
				ID:   fmt.Sprintf("mef-%s", tt.SubscriptionID),
				Href: fmt.Sprintf("/mef/v4/productInventory/product/mef-%s", tt.SubscriptionID),
				Role: "affectedProduct",
				Type: "Product",
			},
		}
	}

	return ticket, nil
}

// PatchTicket updates a trouble ticket.
func (t *Transformer) PatchTicket(ctx context.Context, id string, update MEFTroubleTicketUpdate) (*MEFTroubleTicket, error) {
	telesatID := id
	if len(id) > 7 && id[:7] == "mef-tt-" {
		telesatID = id[7:]
	}

	status := ""
	if update.Status != nil {
		status = mapMEFTicketStatusToTelesat(*update.Status)
	}

	resolutionNote := ""
	if update.ResolutionNote != nil {
		resolutionNote = *update.ResolutionNote
	}

	updated, err := t.telesat.UpdateTroubleTicket(ctx, telesatID, status, resolutionNote)
	if err != nil {
		return nil, fmt.Errorf("failed to update trouble ticket: %w", err)
	}

	return t.telesatTicketToMEF(updated), nil
}

// telesatTicketToMEF converts a Telesat trouble ticket to full MEF format.
func (t *Transformer) telesatTicketToMEF(tt *telesat.TroubleTicket) *MEFTroubleTicket {
	mefStatus := mapTelesatTicketStatusToMEF(tt.Status)
	ticketID := fmt.Sprintf("mef-tt-%s", tt.ID)

	ticket := &MEFTroubleTicket{
		ID:             ticketID,
		Href:           fmt.Sprintf("/mef/v4/troubleTicket/%s", ticketID),
		Status:         mefStatus,
		Severity:       mapTelesatSeverityToMEF(tt.Severity),
		TicketType:     mapTelesatTicketTypeToMEF(tt.Type),
		CreationDate:   tt.CreatedAt,
		LastUpdate:     tt.UpdatedAt,
		ResolutionDate: tt.ResolvedAt,
		Description:    tt.Summary,
		Priority:       severityToPriority(tt.Severity),
		StatusChange: []StatusChange{
			{Status: "acknowledged", ChangeDate: tt.CreatedAt},
		},
	}

	if mefStatus != "acknowledged" {
		ticket.StatusChange = append(ticket.StatusChange,
			StatusChange{Status: mefStatus, ChangeDate: tt.UpdatedAt})
	}

	if tt.ResolutionNote != "" {
		ticket.Note = []Note{
			{
				ID:     "resolution-1",
				Author: "Telesat NOC",
				Date:   tt.UpdatedAt,
				Source: "seller",
				Text:   tt.ResolutionNote,
			},
		}
	}

	if tt.Description != tt.Summary {
		ticket.Note = append(ticket.Note, Note{
			ID:     "detail-1",
			Author: "system",
			Date:   tt.CreatedAt,
			Source: "seller",
			Text:   tt.Description,
		})
	}

	return ticket
}

// --- Status/Severity Mapping Functions ---

func mapTelesatTicketStatusToMEF(status string) string {
	switch status {
	case "submitted":
		return "acknowledged"
	case "acknowledged":
		return "acknowledged"
	case "inProgress":
		return "inProgress"
	case "resolved":
		return "resolved"
	case "closed":
		return "closed"
	case "cancelled":
		return "cancelled"
	default:
		return "acknowledged"
	}
}

func mapMEFTicketStatusToTelesat(status string) string {
	switch status {
	case "acknowledged":
		return "acknowledged"
	case "inProgress":
		return "inProgress"
	case "resolved":
		return "resolved"
	case "closed":
		return "closed"
	case "cancelled":
		return "cancelled"
	default:
		return status
	}
}

func mapTelesatSeverityToMEF(severity string) string {
	switch severity {
	case "critical":
		return "extensive"
	case "major":
		return "significant"
	case "minor":
		return "moderate"
	case "informational":
		return "minor"
	default:
		return "minor"
	}
}

func mapMEFSeverityToTelesat(severity string) string {
	switch severity {
	case "extensive":
		return "critical"
	case "significant":
		return "major"
	case "moderate":
		return "minor"
	case "minor":
		return "informational"
	default:
		return "minor"
	}
}

func mapTelesatTicketTypeToMEF(ticketType string) string {
	switch ticketType {
	case "signalDegradation":
		return "qualityOfServiceProblem"
	case "beamHandover":
		return "connectivityProblem"
	case "terminalFault":
		return "equipmentProblem"
	case "slaViolation":
		return "performanceProblem"
	case "linkDown":
		return "connectivityProblem"
	case "interference":
		return "qualityOfServiceProblem"
	default:
		return "informationRequest"
	}
}

func mapMEFTicketTypeToTelesat(ticketType string) string {
	switch ticketType {
	case "qualityOfServiceProblem":
		return "signalDegradation"
	case "connectivityProblem":
		return "linkDown"
	case "equipmentProblem":
		return "terminalFault"
	case "performanceProblem":
		return "slaViolation"
	case "informationRequest":
		return "informational"
	default:
		return "informational"
	}
}

func severityToPriority(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "major":
		return 1
	case "minor":
		return 2
	case "informational":
		return 3
	default:
		return 3
	}
}
