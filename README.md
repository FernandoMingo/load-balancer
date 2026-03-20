# Load Balancer

This project is a lightweight HTTP load balancer written in Go. It forwards incoming requests to a set of backend services using a round-robin strategy, with middleware support for common production concerns.

## Highlights

- Round-robin request distribution across multiple backends
- Configurable backend list and listening port via CLI flags
- Middleware for panic recovery and rate limiting
- Unit tests for middleware behavior

- `cmd/load-balancer`: application entrypoint (`main`)
- `pkg/loadbalancer`: reusable load balancer package with business logic

## Run

```bash
go run ./cmd/load-balancer -backends "http://localhost:8081,http://localhost:8082" -port 3030
```

## Test

```bash
go test ./... -v
```
