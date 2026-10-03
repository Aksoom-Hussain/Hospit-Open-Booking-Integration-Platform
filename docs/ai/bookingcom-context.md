# Hospit AI Context: Booking.com Adapter

> Structured technical knowledge base and contract specifications for AI agents and developers implementing, querying, and debugging the **Booking.com Demand & Connectivity Adapter** in Hospit.

---

## 1. Adapter Contract Metadata

```yaml
adapter_id: "bookingcom"
name: "Booking.com Demand API v3 & Connectivity Switch"
category: "OTA_WHOLESALE"
primary_protocols: ["REST_JSON", "OTA_XML"]
auth_type: "BEARER_TOKEN_AND_AFFILIATE_ID"
rate_model: "NET_WHOLESALE_AND_COMMISSION"
supported_operations:
  - search_single_hotel: true
  - search_multi_hotel_geo: true
  - quote_price_freeze: true
  - create_booking: true
  - cancel_booking: true
  - push_ari: true
  - pull_reservations: true
rate_key_pattern: "^rk_bk_[0-9]+_.*$"
```

---

## 2. JSON Schema Contracts

### Search Request (Single Hotel vs Multi Hotel)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "HospitSearchRequest",
  "type": "object",
  "required": ["stay", "occupancies", "currency"],
  "properties": {
    "hotelIds": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Optional list of hotel IDs for single-hotel or targeted property shopping. Format: 'bookingcom:1029384'"
    },
    "location": {
      "type": "object",
      "properties": {
        "latitude": { "type": "number" },
        "longitude": { "type": "number" },
        "radiusKm": { "type": "number", "default": 10.0 },
        "city": { "type": "string" },
        "countryCode": { "type": "string" }
      }
    },
    "stay": {
      "type": "object",
      "required": ["checkIn", "checkOut"],
      "properties": {
        "checkIn": { "type": "string", "format": "date" },
        "checkOut": { "type": "string", "format": "date" }
      }
    },
    "occupancies": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["adults"],
        "properties": {
          "adults": { "type": "integer", "minimum": 1 },
          "childrenAges": { "type": "array", "items": { "type": "integer" } }
        }
      }
    },
    "currency": { "type": "string", "example": "USD" }
  }
}
```

---

## 3. Meal Plan Normalization Matrix

| Booking.com Raw Code | Hospit Normalized Enum | Description |
| :--- | :--- | :--- |
| `none` / `ro` | `ROOM_ONLY` | Room only, no meals included |
| `breakfast` / `buffet_breakfast` | `BREAKFAST_INCLUDED` | Bed & Breakfast |
| `half_board` | `HALF_BOARD` | Breakfast and dinner included |
| `full_board` | `FULL_BOARD` | Breakfast, lunch, and dinner included |
| `all_inclusive` / `ultra_all_inclusive` | `ALL_INCLUSIVE` | All meals and beverages included |

---

## 4. Booking Saga State Machine

```
[SEARCH] ──► [QUOTE (Lock Rate 15m)] ──► [RESERVE (POST /v1/bookings)]
                                               │
                                 ┌─────────────┴─────────────┐
                                 ▼                           ▼
                           [CONFIRMED]                   [FAILED]
                         (PNR Generated)             (Auto-Rollback)
                                 │
                                 ▼
                     [CANCEL (POST /v1/bookings/cancel)]
                                 │
                                 ▼
                           [CANCELLED]
```

---

## 5. Error Taxonomy & Troubleshooting

| Error Code | Root Cause | AI Remediation Action |
| :--- | :--- | :--- |
| `401 Unauthorized` | Invalid `BOOKINGCOM_API_KEY` or expired token | Refresh bearer token in credential vault. |
| `404 Hotel Not Found` | Non-existent or inactive Booking.com Hotel ID | Verify property active status on Booking.com Extranet. |
| `409 Rate Changed` | Live price shifted between search and quote | Fetch new quote and present price delta to user. |
| `422 Allotment Exhausted` | Last room sold during booking commitment | Trigger saga auto-rollback and offer alternate room. |
| `504 Gateway Timeout` | Booking.com Demand API latency exceeded 2.5s | Handled gracefully by Goroutine context timeout; skipped in fan-out. |
