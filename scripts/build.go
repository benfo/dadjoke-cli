//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type target struct {
	goos   string
	goarch string
	output string
}

var targets = []target{
	{goos: "windows", goarch: "amd64", output: "dadjoke-windows-amd64.exe"},
	{goos: "windows", goarch: "arm64", output: "dadjoke-windows-arm64.exe"},
	{goos: "darwin", goarch: "amd64", output: "dadjoke-darwin-amd64"},
	{goos: "darwin", goarch: "arm64", output: "dadjoke-darwin-arm64"},
	{goos: "linux", goarch: "amd64", output: "dadjoke-linux-amd64"},
	{goos: "linux", goarch: "arm64", output: "dadjoke-linux-arm64"},
}

func main() {
	distDir := "dist"
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create %s: %v\n", distDir, err)
		os.Exit(1)
	}

	for _, t := range targets {
		outPath := filepath.Join(distDir, t.output)
		fmt.Printf("Building %s/%s -> %s ...\n", t.goos, t.goarch, outPath)

		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", outPath, ".")
		cmd.Env = append(os.Environ(),
			"CGO_ENABLED=0",
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "build failed for %s/%s: %v\n", t.goos, t.goarch, err)
			os.Exit(1)
		}
	}

	fmt.Println("✔ Built all cross-platform binaries in dist/")
}
