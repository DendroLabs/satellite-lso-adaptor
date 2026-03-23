package sonata

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/mapping"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/satellite"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

func setupTestServer() *http.ServeMux {
	tc := telesat.NewClient("http://mock", "")
	sc := satellite.NewClient("http://mock:8090")
	transformer := mapping.NewTransformer(tc, sc)
	mux := http.NewServeMux()
	RegisterRoutes(mux, transformer)
	return mux
}

// --- Product Catalog ---

func TestListProductOfferings(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/productCatalog/productOffering", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var offerings []map[string]any
	json.NewDecoder(w.Body).Decode(&offerings)
	if len(offerings) != 3 {
		t.Errorf("expected 3 offerings, got %d", len(offerings))
	}
}

func TestGetProductOffering(t *testing.T) {
	mux := setupTestServer()

	req := httptest.NewRequest("GET", "/mef/v4/productCatalog/productOffering/telesat-ls-eline", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var offering map[string]any
	json.NewDecoder(w.Body).Decode(&offering)
	if offering["id"] != "telesat-ls-eline" {
		t.Errorf("expected telesat-ls-eline, got %v", offering["id"])
	}
}

func TestGetProductOffering_NotFound(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/productCatalog/productOffering/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Product Inventory ---

func TestListProducts(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/productInventory/product", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []map[string]any
	json.NewDecoder(w.Body).Decode(&products)
	if len(products) == 0 {
		t.Error("expected at least one product")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/productInventory/product/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Quote Management ---

func TestCreateQuote(t *testing.T) {
	mux := setupTestServer()
	body := `{
		"buyerRequestedQuoteLevel": "firm",
		"instantSyncQuote": true,
		"quoteItem": [{
			"id": "1",
			"action": "add",
			"product": {
				"productOffering": {"id": "telesat-ls-eline"},
				"place": [
					{"role": "UNI-A", "latitude": 45.42, "longitude": -75.70},
					{"role": "UNI-Z", "latitude": 48.86, "longitude": 2.35}
				],
				"productConfiguration": {"@type": "AccessElineOvc", "bandwidthMbps": 500}
			},
			"requestedQuoteItemTerm": {"name": "36-month", "duration": {"amount": 36, "units": "months"}}
		}]
	}`

	req := httptest.NewRequest("POST", "/mef/v4/quote", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var quote map[string]any
	json.NewDecoder(w.Body).Decode(&quote)
	if quote["state"] != "approved.orderable" {
		t.Errorf("expected approved.orderable, got %v", quote["state"])
	}
}

func TestCreateQuote_EmptyItems(t *testing.T) {
	mux := setupTestServer()
	body := `{"buyerRequestedQuoteLevel": "firm", "quoteItem": []}`
	req := httptest.NewRequest("POST", "/mef/v4/quote", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateQuote_InvalidJSON(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("POST", "/mef/v4/quote", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListQuotes(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/quote", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetQuote_NotFound(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/quote/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Product Order Management ---

func TestCreateOrder(t *testing.T) {
	mux := setupTestServer()
	body := `{
		"productOrderItem": [{
			"id": "1",
			"action": "add",
			"product": {
				"productOffering": {"id": "telesat-ls-eline"},
				"productConfiguration": {"@type": "AccessElineOvc", "bandwidthMbps": 500}
			},
			"requestedCompletionDate": "2028-05-01T00:00:00Z"
		}]
	}`

	req := httptest.NewRequest("POST", "/mef/v4/productOrder", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	mux := setupTestServer()
	body := `{"productOrderItem": []}`
	req := httptest.NewRequest("POST", "/mef/v4/productOrder", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Trouble Ticket ---

func TestCreateTroubleTicket(t *testing.T) {
	mux := setupTestServer()
	body := `{
		"description": "Link down on Ottawa-Paris circuit",
		"severity": "extensive",
		"ticketType": "connectivityProblem",
		"priority": 0,
		"relatedEntity": [{"id": "mef-sub-001", "role": "affectedProduct"}]
	}`

	req := httptest.NewRequest("POST", "/mef/v4/troubleTicket", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTroubleTicket_MissingDescription(t *testing.T) {
	mux := setupTestServer()
	body := `{"severity": "minor", "ticketType": "informationRequest"}`
	req := httptest.NewRequest("POST", "/mef/v4/troubleTicket", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListTroubleTickets(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/troubleTicket", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var tickets []map[string]any
	json.NewDecoder(w.Body).Decode(&tickets)
	if len(tickets) != 3 {
		t.Errorf("expected 3 tickets, got %d", len(tickets))
	}
}

func TestGetTroubleTicket_NotFound(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/troubleTicket/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPatchTroubleTicket(t *testing.T) {
	mux := setupTestServer()
	body := `{"status": "resolved", "resolutionNote": "Fixed the issue"}`
	req := httptest.NewRequest("PATCH", "/mef/v4/troubleTicket/mef-tt-tt-002", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Performance Monitoring ---

func TestListPerformanceReports(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/performanceMonitoring/performanceReport", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var reports []map[string]any
	json.NewDecoder(w.Body).Decode(&reports)
	if len(reports) != 4 {
		t.Errorf("expected 4 reports, got %d", len(reports))
	}
}

func TestListPerformanceReports_Filtered(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/performanceMonitoring/performanceReport?productId=mef-sub-002", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var reports []map[string]any
	json.NewDecoder(w.Body).Decode(&reports)
	if len(reports) != 2 {
		t.Errorf("expected 2 reports for sub-002, got %d", len(reports))
	}
}

func TestGetPerformanceReport_NotFound(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/mef/v4/performanceMonitoring/performanceReport/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Extensions ---

func TestListTerminals(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/ext/v1/terminals", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListLandingStations(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/ext/v1/landingStations", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLatencyEstimate(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/ext/v1/latencyEstimate?latA=45.42&lonA=-75.70&latZ=48.86&lonZ=2.35", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["pt5Valid"] != true {
		t.Error("Ottawa-Paris should be PT5 valid")
	}
}

func TestLatencyEstimate_InvalidCoords(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/ext/v1/latencyEstimate?latA=abc&lonA=-75.70&latZ=48.86&lonZ=2.35", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Health ---

func TestHealthCheck(t *testing.T) {
	mux := setupTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}
