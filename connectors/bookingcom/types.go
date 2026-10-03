package bookingcom

import "time"

// BookingDemandAvailabilityRequest is the raw payload sent to Demand API v3
type BookingDemandAvailabilityRequest struct {
	CheckIn      string   `json:"checkin"`
	CheckOut     string   `json:"checkout"`
	HotelIDs     []int    `json:"hotel_ids,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	RadiusKm     *float64 `json:"radius_km,omitempty"`
	Adults       int      `json:"adults"`
	ChildrenAges []int    `json:"children_ages,omitempty"`
	Currency     string   `json:"currency"`
}

// BookingDemandAvailabilityResponse is the raw response from Demand API v3
type BookingDemandAvailabilityResponse struct {
	Hotels []BookingHotelResult `json:"result"`
}

type BookingHotelResult struct {
	HotelID      int                `json:"hotel_id"`
	HotelName    string             `json:"hotel_name"`
	StarRating   int                `json:"star_rating"`
	Latitude     float64            `json:"latitude"`
	Longitude    float64            `json:"longitude"`
	Address      string             `json:"address"`
	City         string             `json:"city"`
	CountryCode  string             `json:"country_code"`
	RoomBlocks   []BookingRoomBlock `json:"block"`
}

type BookingRoomBlock struct {
	BlockID           string  `json:"block_id"`
	RoomID            int     `json:"room_id"`
	RoomName          string  `json:"room_name"`
	MealPlan          string  `json:"meal_plan"` // "none", "breakfast", "half_board", "full_board", "all_inclusive"
	Price             float64 `json:"price"`
	NetPrice          float64 `json:"net_price"`
	Currency          string  `json:"currency"`
	Allotment         int     `json:"allotment"`
	IsFreeCancel      bool    `json:"refundable"`
	CancelDeadlineStr string  `json:"refundable_until,omitempty"` // "2026-11-10 23:59:59"
}

// BookingOrderRequest is the raw booking creation payload
type BookingOrderRequest struct {
	Booker struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	} `json:"booker"`
	Blocks []struct {
		BlockID   string `json:"block_id"`
		FirstName string `json:"guest_first_name"`
		LastName  string `json:"guest_last_name"`
	} `json:"blocks"`
	AffiliateID string `json:"affiliate_id"`
	ClientRef   string `json:"customer_reference"`
}

// BookingOrderResponse is the raw response from Order creation
type BookingOrderResponse struct {
	ReservationID int       `json:"reservation_id"`
	PINCode       string    `json:"pincode"`
	Status        string    `json:"status"` // "confirmed", "pending"
	TotalAmount   float64   `json:"total_price"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}
