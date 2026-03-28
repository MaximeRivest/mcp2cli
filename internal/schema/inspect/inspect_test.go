package inspect

import (
	"encoding/json"
	"testing"

	"github.com/maximerivest/mcp2cli/internal/mcp/types"
)

func TestInspectToolPreservesPropertyOrder(t *testing.T) {
	tool := types.Tool{
		Name: "getForecast",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"latitude": {"type": "number", "description": "Latitude"},
				"longitude": {"type": "number", "description": "Longitude"}
			},
			"required": ["latitude", "longitude"]
		}`),
	}

	spec, err := InspectTool(tool)
	if err != nil {
		t.Fatalf("InspectTool: %v", err)
	}
	if spec.CLIName != "get-forecast" {
		t.Fatalf("CLIName = %q, want %q", spec.CLIName, "get-forecast")
	}
	if len(spec.Arguments) != 2 {
		t.Fatalf("len(Arguments) = %d, want 2", len(spec.Arguments))
	}
	if spec.Arguments[0].CLIName != "latitude" || spec.Arguments[1].CLIName != "longitude" {
		t.Fatalf("argument order = %#v", spec.Arguments)
	}
}
