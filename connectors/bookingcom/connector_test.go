package bookingcom

import (
	"context"
	"testing"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

func TestBookingComConnector_SingleHotelSearch(t *testing.T) {
	conn := NewConnector(Config{Sandbox: true})
	ctx := context.Background()

	req := &model.SearchRequest{
		HotelIDs: []string{"bookingcom:1029384"},
		Stay: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Occupancies: []model.Occupancy{
			{Adults: 2},
		},
		Currency: "USD",
	}

	results, err := conn.SearchHotels(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 hotel result, got %d", len(results))
	}

	h := results[0]
	if h.SupplierCode != "bookingcom" {
		t.Errorf("expected supplierCode 'bookingcom', got '%s'", h.SupplierCode)
	}
	if len(h.Rooms) == 0 {
		t.Errorf("expected at least 1 room rate")
	}
}

func TestBookingComConnector_MultiHotelSearch(t *testing.T) {
	conn := NewConnector(Config{Sandbox: true})
	ctx := context.Background()

	req := &model.SearchRequest{
		Location: &model.GeoLocation{
			Latitude:  25.1304,
			Longitude: 55.1171,
			RadiusKm:  10.0,
		},
		Stay: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Occupancies: []model.Occupancy{
			{Adults: 2},
		},
		Currency: "USD",
	}

	results, err := conn.SearchHotels(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) < 2 {
		t.Fatalf("expected at least 2 hotel results for multi-hotel query, got %d", len(results))
	}
}

func TestBookingComConnector_QuoteAndBookingLifecycle(t *testing.T) {
	conn := NewConnector(Config{Sandbox: true})
	ctx := context.Background()

	// 1. Quote
	quoteResp, err := conn.QuoteRate(ctx, &model.QuoteRequest{
		RateKey: "rk_bk_1029384_ocean_bb",
	})
	if err != nil {
		t.Fatalf("quote error: %v", err)
	}
	if quoteResp.TotalAmount <= 0 {
		t.Errorf("expected positive quote amount, got %f", quoteResp.TotalAmount)
	}

	// 2. Booking
	bookResp, err := conn.CreateBooking(ctx, &model.BookingRequest{
		QuoteID: quoteResp.QuoteID,
		RateKey: quoteResp.RateKey,
		Holder: model.BookingHolder{
			FirstName: "Alexander",
			LastName:  "Wright",
			Email:     "alex@example.com",
			Phone:     "+1234567890",
		},
		Rooms: []model.BookedRoom{
			{
				RoomIndex: 0,
				Guests: []model.Guest{
					{FirstName: "Alexander", LastName: "Wright"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("booking error: %v", err)
	}
	if bookResp.Status != model.StatusConfirmed {
		t.Errorf("expected booking status CONFIRMED, got %s", bookResp.Status)
	}

	// 3. Cancel
	cancelResp, err := conn.CancelBooking(ctx, &model.CancelRequest{
		BookingID: bookResp.BookingID,
	})
	if err != nil {
		t.Fatalf("cancel error: %v", err)
	}
	if cancelResp.Status != model.StatusCancelled {
		t.Errorf("expected cancel status CANCELLED, got %s", cancelResp.Status)
	}
}

func TestBookingComConnector_PushARI(t *testing.T) {
	conn := NewConnector(Config{Sandbox: true})
	ctx := context.Background()

	ariResp, err := conn.PushARI(ctx, &model.ARIPushPayload{
		HotelID: "1029384",
		DateRange: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Updates: []model.ARIRecord{
			{
				RoomTypeID: "rm_deluxe_101",
				Date:       "2026-11-15",
				Available:  5,
				Price:      420.00,
			},
		},
	})
	if err != nil {
		t.Fatalf("ari push error: %v", err)
	}
	if !ariResp.Success {
		t.Errorf("expected ARI push success")
	}
}
