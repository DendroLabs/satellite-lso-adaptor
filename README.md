# Satellite LSO Adaptor

MEF LSO Sonata satellite adaptor for the Telesat Lightspeed LEO constellation.

Translates between [MEF LSO Sonata](https://github.com/MEF-GIT/MEF-LSO-Sonata-SDK) standard APIs and Telesat Lightspeed service APIs, enabling telco partners to integrate satellite Carrier Ethernet services using the same standards-based automation they use for terrestrial networks.

## Why This Exists

Telesat Lightspeed is a 198-satellite LEO constellation delivering MEF 3.0 certified Carrier Ethernet services. Telco partners (ISPs, MNOs, enterprise providers) need to integrate Lightspeed into their existing BSS/OSS ordering and provisioning workflows via MEF LSO Sonata APIs.

This adaptor bridges the gap between the MEF standard and satellite-specific concepts like orbital coverage, beam availability, terminal types, and the MEF 23.2.2 PT5 satellite performance tier.

## Architecture

```
Telco BSS/OSS                          Telesat Lightspeed
(MEF LSO Sonata APIs)                  (VNO / Service APIs)
        │                                      │
        ▼                                      ▼
┌─────────────────────────────────────────────────┐
│            Satellite LSO Adaptor                 │
│                                                 │
│  Go API Gateway (:8080)                         │
│  ├── Product Catalog Mapping                    │
│  ├── Serviceability / Coverage Qualification    │
│  ├── Quote Management (pricing, terms)          │
│  ├── Product Order Management (provisioning)    │
│  ├── Trouble Ticket Management (MEF 124.1)      │
│  ├── Performance Monitoring (MEF W133.1)        │
│  ├── Product Inventory Translation              │
│  ├── PT5 SLA Validation (MEF 23.2.2)           │
│  └── Satellite Extension APIs                   │
│                                                 │
│  Python Orbital Engine (:8090)                  │
│  ├── Constellation Model (198 sats, 27 planes) │
│  ├── Coverage Analysis                          │
│  └── Latency Estimation                         │
│                                                 │
│  YANG Models                                    │
│  ├── telesat-lightspeed-types                   │
│  └── telesat-lightspeed-access-eline            │
└─────────────────────────────────────────────────┘
```

## Quick Start

**Prerequisites:** Go 1.22+, Python 3.11+

```bash
# Start the Python orbital engine
python3 python/server.py &

# Start the Go adaptor
go run ./cmd/adaptor
```

The adaptor runs on `:8080`, the orbital engine on `:8090`.

### Docker

```bash
cd deployments
docker compose up
```

## API Endpoints

### MEF LSO Sonata Standard

| Method | Path | Description |
|--------|------|-------------|
| GET | `/mef/v4/productCatalog/productOffering` | List Lightspeed products as MEF offerings |
| GET | `/mef/v4/productCatalog/productOffering/{id}` | Get specific product offering |
| POST | `/mef/v4/serviceability/productOfferingQualification` | Check if a path can be served |
| GET | `/mef/v4/productInventory/product` | List active services as MEF products |
| GET | `/mef/v4/productInventory/product/{id}` | Get specific product |
| POST | `/mef/v4/quote` | Create a quote for a Lightspeed service |
| GET | `/mef/v4/quote` | List quotes |
| GET | `/mef/v4/quote/{id}` | Get specific quote with pricing |
| POST | `/mef/v4/productOrder` | Create a service provisioning order |
| GET | `/mef/v4/productOrder` | List product orders |
| GET | `/mef/v4/productOrder/{id}` | Get specific order with status |
| POST | `/mef/v4/troubleTicket` | Create a trouble ticket |
| GET | `/mef/v4/troubleTicket` | List trouble tickets |
| GET | `/mef/v4/troubleTicket/{id}` | Get specific trouble ticket |
| PATCH | `/mef/v4/troubleTicket/{id}` | Update trouble ticket status |
| GET | `/mef/v4/performanceMonitoring/performanceReport` | List performance reports |
| GET | `/mef/v4/performanceMonitoring/performanceReport/{id}` | Get report with PT5 compliance |

### Satellite Extensions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ext/v1/coverage?latA=&lonA=&latZ=&lonZ=` | Coverage feasibility check |
| GET | `/ext/v1/latencyEstimate?latA=&lonA=&latZ=&lonZ=` | Path latency estimate |
| GET | `/ext/v1/terminals` | Approved user terminals |
| GET | `/ext/v1/landingStations` | Ground landing stations |

### Example: Check Serviceability

```bash
curl -X POST http://localhost:8080/mef/v4/serviceability/productOfferingQualification \
  -H "Content-Type: application/json" \
  -d '{
    "productOfferingId": "telesat-ls-eline",
    "siteA": {"latitude": 45.42, "longitude": -75.70},
    "siteZ": {"latitude": 48.86, "longitude": 2.35},
    "requestedBandwidthMbps": 500,
    "requestedMaxLatencyMs": 50
  }'
```

Returns coverage qualification with PT5 validation, estimated latency (39ms for Ottawa-Paris), nearest landing stations, satellite hop count, and compatible terminal options.

### Example: Create a Quote

```bash
curl -X POST http://localhost:8080/mef/v4/quote \
  -H "Content-Type: application/json" \
  -d '{
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
        "productConfiguration": {
          "@type": "AccessElineOvc",
          "bandwidthMbps": 500
        }
      },
      "requestedQuoteItemTerm": {
        "name": "36-month",
        "duration": {"amount": 36, "units": "months"}
      }
    }]
  }'
```

Returns pricing with monthly recurring charge ($3,500/mo for 500 Mbps @ 36-month term), non-recurring setup charge, installation interval, and contract term details.

### Example: Create a Product Order

```bash
curl -X POST http://localhost:8080/mef/v4/productOrder \
  -H "Content-Type: application/json" \
  -d '{
    "productOrderItem": [{
      "id": "1",
      "action": "add",
      "product": {
        "productOffering": {"id": "telesat-ls-eline"},
        "place": [
          {"role": "UNI-A", "latitude": 45.42, "longitude": -75.70},
          {"role": "UNI-Z", "latitude": 48.86, "longitude": 2.35}
        ],
        "productConfiguration": {
          "@type": "AccessElineOvc",
          "bandwidthMbps": 500
        }
      },
      "requestedCompletionDate": "2028-05-01T00:00:00Z",
      "quoteItem": {"quoteId": "quote-tq-001", "id": "1"}
    }]
  }'
```

Returns order with provisioning status, expected completion date, and reference to the originating quote.

### Example: Create a Trouble Ticket

```bash
curl -X POST http://localhost:8080/mef/v4/troubleTicket \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Packet loss exceeding PT5 threshold on Ottawa-Paris E-Line",
    "severity": "significant",
    "ticketType": "performanceProblem",
    "priority": 1,
    "relatedEntity": [{
      "id": "mef-sub-001",
      "role": "affectedProduct",
      "@referredType": "Product"
    }]
  }'
```

Returns ticket with MEF-standard severity/status mapping, related product reference, satellite-specific ticket type classification (SLA violation, signal degradation, beam handover, link down), and status change history.

### Example: Query Performance Reports

```bash
# List all reports (optionally filter by product)
curl -s 'http://localhost:8080/mef/v4/performanceMonitoring/performanceReport?productId=mef-sub-001'

# Get detailed report with PT5 compliance assessment
curl -s http://localhost:8080/mef/v4/performanceMonitoring/performanceReport/mef-pr-pr-002
```

Returns per-interval performance measurements (latency, jitter, packet loss, throughput, availability) with PT5 compliance assessment per objective. Includes satellite-specific metrics: SNR, beam handovers, and ISL reroutes. Compliance summary: `compliant`, `degraded` (1 violation), or `violated` (2+ violations).

## Standards Alignment

- **MEF 106** — Access E-Line product specification
- **MEF 142** — Product Catalog API
- **MEF 87.1** — Product Offering Qualification (Serviceability)
- **MEF 115.1** — Quote Management API
- **MEF 121.1** — Product Inventory API
- **MEF 123.1** — Product Order Management API
- **MEF 124.1** — Trouble Ticket Management API
- **MEF W133.1** — Performance Monitoring API
- **MEF 23.2.2** — Satellite Performance Tier (PT5)
- **MEF 3.0** — Carrier Ethernet certification

## YANG Models

The `yang/` directory contains NETCONF/RESTCONF-compatible data models for Lightspeed satellite Carrier Ethernet services:

- `telesat-lightspeed-types.yang` — Common types: performance tiers, terminal types, SLA parameters, geographic coordinates
- `telesat-lightspeed-access-eline.yang` — Access E-Line service model with satellite path extensions

## Project Structure

```
├── cmd/adaptor/          # Go API gateway entry point
├── internal/
│   ├── sonata/           # MEF LSO Sonata HTTP handlers + routing
│   ├── telesat/          # Telesat API client, models, mock data
│   ├── mapping/          # API translation: catalog, serviceability, quote, order, ticket, perf, PT5
│   └── satellite/        # Python orbital engine client
├── python/
│   ├── orbital/          # Constellation model, coverage, latency
│   ├── tests/            # pytest test suite
│   └── server.py         # HTTP service wrapper
├── yang/                 # YANG modules
└── deployments/          # Docker, compose
```

## Testing

```bash
# Go tests (mapping layer + HTTP handlers)
go test ./... -v

# Python tests (orbital engine)
python3 -m pytest python/tests/ -v
```

## Roadmap

- [x] Product Catalog mapping (MEF 142)
- [x] Serviceability / Product Offering Qualification (MEF 87.1)
- [x] Product Inventory (MEF 121.1)
- [x] PT5 SLA validation (MEF 23.2.2)
- [x] Satellite coverage and latency estimation
- [x] YANG data models
- [x] Quote Management (MEF 115.1 / Mplify 115.1)
- [x] Product Order Management (MEF 123.1 / Mplify 123.1)
- [x] Trouble Ticket Management (MEF 124.1)
- [x] Performance Monitoring (MEF W133.1)
- [x] Unit and integration tests (Go + Python)

## License

Apache 2.0
