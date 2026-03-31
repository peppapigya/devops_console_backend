package strategies

import "fmt"

func buildSelector(selector interface{}) map[string]interface{} {
	if selector == nil {
		return map[string]interface{}{}
	}

	selectorMap, ok := selector.(map[string]interface{})
	fmt.Printf("selectorMap:%v\n", selectorMap)
	if !ok {
		return map[string]interface{}{}
	}

	result := make(map[string]interface{})

	if namespaces, ok := selectorMap["namespaces"].([]interface{}); ok {
		result["namespaces"] = namespaces
	}

	if labelSelectors, ok := selectorMap["labelSelectors"].(map[string]interface{}); ok {
		result["labelSelectors"] = labelSelectors
	}

	if annotationSelectors, ok := selectorMap["annotationSelectors"].(map[string]interface{}); ok {
		result["annotationSelectors"] = annotationSelectors
	}

	if fieldSelectors, ok := selectorMap["fieldSelectors"].(map[string]interface{}); ok {
		result["fieldSelectors"] = fieldSelectors
	}

	if podSelector, ok := selectorMap["podSelector"].([]interface{}); ok {
		result["podSelector"] = podSelector
	}

	if pods, ok := selectorMap["pods"].(map[string]interface{}); ok {
		result["pods"] = pods
	}

	return result
}
