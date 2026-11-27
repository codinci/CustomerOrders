package graph_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/google/uuid"
)

// requireDomainRejection fails unless err is a real validation or lookup
// failure. A "not implemented" stub error must not count as a rejection.
func requireDomainRejection(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s succeeded, want an error", what)
	}
	if strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("%s: got the stub error, not a rejection: %v", what, err)
	}
}

// createTestCustomer creates a customer through the API and returns its ID.
func createTestCustomer(t *testing.T, c *client.Client) string {
	t.Helper()
	var resp struct {
		CreateCustomer struct {
			ID string `json:"id"`
		} `json:"createCustomer"`
	}
	if err := c.Post(`mutation {
		createCustomer(input: {name: "Grace Hopper", phoneNumber: "+15550199"}) { id }
	}`, &resp); err != nil {
		t.Fatalf("createCustomer: %v", err)
	}
	return resp.CreateCustomer.ID
}

type orderJSON struct {
	ID        string `json:"id"`
	Product   string `json:"product"`
	Quantity  int32  `json:"quantity"`
	Price     string `json:"price"`
	CreatedAt string `json:"createdAt"`
}

// Loop B: placeOrder stores an order that the orders query returns afterwards.
func TestPlaceOrderThenListOrders(t *testing.T) {
	c := newTestClient(t)
	customerID := createTestCustomer(t, c)

	var placed struct {
		PlaceOrder orderJSON `json:"placeOrder"`
	}
	err := c.Post(`mutation($in: PlaceOrderInput!) {
		placeOrder(input: $in) { id product quantity price createdAt }
	}`, &placed, client.Var("in", map[string]any{
		"customerID": customerID,
		"product":    "Widget",
		"quantity":   3,
		"price":      "19.99",
	}))
	if err != nil {
		t.Fatalf("placeOrder: %v", err)
	}

	got := placed.PlaceOrder
	if got.Product != "Widget" || got.Quantity != 3 {
		t.Fatalf("placeOrder returned %+v, want product Widget and quantity 3", got)
	}
	if got.Price != "19.99" {
		t.Fatalf("placeOrder price = %q, want exactly \"19.99\"", got.Price)
	}
	if got.CreatedAt == "" {
		t.Fatalf("placeOrder returned an empty createdAt")
	}

	id, err := uuid.Parse(got.ID)
	if err != nil {
		t.Fatalf("order ID %q is not a UUID: %v", got.ID, err)
	}
	if id.Version() != 7 {
		t.Fatalf("order ID %q is UUID version %d, want 7", got.ID, id.Version())
	}

	var listed struct {
		Orders []orderJSON `json:"orders"`
	}
	if err := c.Post(`{ orders { id product quantity price createdAt } }`, &listed); err != nil {
		t.Fatalf("orders: %v", err)
	}
	if len(listed.Orders) != 1 {
		t.Fatalf("orders returned %d rows, want 1", len(listed.Orders))
	}
	if listed.Orders[0].ID != got.ID {
		t.Fatalf("orders[0].id = %q, want %q", listed.Orders[0].ID, got.ID)
	}
}

// The orders query returns an empty list, not null, before any order exists.
func TestOrdersEmptyOnNewResolver(t *testing.T) {
	c := newTestClient(t)

	var listed struct {
		Orders []orderJSON `json:"orders"`
	}
	if err := c.Post(`{ orders { id } }`, &listed); err != nil {
		t.Fatalf("orders: %v", err)
	}
	if listed.Orders == nil {
		t.Fatalf("orders returned null, want an empty list")
	}
	if len(listed.Orders) != 0 {
		t.Fatalf("orders returned %d rows, want 0", len(listed.Orders))
	}
}

// placeOrder must reject a customerID that does not exist, and store nothing.
func TestPlaceOrderUnknownCustomerFails(t *testing.T) {
	c := newTestClient(t)

	var resp struct {
		PlaceOrder orderJSON `json:"placeOrder"`
	}
	err := c.Post(`mutation($in: PlaceOrderInput!) {
		placeOrder(input: $in) { id }
	}`, &resp, client.Var("in", map[string]any{
		"customerID": uuid.NewString(),
		"product":    "Widget",
		"quantity":   1,
		"price":      "5.00",
	}))
	requireDomainRejection(t, err, "placeOrder with unknown customer")

	var listed struct {
		Orders []orderJSON `json:"orders"`
	}
	if err := c.Post(`{ orders { id } }`, &listed); err != nil {
		t.Fatalf("orders: %v", err)
	}
	if len(listed.Orders) != 0 {
		t.Fatalf("a failed placeOrder stored %d orders, want 0", len(listed.Orders))
	}
}

// Invalid input is rejected: quantity must be positive, price must not be negative.
func TestPlaceOrderValidation(t *testing.T) {
	tests := []struct {
		name     string
		quantity int
		price    string
	}{
		{name: "zero quantity", quantity: 0, price: "5.00"},
		{name: "negative quantity", quantity: -1, price: "5.00"},
		{name: "negative price", quantity: 1, price: "-0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t)
			customerID := createTestCustomer(t, c)

			var resp struct {
				PlaceOrder orderJSON `json:"placeOrder"`
			}
			err := c.Post(`mutation($in: PlaceOrderInput!) {
				placeOrder(input: $in) { id }
			}`, &resp, client.Var("in", map[string]any{
				"customerID": customerID,
				"product":    "Widget",
				"quantity":   tt.quantity,
				"price":      tt.price,
			}))
			requireDomainRejection(t, err, fmt.Sprintf("placeOrder(quantity=%d, price=%s)", tt.quantity, tt.price))
		})
	}
}

// Concurrent placeOrder calls must all be stored. Run with -race to check the locking.
func TestPlaceOrderConcurrent(t *testing.T) {
	const workers = 50

	c := newTestClient(t)
	customerID := createTestCustomer(t, c)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var resp struct {
				PlaceOrder orderJSON `json:"placeOrder"`
			}
			err := c.Post(`mutation($in: PlaceOrderInput!) {
				placeOrder(input: $in) { id }
			}`, &resp, client.Var("in", map[string]any{
				"customerID": customerID,
				"product":    "Widget",
				"quantity":   1,
				"price":      "1.00",
			}))
			if err != nil {
				t.Errorf("placeOrder: %v", err)
			}
		}()
	}
	wg.Wait()

	var listed struct {
		Orders []orderJSON `json:"orders"`
	}
	if err := c.Post(`{ orders { id } }`, &listed); err != nil {
		t.Fatalf("orders: %v", err)
	}
	if len(listed.Orders) != workers {
		t.Fatalf("orders returned %d rows, want %d", len(listed.Orders), workers)
	}
}
