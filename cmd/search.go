package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultpull/internal/audit"
)

var (
	searchKeyContains  string
	searchPathContains string
	searchActionIn     []string
	searchLogFile      string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search audit log entries by key, path, or action",
	Long: `Search through audit log entries using one or more filters.

Filters are combined with AND logic — only entries matching all
provided criteria are returned.

Examples:
  vaultpull search --key DATABASE
  vaultpull search --path secret/app --action added
  vaultpull search --key API --action updated,added`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchKeyContains, "key", "", "Filter entries where the key contains this substring (case-insensitive)")
	searchCmd.Flags().StringVar(&searchPathContains, "path", "", "Filter entries where the secret path contains this substring")
	searchCmd.Flags().StringSliceVar(&searchActionIn, "action", nil, "Filter by action(s): added, updated, unchanged (comma-separated)")
	searchCmd.Flags().StringVar(&searchLogFile, "log", ".vaultpull-audit.log", "Path to the audit log file")
}

func runSearch(cmd *cobra.Command, args []string) error {
	if searchKeyContains == "" && searchPathContains == "" && len(searchActionIn) == 0 {
		return fmt.Errorf("at least one search filter (--key, --path, --action) must be provided")
	}

	q := audit.SearchQuery{
		KeyContains:  searchKeyContains,
		PathContains: searchPathContains,
		ActionIn:     searchActionIn,
	}

	results, err := audit.Search(searchLogFile, q)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No entries matched the search criteria.")
		return nil
	}

	fmt.Printf("Found %d matching entry/entries:\n\n", len(results))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tPATH\tKEY\tACTION")
	fmt.Fprintln(w, "---------\t----\t---\t------")

	for _, e := range results {
		ts := e.Timestamp.Format(time.RFC3339)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ts, e.Path, e.Key, e.Action)
	}

	return w.Flush()
}
