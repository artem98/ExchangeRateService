package utils

import (
	"fmt"
	"strings"
)

func ParseCurrencyPair(input string) (string, string, error) {
	if len(input) != 7 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}
	upper := strings.ToUpper(input)
	parts := strings.Split(upper, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}

	if len(parts[0]) != 3 || len(parts[1]) != 3 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}

	return parts[0], parts[1], nil
}
