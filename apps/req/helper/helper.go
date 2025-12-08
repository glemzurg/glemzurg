package helper

import "encoding/json"

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func JsonPretty(value interface{}) (pretty string) {
	return string(Must(json.MarshalIndent(value, "", "   ")))
}
