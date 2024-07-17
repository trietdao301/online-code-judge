package utils

import "reflect"

func GetFieldName(s interface{}, fieldName string) string {
	t := reflect.TypeOf(s)
	field, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}
	return field.Tag.Get("bson")
}
