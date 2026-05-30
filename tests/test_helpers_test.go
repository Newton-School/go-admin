package admin_test

import "strings"

func stringsContainsFold(value, term string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(term))
}

func boolInValues(value bool, values []string) bool {
	for _, candidate := range values {
		normalized := strings.ToLower(strings.TrimSpace(candidate))
		if value && (normalized == "true" || normalized == "1" || normalized == "on") {
			return true
		}
		if !value && (normalized == "false" || normalized == "0" || normalized == "off") {
			return true
		}
	}
	return false
}
