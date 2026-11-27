package graph_test

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/codinci/customerorders/internal/graph"
)

// newTestClient builds a GraphQL test client around a fresh Resolver, so each
// test starts with empty state and no test depends on another.
func newTestClient(t *testing.T) *client.Client {
	t.Helper()
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: graph.NewResolver(),
	}))
	return client.New(srv)
}
