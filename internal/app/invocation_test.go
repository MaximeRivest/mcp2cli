package app

import "testing"

func TestDetectInvocation(t *testing.T) {
	tests := []struct {
		argv0   string
		exposed string
	}{
		{argv0: "/usr/local/bin/mcptocli", exposed: ""},
		{argv0: "/home/maxime/.local/share/mcptocli/bin/mcp-weather", exposed: "mcp-weather"},
		{argv0: "/home/maxime/.local/share/mcptocli/bin/wea", exposed: "wea"},
	}

	for _, tt := range tests {
		invocation := DetectInvocation(tt.argv0)
		if invocation.ExposedCommandName != tt.exposed {
			t.Fatalf("DetectInvocation(%q) exposed = %q, want %q", tt.argv0, invocation.ExposedCommandName, tt.exposed)
		}
	}
}

func TestRewriteArgsForExposedMode(t *testing.T) {
	invocation := Invocation{ProgramName: "wea", ExposedCommandName: "wea"}

	rewritten := RewriteArgsForExposedMode(invocation, []string{"get-forecast", "--latitude", "1", "--longitude", "2"})
	want := []string{"tool", "get-forecast", "--latitude", "1", "--longitude", "2"}
	if len(rewritten) != len(want) {
		t.Fatalf("rewritten length = %d, want %d", len(rewritten), len(want))
	}
	for i := range want {
		if rewritten[i] != want[i] {
			t.Fatalf("rewritten[%d] = %q, want %q", i, rewritten[i], want[i])
		}
	}

	notRewritten := RewriteArgsForExposedMode(invocation, []string{"tools"})
	if len(notRewritten) != 1 || notRewritten[0] != "tools" {
		t.Fatalf("reserved command should not be rewritten: %#v", notRewritten)
	}
}
