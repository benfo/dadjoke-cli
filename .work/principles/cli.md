1. Use commander for CLI argument and subcommand parsing.
2. Use chalk, boxen, and ora for colors, boxed output, and loading spinners instead of raw ANSI codes.
3. Default invocation outputs the daily dad joke cleanly with no verbose log spam.
4. Support flags like --random or --refresh to bypass daily cache when explicitly requested.
5. Provide an idempotent init command that detects shell and updates user profile safely.
6. Never produce unhandled promise rejections or nonzero exit codes on shell startup that break user shells.
