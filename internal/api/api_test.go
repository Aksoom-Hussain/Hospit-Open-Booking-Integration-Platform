package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/connectors/bookingcom"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/internal/orchestrator"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

func setupTestServer() *Server {
	reg := connector.NewRegistry()
	_ = reg.Register(bookingcom.NewConnector(bookingcom.Config{Sandbox: true}))
	orch := orchestrator.NewOrchestrator(reg, 2*time.Second)
	return NewServer(4000, orch, reg)
}

func TestAPI_Health(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAPI_Search(t *testing.T) {
	srv := setupTestServer()
	body, _ := json.Marshal(model.SearchRequest{
		Location: &model.GeoLocation{
			Latitude:  25.1304,
			Longitude: 55.1171,
		},
		Stay: model.StayDates{
			CheckIn:  "2026-11-15",
			CheckOut: "2026-11-18",
		},
		Occupancies: []model.Occupancy{{Adults: 2}},
		Currency:    "USD",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/hotels/search", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestAPI_Suppliers(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/suppliers", nil)
	w := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
