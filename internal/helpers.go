package internal

import "fmt"

func getModuleName(config map[string]any) string {
	if v, ok := config["module"].(string); ok && v != "" {
		return v
	}
	return "teams"
}

func resolveValue(key string, current, config map[string]any) string {
	if v, ok := current[key].(string); ok && v != "" {
		return v
	}
	if v, ok := config[key].(string); ok && v != "" {
		return v
	}
	return ""
}

func resolveBool(key string, current, config map[string]any) bool {
	for _, m := range []map[string]any{current, config} {
		switch v := m[key].(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1" || v == "yes"
		}
	}
	return false
}

func resolveStringSlice(key string, current, config map[string]any) []string {
	for _, m := range []map[string]any{current, config} {
		switch v := m[key].(type) {
		case []string:
			return v
		case []any:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		var n int64
		fmt.Sscanf(t, "%d", &n)
		return n
	}
	return 0
}

func resolveInt(key string, current, config map[string]any) int {
	if v := toInt64(current[key]); v != 0 {
		return int(v)
	}
	return int(toInt64(config[key]))
}
