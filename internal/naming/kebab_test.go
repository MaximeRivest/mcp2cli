package naming

import "testing"

func TestToKebabCase(t *testing.T) {
	tests := map[string]string{
		"searchFiles":     "search-files",
		"notion_get_self": "notion-get-self",
		"API_DOCS":        "api-docs",
		"maxResults":      "max-results",
	}

	for input, want := range tests {
		if got := ToKebabCase(input); got != want {
			t.Fatalf("ToKebabCase(%q) = %q, want %q", input, got, want)
		}
	}
}
