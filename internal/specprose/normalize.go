// Package specprose normalizes upstream documentation before generation.
package specprose

import "strings"

// Normalize replaces dash punctuation in spec prose. Schema values, examples,
// defaults and enums retain their exact wire values.
func Normalize(value any) {
	switch typed := value.(type) {
	case map[string]any:
		normalizeObject(typed)
	case []any:
		for _, item := range typed {
			Normalize(item)
		}
	}
}

func normalizeObject(value map[string]any) {
	for key, item := range value {
		switch key {
		case "const", "enum", "default", "example", "examples":
			continue
		case "description", "summary", "title":
			prose, ok := item.(string)
			if ok {
				value[key] = strings.NewReplacer(
					" \u2014 ", ". ",
					" \u2013 ", ". ",
					"\u2014", ", ",
					"\u2013", ", ",
					"notify us", "notify BTSE",
					"we can", "BTSE can",
				).Replace(prose)
			}
		}

		Normalize(item)
	}
}
