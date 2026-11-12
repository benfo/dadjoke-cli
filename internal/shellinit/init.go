package shellinit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// SupportedShell represents a shell supported by dadjoke init.
type SupportedShell string

const (
	ShellPowerShell SupportedShell = "powershell"
	ShellBash       SupportedShell = "bash"
	ShellZsh        SupportedShell = "zsh"

	HookStartMarker = "# >>> daily-dadjoke initialize >>>"
	HookEndMarker   = "# <<< daily-dadjoke initialize <<<"
)

// ActionStatus describes the result of ConfigureShellHook.
type ActionStatus string

const (
	ActionInstalled        ActionStatus = "installed"
	ActionUpdated          ActionStatus = "updated"
	ActionAlreadyInstalled ActionStatus = "already-installed"
	ActionUninstalled      ActionStatus = "uninstalled"
	ActionNotInstalled     ActionStatus = "not-installed"
	ActionDryRun           ActionStatus = "dry-run"
)

// Options configures ConfigureShellHook.
type Options struct {
	Shell       SupportedShell
	ProfilePath string
	BinaryPath  string
	DryRun      bool
	Uninstall   bool
	HomeDir     string
	GOOS        string
	Env         map[string]string
}

// Result contains the outcome of ConfigureShellHook.
type Result struct {
	Shell       SupportedShell
	ProfilePath string
	Snippet     string
	Action      ActionStatus
}

// ParseShell normalizes and validates a shell name.
func ParseShell(value string) (SupportedShell, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "powershell", "pwsh", "ps":
		return ShellPowerShell, nil
	case "bash":
		return ShellBash, nil
	case "zsh":
		return ShellZsh, nil
	default:
		return "", fmt.Errorf("shell must be one of: powershell, bash, zsh")
	}
}

// GetDefaultBinaryPath resolves the current executable path, falling back to "dadjoke".
func GetDefaultBinaryPath() string {
	exe, err := os.Executable()
	if err != nil || strings.TrimSpace(exe) == "" {
		return "dadjoke"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil && strings.TrimSpace(resolved) != "" {
		return resolved
	}
	return exe
}

// currentEnvMap converts os.Environ() into a map.
func currentEnvMap() map[string]string {
	env := make(map[string]string)
	for _, entry := range os.Environ() {
		if k, v, ok := strings.Cut(entry, "="); ok {
			env[k] = v
		}
	}
	return env
}

// DetectShell determines the user's shell from environment and OS platform.
func DetectShell(goos string, env map[string]string) SupportedShell {
	if goos == "" {
		goos = runtime.GOOS
	}
	if env == nil {
		env = currentEnvMap()
	}

	shellEnv := strings.ToLower(env["SHELL"])
	if strings.Contains(shellEnv, "zsh") {
		return ShellZsh
	}
	if strings.Contains(shellEnv, "bash") {
		return ShellBash
	}
	if strings.Contains(shellEnv, "pwsh") || strings.Contains(shellEnv, "powershell") {
		return ShellPowerShell
	}
	if goos == "windows" {
		return ShellPowerShell
	}
	return ShellBash
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// ResolveProfilePath returns the standard profile file path for the target shell and OS.
func ResolveProfilePath(shell SupportedShell, homeDir string, goos string, env map[string]string) string {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	if goos == "" {
		goos = runtime.GOOS
	}
	if env == nil {
		env = currentEnvMap()
	}

	if shell == ShellPowerShell {
		if prof := strings.TrimSpace(env["PROFILE"]); prof != "" {
			return prof
		}
		if goos == "windows" {
			pwsh7Dir := filepath.Join(homeDir, "Documents", "PowerShell")
			winPsDir := filepath.Join(homeDir, "Documents", "WindowsPowerShell")
			if !pathExists(pwsh7Dir) && pathExists(winPsDir) {
				return filepath.Join(winPsDir, "Microsoft.PowerShell_profile.ps1")
			}
			return filepath.Join(pwsh7Dir, "Microsoft.PowerShell_profile.ps1")
		}
		return filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	}

	if shell == ShellZsh {
		return filepath.Join(homeDir, ".zshrc")
	}

	// Bash
	bashrc := filepath.Join(homeDir, ".bashrc")
	bashProfile := filepath.Join(homeDir, ".bash_profile")
	if goos == "darwin" && !pathExists(bashrc) && pathExists(bashProfile) {
		return bashProfile
	}
	return bashrc
}

// BuildHookSnippet generates the shell startup hook block for the given shell and binary path.
func BuildHookSnippet(shell SupportedShell, binaryPath string) string {
	if strings.TrimSpace(binaryPath) == "" {
		binaryPath = GetDefaultBinaryPath()
	}

	if shell == ShellPowerShell {
		escapedWinPath := strings.ReplaceAll(binaryPath, `"`, "`\"")
		return strings.Join([]string{
			HookStartMarker,
			fmt.Sprintf(`if (Test-Path "%s") { & "%s" --max-views 3 } elseif (Get-Command dadjoke -ErrorAction SilentlyContinue) { dadjoke --max-views 3 }`, escapedWinPath, escapedWinPath),
			HookEndMarker,
		}, "\n")
	}

	normalizedPath := strings.ReplaceAll(binaryPath, `\`, `/`)
	return strings.Join([]string{
		HookStartMarker,
		fmt.Sprintf(`if [ -x "%s" ]; then "%s" --max-views 3; elif command -v dadjoke >/dev/null 2>&1; then dadjoke --max-views 3; fi`, normalizedPath, normalizedPath),
		HookEndMarker,
	}, "\n")
}

var hookBlockRegex = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(HookStartMarker) + `.*?` + regexp.QuoteMeta(HookEndMarker))
var removeBlockRegex = regexp.MustCompile(`(?s)\r?\n?` + regexp.QuoteMeta(HookStartMarker) + `.*?` + regexp.QuoteMeta(HookEndMarker) + `\r?\n?`)
var multiNewlineRegex = regexp.MustCompile(`\n{3,}`)

// ApplyHookToContent inserts or updates the hook snippet in a profile file's content.
func ApplyHookToContent(existingContent, snippet string) (string, ActionStatus) {
	if hookBlockRegex.MatchString(existingContent) {
		match := hookBlockRegex.FindString(existingContent)
		normMatch := strings.ReplaceAll(match, "\r\n", "\n")
		normSnippet := strings.ReplaceAll(snippet, "\r\n", "\n")
		if normMatch == normSnippet {
			return existingContent, ActionAlreadyInstalled
		}
		updated := hookBlockRegex.ReplaceAllString(existingContent, snippet)
		return updated, ActionUpdated
	}

	trimmed := strings.TrimRight(existingContent, " \t\r\n")
	prefix := ""
	if len(trimmed) > 0 {
		prefix = trimmed + "\n\n"
	}
	return prefix + snippet + "\n", ActionInstalled
}

// RemoveHookFromContent removes the hook block from a profile file's content.
func RemoveHookFromContent(existingContent string) (string, bool) {
	if !hookBlockRegex.MatchString(existingContent) {
		return existingContent, false
	}
	updated := removeBlockRegex.ReplaceAllString(existingContent, "\n")
	updated = strings.ReplaceAll(updated, "\r\n", "\n")
	updated = multiNewlineRegex.ReplaceAllString(updated, "\n\n")
	updated = strings.TrimLeft(updated, " \t\r\n")
	return updated, true
}

// ConfigureShellHook installs, updates, previews, or uninstalls the shell startup hook.
func ConfigureShellHook(opts Options) (Result, error) {
	shell := opts.Shell
	if shell == "" {
		shell = DetectShell(opts.GOOS, opts.Env)
	}

	profilePath := opts.ProfilePath
	if strings.TrimSpace(profilePath) == "" {
		profilePath = ResolveProfilePath(shell, opts.HomeDir, opts.GOOS, opts.Env)
	}

	binaryPath := opts.BinaryPath
	if strings.TrimSpace(binaryPath) == "" {
		binaryPath = GetDefaultBinaryPath()
	}

	snippet := BuildHookSnippet(shell, binaryPath)

	if opts.DryRun {
		return Result{
			Shell:       shell,
			ProfilePath: profilePath,
			Snippet:     snippet,
			Action:      ActionDryRun,
		}, nil
	}

	existing := ""
	if data, err := os.ReadFile(profilePath); err == nil {
		existing = string(data)
	}

	if opts.Uninstall {
		content, removed := RemoveHookFromContent(existing)
		if removed {
			if err := writeProfileFile(profilePath, content); err != nil {
				return Result{}, err
			}
			return Result{
				Shell:       shell,
				ProfilePath: profilePath,
				Snippet:     snippet,
				Action:      ActionUninstalled,
			}, nil
		}
		return Result{
			Shell:       shell,
			ProfilePath: profilePath,
			Snippet:     snippet,
			Action:      ActionNotInstalled,
		}, nil
	}

	content, status := ApplyHookToContent(existing, snippet)
	if status != ActionAlreadyInstalled || !pathExists(profilePath) {
		if err := writeProfileFile(profilePath, content); err != nil {
			return Result{}, err
		}
	}

	return Result{
		Shell:       shell,
		ProfilePath: profilePath,
		Snippet:     snippet,
		Action:      status,
	}, nil
}

func writeProfileFile(profilePath, content string) error {
	dir := filepath.Dir(profilePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(profilePath, []byte(content), 0o644)
}
