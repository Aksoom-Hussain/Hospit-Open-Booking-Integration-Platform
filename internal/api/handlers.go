package api

import (
	"encoding/json"
	"net/http"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/internal/orchestrator"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

type Handler struct {
	orchestrator *orchestrator.Orchestrator
	registry     *connector.Registry
}

func NewHandler(orch *orchestrator.Orchestrator, reg *connector.Registry) *Handler {
	return &Handler{
		orchestrator: orch,
		registry:     reg,
	}
}

// writeJSON writes JSON response with status code
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes standardized error JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

// HandleHealth returns engine health and uptime status
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"engine":  "Hospit Open Booking Integration Platform",
		"version": "1.0.0",
	})
}

// HandleListSuppliers returns all registered and active connectors
func (h *Handler) HandleListSuppliers(w http.ResponseWriter, r *http.Request) {
	connectors := h.registry.ListAll()
	type SupplierInfo struct {
		ID       string                 `json:"id"`
		Name     string                 `json:"name"`
		Category model.SupplierCategory `json:"category"`
	}
	var list []SupplierInfo
	for _, c := range connectors {
		list = append(list, SupplierInfo{
			ID:       c.ID(),
			Name:     c.Name(),
			Category: c.Category(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":     len(list),
		"suppliers": list,
	})
}

// HandleSearch executes parallel multi-threaded hotel availability search (single or multi hotel)
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.orchestrator.Search(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleQuote rechecks live price & locks rate
func (h *Handler) HandleQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.RateKey == "" {
		writeError(w, http.StatusBadRequest, "rateKey is required")
		return
	}

	resp, err := h.orchestrator.Quote(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "quote failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleBooking commits transactional reservation
func (h *Handler) HandleBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.RateKey == "" {
		writeError(w, http.StatusBadRequest, "rateKey is required")
		return
	}

	resp, err := h.orchestrator.Book(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "booking failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleCancel commits cancellation
func (h *Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Supplier ID can be passed in query param or defaults to bookingcom
	supplierID := r.URL.Query().Get("supplier")
	if supplierID == "" {
		supplierID = "bookingcom"
	}

	resp, err := h.orchestrator.Cancel(r.Context(), supplierID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cancellation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandlePushARI synchronizes live Availability, Rates and Inventory for single hotel PMS connect
func (h *Handler) HandlePushARI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.ARIPushPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	supplierID := r.URL.Query().Get("supplier")
	if supplierID == "" {
		supplierID = "bookingcom"
	}

	resp, err := h.orchestrator.PushARI(r.Context(), supplierID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ARI push failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
