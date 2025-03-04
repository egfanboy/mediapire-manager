package utils

import "encoding/json"

// Converts one struct to another
func ConvertStruct[O any, T any](data O) (T, error) {
	var result T

	b, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(b, &result)

	return result, err
}
