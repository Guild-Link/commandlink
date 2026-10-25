package command

import (
	"strconv"
	"strings"
)

func ValueOf(v any) Value {
	return Value{v}
}

func (v Values) Get(name string) Value {
	return v[name]
}

func (v Value) Strings() []string {
	switch value := v.value.(type) {
	case nil:
		return nil
	case []string:
		values := make([]string, len(value))
		for i, item := range value {
			values[i] = strings.TrimSpace(item)
		}
		return values
	}
	return []string{strings.TrimSpace(format(v.value))}
}

func (v Value) String(def ...string) string {
	if values := v.Strings(); len(values) > 0 && values[0] != "" {
		return values[0]
	}
	return getDefault(def, "")
}

func (v Value) Int(def ...int) int {
	return int(v.Float(float64(getDefault(def, 0))))
}

func (v Value) Float(def ...float64) float64 {
	if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
		return f
	}
	return getDefault(def, 0)
}
