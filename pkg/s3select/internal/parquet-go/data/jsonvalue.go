/*
 * Minio Cloud Storage, (C) 2019 Minio, Inc.
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package data

import (
	"encoding/json"
	"fmt"

	"github.com/obstor/obstor/pkg/s3select/internal/parquet-go/gen-go/parquet"
)

type jsonValue struct {
	result interface{} // nil means null/absent; otherwise the unmarshaled JSON value
	path   *string
}

func (v *jsonValue) String() string {
	if v.result == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%v", v.result)
}

func (v *jsonValue) IsNull() bool {
	return v.result == nil
}

func (v *jsonValue) Get(path string) *jsonValue {
	if v.path != nil {
		var result interface{}
		if *v.path == path {
			result = v.result
		}

		return toJSONValue(result)
	}

	if v.result == nil {
		return toJSONValue(nil)
	}

	m, ok := v.result.(map[string]interface{})
	if !ok {
		return toJSONValue(nil)
	}

	val, exists := m[path]
	if !exists {
		return toJSONValue(nil)
	}

	return toJSONValue(val)
}

func (v *jsonValue) GetValue(parquetType parquet.Type, convertedType *parquet.ConvertedType) (interface{}, error) {
	if v.result == nil {
		return nil, nil
	}

	return resultToParquetValue(v.result, parquetType, convertedType)
}

func (v *jsonValue) GetArray() ([]interface{}, error) {
	if v.result == nil {
		return nil, nil
	}

	return resultToArray(v.result)
}

func (v *jsonValue) Range(iterator func(key string, value interface{}) bool) error {
	if v.result == nil {
		return nil
	}

	m, ok := v.result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("result is not Map but %T", v.result)
	}

	for k, val := range m {
		if !iterator(k, val) {
			break
		}
	}
	return nil
}

func toJSONValue(result interface{}) *jsonValue {
	return &jsonValue{
		result: result,
	}
}

func bytesToJSONValue(data []byte) (*jsonValue, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON data")
	}

	return toJSONValue(result), nil
}
