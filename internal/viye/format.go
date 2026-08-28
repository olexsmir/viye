package viye

import "strings"

func FormatOutput[T string | []string | []byte](s T) string {
	if d := decorate(asString(s), "| ", "|"); d != "" {
		return d + "\n"
	}
	return ""
}

func FormatBulletList[T string | []string | []byte](s T) string {
	return decorate(asString(s), "- ", "")
}

func asString[T string | []string | []byte](s T) string {
	switch v := any(s).(type) {
	case string:
		return strings.TrimSuffix(v, "\n")
	case []string:
		return strings.Join(v, "\n")
	case []byte:
		return strings.TrimSuffix(string(v), "\n")
	}
	panic("unreachable")
}

func decorate(s, prefix, blank string) string {
	ls := strings.Split(s, "\n")
	for i, line := range ls {
		if line == "" {
			ls[i] = blank
		} else {
			ls[i] = prefix + line
		}
	}
	return strings.Join(ls, "\n")
}
