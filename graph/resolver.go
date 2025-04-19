package graph

import (
	"sync"
	"example/go-app/graph/model"
)
// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	mu        sync.Mutex
    customers []*model.Customer
    orders    []*model.Order
}
