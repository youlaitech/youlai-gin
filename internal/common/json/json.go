package json

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// BigIntJSONEncoder 自动将 int64 转换为字符串的 JSON 编码器
type BigIntJSONEncoder struct{}

func (e BigIntJSONEncoder) Marshal(v interface{}) ([]byte, error) {
	return marshalWithBigIntConversion(v)
}

// marshalWithBigIntConversion 递归地将结构体中的 int64 字段转换为字符串
func marshalWithBigIntConversion(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	return marshalValue(val)
}

func marshalValue(val reflect.Value) ([]byte, error) {
	if !val.IsValid() {
		return []byte("null"), nil
	}

	// 处理指针
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		return marshalStruct(val)
	case reflect.Slice, reflect.Array:
		return marshalSlice(val)
	case reflect.Map:
		return marshalMap(val)
	case reflect.Int64:
		// 将 int64 转换为字符串
		return []byte(strconv.FormatInt(val.Int(), 10)), nil
	default:
		// 对于其他类型，使用标准 JSON 编码
		return json.Marshal(val.Interface())
	}
}

func marshalStruct(val reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	typ := val.Type()
	fieldCount := 0

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取 JSON tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// 解析 JSON tag
		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// 添加逗号分隔符
		if fieldCount > 0 {
			buf.WriteByte(',')
		}

		// 写入字段名
		buf.WriteByte('"')
		buf.WriteString(fieldName)
		buf.WriteByte('"')
		buf.WriteByte(':')

		// 递归处理字段值
		fieldBytes, err := marshalValue(field)
		if err != nil {
			return nil, err
		}
		buf.Write(fieldBytes)

		fieldCount++
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalSlice(val reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i := 0; i < val.Len(); i++ {
		if i > 0 {
			buf.WriteByte(',')
		}

		elemBytes, err := marshalValue(val.Index(i))
		if err != nil {
			return nil, err
		}
		buf.Write(elemBytes)
	}

	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func marshalMap(val reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	iter := val.MapRange()
	mapCount := 0

	for iter.Next() {
		if mapCount > 0 {
			buf.WriteByte(',')
		}

		// 键
		keyBytes, err := json.Marshal(iter.Key().Interface())
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// 值
		valueBytes, err := marshalValue(iter.Value())
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)

		mapCount++
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
