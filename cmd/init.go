package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/benfo/dadjoke-cli/internal/shellinit"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// InitCmdOptions holds flags for dadjoke init.
type InitCmdOptions struct {
	Shell     string
	Profile   string
	DryRun    bool
	Uninstall bool
}

// NewInitCmd creates the `dadjoke init` Cobra subcommand.
func NewInitCmd() *cobra.Command {
	opts := &InitCmdOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Configure terminal startup hook (PowerShell, Bash, Zsh)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var parsedShell shellinit.SupportedShell
			if strings.TrimSpace(opts.Shell) != "" {
				s, err := shellinit.ParseShell(opts.Shell)
				if err != nil {
					return fmt.Errorf("error: %w", err)
				}
				parsedShell = s
			}

			result, err := shellinit.ConfigureShellHook(shellinit.Options{
				Shell:       parsedShell,
				ProfilePath: opts.Profile,
				DryRun:      opts.DryRun,
				Uninstall:   opts.Uninstall,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			boldStyle := lipgloss.NewStyle().Bold(true)
			cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
			greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
			yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
			dimStyle := lipgloss.NewStyle().Faint(true)
			underlineStyle := lipgloss.NewStyle().Underline(true)

			if result.Action == shellinit.ActionDryRun {
				title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render("🔍 Shell Hook Preview (Dry Run)")
				body := strings.Join([]string{
					title,
					fmt.Sprintf("%s   %s", boldStyle.Render("Shell:"), cyanStyle.Render(string(result.Shell))),
					fmt.Sprintf("%s %s", boldStyle.Render("Profile:"), underlineStyle.Render(result.ProfilePath)),
					"",
					dimStyle.Render(result.Snippet),
				}, "\n")
				box := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("14")).
					Padding(0, 1).
					Render(body)
				_, _ = io.WriteString(out, box+"\n")
				return nil
			}

			if result.Action == shellinit.ActionUninstalled {
				msg := greenStyle.Render(fmt.Sprintf("✔ Removed daily dad joke hook from %s", underlineStyle.Render(result.ProfilePath)))
				_, _ = io.WriteString(out, msg+"\n")
				return nil
			}

			if result.Action == shellinit.ActionNotInstalled {
				msg := yellowStyle.Render(fmt.Sprintf("ℹ No daily dad joke hook found in %s", underlineStyle.Render(result.ProfilePath)))
				_, _ = io.WriteString(out, msg+"\n")
				return nil
			}

			statusLabel := greenStyle.Render("✔ Installed startup hook in your shell profile!")
			if result.Action == shellinit.ActionAlreadyInstalled {
				statusLabel = greenStyle.Render("✔ Startup hook is already up to date!")
			} else if result.Action == shellinit.ActionUpdated {
				statusLabel = greenStyle.Render("✔ Updated startup hook in your shell profile!")
			}

			title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render("🎭 Daily Dad Joke Init")
			summary := strings.Join([]string{
				title,
				statusLabel,
				fmt.Sprintf("%s   %s", boldStyle.Render("Shell:"), cyanStyle.Render(string(result.Shell))),
				fmt.Sprintf("%s %s", boldStyle.Render("Profile:"), underlineStyle.Render(result.ProfilePath)),
				"",
				dimStyle.Render("Open a new terminal session (or reload your profile) to see today's joke!"),
			}, "\n")

			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("10")).
				Padding(0, 1).
				Render(summary)
			_, _ = io.WriteString(out, box+"\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Shell, "shell", "s", "", "Target shell (powershell, bash, zsh)")
	cmd.Flags().StringVar(&opts.Profile, "profile", "", "Custom shell profile file path")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "d", false, "Preview the profile hook snippet without writing to disk")
	cmd.Flags().BoolVarP(&opts.Uninstall, "uninstall", "u", false, "Remove the daily dad joke hook from your shell profile")

	return cmd
}
