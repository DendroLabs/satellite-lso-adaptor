package sonata

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/mapping"
)

type handler struct {
	transformer *mapping.Transformer
}

// --- Product Catalog Handlers ---

func (h *handler) listProductOfferings(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.transformer.GetProductCatalog(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve product catalog", err)
		return
	}
	writeJSON(w, http.StatusOK, catalog)
}

func (h *handler) getProductOffering(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	catalog, err := h.transformer.GetProductCatalog(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve product catalog", err)
		return
	}

	for _, offering := range catalog {
		if offering.ID == id {
			writeJSON(w, http.StatusOK, offering)
			return
		}
	}
	writeError(w, http.StatusNotFound, "product offering not found", nil)
}

// --- Serviceability Handlers ---

func (h *handler) createProductOfferingQualification(w http.ResponseWriter, r *http.Request) {
	var req mapping.ServiceabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	resp, err := h.transformer.CheckServiceability(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "serviceability check failed", err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// --- Product Inventory Handlers ---

func (h *handler) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.transformer.GetInventory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve inventory", err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	products, err := h.transformer.GetInventory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve inventory", err)
		return
	}

	for _, p := range products {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeError(w, http.StatusNotFound, "product not found", nil)
}

// --- Extension Handlers (satellite-specific) ---

func (h *handler) checkCoverage(w http.ResponseWriter, r *http.Request) {
	latA, lonA, latZ, lonZ, err := parseCoordinates(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid coordinates", err)
		return
	}

	req := mapping.ServiceabilityRequest{
		ProductOfferingID: "telesat-ls-eline",
		SiteA:             mapping.MEFSite{Lat: latA, Lon: lonA},
		SiteZ:             mapping.MEFSite{Lat: latZ, Lon: lonZ},
	}
	resp, err := h.transformer.CheckServiceability(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "coverage check failed", err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Coverage)
}

func (h *handler) listTerminals(w http.ResponseWriter, r *http.Request) {
	terminals, err := h.transformer.ListTerminals(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve terminals", err)
		return
	}
	writeJSON(w, http.StatusOK, terminals)
}

func (h *handler) listLandingStations(w http.ResponseWriter, r *http.Request) {
	stations, err := h.transformer.ListLandingStations(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve landing stations", err)
		return
	}
	writeJSON(w, http.StatusOK, stations)
}

func (h *handler) latencyEstimate(w http.ResponseWriter, r *http.Request) {
	latA, lonA, latZ, lonZ, err := parseCoordinates(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid coordinates", err)
		return
	}

	groundDist := mapping.HaversineKm(latA, lonA, latZ, lonZ)
	hops := mapping.EstimateSatelliteHops(groundDist)
	pt5 := mapping.ValidatePT5(latA, lonA, latZ, lonZ, 0, hops)

	writeJSON(w, http.StatusOK, map[string]any{
		"estimatedLatencyMs": pt5.EstimatedLatency,
		"satelliteHops":      hops,
		"groundDistanceKm":   groundDist,
		"pt5Valid":           pt5.Valid,
		"pt5MaxAllowedMs":   pt5.MaxAllowed,
	})
}

// --- Quote Management Handlers (MEF 115.1) ---

func (h *handler) createQuote(w http.ResponseWriter, r *http.Request) {
	var req mapping.MEFQuoteCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if len(req.QuoteItem) == 0 {
		writeError(w, http.StatusBadRequest, "at least one quoteItem is required", nil)
		return
	}

	quote, err := h.transformer.CreateQuote(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "quote creation failed", err)
		return
	}
	writeJSON(w, http.StatusCreated, quote)
}

func (h *handler) listQuotes(w http.ResponseWriter, r *http.Request) {
	quotes, err := h.transformer.ListQuotes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve quotes", err)
		return
	}
	writeJSON(w, http.StatusOK, quotes)
}

func (h *handler) getQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	quote, err := h.transformer.GetQuote(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "quote not found", err)
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

// --- Product Order Management Handlers (MEF 123.1) ---

func (h *handler) createProductOrder(w http.ResponseWriter, r *http.Request) {
	var req mapping.MEFProductOrderCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if len(req.ProductOrderItem) == 0 {
		writeError(w, http.StatusBadRequest, "at least one productOrderItem is required", nil)
		return
	}

	order, err := h.transformer.CreateOrder(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order creation failed", err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *handler) listProductOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.transformer.ListOrders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve orders", err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *handler) getProductOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order, err := h.transformer.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found", err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// --- Trouble Ticket Management Handlers (MEF 124.1) ---

func (h *handler) createTroubleTicket(w http.ResponseWriter, r *http.Request) {
	var req mapping.MEFTroubleTicketCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "description is required", nil)
		return
	}

	ticket, err := h.transformer.CreateTicket(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "trouble ticket creation failed", err)
		return
	}
	writeJSON(w, http.StatusCreated, ticket)
}

func (h *handler) listTroubleTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.transformer.ListTickets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve trouble tickets", err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (h *handler) getTroubleTicket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ticket, err := h.transformer.GetTicket(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "trouble ticket not found", err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (h *handler) patchTroubleTicket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var update mapping.MEFTroubleTicketUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	ticket, err := h.transformer.PatchTicket(r.Context(), id, update)
	if err != nil {
		writeError(w, http.StatusNotFound, "trouble ticket not found", err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

// --- Performance Monitoring Handlers (MEF W133.1) ---

func (h *handler) listPerformanceReports(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.URL.Query().Get("productId")
	reports, err := h.transformer.ListPerformanceReports(r.Context(), subscriptionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve performance reports", err)
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func (h *handler) getPerformanceReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	report, err := h.transformer.GetPerformanceReport(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "performance report not found", err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// --- Health ---

func (h *handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "satellite-lso-adaptor"})
}

// --- Helpers ---

func parseCoordinates(r *http.Request) (latA, lonA, latZ, lonZ float64, err error) {
	latA, err = strconv.ParseFloat(r.URL.Query().Get("latA"), 64)
	if err != nil {
		return
	}
	lonA, err = strconv.ParseFloat(r.URL.Query().Get("lonA"), 64)
	if err != nil {
		return
	}
	latZ, err = strconv.ParseFloat(r.URL.Query().Get("latZ"), 64)
	if err != nil {
		return
	}
	lonZ, err = strconv.ParseFloat(r.URL.Query().Get("lonZ"), 64)
	return
}

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		slog.Error(msg, "err", err)
	}
	writeJSON(w, status, apiError{Error: http.StatusText(status), Message: msg})
}
