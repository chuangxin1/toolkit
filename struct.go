package toolkit

import (
	"net/url"
	"reflect"
	"strconv"
)

// URLValuesStruct convert struct to url.Values
func URLValuesStruct(obj interface{}) url.Values {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	values := url.Values{}
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("form")
		value := format(v.Field(i), v.Field(i).Interface())

		values.Add(key, value)
	}
	return values
}

func format(v reflect.Value, data interface{}) string {
	var s string
	switch v.Kind() {
	case reflect.String:
		s = data.(string)
	case reflect.Int:
		s = strconv.FormatInt(int64(data.(int)), 10)
	case reflect.Uint:
		s = strconv.FormatUint(data.(uint64), 10)
	case reflect.Bool:
		s = strconv.FormatBool(data.(bool))
	case reflect.Float32:
	case reflect.Float64:
		s = strconv.FormatFloat(data.(float64), 'f', -1, 32)
	default:
		s = "" // fmt.Sprintf("unsupported kind %s", v.Type())
	}
	return s
}
