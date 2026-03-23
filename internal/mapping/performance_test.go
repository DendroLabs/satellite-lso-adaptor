package mapping

import (
	"context"
	"testing"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

func TestListPerformanceReports(t *testing.T) {
	tr := newTestTransformer()

	// List all
	reports, err := tr.ListPerformanceReports(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 4 {
		t.Fatalf("expected 4 reports, got %d", len(reports))
	}

	// Filter by subscription
	filtered, err := tr.ListPerformanceReports(context.Background(), "mef-sub-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 reports for sub-001, got %d", len(filtered))
	}
	for _, r := range filtered {
		if r.ServiceRef.ID != "mef-sub-001" {
			t.Errorf("expected service ref mef-sub-001, got %s", r.ServiceRef.ID)
		}
	}
}

func TestListPerformanceReports_Compliance(t *testing.T) {
	tr := newTestTransformer()
	reports, err := tr.ListPerformanceReports(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// pr-002 should be violated (latency 62.3ms exceeds 50ms PT5 threshold)
	for _, r := range reports {
		if r.ID == "mef-pr-pr-002" {
			if r.ComplianceSummary != "violated" {
				t.Errorf("pr-002 compliance should be violated, got %s", r.ComplianceSummary)
			}
		}
		if r.ID == "mef-pr-pr-001" {
			if r.ComplianceSummary != "compliant" {
				t.Errorf("pr-001 compliance should be compliant, got %s", r.ComplianceSummary)
			}
		}
	}
}

func TestGetPerformanceReport(t *testing.T) {
	tr := newTestTransformer()

	report, err := tr.GetPerformanceReport(context.Background(), "mef-pr-pr-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.State != "completed" {
		t.Errorf("expected completed state, got %s", report.State)
	}
	if len(report.PerformanceObjective) != 4 {
		t.Fatalf("expected 4 performance objectives, got %d", len(report.PerformanceObjective))
	}

	// All objectives should be compliant for pr-001
	for _, obj := range report.PerformanceObjective {
		if !obj.Compliant {
			t.Errorf("objective %s should be compliant for pr-001", obj.ObjectiveName)
		}
	}

	// Verify satellite-specific metrics are present
	if report.ResultPayload.Satellite.BeamHandovers == 0 && report.ResultPayload.Satellite.SNRdB == 0 {
		t.Error("expected satellite-specific metrics")
	}

	// Non-existent
	_, err = tr.GetPerformanceReport(context.Background(), "mef-pr-nonexistent")
	if err == nil {
		t.Error("expected error for non-existent report")
	}
}

func TestGetPerformanceReport_WithViolation(t *testing.T) {
	tr := newTestTransformer()

	report, err := tr.GetPerformanceReport(context.Background(), "mef-pr-pr-002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// pr-002 has max latency 62.3ms (exceeds 50ms) and max jitter 12.1ms (exceeds 10ms)
	latencyObj := report.PerformanceObjective[0]
	if latencyObj.Compliant {
		t.Error("latency objective should NOT be compliant (62.3ms > 50ms)")
	}
	if latencyObj.MeasuredValue != 62.3 {
		t.Errorf("expected measured latency 62.3, got %.1f", latencyObj.MeasuredValue)
	}

	jitterObj := report.PerformanceObjective[1]
	if jitterObj.Compliant {
		t.Error("jitter objective should NOT be compliant (12.1ms > 10ms)")
	}
}

func TestBuildPT5Objectives(t *testing.T) {
	metrics := telesat.PerformanceMetrics{
		MaxLatencyMs:    30.0,
		MaxJitterMs:     5.0,
		PacketLossPct:   0.05,
		AvailabilityPct: 99.95,
	}

	// Without subscription (uses PT5 defaults)
	objectives := buildPT5Objectives(metrics, nil)
	if len(objectives) != 4 {
		t.Fatalf("expected 4 objectives, got %d", len(objectives))
	}
	for _, obj := range objectives {
		if !obj.Compliant {
			t.Errorf("all objectives should be compliant with these metrics, but %s is not", obj.ObjectiveName)
		}
	}

	// With subscription that has tighter SLA
	sub := &telesat.Subscription{
		SLA: telesat.SLAParameters{
			MaxLatencyMs:     25.0, // tighter than the 30ms measured
			MaxJitterMs:      10.0,
			MaxPacketLossPct: 0.1,
			Availability:     99.9,
		},
	}
	objectives = buildPT5Objectives(metrics, sub)
	// Latency should now be non-compliant (30 > 25)
	if objectives[0].Compliant {
		t.Error("latency should be non-compliant when threshold is 25ms and measured is 30ms")
	}
}

func TestEvaluateCompliance(t *testing.T) {
	goodMetrics := telesat.PerformanceMetrics{
		MaxLatencyMs:    30.0,
		MaxJitterMs:     5.0,
		PacketLossPct:   0.01,
		AvailabilityPct: 99.99,
	}
	if got := evaluateCompliance(goodMetrics, nil); got != "compliant" {
		t.Errorf("expected compliant, got %s", got)
	}

	// One violation
	degradedMetrics := telesat.PerformanceMetrics{
		MaxLatencyMs:    200.0, // exceeds PT5
		MaxJitterMs:     5.0,
		PacketLossPct:   0.01,
		AvailabilityPct: 99.99,
	}
	if got := evaluateCompliance(degradedMetrics, nil); got != "degraded" {
		t.Errorf("expected degraded, got %s", got)
	}

	// Multiple violations
	violatedMetrics := telesat.PerformanceMetrics{
		MaxLatencyMs:    200.0,
		MaxJitterMs:     50.0,
		PacketLossPct:   1.0,
		AvailabilityPct: 90.0,
	}
	if got := evaluateCompliance(violatedMetrics, nil); got != "violated" {
		t.Errorf("expected violated, got %s", got)
	}
}
