package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func init() {
	schemaSaveCmd := &cobra.Command{
		Use:   "save",
		Short: "Save a schema version to the audit directory",
		RunE:  runSchemaSave,
	}
	schemaSaveCmd.Flags().IntP("version", "v", 1, "Schema version number")
	schemaSaveCmd.Flags().StringSlice("fields", []string{"key", "path", "action"}, "Required fields")
	schemaSaveCmd.Flags().String("dir", ".", "Audit directory")

	schemaValidateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate audit log entries against the saved schema",
		RunE:  runSchemaValidate,
	}
	schemaValidateCmd.Flags().String("log", "audit.log", "Audit log file")
	schemaValidateCmd.Flags().String("dir", ".", "Audit directory")

	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage audit log schema versions",
	}
	schemaCmd.AddCommand(schemaSaveCmd, schemaValidateCmd)
	rootCmd.AddCommand(schemaCmd)
}

func runSchemaSave(cmd *cobra.Command, _ []string) error {
	version, _ := cmd.Flags().GetInt("version")
	fields, _ := cmd.Flags().GetStringSlice("fields")
	dir, _ := cmd.Flags().GetString("dir")

	sv := audit.SchemaVersion{
		Version: version,
		Fields:  fields,
	}
	if err := audit.SaveSchema(dir, sv); err != nil {
		return fmt.Errorf("save schema: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Schema v%d saved (fields: %s)\n", version, strings.Join(fields, ", "))
	return nil
}

func runSchemaValidate(cmd *cobra.Command, _ []string) error {
	logFile, _ := cmd.Flags().GetString("log")
	dir, _ := cmd.Flags().GetString("dir")

	sv, err := audit.LoadSchema(dir)
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	entries, err := audit.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}

	verr := audit.ValidateEntries(entries, sv)
	if verr == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "OK: all entries are valid")
		return nil
	}
	for _, issue := range verr.Issues {
		fmt.Fprintln(cmd.OutOrStdout(), "ISSUE:", issue)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d issue(s) found\n", len(verr.Issues))
	os.Exit(1)
	return nil
}
