package model

import (
	"time"
)

// SupplierCategory identifies the technical class of a partner
type SupplierCategory string

const (
	CategoryBedbank       SupplierCategory = "BEDBANK"
	CategoryGDS           SupplierCategory = "GDS"
	CategoryOTAWholesale  SupplierCategory = "OTA_WHOLESALE"
	CategoryRegionalDMC   SupplierCategory = "REGIONAL_DMC"
	CategoryChannelSwitch SupplierCategory = "CHANNEL_SWITCH"
)

// MealPlan represents normalized board basis across suppliers
type MealPlan string

const (
	MealRoomOnly          MealPlan = "ROOM_ONLY"          // RO / EP
	MealBreakfastIncluded MealPlan = "BREAKFAST_INCLUDED" // BB / CP
	MealHalfBoard         MealPlan = "HALF_BOARD"         // HB / MAP
	MealFullBoard         MealPlan = "FULL_BOARD"         // FB / AP
	MealAllInclusive      MealPlan = "ALL_INCLUSIVE"      // AI
)

// BookingStatus represents the reservation state machine
type BookingStatus string

const (
	StatusPending   BookingStatus = "PENDING"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusFailed    BookingStatus = "FAILED"
	StatusAmended   BookingStatus = "AMENDED"
	StatusCancelled BookingStatus = "CANCELLED"
)

// StayDates defines check-in and check-out dates (YYYY-MM-DD)
type StayDates struct {
	CheckIn  string `json:"checkIn"`
	CheckOut string `json:"checkOut"`
}

// Occupancy defines guest breakdown per room
type Occupancy struct {
	Adults       int   `json:"adults"`
	ChildrenAges []int `json:"childrenAges,omitempty"`
}

// GeoLocation defines spatial coordinates
type GeoLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	RadiusKm    float64 `json:"radiusKm,omitempty"`
	Address     string  `json:"address,omitempty"`
	City        string  `json:"city,omitempty"`
	CountryCode string  `json:"countryCode,omitempty"`
}

// SearchFilter represents search constraint filters
type SearchFilter struct {
	MinStars             int        `json:"minStars,omitempty"`
	MaxStars             int        `json:"maxStars,omitempty"`
	MealPlans            []MealPlan `json:"mealPlans,omitempty"`
	FreeCancellationOnly bool       `json:"freeCancellationOnly,omitempty"`
	Suppliers            []string   `json:"suppliers,omitempty"` // Filter specific suppliers (e.g. ["bookingcom"])
	MaxPrice             float64    `json:"maxPrice,omitempty"`
}

// SearchRequest is the unified request for both single-hotel and multi-hotel queries
type SearchRequest struct {
	// Mode 1: Single Hotel query or specific hotel IDs
	HotelIDs []string `json:"hotelIds,omitempty"` // e.g. ["bookingcom:1029384", "hotel_palm_dubai"]

	// Mode 2: Multi-Hotel geo/location query
	Location *GeoLocation `json:"location,omitempty"`

	Stay              StayDates     `json:"stay"`
	Occupancies       []Occupancy   `json:"occupancies"`
	Filters           *SearchFilter `json:"filters,omitempty"`
	ClientNationality string        `json:"clientNationality,omitempty"` // ISO 2-letter
	Currency          string        `json:"currency"`                    // e.g. "USD", "EUR"
}

// SearchResponse contains aggregated, deduplicated search results
type SearchResponse struct {
	SearchID           string       `json:"searchId"`
	ResultsCount       int          `json:"resultsCount"`
	ExecutionTimeMs    int64        `json:"executionTimeMs"`
	Hotels             []*HotelRate `json:"hotels"`
	SuppliersQueried   []string     `json:"suppliersQueried"`
	SuppliersResponded []string     `json:"suppliersResponded"`
}

// HotelRate represents a hotel property and available room rates
type HotelRate struct {
	HospitHotelID string       `json:"hospitHotelId"`
	SupplierCode  string       `json:"supplierCode"`
	SupplierRef   string       `json:"supplierRef"`
	GIATAID       int          `json:"giataId,omitempty"`
	Name          string       `json:"name"`
	Stars         int          `json:"stars"`
	Location      *GeoLocation `json:"location"`
	LeadRate      *LeadRate    `json:"leadRate"`
	Rooms         []*RoomRate  `json:"rooms,omitempty"`
}

// LeadRate represents the cheapest available room rate for this hotel
type LeadRate struct {
	Amount               float64  `json:"amount"`
	Currency             string   `json:"currency"`
	Supplier             string   `json:"supplier"`
	RoomName             string   `json:"roomName"`
	MealPlan             MealPlan `json:"mealPlan"`
	IsFreeCancellation   bool     `json:"isFreeCancellation"`
	CancellationDeadline string   `json:"cancellationDeadline,omitempty"`
}

// RoomRate represents a specific room option with rate plans
type RoomRate struct {
	RateKey            string              `json:"rateKey"`
	RoomID             string              `json:"roomId"`
	RoomName           string              `json:"roomName"`
	MealPlan           MealPlan            `json:"mealPlan"`
	TotalAmount        float64             `json:"totalAmount"`
	NetAmount          float64             `json:"netAmount,omitempty"`
	TaxAmount          float64             `json:"taxAmount,omitempty"`
	Currency           string              `json:"currency"`
	Allotment          int                 `json:"allotment,omitempty"` // Remaining rooms
	CancellationPolicy *CancellationPolicy `json:"cancellationPolicy"`
}

// CancellationPolicy represents normalized cancellation windows and penalties
type CancellationPolicy struct {
	IsRefundable          bool            `json:"isRefundable"`
	FreeCancellationUntil *time.Time      `json:"freeCancellationUntil,omitempty"`
	PenaltyWindows        []PenaltyWindow `json:"penaltyWindows,omitempty"`
}

// PenaltyWindow defines a specific timeframe and penalty amount
type PenaltyWindow struct {
	From          time.Time `json:"from"`
	To            time.Time `json:"to"`
	PenaltyAmount float64   `json:"penaltyAmount"`
	Currency      string    `json:"currency"`
}

// QuoteRequest rechecks real-time price & freezes rate
type QuoteRequest struct {
	RateKey  string `json:"rateKey"`
	Currency string `json:"currency,omitempty"`
}

// QuoteResponse is the price-freeze guarantee
type QuoteResponse struct {
	QuoteID            string              `json:"quoteId"`
	RateKey            string              `json:"rateKey"`
	SupplierCode       string              `json:"supplierCode"`
	HotelID            string              `json:"hotelId"`
	HotelName          string              `json:"hotelName"`
	RoomName           string              `json:"roomName"`
	MealPlan           MealPlan            `json:"mealPlan"`
	TotalAmount        float64             `json:"totalAmount"`
	NetAmount          float64             `json:"netAmount"`
	Currency           string              `json:"currency"`
	ExpiresAt          time.Time           `json:"expiresAt"`
	CancellationPolicy *CancellationPolicy `json:"cancellationPolicy"`
}

// Guest represents individual room guest
type Guest struct {
	Title     string `json:"title"` // MR, MRS, MS, MSTR
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	IsLead    bool   `json:"isLead,omitempty"`
	Age       int    `json:"age,omitempty"`
}

// BookedRoom represents guest allocation per room
type BookedRoom struct {
	RoomIndex       int     `json:"roomIndex"`
	Guests          []Guest `json:"guests"`
	SpecialRequests string  `json:"specialRequests,omitempty"`
}

// BookingHolder is the primary contact/billing person
type BookingHolder struct {
	Title     string `json:"title"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// BookingRequest creates an immediate reservation
type BookingRequest struct {
	IdempotencyKey string        `json:"idempotencyKey"`
	QuoteID        string        `json:"quoteId"`
	RateKey        string        `json:"rateKey"`
	Holder         BookingHolder `json:"holder"`
	Rooms          []BookedRoom  `json:"rooms"`
	ClientRef      string        `json:"clientRef,omitempty"`
}

// BookingResponse is the confirmed reservation receipt
type BookingResponse struct {
	BookingID              string        `json:"bookingId"`
	Status                 BookingStatus `json:"status"`
	SupplierCode           string        `json:"supplierCode"`
	SupplierReference      string        `json:"supplierReference"`
	HotelConfirmationCode  string        `json:"hotelConfirmationCode,omitempty"`
	TotalAmount            float64       `json:"totalAmount"`
	Currency               string        `json:"currency"`
	Stay                   StayDates     `json:"stay"`
	VoucherURL             string        `json:"voucherUrl,omitempty"`
	CreatedAt              time.Time     `json:"createdAt"`
}

// CancelRequest cancels an existing reservation
type CancelRequest struct {
	BookingID         string `json:"bookingId"`
	SupplierReference string `json:"supplierReference,omitempty"`
	Reason            string `json:"reason,omitempty"`
}

// CancelResponse returns cancellation outcome & penalty
type CancelResponse struct {
	BookingID         string        `json:"bookingId"`
	Status            BookingStatus `json:"status"`
	CancellationRef   string        `json:"cancellationRef"`
	PenaltyAmount     float64       `json:"penaltyAmount"`
	RefundAmount      float64       `json:"refundAmount"`
	Currency          string        `json:"currency"`
	CancelledAt       time.Time     `json:"cancelledAt"`
}

// ARIPushPayload defines real-time Availability, Rate & Inventory sync payload for PMS / Switches
type ARIPushPayload struct {
	HotelID   string     `json:"hotelId"`
	DateRange StayDates  `json:"dateRange"`
	Updates   []ARIRecord `json:"updates"`
}

// ARIRecord represents single room/rate date-level update
type ARIRecord struct {
	RoomTypeID   string   `json:"roomTypeId"`
	RatePlanID   string   `json:"ratePlanId,omitempty"`
	Date         string   `json:"date"` // YYYY-MM-DD
	Available    int      `json:"available"`
	Price        float64  `json:"price,omitempty"`
	Currency     string   `json:"currency,omitempty"`
	Closed       bool     `json:"closed,omitempty"`       // Stop-sell
	MinStay      int      `json:"minStay,omitempty"`
	ClosedOnArrival bool  `json:"closedOnArrival,omitempty"`
}

// ARIPushResult returns confirmation of ARI delivery
type ARIPushResult struct {
	Success       bool      `json:"success"`
	HotelID       string    `json:"hotelId"`
	RecordsSynced int       `json:"recordsSynced"`
	Message       string    `json:"message"`
	SyncedAt      time.Time `json:"syncedAt"`
}
