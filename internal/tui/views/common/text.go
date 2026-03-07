package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func WrapWords(s string, maxWidth int) string {
	if maxWidth <= 0 || lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Keep bullet-separated help chunks intact (e.g. "t: target") when wrapping.
	if strings.Contains(s, " • ") {
		return wrapBulletChunks(s, maxWidth)
	}

	return wrapByWords(s, maxWidth)
}

func wrapBulletChunks(s string, maxWidth int) string {
	chunks := strings.Split(s, " • ")
	if len(chunks) == 0 {
		return s
	}

	lines := []string{chunks[0]}
	for _, chunk := range chunks[1:] {
		last := lines[len(lines)-1]
		candidate := last + " • " + chunk
		if lipgloss.Width(candidate) <= maxWidth {
			lines[len(lines)-1] = candidate
			continue
		}
		lines = append(lines, chunk)
	}

	for i, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			lines[i] = wrapByWords(line, maxWidth)
		}
	}

	return strings.Join(lines, "\n")
}

func wrapByWords(s string, maxWidth int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	lines := []string{words[0]}
	for _, w := range words[1:] {
		last := lines[len(lines)-1]
		if lipgloss.Width(last)+1+lipgloss.Width(w) <= maxWidth {
			lines[len(lines)-1] = last + " " + w
			continue
		}
		lines = append(lines, w)
	}
	return strings.Join(lines, "\n")
}
