package telesat

import "time"

// VNOPool represents a Telesat Lightspeed Virtual Network Operator capacity pool.
// This maps to the dedicated capacity allocation that VNO customers manage independently.
type VNOPool struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Status           string    `json:"status"` // active, provisioning, suspended
	AllocatedMbps    int       `json:"allocatedMbps"`
	UsedMbps         int       `json:"usedMbps"`
	CIR              int       `json:"cir"` // Committed Information Rate (Mbps)
	Region           string    `json:"region"`
	TerminalType     string    `json:"terminalType"`
	CreatedAt        time.Time `json:"createdAt"`
}

// Subscription represents a Telesat Lightspeed point-to-point subscription plan.
type Subscription struct {
	ID              string         `json:"id"`
	Status          string         `json:"status"` // active, provisioning, suspended, terminated
	ServiceType     string         `json:"serviceType"` // e-line, e-access
	BandwidthMbps   int            `json:"bandwidthMbps"`
	CIR             int            `json:"cir"`
	LatencyMs       float64        `json:"latencyMs"`
	EndpointA       ServicePoint   `json:"endpointA"`
	EndpointZ       ServicePoint   `json:"endpointZ"`
	SLA             SLAParameters  `json:"sla"`
	CreatedAt       time.Time      `json:"createdAt"`
}

// ServicePoint represents one end of a Telesat Lightspeed circuit.
type ServicePoint struct {
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	LandingStation string  `json:"landingStation,omitempty"` // nearest landing station ID
	TerminalID     string  `json:"terminalId,omitempty"`
	TerminalType   string  `json:"terminalType,omitempty"`
}

// SLAParameters defines the service level agreement for a Lightspeed service.
// Aligned with MEF 23.2.2 Performance Tier 5 (satellite).
type SLAParameters struct {
	Availability       float64 `json:"availability"`       // e.g., 99.9
	MaxLatencyMs       float64 `json:"maxLatencyMs"`       // PT5 constraint
	MaxJitterMs        float64 `json:"maxJitterMs"`        // PT5 constraint
	MaxPacketLossPct   float64 `json:"maxPacketLossPct"`   // PT5 constraint
	MeanTimeToRepairHr float64 `json:"meanTimeToRepairHr"`
}

// Terminal represents an approved Telesat Lightspeed user terminal.
type Terminal struct {
	ID           string  `json:"id"`
	Manufacturer string  `json:"manufacturer"` // Intellian, Farcast, ThinKom, Viasat
	Model        string  `json:"model"`
	Type         string  `json:"type"`         // AESA, FPA, dual-reflector
	AntennaCm    int     `json:"antennaCm"`    // antenna size in cm
	GainTemp     float64 `json:"gainTemp"`     // G/T in dB/K
	EIRP         float64 `json:"eirp"`         // EIRP in dBW
	MaxThroughput int    `json:"maxThroughput"` // max Mbps
}

// CoverageResult is returned by the satellite context service.
type CoverageResult struct {
	Feasible          bool      `json:"feasible"`
	EstimatedLatencyMs float64  `json:"estimatedLatencyMs"`
	SatelliteHops     int       `json:"satelliteHops"` // number of ISL hops
	NearestLandingA   string    `json:"nearestLandingA"`
	NearestLandingZ   string    `json:"nearestLandingZ"`
	BeamIDA           string    `json:"beamIdA,omitempty"`
	BeamIDZ           string    `json:"beamIdZ,omitempty"`
	PathDescription   string    `json:"pathDescription"`
}

// TroubleTicket represents a trouble ticket in the Telesat operations system.
type TroubleTicket struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"` // submitted, acknowledged, inProgress, resolved, closed, cancelled
	Severity        string    `json:"severity"` // critical, major, minor, informational
	Type            string    `json:"type"` // signalDegradation, beamHandover, terminalFault, slaViolation, linkDown, interference
	SubscriptionID  string    `json:"subscriptionId"`
	Summary         string    `json:"summary"`
	Description     string    `json:"description"`
	AffectedSite    string    `json:"affectedSite,omitempty"` // UNI-A or UNI-Z
	ResolutionNote  string    `json:"resolutionNote,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	ResolvedAt      *time.Time `json:"resolvedAt,omitempty"`
}

// PerformanceReport represents a performance measurement report from Telesat's monitoring system.
type PerformanceReport struct {
	ID             string              `json:"id"`
	SubscriptionID string              `json:"subscriptionId"`
	ReportPeriod   string              `json:"reportPeriod"` // 15min, 1hour, 24hour
	StartTime      time.Time           `json:"startTime"`
	EndTime        time.Time           `json:"endTime"`
	Metrics        PerformanceMetrics  `json:"metrics"`
	Status         string              `json:"status"` // complete, partial
}

// PerformanceMetrics contains measured network performance values.
type PerformanceMetrics struct {
	AvgLatencyMs       float64 `json:"avgLatencyMs"`
	MaxLatencyMs       float64 `json:"maxLatencyMs"`
	MinLatencyMs       float64 `json:"minLatencyMs"`
	AvgJitterMs        float64 `json:"avgJitterMs"`
	MaxJitterMs        float64 `json:"maxJitterMs"`
	PacketLossPct      float64 `json:"packetLossPct"`
	AvgThroughputMbps  float64 `json:"avgThroughputMbps"`
	MaxThroughputMbps  float64 `json:"maxThroughputMbps"`
	AvailabilityPct    float64 `json:"availabilityPct"`
	SNRdB              float64 `json:"snrDb"`
	BeamHandovers      int     `json:"beamHandovers"`
	ISLReroutes        int     `json:"islReroutes"`
}

// LandingStation represents a Telesat ground landing station with fiber backhaul.
type LandingStation struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"` // operational, planned, under-construction
}
