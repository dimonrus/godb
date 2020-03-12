package godb

import "reflect"

// Model column in db
func ModelColumn(model IModel, field interface{}) string {
	if model == nil {
		return ""
	}
	cte := reflect.ValueOf(field)
	if cte.Kind() != reflect.Ptr {
		return ""
	}
	ve := reflect.ValueOf(model).Elem()
	te := reflect.TypeOf(model).Elem()
	for i := 0; i < ve.NumField(); i++ {
		if ve.Field(i).Addr().Pointer() == cte.Elem().Addr().Pointer() {
			return te.Field(i).Tag.Get("column")
		}
	}
	return ""
}
