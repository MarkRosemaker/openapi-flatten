package jsonutil

import (
	"encoding/json/v2"
	"os"
)

// ReadFile reads a json file and unmarshals it.
func ReadFile[T any](name string, opts ...json.Options) (T, error) {
	var v T

	f, err := os.Open(name)
	if err != nil {
		return v, err
	}
	defer f.Close()

	if err := json.UnmarshalRead(f, &v, opts...); err != nil {
		return v, err
	}

	return v, nil
}
