// 工具类
package utils

import "encoding/json"

func Struct2Map(s interface{}) map[string]string {
	j, _ := json.Marshal(s)
	m := make(map[string]string)
	json.Unmarshal(j, &m)
	return m
}
