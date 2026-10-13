// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "encoding/json"

func MapToStruct[T any](data map[string]any) (*T, error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var t T
	err = json.Unmarshal(jsonStr, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func MapsToStructs[T any](data []map[string]any) ([]*T, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var toReturn []*T
	for _, d := range data {
		t, err := MapToStruct[T](d)
		if err != nil {
			return nil, err
		}

		toReturn = append(toReturn, t)
	}

	return toReturn, nil
}

func StructToMap[T any](data *T) (map[string]any, error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var toReturn map[string]any
	if err := json.Unmarshal(jsonStr, &toReturn); err != nil {
		return nil, err
	}

	return toReturn, nil
}

func StructsToMaps[T any](data []*T) ([]map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var toReturn []map[string]any
	for _, d := range data {
		m, err := StructToMap(d)
		if err != nil {
			return nil, err
		}

		toReturn = append(toReturn, m)
	}

	return toReturn, nil
}
