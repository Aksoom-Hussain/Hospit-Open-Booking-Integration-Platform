package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/model"
)

// Orchestrator coordinates concurrent multi-threaded supplier fan-out and booking sagas
type Orchestrator struct {
	registry       *connector.Registry
	defaultTimeout time.Duration
}

// NewOrchestrator initializes a new multi-threaded search orchestrator
func NewOrchestrator(registry *connector.Registry, timeout time.Duration) *Orchestrator {
	if timeout == 0 {
		timeout = 2500 * time.Millisecond
	}
	return &Orchestrator{
		registry:       registry,
		defaultTimeout: timeout,
	}
}

// supplierSearchResult holds async result per worker
type supplierSearchResult struct {
	supplierID string
	hotels     []*model.HotelRate
	err        error
}

// Search executes concurrent Goroutine fan-out across all configured suppliers
// Handles both Single Hotel (targeted property query) and Multi-Hotel (geo/destination query)
func (o *Orchestrator) Search(ctx context.Context, req *model.SearchRequest) (*model.SearchResponse, error) {
	startTime := time.Now()

	// Create bounded context with timeout
	ctx, cancel := context.WithTimeout(ctx, o.defaultTimeout)
	defer cancel()

	allConnectors := o.registry.ListAll()
	var activeConnectors []connector.SupplierConnector

	// Filter by requested suppliers if specified
	if req.Filters != nil && len(req.Filters.Suppliers) > 0 {
		filterMap := make(map[string]bool)
		for _, s := range req.Filters.Suppliers {
			filterMap[strings.ToLower(strings.TrimSpace(s))] = true
		}
		for _, c := range allConnectors {
			if filterMap[strings.ToLower(c.ID())] {
				activeConnectors = append(activeConnectors, c)
			}
		}
	} else {
		activeConnectors = allConnectors
	}

	if len(activeConnectors) == 0 {
		return &model.SearchResponse{
			SearchID:           fmt.Sprintf("srch_%d", time.Now().UnixNano()),
			ResultsCount:       0,
			ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
			Hotels:             []*model.HotelRate{},
			SuppliersQueried:   []string{},
			SuppliersResponded: []string{},
		}, nil
	}

	resultsChan := make(chan supplierSearchResult, len(activeConnectors))
	var wg sync.WaitGroup

	var queriedSuppliers []string
	for _, c := range activeConnectors {
		queriedSuppliers = append(queriedSuppliers, c.ID())
		wg.Add(1)

		// Spawn concurrent goroutine per supplier
		go func(conn connector.SupplierConnector) {
			defer wg.Done()
			hotels, err := conn.SearchHotels(ctx, req)
			select {
			case resultsChan <- supplierSearchResult{supplierID: conn.ID(), hotels: hotels, err: err}:
			case <-ctx.Done():
				// Context cancelled or timed out
			}
		}(c)
	}

	// Wait in background and close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allHotels []*model.HotelRate
	var respondedSuppliers []string

	// Collect results as they arrive
	for res := range resultsChan {
		if res.err == nil && len(res.hotels) > 0 {
			respondedSuppliers = append(respondedSuppliers, res.supplierID)
			allHotels = append(allHotels, res.hotels...)
		}
	}

	return &model.SearchResponse{
		SearchID:           fmt.Sprintf("srch_%d", time.Now().UnixNano()),
		ResultsCount:       len(allHotels),
		ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
		Hotels:             allHotels,
		SuppliersQueried:   queriedSuppliers,
		SuppliersResponded: respondedSuppliers,
	}, nil
}

// Quote verifies rate with the corresponding supplier
func (o *Orchestrator) Quote(ctx context.Context, req *model.QuoteRequest) (*model.QuoteResponse, error) {
	conn := o.resolveConnectorFromKey(req.RateKey)
	if conn == nil {
		return nil, fmt.Errorf("unable to resolve supplier connector from rate key: %s", req.RateKey)
	}
	return conn.QuoteRate(ctx, req)
}

// Book commits transactional reservation with supplier
func (o *Orchestrator) Book(ctx context.Context, req *model.BookingRequest) (*model.BookingResponse, error) {
	conn := o.resolveConnectorFromKey(req.RateKey)
	if conn == nil {
		return nil, fmt.Errorf("unable to resolve supplier connector from rate key: %s", req.RateKey)
	}
	return conn.CreateBooking(ctx, req)
}

// Cancel cancels existing booking
func (o *Orchestrator) Cancel(ctx context.Context, supplierID string, req *model.CancelRequest) (*model.CancelResponse, error) {
	conn, ok := o.registry.Get(supplierID)
	if !ok {
		return nil, fmt.Errorf("supplier connector not found: %s", supplierID)
	}
	return conn.CancelBooking(ctx, req)
}

// PushARI delivers availability, rates and inventory to channel switch
func (o *Orchestrator) PushARI(ctx context.Context, supplierID string, payload *model.ARIPushPayload) (*model.ARIPushResult, error) {
	conn, ok := o.registry.Get(supplierID)
	if !ok {
		return nil, fmt.Errorf("supplier connector not found: %s", supplierID)
	}
	return conn.PushARI(ctx, payload)
}

// resolveConnectorFromKey extracts supplier code prefix from rateKey (e.g. "rk_bk_..." -> "bookingcom")
func (o *Orchestrator) resolveConnectorFromKey(rateKey string) connector.SupplierConnector {
	if strings.HasPrefix(rateKey, "rk_bk_") {
		conn, _ := o.registry.Get("bookingcom")
		return conn
	}
	if strings.HasPrefix(rateKey, "rk_hb_") {
		conn, _ := o.registry.Get("hotelbeds")
		return conn
	}
	if strings.HasPrefix(rateKey, "rk_exp_") {
		conn, _ := o.registry.Get("expedia")
		return conn
	}
	// Default fallback to first registered connector
	all := o.registry.ListAll()
	if len(all) > 0 {
		return all[0]
	}
	return nil
}
