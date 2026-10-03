# Booking.com Integration Specification (Hospit Engine)

> Comprehensive technical documentation for **Booking.com Demand API (v3)** & **Connectivity API (Two-Way ARI Switch)** integration in Hospit.

---

## 1. Overview & Dual-Mode Support

Hospit supports Booking.com in two complementary operational modes:

| Mode | Target User | Protocol | Primary Operations |
| :--- | :--- | :--- | :--- |
| **Multi-Hotel Mode (Demand API)** | OTAs, Wholesalers, Metasearch engines, B2B Aggregators | REST JSON (Demand API v3.1) | Geographic search, parallel rate shopping, room quote, instant reservation |
| **Single-Hotel Mode (Connectivity API)** | Individual Hotels, Boutique Resorts, PMS/CRS integrations | REST & OTA XML Switch | Two-way ARI push (Rates, Allotment, Stop-Sells), direct reservation pull |

---

## 2. Authentication & Environment Configuration

### Credentials
```env
# Booking.com Demand & Connectivity Environment
BOOKINGCOM_API_KEY=your_demand_api_bearer_token
BOOKINGCOM_AFFILIATE_ID=your_affiliate_id
BOOKINGCOM_SANDBOX=true # Set to false for live production
```

### Environment URLs:
- **Production (Demand API)**: `https://demandapi.booking.com/3.1`
- **Sandbox (Demand API)**: `https://demandapi-sandbox.booking.com/3.1`
- **Connectivity API**: `https://supply-xml.booking.com/`

---

## 3. Supported Flows & Endpoints

### 3.1 Single Hotel Search vs Multi-Hotel Search

#### Flow A: Single Hotel Query (Targeted Property)
When a customer or PMS is booking for a specific hotel, specify `hotelIds`:
```bash
curl -X POST http://localhost:4000/v1/hotels/search \
  -H "Content-Type: application/json" \
  -d '{
    "hotelIds": ["bookingcom:1029384"],
    "stay": { "checkIn": "2026-11-15", "checkOut": "2026-11-18" },
    "occupancies": [{ "adults": 2 }],
    "currency": "USD"
  }'
```

#### Flow B: Multi-Hotel Geo Search
When searching across an entire city or coordinate radius:
```bash
curl -X POST http://localhost:4000/v1/hotels/search \
  -H "Content-Type: application/json" \
  -d '{
    "location": {
      "latitude": 25.1304,
      "longitude": 55.1171,
      "radiusKm": 10.0,
      "city": "Dubai",
      "countryCode": "AE"
    },
    "stay": { "checkIn": "2026-11-15", "checkOut": "2026-11-18" },
    "occupancies": [{ "adults": 2, "childrenAges": [5] }],
    "currency": "USD"
  }'
```

---

### 3.2 Price Quote & Rate Freeze (`POST /v1/hotels/quote`)

Rechecks room availability with Booking.com and returns a 15-minute price guarantee with precise cancellation windows:
```json
{
  "quoteId": "quot_bk_172793849102",
  "rateKey": "rk_bk_1029381_ocean_bb",
  "supplierCode": "bookingcom",
  "hotelName": "Atlantis, The Palm",
  "roomName": "Ocean King Room",
  "mealPlan": "BREAKFAST_INCLUDED",
  "totalAmount": 390.00,
  "currency": "USD",
  "cancellationPolicy": {
    "isRefundable": true,
    "freeCancellationUntil": "2026-11-10T23:59:59Z",
    "penaltyWindows": [
      {
        "from": "2026-11-11T00:00:00Z",
        "to": "2026-11-15T14:00:00Z",
        "penaltyAmount": 390.00,
        "currency": "USD"
      }
    ]
  }
}
```

---

### 3.3 Transactional Booking Dispatch (`POST /v1/bookings`)

Executes transactional order creation with guest allocation:
```json
{
  "quoteId": "quot_bk_172793849102",
  "rateKey": "rk_bk_1029381_ocean_bb",
  "holder": {
    "title": "MR",
    "firstName": "Alexander",
    "lastName": "Wright",
    "email": "alex.wright@example.com",
    "phone": "+447700900123"
  },
  "rooms": [
    {
      "roomIndex": 0,
      "guests": [
        { "title": "MR", "firstName": "Alexander", "lastName": "Wright" },
        { "title": "MRS", "firstName": "Elena", "lastName": "Wright" }
      ]
    }
  ]
}
```

---

### 3.4 Two-Way ARI Push (`POST /v1/ari/push?supplier=bookingcom`)

For PMS properties to update real-time rates, inventory allotment, and stop-sells:
```json
{
  "hotelId": "1029381",
  "dateRange": {
    "checkIn": "2026-11-15",
    "checkOut": "2026-11-18"
  },
  "updates": [
    {
      "roomTypeId": "rm_deluxe_101",
      "ratePlanID": "standard_bb",
      "date": "2026-11-15",
      "available": 8,
      "price": 390.00,
      "currency": "USD",
      "closed": false,
      "minStay": 2
    }
  ]
}
```
