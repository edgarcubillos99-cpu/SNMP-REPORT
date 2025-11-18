package rules

// Evaluar si un valor cumple con una regla dada
func IsRisk(value float64, rule string, limit float64) bool {
	// Evaluar según la regla
	switch rule {
	case ">":
		return value > limit
	case "<":
		return value < limit
	case "==":
		return value == limit
	}
	return false
}
