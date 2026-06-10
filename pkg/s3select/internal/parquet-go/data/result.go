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
	"fmt"
	"math"

	"github.com/obstor/obstor/pkg/s3select/internal/parquet-go/gen-go/parquet"
)

func resultToBool(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case bool:
		return v, nil
	}

	return nil, fmt.Errorf("result is not Bool but %T", result)
}

func resultToInt32(result interface{}) (value interface{}, err error) {
	if value, err = resultToInt64(result); err != nil {
		return nil, err
	}

	if value.(int64) < math.MinInt32 || value.(int64) > math.MaxInt32 {
		return nil, fmt.Errorf("int32 overflow")
	}

	return int32(value.(int64)), nil
}

func resultToInt64(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	}

	return nil, fmt.Errorf("result is not Number but %T", result)
}

func resultToFloat(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case float64:
		return float32(v), nil
	}

	return nil, fmt.Errorf("result is not float32 but %T", result)
}

func resultToDouble(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case float64:
		return v, nil
	}

	return nil, fmt.Errorf("result is not float64 but %T", result)
}

func resultToBytes(result interface{}) (interface{}, error) {
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("result is not byte array but %T", result)
	}

	data := []byte{}
	for i, r := range arr {
		f, ok := r.(float64)
		if !ok {
			return nil, fmt.Errorf("result[%v] is not byte but %T", i, r)
		}

		value := uint64(f)
		if value > math.MaxUint8 {
			return nil, fmt.Errorf("byte overflow in result[%v]", i)
		}

		data = append(data, byte(value))
	}

	return data, nil
}

func resultToString(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case string:
		return v, nil
	}

	return nil, fmt.Errorf("result is not String but %T", result)
}

func resultToUint8(result interface{}) (value interface{}, err error) {
	if value, err = resultToUint64(result); err != nil {
		return nil, err
	}

	if value.(uint64) > math.MaxUint8 {
		return nil, fmt.Errorf("uint8 overflow")
	}

	return uint8(value.(uint64)), nil
}

func resultToUint16(result interface{}) (value interface{}, err error) {
	if value, err = resultToUint64(result); err != nil {
		return nil, err
	}

	if value.(uint64) > math.MaxUint16 {
		return nil, fmt.Errorf("uint16 overflow")
	}

	return uint16(value.(uint64)), nil
}

func resultToUint32(result interface{}) (value interface{}, err error) {
	if value, err = resultToUint64(result); err != nil {
		return nil, err
	}

	if value.(uint64) > math.MaxUint32 {
		return nil, fmt.Errorf("uint32 overflow")
	}

	return uint32(value.(uint64)), nil
}

func resultToUint64(result interface{}) (value interface{}, err error) {
	switch v := result.(type) {
	case float64:
		return uint64(v), nil
	}

	return nil, fmt.Errorf("result is not Number but %T", result)
}

func resultToInt8(result interface{}) (value interface{}, err error) {
	if value, err = resultToInt64(result); err != nil {
		return nil, err
	}

	if value.(int64) < math.MinInt8 || value.(int64) > math.MaxInt8 {
		return nil, fmt.Errorf("int8 overflow")
	}

	return int8(value.(int64)), nil
}

func resultToInt16(result interface{}) (value interface{}, err error) {
	if value, err = resultToInt64(result); err != nil {
		return nil, err
	}

	if value.(int64) < math.MinInt16 || value.(int64) > math.MaxInt16 {
		return nil, fmt.Errorf("int16 overflow")
	}

	return int16(value.(int64)), nil
}

func stringToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT96, parquet.Type_BYTE_ARRAY, parquet.Type_FIXED_LEN_BYTE_ARRAY:
		return []byte(value.(string)), nil
	}

	return nil, fmt.Errorf("string cannot be converted to parquet type %v", parquetType)
}

func uint8ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(uint8)), nil
	case parquet.Type_INT64:
		return int64(value.(uint8)), nil
	}

	return nil, fmt.Errorf("uint8 cannot be converted to parquet type %v", parquetType)
}

func uint16ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(uint16)), nil
	case parquet.Type_INT64:
		return int64(value.(uint16)), nil
	}

	return nil, fmt.Errorf("uint16 cannot be converted to parquet type %v", parquetType)
}

func uint32ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(uint32)), nil
	case parquet.Type_INT64:
		return int64(value.(uint32)), nil
	}

	return nil, fmt.Errorf("uint32 cannot be converted to parquet type %v", parquetType)
}

func uint64ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(uint64)), nil
	case parquet.Type_INT64:
		return int64(value.(uint64)), nil
	}

	return nil, fmt.Errorf("uint64 cannot be converted to parquet type %v", parquetType)
}

func int8ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(int8)), nil
	case parquet.Type_INT64:
		return int64(value.(int8)), nil
	}

	return nil, fmt.Errorf("int8 cannot be converted to parquet type %v", parquetType)
}

func int16ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(int16)), nil
	case parquet.Type_INT64:
		return int64(value.(int16)), nil
	}

	return nil, fmt.Errorf("int16 cannot be converted to parquet type %v", parquetType)
}

func int32ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return value.(int32), nil
	case parquet.Type_INT64:
		return int64(value.(int32)), nil
	}

	return nil, fmt.Errorf("int32 cannot be converted to parquet type %v", parquetType)
}

func int64ToParquetValue(value interface{}, parquetType parquet.Type) (interface{}, error) {
	switch parquetType {
	case parquet.Type_INT32:
		return int32(value.(int64)), nil
	case parquet.Type_INT64:
		return value.(int64), nil
	}

	return nil, fmt.Errorf("int64 cannot be converted to parquet type %v", parquetType)
}

func resultToParquetValueByConvertedValue(result interface{}, convertedType parquet.ConvertedType, parquetType parquet.Type) (value interface{}, err error) {
	if result == nil {
		return nil, nil
	}

	switch convertedType {
	case parquet.ConvertedType_UTF8:
		if value, err = resultToString(result); err != nil {
			return nil, err
		}
		return stringToParquetValue(value, parquetType)
	case parquet.ConvertedType_UINT_8:
		if value, err = resultToUint8(result); err != nil {
			return nil, err
		}
		return uint8ToParquetValue(value, parquetType)
	case parquet.ConvertedType_UINT_16:
		if value, err = resultToUint16(result); err != nil {
			return nil, err
		}
		return uint16ToParquetValue(value, parquetType)
	case parquet.ConvertedType_UINT_32:
		if value, err = resultToUint32(result); err != nil {
			return nil, err
		}
		return uint32ToParquetValue(value, parquetType)
	case parquet.ConvertedType_UINT_64:
		if value, err = resultToUint64(result); err != nil {
			return nil, err
		}
		return uint64ToParquetValue(value, parquetType)
	case parquet.ConvertedType_INT_8:
		if value, err = resultToInt8(result); err != nil {
			return nil, err
		}
		return int8ToParquetValue(value, parquetType)
	case parquet.ConvertedType_INT_16:
		if value, err = resultToInt16(result); err != nil {
			return nil, err
		}
		return int16ToParquetValue(value, parquetType)
	case parquet.ConvertedType_INT_32:
		if value, err = resultToInt32(result); err != nil {
			return nil, err
		}
		return int32ToParquetValue(value, parquetType)
	case parquet.ConvertedType_INT_64:
		if value, err = resultToInt64(result); err != nil {
			return nil, err
		}
		return int64ToParquetValue(value, parquetType)
	}

	return nil, fmt.Errorf("unsupported converted type %v", convertedType)
}

func resultToParquetValue(result interface{}, parquetType parquet.Type, convertedType *parquet.ConvertedType) (interface{}, error) {
	if convertedType != nil {
		return resultToParquetValueByConvertedValue(result, *convertedType, parquetType)
	}

	if result == nil {
		return nil, nil
	}

	switch parquetType {
	case parquet.Type_BOOLEAN:
		return resultToBool(result)
	case parquet.Type_INT32:
		return resultToInt32(result)
	case parquet.Type_INT64:
		return resultToInt64(result)
	case parquet.Type_FLOAT:
		return resultToFloat(result)
	case parquet.Type_DOUBLE:
		return resultToDouble(result)
	case parquet.Type_INT96, parquet.Type_BYTE_ARRAY, parquet.Type_FIXED_LEN_BYTE_ARRAY:
		return resultToBytes(result)
	}

	return nil, fmt.Errorf("unknown parquet type %v", parquetType)
}

func resultToArray(result interface{}) ([]interface{}, error) {
	if result == nil {
		return nil, nil
	}

	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("result is not Array but %T", result)
	}

	return arr, nil
}
