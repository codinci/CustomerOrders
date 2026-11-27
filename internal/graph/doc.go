// Package graph implements the GraphQL API for customer orders.
//
// The schema lives in schema.graphqls. Running go generate rebuilds
// generated.go, the model package and the resolver stubs from that schema.
// Do not edit generated.go by hand.
//
//go:generate go tool gqlgen generate
package graph
