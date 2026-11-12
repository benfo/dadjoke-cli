package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/benfourie/daily-dadjoke/internal/joke"
	"github.com/spf13/cobra"
)

// Version is the CLI semantic version.
const Version = "1.0.0"

// RootOptions holds flags for the root dadjoke command.
type RootOptions struct {
	Refresh   bool
	Random    bool
	Plain     bool
	Force     bool
	MaxViews  int
	CachePath string
}

// NewRootCmd creates the Cobra root command for dadjoke.
func NewRootCmd() *cobra.Command {
	opts := &RootOptions{}

	cmd := &cobra.Command{
		Use:     "dadjoke",
		Short:   "Daily Dad Joke CLI - Welcome your terminal sessions with humor!",
		Version: Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var maxViewsPtr *int
			if cmd.Flags().Changed("max-views") {
				if opts.MaxViews < 0 {
					return fmt.Errorf("error: max-views must be a non-negative integer (0 for unlimited)")
				}
				val := opts.MaxViews
				maxViewsPtr = &val
			}

			var res joke.Result
			title := "Today's Dad Joke"

			if opts.Random {
				res = joke.GetRandomJoke(joke.FetchOptions{})
				title = "Random Dad Joke"
			} else {
				res = joke.GetDailyJoke(joke.DailyJokeOptions{
					ForceRefresh: opts.Refresh,
					ForceShow:    opts.Force,
					MaxViews:     maxViewsPtr,
					CachePath:    opts.CachePath,
				})
				if opts.Refresh {
					title = "Today's Dad Joke (Refreshed)"
				}
			}

			if res.Joke != "" && !res.Hidden {
				formatted := joke.FormatJoke(res.Joke, joke.FormatOptions{
					Plain: opts.Plain,
					Title: title,
				})
				out := cmd.OutOrStdout()
				_, _ = io.WriteString(out, formatted+"\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Refresh, "refresh", "r", false, "Force refresh today's joke from the API")
	cmd.Flags().BoolVar(&opts.Random, "random", false, "Fetch a new random joke immediately without updating daily cache")
	cmd.Flags().BoolVarP(&opts.Plain, "plain", "p", false, "Print plain text joke without box or color formatting")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Always show today's joke even if daily session limit is reached")
	cmd.Flags().IntVarP(&opts.MaxViews, "max-views", "m", joke.DefaultMaxDailyViews, "Max terminal sessions per day to display the joke (default: 3, 0 for unlimited)")

	cmd.AddCommand(NewInitCmd())

	return cmd
}

// Execute runs the root command and exits cleanly unless an explicit argument validation error occurs.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
