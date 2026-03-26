package executor

import (
	"strconv"
)

// getString 从map中获取字符串，如果不存在或类型不匹配则返回默认值
func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

// getUint64 从map中获取uint64，如果不存在或类型不匹配则返回默认值
func getUint64(m map[string]interface{}, key string, defaultVal uint64) uint64 {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case uint64:
		return val
	case float64:
		return uint64(val)
	case int:
		return uint64(val)
	case int64:
		return uint64(val)
	case string:
		u, _ := strconv.ParseUint(val, 10, 64)
		return u
	default:
		return defaultVal
	}
}

// getMap 从map中获取子map，如果不存在或类型不匹配则返回空map
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}

// truncateString 截断字符串并添加省略号
func truncateString(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
