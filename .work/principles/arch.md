1. Write all source and test code in idiomatic Go with standard go test coverage.
2. Use spf13/cobra and charmbracelet/lipgloss instead of custom CLI/styling boilerplate.
3. Cache daily joke locally in ~/.dadjoke-cache.json with the current YYYY-MM-DD date key.
4. Fail silently or fall back to an offline/cached joke on network error so shell startup is never blocked.
5. Keep startup latency under 5ms when reading from the daily cache.
