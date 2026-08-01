# skycast

A terminal weather app.

## Install

**From source** (requires Node.js 22+ and [Bun](https://bun.sh)):

```bash
git clone https://github.com/ernilambar/skycast.git
cd skycast
bun install
bun run compile
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
| `-p, --plain` | Output simple plain text without ASCII borders |

### Examples

```bash
skycast
skycast Tokyo
skycast "New York" --forecast --units imperial
skycast London --plain
```

## Development

After cloning and `bun install`:

```bash
bun run test        # run tests (node:test)
bun run lint         # check code style (standard)
bun run format       # auto-fix code style (standard --fix)
bun run compile      # build a standalone binary (./skycast)
```

### Manual Testing

```bash
bun src/index.mjs Tokyo --forecast
bun src/index.mjs "New York" --units imperial --plain
bun src/index.mjs
```

## Copyright and License

This project is licensed under the [MIT](http://opensource.org/licenses/MIT).

2026 &copy; [Nilambar Sharma](https://www.nilambar.net).
