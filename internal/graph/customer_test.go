package graph_test

import (
	"testing"

	"github.com/google/uuid"
)

// Loop A: createCustomer returns a customer with a UUIDv7 ID, and the
// customers query returns it afterwards.
func TestCreateCustomerThenListCustomers(t *testing.T) {
	c := newTestClient(t)

	var created struct {
		CreateCustomer struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			PhoneNumber string `json:"phoneNumber"`
		} `json:"createCustomer"`
	}
	err := c.Post(`mutation {
		createCustomer(input: {name: "Ada Lovelace", phoneNumber: "+15550100"}) {
			id name phoneNumber
		}
	}`, &created)
	if err != nil {
		t.Fatalf("createCustomer: %v", err)
	}

	got := created.CreateCustomer
	if got.Name != "Ada Lovelace" || got.PhoneNumber != "+15550100" {
		t.Fatalf("createCustomer returned %+v, want the input name and phone", got)
	}

	id, err := uuid.Parse(got.ID)
	if err != nil {
		t.Fatalf("customer ID %q is not a UUID: %v", got.ID, err)
	}
	if id.Version() != 7 {
		t.Fatalf("customer ID %q is UUID version %d, want 7", got.ID, id.Version())
	}

	var listed struct {
		Customers []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"customers"`
	}
	if err := c.Post(`{ customers { id name } }`, &listed); err != nil {
		t.Fatalf("customers: %v", err)
	}
	if len(listed.Customers) != 1 {
		t.Fatalf("customers returned %d rows, want 1", len(listed.Customers))
	}
	if listed.Customers[0].ID != got.ID {
		t.Fatalf("customers[0].id = %q, want %q", listed.Customers[0].ID, got.ID)
	}
}
