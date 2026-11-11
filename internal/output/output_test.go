package output

import (
	"bytes"
	"testing"
)

func TestPrintJSONV2Output(t *testing.T) {
	var out bytes.Buffer
	value := struct {
		Items   []string `json:"items"`
		Label   string   `json:"label,omitempty"`
		Enabled bool     `json:"enabled"`
	}{}
	if err := (&Writer{Out: &out}).PrintJSON(value); err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"items\": [],\n  \"enabled\": false\n}\n"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}
