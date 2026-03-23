package mapping

import (
	"context"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/satellite"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

// MEF LSO Sonata Serviceability API models.
// Serviceability answers: "Can you deliver service between these two points?"

// ServiceabilityRequest represents an MEF Product Offering Qualification request.
type ServiceabilityRequest struct {
	ProductOfferingID string  `json:"productOfferingId"`
	SiteA             MEFSite `json:"siteA"`
	SiteZ             MEFSite `json:"siteZ"`
	RequestedBandwidth int    `json:"requestedBandwidthMbps,omitempty"`
	RequestedLatency   float64 `json:"requestedMaxLatencyMs,omitempty"`
}

// ServiceabilityResponse is the MEF Product Offering Qualification response.
type ServiceabilityResponse struct {
	ID                    string              `json:"id"`
	State                 string              `json:"state"` // done, inProgress, terminatedWithError
	Qualified             bool                `json:"qualified"`
	ProductOfferingID     string              `json:"productOfferingId"`
	ServiceConfidence     string              `json:"serviceConfidenceReason,omitempty"`
	AlternateProductIDs   []string            `json:"alternateProductOfferingIds,omitempty"`
	EstimatedLatencyMs    float64             `json:"estimatedLatencyMs,omitempty"`
	AvailableBandwidthMbps int               `json:"availableBandwidthMbps,omitempty"`
	PT5Validation         PT5ValidationResult `json:"pt5Validation"`
	Coverage              *CoverageDetail     `json:"coverageDetail,omitempty"`
	TerminalOptions       []telesat.Terminal  `json:"terminalOptions,omitempty"`
}

// CoverageDetail provides satellite-specific serviceability information
// that goes beyond standard MEF responses.
type CoverageDetail struct {
	NearestLandingStationA string  `json:"nearestLandingStationA"`
	NearestLandingStationZ string  `json:"nearestLandingStationZ"`
	EstimatedSatelliteHops int     `json:"estimatedSatelliteHops"`
	PathDescription        string  `json:"pathDescription"`
}

// CheckServiceability determines if Telesat Lightspeed can serve a requested connection.
func (t *Transformer) CheckServiceability(ctx context.Context, req ServiceabilityRequest) (*ServiceabilityResponse, error) {
	resp := &ServiceabilityResponse{
		ID:                "poq-" + req.ProductOfferingID,
		State:             "done",
		ProductOfferingID: req.ProductOfferingID,
	}

	// Check satellite coverage via Python orbital service
	covReq := satellite.CoverageRequest{
		LatA: req.SiteA.Lat, LonA: req.SiteA.Lon,
		LatZ: req.SiteZ.Lat, LonZ: req.SiteZ.Lon,
	}
	covResp, err := t.satellite.CheckCoverage(ctx, covReq)

	// If the orbital service is unavailable, fall back to geometric estimation
	if err != nil {
		groundDist := HaversineKm(req.SiteA.Lat, req.SiteA.Lon, req.SiteZ.Lat, req.SiteZ.Lon)
		hops := EstimateSatelliteHops(groundDist)
		pt5 := ValidatePT5(req.SiteA.Lat, req.SiteA.Lon, req.SiteZ.Lat, req.SiteZ.Lon,
			req.RequestedLatency, hops)

		resp.Qualified = pt5.Valid
		resp.EstimatedLatencyMs = pt5.EstimatedLatency
		resp.PT5Validation = pt5
		resp.Coverage = &CoverageDetail{
			EstimatedSatelliteHops: hops,
			PathDescription:        "Estimated via geometric model (orbital service unavailable)",
		}
	} else {
		pt5 := ValidatePT5(req.SiteA.Lat, req.SiteA.Lon, req.SiteZ.Lat, req.SiteZ.Lon,
			req.RequestedLatency, covResp.SatelliteHops)

		resp.Qualified = covResp.Feasible && pt5.Valid
		resp.EstimatedLatencyMs = covResp.EstimatedLatencyMs
		resp.PT5Validation = pt5
		resp.Coverage = &CoverageDetail{
			NearestLandingStationA: covResp.NearestLandingA,
			NearestLandingStationZ: covResp.NearestLandingZ,
			EstimatedSatelliteHops: covResp.SatelliteHops,
			PathDescription:        covResp.PathDescription,
		}
	}

	// Attach available terminal options
	terminals, _ := t.telesat.ListTerminals(ctx)
	resp.TerminalOptions = terminals

	if !resp.Qualified {
		resp.ServiceConfidence = "notAvailable"
	} else {
		resp.ServiceConfidence = "available"
	}

	return resp, nil
}
