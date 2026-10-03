package connector

import (
	"context"
	"fmt"
	"sync"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

// SupplierConnector is the standard pluggable Go interface for any hotel supplier
type SupplierConnector interface {
	// ID returns unique lowercase code (e.g. "bookingcom", "hotelbeds", "expedia")
	ID() string

	// Name returns human-readable name (e.g. "Booking.com Demand & Connectivity")
	Name() string

	// Category returns supplier type (Bedbank, GDS, OTAWholesale, ChannelSwitch, etc.)
	Category() model.SupplierCategory

	// SearchHotels executes availability search (supports single hotel ID or multi-hotel location)
	SearchHotels(ctx context.Context, req *model.SearchRequest) ([]*model.HotelRate, error)

	// QuoteRate verifies real-time price & returns binding cancellation policy
	QuoteRate(ctx context.Context, req *model.QuoteRequest) (*model.QuoteResponse, error)

	// CreateBooking commits reservation with supplier
	CreateBooking(ctx context.Context, req *model.BookingRequest) (*model.BookingResponse, error)

	// CancelBooking cancels confirmed reservation
	CancelBooking(ctx context.Context, req *model.CancelRequest) (*model.CancelResponse, error)

	// PushARI pushes live Availability, Rates, and Inventory (for PMS / Channel Switches)
	PushARI(ctx context.Context, payload *model.ARIPushPayload) (*model.ARIPushResult, error)
}

// Registry manages thread-safe registration and discovery of supplier connectors
type Registry struct {
	mu         sync.RWMutex
	connectors map[string]SupplierConnector
}

// NewRegistry initializes an empty connector registry
func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]SupplierConnector),
	}
}

// Register registers a new connector
func (r *Registry) Register(c SupplierConnector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := c.ID()
	if _, exists := r.connectors[id]; exists {
		return fmt.Errorf("connector already registered: %s", id)
	}
	r.connectors[id] = c
	return nil
}

// Get retrieves a connector by ID
func (r *Registry) Get(id string) (SupplierConnector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.connectors[id]
	return c, ok
}

// ListAll returns all registered connectors
func (r *Registry) ListAll() []SupplierConnector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]SupplierConnector, 0, len(r.connectors))
	for _, c := range r.connectors {
		list = append(list, c)
	}
	return list
}
