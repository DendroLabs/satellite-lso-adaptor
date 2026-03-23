package satellite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client communicates with the Python orbital mechanics service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CoverageRequest asks whether two geographic points can be connected via Lightspeed.
type CoverageRequest struct {
	LatA float64 `json:"latA"`
	LonA float64 `json:"lonA"`
	LatZ float64 `json:"latZ"`
	LonZ float64 `json:"lonZ"`
}

// CoverageResponse contains the orbital analysis result.
type CoverageResponse struct {
	Feasible           bool    `json:"feasible"`
	EstimatedLatencyMs float64 `json:"estimatedLatencyMs"`
	SatelliteHops      int     `json:"satelliteHops"`
	NearestLandingA    string  `json:"nearestLandingA"`
	NearestLandingZ    string  `json:"nearestLandingZ"`
	PathDescription    string  `json:"pathDescription"`
}

// CheckCoverage calls the Python orbital service to determine if a path is feasible.
func (c *Client) CheckCoverage(ctx context.Context, req CoverageRequest) (*CoverageResponse, error) {
	url := fmt.Sprintf("%s/coverage?latA=%f&lonA=%f&latZ=%f&lonZ=%f",
		c.baseURL, req.LatA, req.LonA, req.LatZ, req.LonZ)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating coverage request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling satellite service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("satellite service returned %d", resp.StatusCode)
	}

	var result CoverageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding coverage response: %w", err)
	}
	return &result, nil
}

// LatencyEstimate returns estimated one-way latency between two points.
func (c *Client) LatencyEstimate(ctx context.Context, latA, lonA, latZ, lonZ float64) (float64, error) {
	cov, err := c.CheckCoverage(ctx, CoverageRequest{
		LatA: latA, LonA: lonA,
		LatZ: latZ, LonZ: lonZ,
	})
	if err != nil {
		return 0, err
	}
	return cov.EstimatedLatencyMs, nil
}
