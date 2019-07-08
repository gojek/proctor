package utility

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func MergeMaps(mapOne, mapTwo map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range mapOne {
		result[k] = v
	}
	for k, v := range mapTwo {
		result[k] = v
	}
	return result
}

func MapToString(someMap map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range someMap {
		_, _ = fmt.Fprintf(b, "%s=\"%s\",", key, value)
	}
	return strings.TrimRight(b.String(), ",")
}

func DeserializeMap(encodedMap string) (map[string]string, error) {
	var mapStringToString map[string]string
	if encodedMap == "" {
		return mapStringToString, nil
	}

	decodedMap, err := base64.StdEncoding.DecodeString(encodedMap)
	if err != nil {
		return mapStringToString, err
	}

	err = json.Unmarshal(decodedMap, &mapStringToString)
	return mapStringToString, err
}
