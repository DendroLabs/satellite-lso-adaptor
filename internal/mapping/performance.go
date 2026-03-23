package mapping

import (
	"context"
	"fmt"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// MEF LSO Sonata Performance Monitoring models
// Aligned with MEF W133.1 (Performance Monitoring) concepts.

// MEFPerformanceReport represents a performance report in MEF format.
type MEFPerformanceReport struct {
	ID                    string                  `json:"id"`
	Href                  string                  `json:"href"`
	State                 string                  `json:"state"` // completed, inProgress, failed
	Description           string                  `json:"description"`
	ReportingPeriod       string                  `json:"reportingPeriod"` // 15min, 1hour, 24hour
	GranularityPeriod     string                  `json:"granularityPeriod"`
	ReportingStartDate    time.Time               `json:"reportingStartDate"`
	ReportingEndDate      time.Time               `json:"reportingEndDate"`
	ServiceRef            ServiceRef              `json:"serviceRef"`
	PerformanceObjective  []PerformanceObjective  `json:"performanceObjective"`
	ResultPayload         ResultPayload           `json:"resultPayload"`
}

// MEFPerformanceReportFind is the abbreviated report returned in list responses.
type MEFPerformanceReportFind struct {
	ID                 string    `json:"id"`
	Href               string    `json:"href"`
	State              string    `json:"state"`
	ReportingPeriod    string    `json:"reportingPeriod"`
	ReportingStartDate time.Time `json:"reportingStartDate"`
	ReportingEndDate   time.Time `json:"reportingEndDate"`
	ServiceRef         ServiceRef `json:"serviceRef"`
	ComplianceSummary  string    `json:"complianceSummary"` // compliant, degraded, violated
}

// ServiceRef references the service being measured.
type ServiceRef struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Name string `json:"name,omitempty"`
}

// PerformanceObjective defines a target metric and whether it was met.
type PerformanceObjective struct {
	ObjectiveName  string  `json:"objectiveName"`
	Metric         string  `json:"metric"`
	ThresholdValue float64 `json:"thresholdValue"`
	MeasuredValue  float64 `json:"measuredValue"`
	Unit           string  `json:"unit"`
	Compliant      bool    `json:"compliant"`
}

// ResultPayload contains the full performance measurement results.
type ResultPayload struct {
	Latency      LatencyResult      `json:"latency"`
	Jitter       JitterResult       `json:"jitter"`
	PacketLoss   PacketLossResult   `json:"packetLoss"`
	Throughput   ThroughputResult   `json:"throughput"`
	Availability AvailabilityResult `json:"availability"`
	Satellite    SatelliteResult    `json:"satellite"`
}

// LatencyResult contains latency measurements.
type LatencyResult struct {
	AvgOneWayDelayMs float64 `json:"avgOneWayDelayMs"`
	MaxOneWayDelayMs float64 `json:"maxOneWayDelayMs"`
	MinOneWayDelayMs float64 `json:"minOneWayDelayMs"`
	PercentileP95Ms  float64 `json:"percentileP95Ms"`
}

// JitterResult contains jitter (delay variation) measurements.
type JitterResult struct {
	AvgDelayVariationMs float64 `json:"avgDelayVariationMs"`
	MaxDelayVariationMs float64 `json:"maxDelayVariationMs"`
}

// PacketLossResult contains packet loss measurements.
type PacketLossResult struct {
	FrameLossRatioPct float64 `json:"frameLossRatioPct"`
}

// ThroughputResult contains throughput measurements.
type ThroughputResult struct {
	AvgThroughputMbps float64 `json:"avgThroughputMbps"`
	MaxThroughputMbps float64 `json:"maxThroughputMbps"`
	UtilizationPct    float64 `json:"utilizationPct"`
}

// AvailabilityResult contains service availability measurements.
type AvailabilityResult struct {
	AvailabilityPct float64 `json:"availabilityPct"`
}

// SatelliteResult contains satellite-specific performance metrics.
type SatelliteResult struct {
	SNRdB         float64 `json:"snrDb"`
	BeamHandovers int     `json:"beamHandovers"`
	ISLReroutes   int     `json:"islReroutes"`
}

// --- Performance Monitoring Transformation Methods ---

// ListPerformanceReports returns performance reports in MEF format.
func (t *Transformer) ListPerformanceReports(ctx context.Context, subscriptionID string) ([]MEFPerformanceReportFind, error) {
	// Strip MEF prefix if present
	telesatSubID := subscriptionID
	if len(subscriptionID) > 4 && subscriptionID[:4] == "mef-" {
		telesatSubID = subscriptionID[4:]
	}

	reports, err := t.telesat.ListPerformanceReports(ctx, telesatSubID)
	if err != nil {
		return nil, fmt.Errorf("fetching performance reports: %w", err)
	}

	results := make([]MEFPerformanceReportFind, 0, len(reports))
	for _, r := range reports {
		sub, _ := t.telesat.GetSubscription(ctx, r.SubscriptionID)
		compliance := evaluateCompliance(r.Metrics, sub)

		results = append(results, MEFPerformanceReportFind{
			ID:                 fmt.Sprintf("mef-pr-%s", r.ID),
			Href:               fmt.Sprintf("/mef/v4/performanceMonitoring/performanceReport/mef-pr-%s", r.ID),
			State:              mapReportStatus(r.Status),
			ReportingPeriod:    r.ReportPeriod,
			ReportingStartDate: r.StartTime,
			ReportingEndDate:   r.EndTime,
			ServiceRef: ServiceRef{
				ID:   fmt.Sprintf("mef-%s", r.SubscriptionID),
				Href: fmt.Sprintf("/mef/v4/productInventory/product/mef-%s", r.SubscriptionID),
			},
			ComplianceSummary: compliance,
		})
	}
	return results, nil
}

// GetPerformanceReport returns a single performance report in full MEF format.
func (t *Transformer) GetPerformanceReport(ctx context.Context, id string) (*MEFPerformanceReport, error) {
	telesatID := id
	if len(id) > 7 && id[:7] == "mef-pr-" {
		telesatID = id[7:]
	}

	report, err := t.telesat.GetPerformanceReport(ctx, telesatID)
	if err != nil {
		return nil, fmt.Errorf("performance report not found: %w", err)
	}

	sub, _ := t.telesat.GetSubscription(ctx, report.SubscriptionID)
	return t.telesatReportToMEF(report, sub), nil
}

// telesatReportToMEF converts a Telesat performance report to full MEF format.
func (t *Transformer) telesatReportToMEF(report *telesat.PerformanceReport, sub *telesat.Subscription) *MEFPerformanceReport {
	reportID := fmt.Sprintf("mef-pr-%s", report.ID)
	m := report.Metrics

	// Build PT5 performance objectives
	objectives := buildPT5Objectives(m, sub)

	// Estimate P95 latency (approximate from avg/max)
	p95Latency := m.AvgLatencyMs + (m.MaxLatencyMs-m.AvgLatencyMs)*0.8

	mefReport := &MEFPerformanceReport{
		ID:                 reportID,
		Href:               fmt.Sprintf("/mef/v4/performanceMonitoring/performanceReport/%s", reportID),
		State:              mapReportStatus(report.Status),
		Description:        fmt.Sprintf("Performance report for %s (%s period)", report.SubscriptionID, report.ReportPeriod),
		ReportingPeriod:    report.ReportPeriod,
		GranularityPeriod:  report.ReportPeriod,
		ReportingStartDate: report.StartTime,
		ReportingEndDate:   report.EndTime,
		ServiceRef: ServiceRef{
			ID:   fmt.Sprintf("mef-%s", report.SubscriptionID),
			Href: fmt.Sprintf("/mef/v4/productInventory/product/mef-%s", report.SubscriptionID),
		},
		PerformanceObjective: objectives,
		ResultPayload: ResultPayload{
			Latency: LatencyResult{
				AvgOneWayDelayMs: m.AvgLatencyMs,
				MaxOneWayDelayMs: m.MaxLatencyMs,
				MinOneWayDelayMs: m.MinLatencyMs,
				PercentileP95Ms:  p95Latency,
			},
			Jitter: JitterResult{
				AvgDelayVariationMs: m.AvgJitterMs,
				MaxDelayVariationMs: m.MaxJitterMs,
			},
			PacketLoss: PacketLossResult{
				FrameLossRatioPct: m.PacketLossPct,
			},
			Throughput: ThroughputResult{
				AvgThroughputMbps: m.AvgThroughputMbps,
				MaxThroughputMbps: m.MaxThroughputMbps,
				UtilizationPct:    (m.AvgThroughputMbps / m.MaxThroughputMbps) * 100,
			},
			Availability: AvailabilityResult{
				AvailabilityPct: m.AvailabilityPct,
			},
			Satellite: SatelliteResult{
				SNRdB:         m.SNRdB,
				BeamHandovers: m.BeamHandovers,
				ISLReroutes:   m.ISLReroutes,
			},
		},
	}

	return mefReport
}

// buildPT5Objectives creates performance objectives based on the subscription's SLA and PT5 thresholds.
func buildPT5Objectives(m telesat.PerformanceMetrics, sub *telesat.Subscription) []PerformanceObjective {
	maxLatency := PT5MaxOneWayDelay
	maxJitter := PT5MaxDelayVariation
	maxLoss := PT5MaxFrameLossRatio
	minAvailability := 99.9

	if sub != nil {
		maxLatency = sub.SLA.MaxLatencyMs
		maxJitter = sub.SLA.MaxJitterMs
		maxLoss = sub.SLA.MaxPacketLossPct
		minAvailability = sub.SLA.Availability
	}

	return []PerformanceObjective{
		{
			ObjectiveName:  "PT5 One-Way Delay",
			Metric:         "oneWayDelay",
			ThresholdValue: maxLatency,
			MeasuredValue:  m.MaxLatencyMs,
			Unit:           "ms",
			Compliant:      m.MaxLatencyMs <= maxLatency,
		},
		{
			ObjectiveName:  "PT5 Delay Variation",
			Metric:         "delayVariation",
			ThresholdValue: maxJitter,
			MeasuredValue:  m.MaxJitterMs,
			Unit:           "ms",
			Compliant:      m.MaxJitterMs <= maxJitter,
		},
		{
			ObjectiveName:  "PT5 Frame Loss Ratio",
			Metric:         "frameLossRatio",
			ThresholdValue: maxLoss,
			MeasuredValue:  m.PacketLossPct,
			Unit:           "percent",
			Compliant:      m.PacketLossPct <= maxLoss,
		},
		{
			ObjectiveName:  "Service Availability",
			Metric:         "availability",
			ThresholdValue: minAvailability,
			MeasuredValue:  m.AvailabilityPct,
			Unit:           "percent",
			Compliant:      m.AvailabilityPct >= minAvailability,
		},
	}
}

// evaluateCompliance checks if metrics meet SLA thresholds.
func evaluateCompliance(m telesat.PerformanceMetrics, sub *telesat.Subscription) string {
	objectives := buildPT5Objectives(m, sub)
	violations := 0
	for _, o := range objectives {
		if !o.Compliant {
			violations++
		}
	}
	switch {
	case violations == 0:
		return "compliant"
	case violations == 1:
		return "degraded"
	default:
		return "violated"
	}
}

func mapReportStatus(status string) string {
	switch status {
	case "complete":
		return "completed"
	case "partial":
		return "inProgress"
	default:
		return "completed"
	}
}
