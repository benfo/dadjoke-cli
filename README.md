# dadjoke-cli

A fast, zero-runtime-dependency Go CLI that displays a cached daily dad joke when you open a terminal.

- **Instant startup (<5ms):** Caches today's joke from [icanhazdadjoke.com](https://icanhazdadjoke.com) in `~/.dadjoke-cache.json`.
- **No terminal spam:** Shows today's joke for the first 3 terminal sessions per day by default, then stays quiet.
- **Safe for shell profiles:** Falls back to built-in jokes offline and never exits non-zero on network errors.

## Installation

### 1. Install the binary

**From GitHub Releases (recommended):**

Download the binary for your OS and architecture from [Releases](https://github.com/benfo/dadjoke-cli/releases), rename it to `dadjoke` (`dadjoke.exe` on Windows), and place it in a directory on your `PATH`.

```bash
# Linux / macOS: make executable
chmod +x dadjoke
```

**With Go:**

```bash
go install github.com/benfo/dadjoke-cli@latest
```

### 2. Hook your shell startup

Auto-detect your shell (`powershell`, `bash`, or `zsh`) and add the startup hook to your profile:

```bash
dadjoke init
```

Preview or remove the hook at any time:

```bash
dadjoke init --dry-run
dadjoke init --uninstall
```

## Usage

```bash
dadjoke                  # Show today's joke (up to daily view limit)
dadjoke --force          # Show today's joke even after hitting the daily view limit
dadjoke --refresh        # Fetch and cache a new daily joke
dadjoke --random         # Fetch a one-off random joke without changing today's cache
dadjoke --plain          # Print plain text without box or colors
```

### Flags

| Flag | Short | Description |
| :--- | :--- | :--- |
| `--force` | `-f` | Always show today's joke even if daily session limit is reached |
| `--refresh` | `-r` | Force refresh today's joke from the API |
| `--random` | | Fetch a random joke without updating the daily cache |
| `--plain` | `-p` | Print plain text without rounded box or ANSI colors |
| `--max-views <n>` | `-m` | Max terminal sessions per day to display the joke (default `3`, `0` for unlimited) |

### `dadjoke init` Flags

| Flag | Short | Description |
| :--- | :--- | :--- |
| `--shell <name>` | `-s` | Target shell (`powershell`, `bash`, `zsh`) |
| `--profile <path>` | | Custom shell profile path |
| `--dry-run` | `-d` | Preview the profile snippet without writing to disk |
| `--uninstall` | `-u` | Remove the startup hook from your shell profile |

## Building from Source

```bash
# Build for current platform
go build -o dadjoke .

# Build all cross-platform binaries into dist/
go run ./scripts/build.go

# Run tests
go test ./...
```
