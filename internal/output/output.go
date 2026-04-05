// Package output handles formatting CLI output as JSON or text tables.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Format represents an output format.
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatCSV   Format = "csv"
)

// Printer handles formatted output.
type Printer struct {
	Format Format
	Writer io.Writer
	Pretty bool
}

// NewPrinter creates a new Printer writing to stdout.
func NewPrinter(format Format, pretty bool) *Printer {
	return NewPrinterWithWriter(format, pretty, os.Stdout)
}

// NewPrinterWithWriter creates a new Printer writing to the given writer.
func NewPrinterWithWriter(format Format, pretty bool, w io.Writer) *Printer {
	return &Printer{
		Format: format,
		Writer: w,
		Pretty: pretty,
	}
}

// PrintJSON outputs raw JSON bytes, optionally pretty-printing them.
func (p *Printer) PrintJSON(data []byte) error {
	if p.Pretty {
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			// Fall back to raw output if we can't parse
			_, err := p.Writer.Write(data)
			fmt.Fprintln(p.Writer)
			return err
		}
		enc := json.NewEncoder(p.Writer)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	_, err := p.Writer.Write(data)
	fmt.Fprintln(p.Writer)
	return err
}

// PrintResult outputs the result in the configured format.
// For JSON format, it outputs the raw API response.
// For table/csv format, it uses the provided headers and row extractor.
func (p *Printer) PrintResult(data []byte, headers []string, extractRows func([]byte) ([][]string, error)) error {
	switch p.Format {
	case FormatTable:
		if extractRows == nil {
			return p.PrintJSON(data)
		}
		rows, err := extractRows(data)
		if err != nil {
			return err
		}
		return p.printTable(headers, rows)
	case FormatCSV:
		if extractRows == nil {
			return p.PrintJSON(data)
		}
		rows, err := extractRows(data)
		if err != nil {
			return err
		}
		return p.printCSV(headers, rows)
	default:
		return p.PrintJSON(data)
	}
}

func (p *Printer) printTable(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(p.Writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return w.Flush()
}

func (p *Printer) printCSV(headers []string, rows [][]string) error {
	fmt.Fprintln(p.Writer, strings.Join(headers, ","))
	for _, row := range rows {
		escaped := make([]string, len(row))
		for i, field := range row {
			if strings.ContainsAny(field, ",\"\n") {
				escaped[i] = `"` + strings.ReplaceAll(field, `"`, `""`) + `"`
			} else {
				escaped[i] = field
			}
		}
		fmt.Fprintln(p.Writer, strings.Join(escaped, ","))
	}
	return nil
}
