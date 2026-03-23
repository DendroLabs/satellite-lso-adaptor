package telesat

import "time"

func mockQuoteRequests() []QuoteRequest {
	return []QuoteRequest{
		{
			ID:              "tq-001",
			Status:          "priced",
			ServiceType:     "e-line",
			BandwidthMbps:   500,
			TermMonths:      36,
			EndpointA:       ServicePoint{Latitude: 45.4215, Longitude: -75.6972, LandingStation: "ls-gatineau", TerminalType: "intellian-aesa"},
			EndpointZ:       ServicePoint{Latitude: 48.8566, Longitude: 2.3522, LandingStation: "ls-france", TerminalType: "intellian-aesa"},
			MonthlyPriceUSD: 4500.00,
			SetupPriceUSD:   15000.00,
			ValidUntil:      time.Date(2028, 6, 30, 0, 0, 0, 0, time.UTC),
			CreatedAt:       time.Date(2028, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:              "tq-002",
			Status:          "priced",
			ServiceType:     "e-access",
			BandwidthMbps:   1000,
			TermMonths:      12,
			EndpointA:       ServicePoint{Latitude: 45.4215, Longitude: -75.6972, LandingStation: "ls-gatineau", TerminalType: "intellian-aesa"},
			EndpointZ:       ServicePoint{Latitude: 62.4540, Longitude: -114.3718, TerminalType: "farcast-fpa"},
			MonthlyPriceUSD: 8200.00,
			SetupPriceUSD:   22000.00,
			ValidUntil:      time.Date(2028, 5, 15, 0, 0, 0, 0, time.UTC),
			CreatedAt:       time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}

func mockOrderRequests() []OrderRequest {
	completionDate := time.Date(2028, 4, 10, 0, 0, 0, 0, time.UTC)
	return []OrderRequest{
		{
			ID:             "to-001",
			Status:         "active",
			QuoteID:        "tq-001",
			ServiceType:    "e-line",
			BandwidthMbps:  500,
			EndpointA:      ServicePoint{Latitude: 45.4215, Longitude: -75.6972, LandingStation: "ls-gatineau", TerminalType: "intellian-aesa"},
			EndpointZ:      ServicePoint{Latitude: 48.8566, Longitude: 2.3522, LandingStation: "ls-france", TerminalType: "intellian-aesa"},
			RequestedDate:  time.Date(2028, 3, 20, 0, 0, 0, 0, time.UTC),
			CompletionDate: &completionDate,
			CreatedAt:      time.Date(2028, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            "to-002",
			Status:        "provisioning",
			QuoteID:       "tq-002",
			ServiceType:   "e-access",
			BandwidthMbps: 1000,
			EndpointA:     ServicePoint{Latitude: 45.4215, Longitude: -75.6972, LandingStation: "ls-gatineau", TerminalType: "intellian-aesa"},
			EndpointZ:     ServicePoint{Latitude: 62.4540, Longitude: -114.3718, TerminalType: "farcast-fpa"},
			RequestedDate: time.Date(2028, 4, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:     time.Date(2028, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	}
}
