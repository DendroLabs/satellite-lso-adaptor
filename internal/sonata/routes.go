package sonata

import (
	"net/http"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/mapping"
)

// RegisterRoutes mounts the MEF LSO Sonata API endpoints on the given mux.
// These follow the MEF LSO Sonata SDK URL patterns.
func RegisterRoutes(mux *http.ServeMux, transformer *mapping.Transformer) {
	h := &handler{transformer: transformer}

	// Product Catalog (MEF 142)
	mux.HandleFunc("GET /mef/v4/productCatalog/productOffering", h.listProductOfferings)
	mux.HandleFunc("GET /mef/v4/productCatalog/productOffering/{id}", h.getProductOffering)

	// Serviceability / Product Offering Qualification (MEF 87.1)
	mux.HandleFunc("POST /mef/v4/serviceability/productOfferingQualification", h.createProductOfferingQualification)

	// Product Inventory (MEF 121.1)
	mux.HandleFunc("GET /mef/v4/productInventory/product", h.listProducts)
	mux.HandleFunc("GET /mef/v4/productInventory/product/{id}", h.getProduct)

	// Quote Management (MEF 115.1 / Mplify 115.1)
	mux.HandleFunc("POST /mef/v4/quote", h.createQuote)
	mux.HandleFunc("GET /mef/v4/quote", h.listQuotes)
	mux.HandleFunc("GET /mef/v4/quote/{id}", h.getQuote)

	// Product Order Management (MEF 123.1 / Mplify 123.1)
	mux.HandleFunc("POST /mef/v4/productOrder", h.createProductOrder)
	mux.HandleFunc("GET /mef/v4/productOrder", h.listProductOrders)
	mux.HandleFunc("GET /mef/v4/productOrder/{id}", h.getProductOrder)

	// Trouble Ticket Management (MEF 124.1)
	mux.HandleFunc("POST /mef/v4/troubleTicket", h.createTroubleTicket)
	mux.HandleFunc("GET /mef/v4/troubleTicket", h.listTroubleTickets)
	mux.HandleFunc("GET /mef/v4/troubleTicket/{id}", h.getTroubleTicket)
	mux.HandleFunc("PATCH /mef/v4/troubleTicket/{id}", h.patchTroubleTicket)

	// Performance Monitoring (MEF W133.1)
	mux.HandleFunc("GET /mef/v4/performanceMonitoring/performanceReport", h.listPerformanceReports)
	mux.HandleFunc("GET /mef/v4/performanceMonitoring/performanceReport/{id}", h.getPerformanceReport)

	// Health check
	mux.HandleFunc("GET /health", h.healthCheck)

	// Satellite-specific extensions (not part of MEF standard)
	mux.HandleFunc("GET /ext/v1/coverage", h.checkCoverage)
	mux.HandleFunc("GET /ext/v1/terminals", h.listTerminals)
	mux.HandleFunc("GET /ext/v1/landingStations", h.listLandingStations)
	mux.HandleFunc("GET /ext/v1/latencyEstimate", h.latencyEstimate)
}
