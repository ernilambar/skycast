# skycast

Terminal weather CLI written in Go using Cobra. Fetches weather data from Open-Meteo API with IP-based geolocation fallback.

## Setup

```bash
git clone https://github.com/ernilambar/skycast.git
cd skycast
go build ./...
```

Requires Go 1.22+.

## Commands

```bash
go build ./...        # compile all packages
go run . [city]       # run locally (e.g., `go run . Tokyo --forecast`)
go test ./...         # run tests
go vet ./...          # static analysis
gofmt -l .            # check formatting (must return no output)
```

## Conventions

- **Separation of concerns**: `internal/weather` handles HTTP calls and returns plain structs only — no formatting. `internal/render` handles all presentation (ANSI colors, icons, layout).
- **Data flow**: `main.go` orchestrates: resolve location → fetch weather → render. The `run()` function ties these stages together.
- **Error handling**: Network errors return user-friendly messages (e.g., "unable to reach the weather service") without exposing internal details.
- **Testing**: HTTP clients are mocked via `stubTransport` — tests never make real network calls. Use `withMockResponse()` and `withMockResponsesByHost()` helpers.
- **No external test dependencies**: Tests use only the standard library (`testing`, `net/http`).

## Quality Gate

Before declaring a task complete, run and verify all pass (exit code 0):

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .
```
