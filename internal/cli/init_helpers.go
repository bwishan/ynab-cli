package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
	"github.com/bwishan/ynab-cli/internal/output"
)

// initPrinterAndConfig sets up output formatting and loads stored config.
func (a *App) initPrinterAndConfig(cmd *cobra.Command) error {
	formatStr, _ := cmd.Flags().GetString("output")
	format := output.Format(formatStr)
	switch format {
	case output.FormatJSON, output.FormatTable, output.FormatCSV:
	default:
		return fmt.Errorf("invalid output format %q (valid: json, table, csv)", formatStr)
	}

	pretty, _ := cmd.Flags().GetBool("pretty")
	a.printer = output.NewPrinter(format, pretty)

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.cfg = cfg
	return nil
}
