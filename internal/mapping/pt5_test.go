package mapping

import (
	"math"
	"testing"
)

func TestHaversineKm(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		wantMin  float64
		wantMax  float64
	}{
		{
			name:    "Ottawa to Paris",
			lat1:    45.4215, lon1: -75.6972,
			lat2:    48.8566, lon2: 2.3522,
			wantMin: 5600, wantMax: 5700,
		},
		{
			name:    "Ottawa to Yellowknife",
			lat1:    45.4215, lon1: -75.6972,
			lat2:    62.4540, lon2: -114.3718,
			wantMin: 2800, wantMax: 3200,
		},
		{
			name:    "same point",
			lat1:    45.0, lon1: -75.0,
			lat2:    45.0, lon2: -75.0,
			wantMin: 0, wantMax: 0.001,
		},
		{
			name:    "antipodal points",
			lat1:    0.0, lon1: 0.0,
			lat2:    0.0, lon2: 180.0,
			wantMin: 20000, wantMax: 20100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HaversineKm(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("HaversineKm() = %.1f, want between %.1f and %.1f", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEstimateSatelliteHops(t *testing.T) {
	tests := []struct {
		name     string
		distKm   float64
		wantHops int
	}{
		{"short distance", 500, 1},
		{"zero distance", 0, 1},
		{"medium distance", 5000, 2},
		{"long distance", 10000, 4},
		{"very long distance", 20000, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateSatelliteHops(tt.distKm)
			if got != tt.wantHops {
				t.Errorf("EstimateSatelliteHops(%.0f) = %d, want %d", tt.distKm, got, tt.wantHops)
			}
		})
	}
}

func TestEstimateOneWayLatency(t *testing.T) {
	// Ottawa to Paris — should be roughly 25-45ms for LEO
	latency := EstimateOneWayLatency(45.4215, -75.6972, 48.8566, 2.3522, 2)
	if latency < 10 || latency > 80 {
		t.Errorf("Ottawa-Paris latency = %.1fms, want between 10 and 80ms", latency)
	}

	// Same city — should be very low, just uplink+downlink
	local := EstimateOneWayLatency(45.0, -75.0, 45.01, -75.01, 1)
	if local < 2 || local > 20 {
		t.Errorf("local latency = %.1fms, want between 2 and 20ms", local)
	}

	// More hops should mean more latency
	oneHop := EstimateOneWayLatency(45.0, -75.0, 50.0, -70.0, 1)
	threeHop := EstimateOneWayLatency(45.0, -75.0, 50.0, -70.0, 3)
	if threeHop <= oneHop {
		t.Errorf("3-hop latency (%.1fms) should exceed 1-hop (%.1fms)", threeHop, oneHop)
	}
}

func TestValidatePT5(t *testing.T) {
	tests := []struct {
		name               string
		latA, lonA         float64
		latZ, lonZ         float64
		requestedMax       float64
		hops               int
		wantValid          bool
		wantViolationCount int
	}{
		{
			name:               "Ottawa-Paris within PT5",
			latA: 45.42, lonA: -75.70,
			latZ: 48.86, lonZ: 2.35,
			hops:               2,
			wantValid:          true,
			wantViolationCount: 0,
		},
		{
			name:               "Ottawa-Paris with tight requested max",
			latA: 45.42, lonA: -75.70,
			latZ: 48.86, lonZ: 2.35,
			requestedMax:       10.0, // impossibly tight
			hops:               2,
			wantValid:          false,
			wantViolationCount: 1,
		},
		{
			name:               "antipodal path with tight requested max",
			latA: 0.0, lonA: 0.0,
			latZ: 0.0, lonZ: 180.0,
			requestedMax:       50.0, // tighter than the ~100ms actual
			hops:               8,
			wantValid:          false,
			wantViolationCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePT5(tt.latA, tt.lonA, tt.latZ, tt.lonZ, tt.requestedMax, tt.hops)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v (latency=%.1fms, max=%.1fms)",
					result.Valid, tt.wantValid, result.EstimatedLatency, result.MaxAllowed)
			}
			if len(result.Violations) != tt.wantViolationCount {
				t.Errorf("got %d violations, want %d: %v",
					len(result.Violations), tt.wantViolationCount, result.Violations)
			}
			if result.EstimatedLatency <= 0 {
				t.Error("estimated latency should be positive")
			}
			if math.IsNaN(result.EstimatedLatency) || math.IsInf(result.EstimatedLatency, 0) {
				t.Error("estimated latency should be finite")
			}
		})
	}
}
