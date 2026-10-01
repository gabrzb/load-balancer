# HTTP Load Balancer

A Go learning project that accepts requests on port `8000` and distributes them across HTTP backends in round-robin order. It uses only the standard library.

The load balancer checks connectivity to each backend at startup and every 20 seconds. It skips unavailable backends, returns `503` when none are available, and marks a backend as unavailable if the proxy fails (returning `502` for that request).

## Requirements

- Go 1.26.1 or later.

## Run

From the project root, start the following commands in three separate terminals:

```sh
go run ./backend 8081
```

```sh
go run ./backend 8082
```

```sh
go run . http://localhost:8081 http://localhost:8082
```

Send a few requests to the load balancer:

```sh
curl http://localhost:8000/api/visits
```

The response is JSON such as `{"instance":"8081","visits":1}`. The `instance` field alternates between `8081` and `8082`. Each process keeps its own in-memory counter, which resets when the process restarts.

The sample backend also responds to `GET /health` with status `204`. If no backend address is provided to the load balancer, it uses `http://localhost:8080`. If no port is provided to the sample backend, it listens on port `8080`.

## Verify

```sh
go test ./...
go vet ./...
```
