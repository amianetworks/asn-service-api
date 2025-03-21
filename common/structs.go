// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "encoding/json"

func MapToStruct[T any](data map[string]interface{}) (*T, error) {
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

func StructToMap[T any](t *T) map[string]interface{} {
	jsonStr, _ := json.Marshal(t)
	var data map[string]interface{}
	_ = json.Unmarshal(jsonStr, &data)
	return data
}
