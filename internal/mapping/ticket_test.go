package mapping

import (
	"context"
	"testing"
)

func TestCreateTicket(t *testing.T) {
	tr := newTestTransformer()
	req := MEFTroubleTicketCreate{
		Description: "Service degradation on Ottawa-Paris link",
		Severity:    "significant",
		TicketType:  "performanceProblem",
		Priority:    1,
		RelatedEntity: []RelatedEntity{
			{ID: "mef-sub-001", Role: "affectedProduct", Type: "Product"},
		},
	}

	ticket, err := tr.CreateTicket(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ticket.Status != "acknowledged" {
		t.Errorf("expected acknowledged status, got %s", ticket.Status)
	}
	if ticket.Severity != "significant" {
		t.Errorf("expected significant severity, got %s", ticket.Severity)
	}
	if ticket.Priority != 1 {
		t.Errorf("expected priority 1, got %d", ticket.Priority)
	}
	if len(ticket.RelatedEntity) != 1 {
		t.Fatalf("expected 1 related entity, got %d", len(ticket.RelatedEntity))
	}
	if len(ticket.StatusChange) == 0 {
		t.Error("expected at least one status change")
	}
}

func TestListTickets(t *testing.T) {
	tr := newTestTransformer()
	tickets, err := tr.ListTickets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tickets) != 3 {
		t.Fatalf("expected 3 tickets, got %d", len(tickets))
	}

	// Verify severity mapping
	for _, tt := range tickets {
		if tt.Severity == "" {
			t.Errorf("ticket %s should have a severity", tt.ID)
		}
		if tt.TicketType == "" {
			t.Errorf("ticket %s should have a type", tt.ID)
		}
	}
}

func TestGetTicket(t *testing.T) {
	tr := newTestTransformer()

	ticket, err := tr.GetTicket(context.Background(), "mef-tt-tt-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != "resolved" {
		t.Errorf("expected resolved status, got %s", ticket.Status)
	}
	if ticket.ResolutionDate == nil {
		t.Error("resolved ticket should have resolution date")
	}
	if len(ticket.RelatedEntity) == 0 {
		t.Error("expected related entity for subscription")
	}
	if len(ticket.Note) == 0 {
		t.Error("resolved ticket should have resolution note")
	}

	// Non-existent
	_, err = tr.GetTicket(context.Background(), "mef-tt-nonexistent")
	if err == nil {
		t.Error("expected error for non-existent ticket")
	}
}

func TestPatchTicket(t *testing.T) {
	tr := newTestTransformer()
	status := "inProgress"
	note := "Investigating the issue"
	update := MEFTroubleTicketUpdate{
		Status:         &status,
		ResolutionNote: &note,
	}

	ticket, err := tr.PatchTicket(context.Background(), "mef-tt-tt-003", update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != "inProgress" {
		t.Errorf("expected inProgress status, got %s", ticket.Status)
	}
}

func TestSeverityMapping(t *testing.T) {
	tests := []struct {
		mef     string
		telesat string
	}{
		{"extensive", "critical"},
		{"significant", "major"},
		{"moderate", "minor"},
		{"minor", "informational"},
	}

	for _, tt := range tests {
		t.Run(tt.mef, func(t *testing.T) {
			got := mapMEFSeverityToTelesat(tt.mef)
			if got != tt.telesat {
				t.Errorf("MEF %s -> got %s, want %s", tt.mef, got, tt.telesat)
			}
			// Round-trip
			back := mapTelesatSeverityToMEF(got)
			if back != tt.mef {
				t.Errorf("round-trip: %s -> %s -> %s, want %s", tt.mef, got, back, tt.mef)
			}
		})
	}
}

func TestTicketTypeMapping(t *testing.T) {
	tests := []struct {
		telesat string
		mef     string
	}{
		{"signalDegradation", "qualityOfServiceProblem"},
		{"beamHandover", "connectivityProblem"},
		{"terminalFault", "equipmentProblem"},
		{"slaViolation", "performanceProblem"},
		{"linkDown", "connectivityProblem"},
		{"interference", "qualityOfServiceProblem"},
	}

	for _, tt := range tests {
		t.Run(tt.telesat, func(t *testing.T) {
			got := mapTelesatTicketTypeToMEF(tt.telesat)
			if got != tt.mef {
				t.Errorf("got %s, want %s", got, tt.mef)
			}
		})
	}
}

func TestSeverityToPriority(t *testing.T) {
	tests := []struct {
		severity string
		priority int
	}{
		{"critical", 0},
		{"major", 1},
		{"minor", 2},
		{"informational", 3},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := severityToPriority(tt.severity)
			if got != tt.priority {
				t.Errorf("got %d, want %d", got, tt.priority)
			}
		})
	}
}
