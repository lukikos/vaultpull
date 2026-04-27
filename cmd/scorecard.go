package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpull/internal/audit"
)

func init() {
	var logFile string

	cmd := &cobra.Command{
		Use:   "scorecard",
		Short: "Show health scores for each synced secret path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScorecard(logFile)
		},
	}

	cmd.Flags().StringVar(&logFile, "log", ".vaultpull-audit.log", "path to audit log file")
	rootCmd.AddCommand(cmd)
}

func runScorecard(logFile string) error {
	entries, err := audit.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	results := audit.Scorecard(entries)
	if len(results) == 0 {
		fmt.Println("No audit entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SCORE\tPATH\tTOTAL\tRECENT(7d)\tLAST SYNCED\tNOTES")
	fmt.Fprintln(w, "-----\t----\t-----\t----------\t-----------\t-----")

	for _, r := range results {
		lastSynced := r.LastSyncedAt.Format("2006-01-02 15:04")
		notes := "-"
		if len(r.Notes) > 0 {
			notes = r.Notes[0]
			if len(r.Notes) > 1 {
				notes += fmt.Sprintf(" (+%d)", len(r.Notes)-1)
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\t%s\n",
			r.Score, r.Path, r.TotalSyncs, r.RecentSyncs, lastSynced, notes)
	}

	return w.Flush()
}
