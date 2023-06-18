package utils

import "fmt"

// MergeMapsShallow merges any number of string maps into a single string map
// this cascades so later maps have precedence over earlier maps
// https://stackoverflow.com/a/39406305
func MergeMapsShallow(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MergeMaps deeply merges maps recursively overwriting duplicates
// https://stackoverflow.com/a/62954592
func MergeMapsInterface(a, b map[interface{}]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		// If you use map[string]interface{}, ok is always false here.
		// Because yaml.Unmarshal will give you map[interface{}]interface{}.
		if v, ok := v.(map[interface{}]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[interface{}]interface{}); ok {
					out[k] = MergeMapsInterface(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func MergeDicts(map1, map2 map[string]interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})

	// Copy values from map1 to mergedMap
	for key, value := range map1 {
		mergedMap[key] = value
	}

	// Copy values from map2 to mergedMap
	for key, value := range map2 {
		// If the key already exists in mergedMap and both values are maps, merge them recursively
		if existingValue, ok := mergedMap[key]; ok {
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if valueMap, ok := value.(map[string]interface{}); ok {
					mergedMap[key] = MergeDicts(existingMap, valueMap)
					continue
				}
			}
		}
		mergedMap[key] = value
	}

	return mergedMap
}

func ConvertMap(originalMap map[interface{}]interface{}) map[string]interface{} {
	convertedMap := make(map[string]interface{})

	for key, value := range originalMap {
		// Convert the key to a string
		stringKey, ok := key.(string)
		if !ok {
			// Handle the case where the key is not a string
			fmt.Printf("Skipping key %v: not a string\n", key)
			continue
		}

		// Convert the value to the expected type
		switch v := value.(type) {
		case map[interface{}]interface{}:
			// If the value is a nested map, recursively convert it
			convertedMap[stringKey] = ConvertMap(v)
		default:
			// Otherwise, directly assign the value to the converted map
			convertedMap[stringKey] = v
		}
	}

	return convertedMap
}

// func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
// 	result := make(map[string]interface{})
// 	for _, m := range maps {
// 		MergeMapsInterface(result, m)
// 	}
// 	return result
// }
