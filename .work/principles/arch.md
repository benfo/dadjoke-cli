1. Write all source and test code in strict TypeScript.
2. Use established libraries instead of writing custom boilerplate (e.g. commander, chalk, boxen, ora).
3. Cache daily joke locally with the current date key.
4. Fail silently or fall back to an offline/cached joke on network error so shell startup is never blocked.
5. Keep startup latency low when reading from the daily cache.
