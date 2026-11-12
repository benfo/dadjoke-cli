package shellinit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShellAndResolveProfilePath(t *testing.T) {
	if s := DetectShell("windows", map[string]string{}); s != ShellPowerShell {
		t.Errorf("expected powershell on windows, got %s", s)
	}
	if s := DetectShell("darwin", map[string]string{"SHELL": "/bin/zsh"}); s != ShellZsh {
		t.Errorf("expected zsh when SHELL=/bin/zsh, got %s", s)
	}
	if s := DetectShell("linux", map[string]string{"SHELL": "/bin/bash"}); s != ShellBash {
		t.Errorf("expected bash when SHELL=/bin/bash, got %s", s)
	}

	home := "/home/tester"
	if p := ResolveProfilePath(ShellZsh, home, "linux", map[string]string{}); !strings.HasSuffix(p, ".zshrc") {
		t.Errorf("expected .zshrc path, got %s", p)
	}
	if p := ResolveProfilePath(ShellPowerShell, home, "windows", map[string]string{"PROFILE": `C:\Custom\profile.ps1`}); p != `C:\Custom\profile.ps1` {
		t.Errorf("expected custom PROFILE env path, got %s", p)
	}
}

func TestConfigureShellHook_InstallIdempotentUpdateAndUninstall(t *testing.T) {
	tempDir := t.TempDir()
	profilePath := filepath.Join(tempDir, ".bashrc")

	// 1. Dry run should not create the file
	dryRes, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		BinaryPath:  "/usr/local/bin/dadjoke",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("unexpected dry-run error: %v", err)
	}
	if dryRes.Action != ActionDryRun {
		t.Errorf("expected dry-run action, got %s", dryRes.Action)
	}
	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Errorf("expected profile file not to exist after dry-run")
	}

	// 2. First install
	res1, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		BinaryPath:  "/usr/local/bin/dadjoke",
	})
	if err != nil {
		t.Fatalf("unexpected install error: %v", err)
	}
	if res1.Action != ActionInstalled {
		t.Errorf("expected installed action, got %s", res1.Action)
	}
	if !strings.Contains(res1.Snippet, "--max-views 3") {
		t.Errorf("expected hook snippet to pass --max-views 3, got %s", res1.Snippet)
	}

	// 3. Second install with same binary -> already-installed
	res2, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		BinaryPath:  "/usr/local/bin/dadjoke",
	})
	if err != nil {
		t.Fatalf("unexpected 2nd install error: %v", err)
	}
	if res2.Action != ActionAlreadyInstalled {
		t.Errorf("expected already-installed, got %s", res2.Action)
	}

	// 4. Install with new binary path -> updated
	res3, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		BinaryPath:  "/opt/dadjoke/bin/dadjoke",
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if res3.Action != ActionUpdated {
		t.Errorf("expected updated action, got %s", res3.Action)
	}

	// 5. Uninstall -> uninstalled
	res4, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		Uninstall:   true,
	})
	if err != nil {
		t.Fatalf("unexpected uninstall error: %v", err)
	}
	if res4.Action != ActionUninstalled {
		t.Errorf("expected uninstalled action, got %s", res4.Action)
	}

	// 6. Second uninstall -> not-installed
	res5, err := ConfigureShellHook(Options{
		Shell:       ShellBash,
		ProfilePath: profilePath,
		Uninstall:   true,
	})
	if err != nil {
		t.Fatalf("unexpected 2nd uninstall error: %v", err)
	}
	if res5.Action != ActionNotInstalled {
		t.Errorf("expected not-installed action, got %s", res5.Action)
	}
}
