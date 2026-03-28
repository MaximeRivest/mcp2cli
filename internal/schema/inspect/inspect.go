package inspect

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/maximerivest/mcp2cli/internal/mcp/types"
	"github.com/maximerivest/mcp2cli/internal/naming"
)

// ToolSpec is a schema-derived view of a tool suitable for CLI rendering and parsing.
type ToolSpec struct {
	ToolName           string
	CLIName            string
	Description        string
	Arguments          []ArgSpec
	SupportsCLIParsing bool
	RawSchema          json.RawMessage
}

// ArgSpec describes one tool argument derived from JSON Schema.
type ArgSpec struct {
	Name        string
	CLIName     string
	Type        string
	ItemType    string
	Description string
	Required    bool
	HasDefault  bool
	Default     any
}

// InspectTool derives a CLI-oriented tool spec from the MCP tool definition.
func InspectTool(tool types.Tool) (*ToolSpec, error) {
	spec := &ToolSpec{
		ToolName:           tool.Name,
		CLIName:            naming.ToKebabCase(tool.Name),
		Description:        tool.Description,
		SupportsCLIParsing: true,
		RawSchema:          tool.InputSchema,
	}
	if len(tool.InputSchema) == 0 {
		return spec, nil
	}

	rootFields, err := parseOrderedObject(tool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("parse input schema: %w", err)
	}

	rootType := decodeStringField(rootFields, "type")
	if rootType != "" && rootType != "object" {
		spec.SupportsCLIParsing = false
		return spec, nil
	}

	required := decodeStringSliceField(rootFields, "required")
	requiredSet := map[string]struct{}{}
	for _, name := range required {
		requiredSet[name] = struct{}{}
	}

	propertiesRaw := findRawField(rootFields, "properties")
	if len(propertiesRaw) == 0 {
		return spec, nil
	}
	propertyFields, err := parseOrderedObject(propertiesRaw)
	if err != nil {
		spec.SupportsCLIParsing = false
		return spec, nil
	}

	arguments := make([]ArgSpec, 0, len(propertyFields))
	for _, field := range propertyFields {
		propFields, err := parseOrderedObject(field.Value)
		if err != nil {
			spec.SupportsCLIParsing = false
			continue
		}
		arg := ArgSpec{
			Name:        field.Name,
			CLIName:     naming.ToKebabCase(field.Name),
			Type:        decodeStringField(propFields, "type"),
			Description: decodeStringField(propFields, "description"),
		}
		_, arg.Required = requiredSet[field.Name]

		if defaultRaw := findRawField(propFields, "default"); len(defaultRaw) > 0 {
			arg.HasDefault = true
			_ = json.Unmarshal(defaultRaw, &arg.Default)
		}

		if arg.Type == "array" {
			itemsRaw := findRawField(propFields, "items")
			if len(itemsRaw) > 0 {
				itemsFields, err := parseOrderedObject(itemsRaw)
				if err == nil {
					arg.ItemType = decodeStringField(itemsFields, "type")
				}
			}
			if arg.ItemType == "" {
				spec.SupportsCLIParsing = false
			}
		}

		if arg.Type == "" {
			arg.Type = "object"
		}
		if !isSupportedArgType(arg) {
			spec.SupportsCLIParsing = false
		}
		arguments = append(arguments, arg)
	}

	for i := range arguments {
		for j := i + 1; j < len(arguments); j++ {
			if arguments[i].CLIName == arguments[j].CLIName {
				spec.SupportsCLIParsing = false
			}
		}
	}

	spec.Arguments = arguments
	return spec, nil
}

// FindArgument returns the argument matching a CLI flag name.
func (s *ToolSpec) FindArgument(cliName string) (ArgSpec, bool) {
	for _, arg := range s.Arguments {
		if arg.CLIName == cliName || arg.Name == cliName {
			return arg, true
		}
	}
	return ArgSpec{}, false
}

// PositionalArguments returns required scalar arguments eligible for positional parsing.
func (s *ToolSpec) PositionalArguments() []ArgSpec {
	args := make([]ArgSpec, 0, len(s.Arguments))
	for _, arg := range s.Arguments {
		if arg.Required && isPositionalEligible(arg) {
			args = append(args, arg)
		}
	}
	return args
}

// UsageParts returns usage-friendly flag specs in schema order.
func (s *ToolSpec) UsageParts() []string {
	parts := make([]string, 0, len(s.Arguments))
	for _, arg := range s.Arguments {
		placeholder := Placeholder(arg)
		part := fmt.Sprintf("--%s <%s>", arg.CLIName, placeholder)
		if !arg.Required {
			part = "[" + part + "]"
		}
		parts = append(parts, part)
	}
	return parts
}

// Placeholder returns a user-friendly placeholder string for the argument type.
func Placeholder(arg ArgSpec) string {
	switch arg.Type {
	case "string":
		return "string"
	case "number":
		return "float"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "array":
		if arg.ItemType != "" {
			return arg.ItemType
		}
		return "json"
	case "object":
		return "json"
	default:
		return "value"
	}
}

func isPositionalEligible(arg ArgSpec) bool {
	switch arg.Type {
	case "string", "number", "integer":
		return true
	default:
		return false
	}
}

func isSupportedArgType(arg ArgSpec) bool {
	switch arg.Type {
	case "string", "number", "integer", "boolean", "object":
		return true
	case "array":
		return arg.ItemType == "string" || arg.ItemType == "number" || arg.ItemType == "integer" || arg.ItemType == "boolean"
	default:
		return false
	}
}

type orderedField struct {
	Name  string
	Value json.RawMessage
}

func parseOrderedObject(raw json.RawMessage) ([]orderedField, error) {
	dec := json.NewDecoder(bytesReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("expected JSON object")
	}

	fields := []orderedField{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected string object key")
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		fields = append(fields, orderedField{Name: key, Value: value})
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return fields, nil
}

func decodeStringField(fields []orderedField, name string) string {
	raw := findRawField(fields, name)
	if len(raw) == 0 {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return value
}

func decodeStringSliceField(fields []orderedField, name string) []string {
	raw := findRawField(fields, name)
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return values
}

func findRawField(fields []orderedField, name string) json.RawMessage {
	for _, field := range fields {
		if field.Name == name {
			return field.Value
		}
	}
	return nil
}

func bytesReader(raw []byte) *bytes.Reader { return bytes.NewReader(raw) }
