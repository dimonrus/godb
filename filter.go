package godb

// SQL Filter

import (
	"fmt"
	"strings"
)

// Order
type sqlOrder struct {
	Field     string
	Direction string
}

// SQL Expression
type sqlExpression struct {
	Expression string
}

// SQL Pagination limit offset
type sqlPagination struct {
	Limit  int
	Offset int
}

// Filter struct
type SqlFilter struct {
	expression []sqlExpression
	orders     []sqlOrder
	pagination sqlPagination
	group      []string
	arguments  []interface{}
}

//Add or filter
func (f *SqlFilter) AddOrFilters(filter ...*SqlFilter) *SqlFilter {
	args := make([]interface{}, 0)
	conditions := make([]string, 0, len(filter))
	for i := range filter {
		if len(filter[i].GetArguments()) > 0 {
			args = append(args, filter[i].GetArguments()...)
		}
		conditions = append(conditions, filter[i].String())
	}
	return f.AddExpression("("+strings.Join(conditions, " OR ")+")", args)
}

// Add fileld to filter
func (f *SqlFilter) AddFiledFilter(field string, condition string, value interface{}) *SqlFilter {
	expression := field + " " + condition + " ?"
	f.expression = append(f.expression, sqlExpression{
		Expression: expression,
	})
	if value != nil {
		f.arguments = append(f.arguments, value)
	}
	return f
}

// Add in Filter
func (f *SqlFilter) AddInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	f.expression = append(f.expression, sqlExpression{
		Expression: fmt.Sprintf("%s IN (%s)", field, strings.Join(condition, ",")),
	})
	f.arguments = append(f.arguments, values...)
	return f
}

// Add not in filter
func (f *SqlFilter) AddNotInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	f.expression = append(f.expression, sqlExpression{
		Expression: fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(condition, ",")),
	})
	f.arguments = append(f.arguments, values...)
	return f
}

// Add filter expression
func (f *SqlFilter) AddExpression(expression string, values []interface{}) *SqlFilter {
	f.expression = append(f.expression, sqlExpression{
		Expression: expression,
	})
	if values != nil && len(values) > 0 {
		f.arguments = append(f.arguments, values...)
	}
	return f
}

// Add Order
func (f *SqlFilter) AddOrder(field string, direction string) *SqlFilter {
	f.orders = append(f.orders, sqlOrder{Field: field, Direction: direction})
	return f
}

// Reset Order
func (f *SqlFilter) ResetOrder() *SqlFilter {
	f.orders = []sqlOrder{}
	return f
}

// Add Group
func (f *SqlFilter) GroupBy(fields ...string) *SqlFilter {
	f.group = append(f.group, fields...)
	return f
}

// Reset Group
func (f *SqlFilter) ResetGroupBy() *SqlFilter {
	f.group = []string{}
	return f
}

// Set pagination
func (f *SqlFilter) SetPagination(limit int, offset int) *SqlFilter {
	f.pagination = sqlPagination{Limit: limit, Offset: offset}
	return f
}

// Get arguments
func (f *SqlFilter) GetArguments() []interface{} {
	return f.arguments
}

// Make SQL query
func (f SqlFilter) String() string {
	var expressionFilters []string
	var orders []string
	var result = make([]string, 0, 4)

	// Prepare conditions
	for _, value := range f.expression {
		expressionFilters = append(expressionFilters, value.Expression)
	}
	if len(expressionFilters) > 0 {
		result = append(result, strings.Join(expressionFilters, " AND "))
	}

	// Prepare groups
	if len(f.group) > 0 {
		result = append(result, "GROUP BY "+strings.Join(f.group, ", "))
	}

	// Prepare orders
	for _, value := range f.orders {
		orders = append(orders, value.Field+" "+value.Direction)
	}
	if len(orders) > 0 {
		result = append(result, "ORDER BY "+strings.Join(orders, ", "))
	}

	// Prepare pagination
	if f.pagination.Limit > 0 {
		result = append(result, fmt.Sprintf("LIMIT %v OFFSET %v", f.pagination.Limit, f.pagination.Offset))
	}

	return strings.Join(result, " ")
}

// Get query with WHERE
func (f SqlFilter) GetWithWhere() string {
	if len(f.expression) > 0 {
		return "WHERE " + f.String()
	}

	return f.String()
}

// New SQL Filter with pagination
func NewSqlFilter() *SqlFilter {
	return &SqlFilter{}
}
