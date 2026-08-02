# skycast

A terminal weather app.

## Install

**Prebuilt binary** (macOS):

```bash
curl -L -o skycast https://github.com/ernilambar/skycast/releases/latest/download/skycast-darwin-arm64 # Apple Silicon
# or: skycast-darwin-amd64 for Intel Macs
xattr -d com.apple.quarantine skycast 2>/dev/null || true
chmod +x skycast
sudo mv skycast /usr/local/bin/
```

**From source** (requires Go 1.22+):

```bash
git clone https://github.com/ernilambar/skycast.git
cd skycast
go build -o skycast .
sudo mv skycast /usr/local/bin/
```

## Usage

```bash
skycast [city] [options]
```

| Option | Description |
| --- | --- |
| `[city]` | City name to fetch weather for (defaults to IP location) |
| `-f, --forecast` | Show 5-day forecast |
| `-u, --units <type>` | Temperature units: `metric` (C) or `imperial` (F). Default: `metric` |
| `-p, --plain` | Output simple plain text without colors or icons |

### Examples

```bash
skycast
skycast Tokyo
skycast "New York" --forecast --units imperial
skycast London --plain
```

## Development

```bash
go build ./...    # compile all packages
go test ./...     # run tests
go vet ./...      # static analysis
gofmt -l .        # check formatting
```

### Manual Testing

```bash
go run . Tokyo --forecast
go run . "New York" --units imperial --plain
go run .
```

## Copyright and License

This project is licensed under the [MIT](http://opensource.org/licenses/MIT).

2026 &copy; [Nilambar Sharma](https://www.nilambar.net).
