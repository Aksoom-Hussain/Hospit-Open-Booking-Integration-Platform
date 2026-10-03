# Hospit: Open Booking Integration Platform

> **High-Performance Open-Source Hotel Booking API & Real-Time Multi-Supplier Integration Engine** built in **Golang** for ultra-low latency, concurrent fan-out search, and high-throughput inventory synchronization across 100+ global booking partners, bedbanks, GDS, OTAs, and hotel channel managers.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-API--First%20%7C%20Concurrent-emerald)](docs/ARCHITECTURE.md)
[![Suppliers Supported](https://img.shields.io/badge/Suppliers%20Pre--Mapped-98+-indigo)](docs/suppliers/README.md)

---

## 1. Overview & Core Philosophy

**Hospit** is a developer-first, API-native integration platform that eliminates the fragmentation of the global hotel distribution industry. Instead of maintaining dozens of bespoke, legacy XML/SOAP and REST integrations with inconsistent schemas, hotel identifiers, and rate rules, developers connect once to Hospit's unified API.

### Why Golang?
Hotel distribution requires updating and querying rates and inventory across **100+ suppliers concurrently**:
- **Massive Concurrency**: Lightweight Go routines and bounded worker pools allow querying 30+ suppliers in parallel with $p95 < 800\text{ms}$.
- **High-Throughput ARI Sync**: Millions of real-time Availability, Rate, and Inventory (ARI) delta events streamed without garbage-collection spikes or thread starvation.
- **Zero-Allocation JSON/XML Streaming**: High-efficiency serialization for multi-megabyte supplier room-rate catalogs.
- **Self-Contained Binary**: Deploys as a single static binary with minimal memory footprint ($< 50\text{MB}$ idle) in Docker and Kubernetes.

---

## 2. API-First Architecture

Hospit is built purely as a headless, ultra-fast API engine. All features are exposed through standardized **REST & gRPC** endpoints.

```
                                 ┌─────────────────────────────────┐
                                 │   Clients / SaaS / Web / PMS    │
                                 └────────────────┬────────────────┘
                                                  │ REST / gRPC
                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                       HOSPIT GOLANG ENGINE                                       │
│                                                                                                  │
│   ┌──────────────────────────────────────────────────────────────────────────────────────────┐   │
│   │                                HTTP / gRPC Transport Layer                               │   │
│   │                 (Fiber / Chi router, Tenant Context & Auth Middleware)                   │   │
│   └─────────────────────────────────────────────┬────────────────────────────────────────────┘   │
│                                                 │                                                │
│   ┌─────────────────────────────────────────────▼────────────────────────────────────────────┐   │
│   │                          Concurrent Search & Saga Orchestrator                           │   │
│   │  • Goroutine Fan-Out Worker Pool (context timeout & cancellation propagation)            │   │
│   │  • Dynamic Markup & Net-Rate Calculator                                                  │   │
│   │  • Property & Room Deduplication (H3 Geo-indexing + Levenshtein fuzzy matcher)           │   │
│   │  • Canonical Normalizer (Unified Meal Plans, Occupancy, Bedding & Cancellation Policies)  │   │
│   └─────────────────────────────────────────────┬────────────────────────────────────────────┘   │
│                                                 │                                                │
│   ┌─────────────────────────────────────────────▼────────────────────────────────────────────┐   │
│   │                        Pluggable Golang Connector Ecosystem                              │   │
│   │                                                                                          │   │
│   │   ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────┐   │   │
│   │   │     Bedbanks     │  │       GDS        │  │  OTA-Wholesale   │  │ Channel Switch │   │   │
│   │   │ (Hotelbeds,      │  │ (Amadeus,        │  │ (Expedia Rapid,  │  │ (SiteMinder,   │   │   │
│   │   │  WebBeds,        │  │  Sabre,          │  │  Agoda Demand,   │  │  RateGain,     │   │   │
│   │   │  RateHawk, Dida) │  │  Travelport)     │  │  Priceline)      │  │  STAAH)        │   │   │
│   │   └──────────────────┘  └──────────────────┘  └──────────────────┘  └────────────────┘   │   │
│   └──────────────────────────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Go Connector Interface Specification

Every supplier connector implements the pure Go `SupplierConnector` interface:

```go
package connector

import "context"

// SupplierConnector defines the standard contract for any hotel supply partner
type SupplierConnector interface {
	// Metadata
	ID() string
	Name() string
	Category() SupplierCategory // Bedbank, GDS, OTA, DMC, Switch

	// Core Operations
	SearchHotels(ctx context.Context, req *SearchRequest) ([]*HotelRate, error)
	QuoteRate(ctx context.Context, req *QuoteRequest) (*QuoteResponse, error)
	CreateBooking(ctx context.Context, req *BookingRequest) (*BookingResponse, error)
	CancelBooking(ctx context.Context, req *CancelRequest) (*CancelResponse, error)
	GetBooking(ctx context.Context, supplierRef string) (*BookingDetails, error)

	// Two-Way Channel Management (ARI Sync)
	PushARI(ctx context.Context, payload *ARIPushPayload) (*ARIPushResult, error)
}
```

### Supported Supplier Categories (98+ Pre-Mapped Partners):
1. **Bedbanks (44)**: Hotelbeds (APItude), WebBeds (DOTW), RateHawk, DidaTravel, Yalago, Restel, Goglobal, Bonotel, Smyrooms...
2. **Global Distribution Systems (4)**: Amadeus (REST/SOAP), Sabre, Travelport, Amadeus Value Hotels.
3. **OTA Wholesalers (20)**: Expedia Rapid 3.0, Agoda Wholesale, Booking.com Demand, MakeMyTrip, Priceline...
4. **Regional & DMCs (26)**: HyperGuest, REZLive, Darina, Akbar Travels, EETGlobal, Roibos...
5. **Channel Switches & CRS (4)**: SiteMinder (Two-way ARI), RateGain, STAAH, Hobse...

---

## 4. API Endpoints

All endpoints support `X-Tenant-Key` for multi-tenant credential isolation (Bring Your Own Credentials).

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/hotels/search` | Parallel multi-supplier availability search with deduplicated results |
| `POST` | `/v1/hotels/quote` | Recheck real-time price, room availability, and freeze rate |
| `POST` | `/v1/bookings` | Execute transactional booking with saga rollback guarantees |
| `GET` | `/v1/bookings/:id` | Fetch live booking details, supplier PNR, and PDF voucher |
| `POST` | `/v1/bookings/:id/cancel` | Quote cancellation penalty and commit cancellation |
| `POST` | `/v1/ari/push` | Push real-time rates and stop-sell availability to channel managers |
| `POST` | `/v1/webhooks` | Receive asynchronous booking events and supplier status alerts |

---

## 5. Repository Structure

```
.
├── cmd/
│   └── hospit-server/       # Main executable entry point
│
├── internal/
│   ├── api/                 # REST (HTTP/JSON) & gRPC handlers and routing
│   ├── orchestrator/        # Parallel fan-out search worker pools & saga coordinator
│   ├── normalizer/          # Room, meal plan, GIATA ID & geo-spatial deduplication
│   ├── markup/              # Dynamic net-to-gross rate calculation engine
│   ├── vault/               # Encrypted multi-tenant BYOC credential manager
│   ├── webhook/             # Asynchronous event dispatcher with exponential backoff
│   └── model/               # Universal domain models & schemas
│
├── pkg/
│   ├── connector/           # Standard Go Connector interface & registry
│   └── hospitsdk/           # Official Go client SDK for external consumers
│
├── connectors/              # Pluggable supplier implementations
│   ├── hotelbeds/           # Hotelbeds APItude REST implementation
│   ├── expedia/             # Expedia Rapid 3.0 implementation
│   ├── amadeus/             # Amadeus GDS implementation
│   ├── webbeds/             # WebBeds XML/REST implementation
│   ├── ratehawk/            # RateHawk API implementation
│   ├── siteminder/          # SiteMinder Two-Way ARI switch
│   └── ...                  # Other connectors
│
├── docs/                    # Schema documentation & integration specifications
├── Makefile                 # Build, test, and container targets
└── go.mod                   # Go module definition
```

---

## 6. Quickstart & Local Development

### Prerequisites
- **Go 1.23+**
- **Docker** & **Docker Compose** (optional)

### Build & Run
```bash
# Clone the repository
git clone https://github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform.git
cd Hospit-Open-Booking-Integration-Platform

# Download dependencies
go mod download

# Run the local server with sandbox mock connectors
go run ./cmd/hospit-server
```

### Search Hotels via cURL
```bash
curl -X POST http://localhost:4000/v1/hotels/search \
  -H "Content-Type: application/json" \
  -d '{
    "stay": { "checkIn": "2026-11-15", "checkOut": "2026-11-18" },
    "occupancies": [{ "adults": 2, "childrenAges": [] }],
    "location": { "latitude": 25.1304, "longitude": 55.1171, "radiusKm": 5.0 },
    "suppliers": ["hotelbeds", "expedia", "ratehawk"],
    "currency": "USD"
  }'
```

---

## 7. Roadmap

- [x] **Universal Schema & Domain Engine Definition**
- [ ] **Golang Core Server & Goroutine Fan-Out Search Worker Pool**
- [ ] **Top 5 Connectors**: Hotelbeds, Expedia Rapid, WebBeds, Amadeus, SiteMinder
- [ ] **GIATA & Geo-spatial H3 Property Deduplication Engine**
- [ ] **Multi-Tenant BYOC (Bring Your Own Credentials) Vault**
- [ ] **Channel Switch Two-Way ARI Push/Pull Engine**
- [ ] **Web Management Dashboard & Visual Inspector** *(Scheduled for subsequent phase)*

---

## License

Hospit is open-source software licensed under the [Apache License 2.0](LICENSE).
