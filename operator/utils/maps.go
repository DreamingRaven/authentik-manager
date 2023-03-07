package utils

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

// func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
// 	result := make(map[string]interface{})
// 	for _, m := range maps {
// 		MergeMapsInterface(result, m)
// 	}
// 	return result
// }
