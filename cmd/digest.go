package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	var save bool

	digestCmd := &cobra.Command{
		Use:   "digest",
		Short: "Compute and display SHA-256 digests of current secret state per path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDigest(save)
		},
	}

	digestCmd.Flags().BoolVar(&save, "save", false, "Persist the digest report to disk")
	rootCmd.AddCommand(digestCmd)
}

func runDigest(save bool) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("digest: get working dir: %w", err)
	}

	report, err := audit.ComputeDigests(dir)
	if err != nil {
		return fmt.Errorf("digest: compute: %w", err)
	}

	if len(report.Entries) == 0 {
		fmt.Println("No audit entries found — nothing to digest.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tKEYS\tDIGEST\tCOMPUTED AT")
	fmt.Fprintln(w, "----\t-----\t------\t-----------")
	for _, e := range report.Entries {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			e.Path,
			e.KeyCount,
			e.Digest[:16]+"...",
			e.ComputedAt.Format("2006-01-02 15:04:05"),
		)
	}
	w.Flush()

	if save {
		if err := audit.SaveDigests(dir, report); err != nil {
			return fmt.Errorf("digest: save: %w", err)
		}
		fmt.Printf("\nDigest report saved (%d path(s))\n", len(report.Entries))
	}

	return nil
}
