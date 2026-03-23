package telesat

import "time"

func mockTroubleTickets() []TroubleTicket {
	resolvedAt := time.Date(2028, 3, 18, 14, 30, 0, 0, time.UTC)
	return []TroubleTicket{
		{
			ID:             "tt-001",
			Status:         "resolved",
			Severity:       "major",
			Type:           "slaViolation",
			SubscriptionID: "sub-001",
			Summary:        "Latency exceeded PT5 threshold on Ottawa-Paris E-Line",
			Description:    "Observed one-way latency of 62ms on sub-001 for 15 minutes during beam handover event. PT5 maximum is 50ms. Affected traffic between 14:00-14:15 UTC.",
			AffectedSite:   "UNI-A",
			ResolutionNote: "Beam handover optimization applied. Routing updated to prefer lower-latency ISL path via Allan Park uplink.",
			CreatedAt:      time.Date(2028, 3, 18, 14, 5, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2028, 3, 18, 14, 30, 0, 0, time.UTC),
			ResolvedAt:     &resolvedAt,
		},
		{
			ID:             "tt-002",
			Status:         "inProgress",
			Severity:       "minor",
			Type:           "signalDegradation",
			SubscriptionID: "sub-002",
			Summary:        "Intermittent signal degradation at Yellowknife terminal",
			Description:    "FPA terminal at UNI-Z reporting 3dB SNR drop during high-elevation passes. Throughput reduced to 800 Mbps from 1000 Mbps CIR. Suspected thermal issue on terminal RF chain.",
			AffectedSite:   "UNI-Z",
			CreatedAt:      time.Date(2028, 3, 20, 9, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2028, 3, 21, 11, 0, 0, 0, time.UTC),
		},
		{
			ID:             "tt-003",
			Status:         "acknowledged",
			Severity:       "critical",
			Type:           "linkDown",
			SubscriptionID: "sub-001",
			Summary:        "Complete link loss on Ottawa-Paris circuit",
			Description:    "Total connectivity loss on sub-001. Both UNI-A and UNI-Z terminals reporting no satellite lock. Ground station telemetry shows nominal constellation status. Investigating terminal-side issue.",
			CreatedAt:      time.Date(2028, 3, 22, 7, 30, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2028, 3, 22, 7, 45, 0, 0, time.UTC),
		},
	}
}
