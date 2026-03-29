package cli

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI escape codes for styling.
const (
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiItalic  = "\033[3m"
	ansiReset   = "\033[0m"
	ansiCyan    = "\033[36m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiWhite   = "\033[97m"
)

// colorEnabled returns true if stdout supports ANSI colors.
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// style helpers — return input unchanged if color is disabled.

func bold(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiBold + s + ansiReset
}

func dim(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiDim + s + ansiReset
}

func cyan(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiCyan + s + ansiReset
}

func green(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiGreen + s + ansiReset
}

func yellow(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiYellow + s + ansiReset
}

func boldCyan(s string) string {
	if !colorEnabled() {
		return s
	}
	return ansiBold + ansiCyan + s + ansiReset
}

// termWidth returns the terminal width, defaulting to 80.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// wordWrap wraps text to fit within maxWidth, adding indent to continuation lines.
func wordWrap(text string, maxWidth int, indent string) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	contentWidth := maxWidth - len(indent)
	if contentWidth < 20 {
		contentWidth = 20
	}

	var result strings.Builder
	paragraphs := strings.Split(text, "\n")
	for i, para := range paragraphs {
		if i > 0 {
			result.WriteString("\n")
		}
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		// Don't wrap short lines or bullet points
		if len(para) <= contentWidth {
			result.WriteString(indent + para)
			continue
		}
		// Word wrap
		words := strings.Fields(para)
		lineLen := 0
		firstWord := true
		for _, word := range words {
			if firstWord {
				result.WriteString(indent + word)
				lineLen = len(indent) + len(word)
				firstWord = false
			} else if lineLen+1+len(word) > maxWidth {
				result.WriteString("\n" + indent + word)
				lineLen = len(indent) + len(word)
			} else {
				result.WriteString(" " + word)
				lineLen += 1 + len(word)
			}
		}
	}
	return result.String()
}
