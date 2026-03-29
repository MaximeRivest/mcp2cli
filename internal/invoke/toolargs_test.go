package invoke

import (
	"testing"

	"github.com/maximerivest/mcptocli/internal/schema/inspect"
)

func TestParseToolArgumentsFlagsAndPositionals(t *testing.T) {
	spec := &inspect.ToolSpec{
		CLIName:            "get-forecast",
		SupportsCLIParsing: true,
		Arguments: []inspect.ArgSpec{
			{Name: "latitude", CLIName: "latitude", Type: "number", Required: true},
			{Name: "longitude", CLIName: "longitude", Type: "number", Required: true},
			{Name: "units", CLIName: "units", Type: "string", HasDefault: true, Default: "metric"},
		},
	}

	args, err := ParseToolArguments(spec, []string{"37.7", "--longitude", "-122.4"})
	if err != nil {
		t.Fatalf("ParseToolArguments: %v", err)
	}
	if args["latitude"] != 37.7 {
		t.Fatalf("latitude = %#v", args["latitude"])
	}
	if args["longitude"] != -122.4 {
		t.Fatalf("longitude = %#v", args["longitude"])
	}
	if args["units"] != "metric" {
		t.Fatalf("units = %#v", args["units"])
	}
}

func TestParseToolArgumentsRepeatedArrayFlag(t *testing.T) {
	spec := &inspect.ToolSpec{
		CLIName:            "search",
		SupportsCLIParsing: true,
		Arguments:          []inspect.ArgSpec{{Name: "tag", CLIName: "tag", Type: "array", ItemType: "string"}},
	}

	args, err := ParseToolArguments(spec, []string{"--tag", "cli", "--tag", "go"})
	if err != nil {
		t.Fatalf("ParseToolArguments: %v", err)
	}
	values, ok := args["tag"].([]any)
	if !ok || len(values) != 2 || values[0] != "cli" || values[1] != "go" {
		t.Fatalf("tag values = %#v", args["tag"])
	}
}
