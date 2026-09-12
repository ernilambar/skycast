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
go build ./...              # compile all packages
go run . [city]             # run locally (e.g., `go run . Tokyo --forecast`)
go test ./...               # run tests
go test -race -cover ./...  # run tests with the race detector and coverage
go vet ./...                # static analysis
gofmt -l .                  # check formatting (must return no output)
```

## Conventions

- **Separation of concerns**: `internal/weather` handles HTTP calls and returns plain structs only — no formatting. `internal/render` handles all presentation (ANSI colors, icons, layout).
- **Data flow**: `main.go` orchestrates: resolve location → fetch weather → render. The `run()` function ties these stages together and takes an `io.Writer` so output is testable.
- **Error handling**: Network errors return user-friendly messages (e.g., "unable to reach the weather service") without exposing internal details. The `weather` package exposes sentinel errors (`ErrGeocodingService`, `ErrWeatherService`, `ErrIPDetection`, …) so callers and tests match them with `errors.Is`.
- **Testing**: HTTP clients are mocked via a `stubTransport` that records every request — tests never make real network calls and can assert on the outgoing URL, host and query params. Use the `withMockResponse()`, `withMockError()` and `withMockResponsesByHost()` helpers. `main` uses package-level function seams (`geocodeCity`, `locateByIP`, `fetchWeather`, `newSpinner`) that tests override. Render functions take an `io.Writer`; the spinner accepts a custom output and interval.
- **No external test dependencies**: Tests use only the standard library (`testing`, `net/http`).

## Quality Gate

Before declaring a task complete, run and verify all pass (exit code 0):

```bash
go build ./...
go test -race -cover ./...
go vet ./...
gofmt -l .
```

