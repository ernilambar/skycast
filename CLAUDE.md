# skycast

Terminal weather CLI written in Go (module `github.com/ernilambar/skycast`, Cobra for CLI).

## Architecture

- `main.go` — Cobra root command, flag parsing (`--forecast`, `--units`, `--plain`), orchestrates `run()`: resolve location → fetch weather → render.
- `internal/weather` — HTTP calls to Open-Meteo (forecast + geocoding APIs) and IP-based geolocation (`ipwho.is`, falls back to `ip-api.com`). Returns plain data structs (`Location`, `CurrentWeather`, `DailyForecast`, `Weather`); no formatting/presentation logic.
- `internal/render` — All presentation: ANSI color helpers (`ansi.go`), ASCII-art weather icons + emoji condition icons (`icons.go`), and the `CurrentWeather`/`Forecast` print functions (`render.go`) with a `plain` (no color/emoji) vs. default (colored + emoji) mode. Both modes render the same data fields; they differ only in styling.
- `internal/spinner` — Hand-rolled terminal spinner (stderr) shown while network calls are in flight.

## Quality gates

Before marking a task complete, run:

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .   # must return no output
```
