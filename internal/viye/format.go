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

func FormatBulletListString(s string) string {
	return FormatBulletList(strings.Split(s, "\n"))
}

func FormatBulletList(s []string) string {
	for i, line := range s {
		if line != "" {
			s[i] = "- " + line
		}
	}
	return strings.Join(s, "\n")
}
