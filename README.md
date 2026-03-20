# Load Balancer

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
