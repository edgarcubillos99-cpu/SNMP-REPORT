package rules

import "strings"

// Evalúa reglas expresadas como cadena: >, <, >=, <=, ==, !=, between, outside
func IsRisk(value float64, rule string, limit float64, extra ...float64) bool {

	r := strings.TrimSpace(strings.ToLower(rule))

	switch r {

	case ">":
		return value > limit

	case "<":
		return value < limit

	case ">=":
		return value >= limit

	case "<=":
		return value <= limit

	case "==", "=":
		return value == limit

	case "!=":
		return value != limit

	// entre 2 números
	case "between":
		if len(extra) < 1 {
			return false
		}
		max := extra[0]
		return value >= limit && value <= max

	// fuera del rango
	case "outside":
		if len(extra) < 1 {
			return false
		}
		max := extra[0]
		return value < limit || value > max
	}

	return false
}
