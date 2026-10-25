package command

import (
	"fmt"
	"strconv"
)

func getDefault[T any](values []T, fallback T) T {
	if len(values) > 0 {
		return values[0]
	}
	return fallback
}

func format(value any) string {
	if f, ok := value.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprint(value)
}
