# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Module: `github.com/codinci/customerorders`, Go 1.24.

```bash
go build ./...                                  # build everything
go vet ./...                                    # the only linter configured; no golangci config exists
gofmt -w internal cmd                           # format before committing
go test ./...                                   # all tests
go test ./internal/graph/ -run TestName -count=1  # single test (-count=1 skips the cache)
CGO_ENABLED=1 go test -race ./...               # race detector; needs cgo (gcc) on Windows
go generate ./internal/graph/...                # regenerate gqlgen code after editing the schema
go run ./cmd/server                             # serve on :8080 (PORT overrides); playground at /
```

`gqlgen` is pinned as a Go tool in `go.mod` (`tool github.com/99designs/gqlgen`), so use `go tool gqlgen` or `go generate`, not a globally installed binary.

## Architecture

**Code generation pipeline.** `internal/graph/schema.graphqls` is the source of truth. `gqlgen.yml` drives `go generate`, which writes:
- `internal/graph/generated.go`: never edit by hand.
- `internal/graph/model/models_gen.go`: generated models.
- `internal/graph/schema.resolvers.go`: resolver method stubs. Hand-written bodies in this file survive regeneration, but a new schema field adds a new stub here.

Hand-written code: `internal/graph/resolver.go` (state), `internal/graph/scalar/decimal.go` (the `Decimal` scalar), `cmd/server/main.go` (wiring and HTTP).

Config decisions that shape the code:
- `omit_resolver_fields: true` with `Customer.orders` and `Order.customer` as field resolvers. Models therefore hold no pointer cycles, and relations are loaded by resolvers.
- `Decimal` is `scalar.Decimal`, a wrapper over `shopspring/decimal` that implements `MarshalGQL`/`UnmarshalGQL`. Money never goes through `float64`; it is serialized as a string.
- `Time` maps to gqlgen's `graphql.Time`. `Int` is 32-bit.

**Concurrency model.** gqlgen runs resolvers on separate goroutines, and it serializes the response *after* a resolver returns. `Resolver` (in `resolver.go`) therefore:
- guards `customers` and `orders` with `mu sync.RWMutex` (`RLock` for reads, `Lock` for writes), and
- stores `customerRecord`/`orderRecord` values that reference each other by ID, not by pointer.

Resolvers must return freshly built `model` values, never pointers into the stored slices. A lock that is released before serialization does not protect a returned pointer.

**Request flow.** `cmd/server/main.go` builds a gqlgen `handler` around `graph.NewExecutableSchema` with `graph.NewResolver()`, mounts `/query`, and serves the playground at `/`. Both the handler and the resolver are wired only there.

**Tests.** `internal/graph/testclient_test.go` provides `newTestClient(t)`, which builds a fresh `Resolver` per test using gqlgen's `client` package, so tests share no state. Tests are black-box (`package graph_test`) and send GraphQL strings.

## Project plan and workflow

`PLAN.md` is the learning roadmap and is git-ignored (see `.gitignore`); it is local-only. Work follows its phases in order. Phase 0 is in progress on branch `phase-0`.

The user codes the key logic (resolvers, locking, validation) and Claude writes the scaffolding and failing tests first. Do not implement the user's assigned functions when asked to hand them over; leave stubs returning `ErrNotImplemented`. Phase 1 replaces the in-memory slices with a `Store` interface, so keep the resolver free of storage detail where practical.
