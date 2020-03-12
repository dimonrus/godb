package godb

import (
	"github.com/dimonrus/porterr"
	"reflect"
	"strings"
)

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

// Model update query
func ModelUpdateQuery(model IModel, condition *Condition, fields ...interface{}) (sql string, params []interface{}, e porterr.IError) {
	if model == nil {
		e = porterr.New(porterr.PortErrorArgument, "Model is nil, check your logic")
		return
	}
	if fields == nil {
		e = porterr.New(porterr.PortErrorArgument, "Fields is empty, nothing for update")
		return
	}
	params = make([]interface{}, 0)
	var columns []string
	ve := reflect.ValueOf(model).Elem()
	te := reflect.TypeOf(model).Elem()
	for i := 0; i < ve.NumField(); i++ {
		for _, v := range fields {
			cte := reflect.ValueOf(v)
			if cte.Kind() != reflect.Ptr {
				e = porterr.New(porterr.PortErrorArgument, "Fields must be an interfaces")
				return
			}
			if ve.Field(i).Addr().Pointer() == cte.Elem().Addr().Pointer() {
				columns = append(columns, te.Field(i).Tag.Get("column")+" = ?")
				params = append(params, v)
			}
		}
	}
	if len(columns) > 0 {
		sql = "UPDATE " + model.Table() + " SET " + strings.Join(columns, ",")
		if condition != nil && !condition.IsEmpty() {
			sql += " WHERE " + condition.String()
			params = append(params, condition.GetArguments()...)
		}
	} else {
		e = porterr.New(porterr.PortErrorArgument, "No columns found in model")
	}
	return sql, params, e
}
