package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSON_RawOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf, Pretty: false}

	data := []byte(`{"data":{"user":{"id":"abc-123"}}}`)
	if err := p.PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	got := buf.String()
	// Should output raw JSON followed by a newline
	if got != `{"data":{"user":{"id":"abc-123"}}}`+"\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestPrintJSON_PrettyOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf, Pretty: true}

	data := []byte(`{"key":"val","num":42}`)
	if err := p.PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	got := buf.String()
	// Pretty output should be indented
	if !strings.Contains(got, "  ") {
		t.Errorf("expected indented output, got: %q", got)
	}

	// Should be valid JSON
	var v interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(got)), &v); err != nil {
		t.Errorf("pretty output is not valid JSON: %v", err)
	}

	// Verify values preserved
	m := v.(map[string]interface{})
	if m["key"] != "val" {
		t.Errorf("expected key=val, got %v", m["key"])
	}
}

func TestPrintResult_FormatJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf, Pretty: false}

	data := []byte(`{"hello":"world"}`)
	headers := []string{"COL1"}
	extractRows := func(d []byte) ([][]string, error) {
		return [][]string{{"row1"}}, nil
	}

	// With FormatJSON, extractRows should be ignored and raw JSON printed
	if err := p.PrintResult(data, headers, extractRows); err != nil {
		t.Fatalf("PrintResult returned error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != `{"hello":"world"}` {
		t.Errorf("expected raw JSON, got: %q", got)
	}
}

func TestPrintResult_FormatTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Writer: &buf, Pretty: false}

	data := []byte(`{}`)
	headers := []string{"NAME", "VALUE"}
	extractRows := func(d []byte) ([][]string, error) {
		return [][]string{
			{"alice", "100"},
			{"bob", "200"},
		}, nil
	}

	if err := p.PrintResult(data, headers, extractRows); err != nil {
		t.Fatalf("PrintResult returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d: %q", len(lines), got)
	}

	// Header line should contain both column names
	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "VALUE") {
		t.Errorf("header missing columns: %q", lines[0])
	}
	if !strings.Contains(lines[1], "alice") || !strings.Contains(lines[1], "100") {
		t.Errorf("row 1 missing data: %q", lines[1])
	}
	if !strings.Contains(lines[2], "bob") || !strings.Contains(lines[2], "200") {
		t.Errorf("row 2 missing data: %q", lines[2])
	}
}

func TestPrintResult_FormatCSV(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatCSV, Writer: &buf, Pretty: false}

	data := []byte(`{}`)
	headers := []string{"ID", "NAME"}
	extractRows := func(d []byte) ([][]string, error) {
		return [][]string{
			{"1", "Alice"},
			{"2", "Bob"},
		}, nil
	}

	if err := p.PrintResult(data, headers, extractRows); err != nil {
		t.Fatalf("PrintResult returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "ID,NAME" {
		t.Errorf("expected header 'ID,NAME', got %q", lines[0])
	}
	if lines[1] != "1,Alice" {
		t.Errorf("expected '1,Alice', got %q", lines[1])
	}
	if lines[2] != "2,Bob" {
		t.Errorf("expected '2,Bob', got %q", lines[2])
	}
}

func TestCSV_Escaping(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{"comma", "hello,world", `"hello,world"`},
		{"quote", `say "hi"`, `"say ""hi"""`},
		{"newline", "line1\nline2", "\"line1\nline2\""},
		{"plain", "simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := &Printer{Format: FormatCSV, Writer: &buf, Pretty: false}

			data := []byte(`{}`)
			headers := []string{"FIELD"}
			extractRows := func(d []byte) ([][]string, error) {
				return [][]string{{tt.field}}, nil
			}

			if err := p.PrintResult(data, headers, extractRows); err != nil {
				t.Fatalf("PrintResult returned error: %v", err)
			}

			// For newline test, the "lines" split will break it, so check full output
			got := buf.String()
			// Should start with header
			if !strings.HasPrefix(got, "FIELD\n") {
				t.Errorf("expected header line, got: %q", got)
			}
			// The data portion is everything after the header line
			dataLine := got[len("FIELD\n"):]
			dataLine = strings.TrimRight(dataLine, "\n")
			if dataLine != tt.expected {
				t.Errorf("field %q: expected %q, got %q", tt.field, tt.expected, dataLine)
			}
		})
	}
}

func TestTableCSV_NilExtractRows_FallsBackToJSON(t *testing.T) {
	for _, format := range []Format{FormatTable, FormatCSV} {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			p := &Printer{Format: format, Writer: &buf, Pretty: false}

			data := []byte(`{"fallback":true}`)
			if err := p.PrintResult(data, nil, nil); err != nil {
				t.Fatalf("PrintResult returned error: %v", err)
			}

			got := strings.TrimSpace(buf.String())
			if got != `{"fallback":true}` {
				t.Errorf("expected JSON fallback, got: %q", got)
			}
		})
	}
}

func TestEmptyRows_StillPrintHeaders(t *testing.T) {
	for _, format := range []Format{FormatTable, FormatCSV} {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			p := &Printer{Format: format, Writer: &buf, Pretty: false}

			data := []byte(`{}`)
			headers := []string{"COL_A", "COL_B"}
			extractRows := func(d []byte) ([][]string, error) {
				return [][]string{}, nil
			}

			if err := p.PrintResult(data, headers, extractRows); err != nil {
				t.Fatalf("PrintResult returned error: %v", err)
			}

			got := buf.String()
			if !strings.Contains(got, "COL_A") || !strings.Contains(got, "COL_B") {
				t.Errorf("expected headers in output, got: %q", got)
			}
		})
	}
}
