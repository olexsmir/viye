package viye

import "strings"

func FormatOutput(s string) string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = "|"
		} else {
			lines[i] = "| " + line
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func FormatBulletlist(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = "- " + line
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
