package joke

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FormatOptions configures how a joke is rendered for the terminal.
type FormatOptions struct {
	Plain bool
	Title string
}

// FormatJoke renders the joke either as plain text or inside a styled Lipgloss rounded box.
func FormatJoke(jokeText string, opts FormatOptions) string {
	if opts.Plain {
		return jokeText
	}

	title := opts.Title
	if strings.TrimSpace(title) == "" {
		title = "Today's Dad Joke"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")) // Bright Yellow

	jokeStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("14")) // Bright Cyan

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")).
		Padding(0, 1)

	header := titleStyle.Render(fmt.Sprintf("🎭 %s", title))
	body := jokeStyle.Render(fmt.Sprintf("%q", jokeText))

	content := fmt.Sprintf("%s\n%s", header, body)
	return boxStyle.Render(content)
}
