1. Use spf13/cobra for CLI argument and subcommand parsing.
2. Use charmbracelet/lipgloss for colors and rounded boxed output instead of raw ANSI codes.
3. Default invocation outputs the daily dad joke cleanly with no verbose log spam.
4. Support flags like --random, --refresh, --plain, --force, and --max-views to control joke output.
5. Provide an idempotent init command that detects shell and updates user profile safely.
6. Never produce panics or nonzero exit codes on shell startup that break user shells.
