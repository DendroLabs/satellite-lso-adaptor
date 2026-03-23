package telesat

import "time"

// Mock data based on publicly documented Telesat Lightspeed specifications.
// These will be replaced with real API calls when Telesat API access is available.

func mockPools() []VNOPool {
	return []VNOPool{
		{
			ID:            "vno-pool-001",
			Name:          "Enterprise Canada East",
			Status:        "active",
			AllocatedMbps: 2000,
			UsedMbps:      1200,
			CIR:           1500,
			Region:        "NA-EAST",
			TerminalType:  "intellian-aesa",
			CreatedAt:     time.Date(2028, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            "vno-pool-002",
			Name:          "Maritime Atlantic",
			Status:        "active",
			AllocatedMbps: 5000,
			UsedMbps:      3100,
			CIR:           3500,
			Region:        "ATL",
			TerminalType:  "farcast-fpa",
			CreatedAt:     time.Date(2028, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            "vno-pool-003",
			Name:          "Government Arctic",
			Status:        "active",
			AllocatedMbps: 7500,
			UsedMbps:      2000,
			CIR:           5000,
			Region:        "ARCTIC",
			TerminalType:  "intellian-aesa",
			CreatedAt:     time.Date(2028, 1, 20, 0, 0, 0, 0, time.UTC),
		},
	}
}

func mockSubscriptions() []Subscription {
	return []Subscription{
		{
			ID:            "sub-001",
			Status:        "active",
			ServiceType:   "e-line",
			BandwidthMbps: 500,
			CIR:           500,
			LatencyMs:     24.2,
			EndpointA: ServicePoint{
				Latitude:       45.4215,
				Longitude:      -75.6972,
				LandingStation: "ls-gatineau",
				TerminalType:   "intellian-aesa",
			},
			EndpointZ: ServicePoint{
				Latitude:       48.8566,
				Longitude:      2.3522,
				LandingStation: "ls-france",
				TerminalType:   "intellian-aesa",
			},
			SLA: SLAParameters{
				Availability:       99.9,
				MaxLatencyMs:       50.0,
				MaxJitterMs:        10.0,
				MaxPacketLossPct:   0.1,
				MeanTimeToRepairHr: 4.0,
			},
			CreatedAt: time.Date(2028, 2, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            "sub-002",
			Status:        "active",
			ServiceType:   "e-access",
			BandwidthMbps: 1000,
			CIR:           750,
			LatencyMs:     22.8,
			EndpointA: ServicePoint{
				Latitude:       45.4215,
				Longitude:      -75.6972,
				LandingStation: "ls-gatineau",
				TerminalType:   "intellian-aesa",
			},
			EndpointZ: ServicePoint{
				Latitude:       62.4540,
				Longitude:      -114.3718,
				TerminalType:   "farcast-fpa",
			},
			SLA: SLAParameters{
				Availability:       99.9,
				MaxLatencyMs:       40.0,
				MaxJitterMs:        8.0,
				MaxPacketLossPct:   0.05,
				MeanTimeToRepairHr: 4.0,
			},
			CreatedAt: time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}

// approvedTerminals returns the publicly documented approved terminal types.
func approvedTerminals() []Terminal {
	return []Terminal{
		{
			ID:            "term-intellian-aesa",
			Manufacturer:  "Intellian",
			Model:         "AESA LEO Terminal",
			Type:          "AESA",
			AntennaCm:     60,
			GainTemp:      19.0,
			EIRP:          60.0,
			MaxThroughput: 7500,
		},
		{
			ID:            "term-farcast-fpa",
			Manufacturer:  "Farcast",
			Model:         "Enterprise FPA",
			Type:          "FPA",
			AntennaCm:     50,
			GainTemp:      17.5,
			EIRP:          55.0,
			MaxThroughput: 5000,
		},
		{
			ID:            "term-thinkom-ka2517",
			Manufacturer:  "ThinKom",
			Model:         "ThinAir Ka2517",
			Type:          "VICTS",
			AntennaCm:     43,
			GainTemp:      16.0,
			EIRP:          52.0,
			MaxThroughput: 3000,
		},
		{
			ID:            "term-viasat-gm40",
			Manufacturer:  "Viasat",
			Model:         "GM-40",
			Type:          "dual-reflector",
			AntennaCm:     100,
			GainTemp:      19.0,
			EIRP:          60.0,
			MaxThroughput: 7500,
		},
	}
}

// knownLandingStations returns publicly documented Telesat Lightspeed ground stations.
func knownLandingStations() []LandingStation {
	return []LandingStation{
		{
			ID:        "ls-gatineau",
			Name:      "Gatineau Technical Operations Centre",
			Country:   "Canada",
			Latitude:  45.4765,
			Longitude: -75.7013,
			Status:    "operational",
		},
		{
			ID:        "ls-france",
			Name:      "France Landing Station",
			Country:   "France",
			Latitude:  48.8566,
			Longitude: 2.3522,
			Status:    "under-construction",
		},
		{
			ID:        "ls-australia",
			Name:      "Australia Landing Station",
			Country:   "Australia",
			Latitude:  -33.8688,
			Longitude: 151.2093,
			Status:    "under-construction",
		},
		{
			ID:        "ls-allan-park",
			Name:      "Allan Park Teleport",
			Country:   "Canada",
			Latitude:  44.2167,
			Longitude: -80.8167,
			Status:    "operational",
		},
	}
}
