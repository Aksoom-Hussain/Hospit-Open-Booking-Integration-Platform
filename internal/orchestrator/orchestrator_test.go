package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/connectors/bookingcom"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

func TestOrchestrator_ConcurrentSearch(t *testing.T) {
	registry := connector.NewRegistry()
	_ = registry.Register(bookingcom.NewConnector(bookingcom.Config{Sandbox: true}))

	orch := NewOrchestrator(registry, 2*time.Second)

	// Test Multi-Hotel Search
	resp, err := orch.Search(context.Background(), &model.SearchRequest{
		Location: &model.GeoLocation{
			Latitude:  25.1304,
			Longitude: 55.1171,
			RadiusKm:  10.0,
		},
		Stay: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Occupancies: []model.Occupancy{{Adults: 2}},
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("search error: %v", err)
	}

	if resp.ResultsCount == 0 {
		t.Errorf("expected search results, got 0")
	}

	// Test Single-Hotel Search
	singleResp, err := orch.Search(context.Background(), &model.SearchRequest{
		HotelIDs: []string{"bookingcom:1029384"},
		Stay: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Occupancies: []model.Occupancy{{Adults: 2}},
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("single hotel search error: %v", err)
	}
	if singleResp.ResultsCount != 1 {
		t.Errorf("expected 1 result for single hotel, got %d", singleResp.ResultsCount)
	}
}
