package mapping

import (
	"fmt"
	"math"
)

// PT5 Performance Tier constraints from MEF 23.2.2 satellite amendment.
// These define the maximum allowed values for satellite Carrier Ethernet services.
const (
	// Lightspeed constellation altitude in km
	ConstellationAltitudeKm = 1325.0

	// Speed of light in km/ms
	SpeedOfLightKmMs = 299.792

	// Atmospheric propagation adjustment factor (MEF 23.2.2)
	AtmosphericAdjustment = 0.164 // 16.4%

	// Signal propagation delay through atmosphere (ms/km) per MEF 23.2.2
	AtmosphericDelayMsPerKm = 0.0033

	// Optical ISL speed factor (slightly slower than vacuum c)
	ISLSpeedFactor = 0.97

	// Maximum one-way delay for PT5 (ms)
	PT5MaxOneWayDelay = 150.0

	// Maximum delay variation (jitter) for PT5 (ms)
	PT5MaxDelayVariation = 30.0

	// Maximum frame loss ratio for PT5 (percentage)
	PT5MaxFrameLossRatio = 0.1

	// Typical processing delay per satellite hop (ms)
	ProcessingDelayPerHop = 0.5
)

// PT5ValidationResult contains the result of validating service parameters against PT5.
type PT5ValidationResult struct {
	Valid            bool     `json:"valid"`
	EstimatedLatency float64  `json:"estimatedLatencyMs"`
	MaxAllowed       float64  `json:"maxAllowedLatencyMs"`
	Violations       []string `json:"violations,omitempty"`
}

// ValidatePT5 checks whether a requested service can meet PT5 performance objectives.
func ValidatePT5(latA, lonA, latZ, lonZ float64, requestedMaxLatency float64, satelliteHops int) PT5ValidationResult {
	result := PT5ValidationResult{
		Valid:      true,
		MaxAllowed: PT5MaxOneWayDelay,
	}

	estimatedLatency := EstimateOneWayLatency(latA, lonA, latZ, lonZ, satelliteHops)
	result.EstimatedLatency = estimatedLatency

	if estimatedLatency > PT5MaxOneWayDelay {
		result.Valid = false
		result.Violations = append(result.Violations,
			fmt.Sprintf("estimated latency %.1fms exceeds PT5 max %.1fms", estimatedLatency, PT5MaxOneWayDelay))
	}

	if requestedMaxLatency > 0 && estimatedLatency > requestedMaxLatency {
		result.Valid = false
		result.Violations = append(result.Violations,
			fmt.Sprintf("estimated latency %.1fms exceeds requested max %.1fms", estimatedLatency, requestedMaxLatency))
	}

	return result
}

// EstimateOneWayLatency calculates the estimated one-way latency for a Lightspeed path.
//
// The calculation models:
// 1. Uplink: ground → satellite (slant range based on elevation angle)
// 2. ISL hops: satellite → satellite via optical inter-satellite links
// 3. Downlink: satellite → ground (slant range)
// 4. Processing delay at each satellite hop
// 5. Atmospheric propagation adjustment per MEF 23.2.2
func EstimateOneWayLatency(latA, lonA, latZ, lonZ float64, satelliteHops int) float64 {
	// Ground distance between endpoints (great circle, km)
	groundDist := HaversineKm(latA, lonA, latZ, lonZ)

	// Minimum elevation angle for LEO at this altitude (typically ~40° for Lightspeed)
	elevAngle := 40.0
	slantRange := slantRangeKm(ConstellationAltitudeKm, elevAngle)

	// Uplink + downlink propagation (2 slant ranges)
	uplinkDownlinkMs := (2 * slantRange) / SpeedOfLightKmMs

	// Atmospheric delay for uplink + downlink
	atmosphericMs := 2 * slantRange * AtmosphericDelayMsPerKm * AtmosphericAdjustment

	// ISL propagation: approximate the inter-satellite path length
	// For LEO at 1325km, the ISL arc roughly follows the ground distance
	// but at orbital altitude
	islDistKm := float64(satelliteHops) * (groundDist / float64(max(satelliteHops, 1))) *
		(1 + ConstellationAltitudeKm/earthRadiusKm)
	islMs := islDistKm / (SpeedOfLightKmMs * ISLSpeedFactor)

	// Processing delay at each satellite
	processingMs := float64(satelliteHops) * ProcessingDelayPerHop

	return uplinkDownlinkMs + atmosphericMs + islMs + processingMs
}

// EstimateSatelliteHops estimates the number of ISL hops for a given ground distance.
// Lightspeed satellites have 4x 10 Gbps optical ISLs per satellite.
// Average inter-satellite spacing at 1325km altitude with 198 satellites in 27 planes
// is approximately 2000-3000km per hop.
func EstimateSatelliteHops(groundDistKm float64) int {
	avgHopDistKm := 2500.0
	hops := int(math.Ceil(groundDistKm / avgHopDistKm))
	if hops < 1 {
		hops = 1
	}
	return hops
}

const earthRadiusKm = 6371.0

// HaversineKm calculates the great-circle distance between two lat/lon points.
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// slantRangeKm calculates the distance from ground to satellite at a given elevation angle.
func slantRangeKm(altitudeKm, elevDeg float64) float64 {
	elevRad := toRad(elevDeg)
	r := earthRadiusKm
	h := altitudeKm
	return -r*math.Sin(elevRad) + math.Sqrt(r*r*math.Sin(elevRad)*math.Sin(elevRad)+2*r*h+h*h)
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}
