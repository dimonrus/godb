package godb

// SQL Filter

import (
	"fmt"
	"strings"
)

// Order
type sqlOrderFilter struct {
	Field     string
	Direction string
}

// SQL Expression
type sqlExpression struct {
	Expression string
}

// SQL Pagination limit offset
type sqlPaginationFilter struct {
	Limit  int
	Offset int
}

// Filter struct
type SqlFilter struct {
	expression []sqlExpression
	orders     []sqlOrderFilter
	pagination sqlPaginationFilter
	group      []string
	arguments  []interface{}
}

//Add or filter
func (f *SqlFilter) AddOrFilters(filter ...*SqlFilter) *SqlFilter {
	args := make([]interface{}, 0)
	conditions := make([]string, 0, len(filter))
	for i := range filter {
		if filter[i].GetArguments() != nil && len(filter[i].GetArguments()) > 0 {
			args = append(args, filter[i].GetArguments())
		}
		conditions = append(conditions, filter[i].String())
	}
	return f.AddExpression("("+strings.Join(conditions, " OR ")+")", args)
}

// Add fileld to filter
func (f *SqlFilter) AddFiledFilter(field string, condition string, value interface{}) *SqlFilter {
	expression := field + " " + condition + " ?"
	expr := sqlExpression{
		Expression: expression,
	}
	f.expression = append(f.expression, expr)
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
	expression := sqlExpression{
		Expression: fmt.Sprintf("%s IN (%s)", field, strings.Join(condition, ",")),
	}
	f.expression = append(f.expression, expression)
	f.arguments = append(f.arguments, values...)
	return f
}

// Add not in filter
func (f *SqlFilter) AddNotInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	expression := sqlExpression{
		Expression: fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(condition, ",")),
	}
	f.expression = append(f.expression, expression)
	f.arguments = append(f.arguments, values...)
	return f
}

// Add filter expression
func (f *SqlFilter) AddExpression(expression string, values []interface{}) *SqlFilter {
	expr := sqlExpression{
		Expression: expression,
	}
	f.expression = append(f.expression, expr)
	if values != nil && len(values) > 0 {
		f.arguments = append(f.arguments, values...)
	}

	return f
}

// Add Order
func (f *SqlFilter) AddOrder(field string, direction string) *SqlFilter {
	f.orders = append(f.orders, sqlOrderFilter{Field: field, Direction: direction})
	return f
}

// Reset Order
func (f *SqlFilter) ResetOrder() *SqlFilter {
	f.orders = []sqlOrderFilter{}
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
	f.pagination = sqlPaginationFilter{Limit: limit, Offset: offset}
	return f
}

// Get arguments
func (f *SqlFilter) GetArguments() []interface{} {
	return f.arguments
}

// Make SQL query
func (f SqlFilter) String() string {
	var conditionFilters []string
	var expressionFilters []string
	var orders []string
	var pagination string
	var group string

	for _, value := range f.orders {
		orders = append(orders, value.Field+" "+value.Direction)
	}

	if len(f.group) > 0 {
		group = "GROUP BY " + strings.Join(f.group, ", ")
	}

	if len(orders) > 0 {
		pagination = "ORDER BY " + strings.Join(orders, ", ") + " "
	}

	if f.pagination.Limit > 0 {
		pagination = fmt.Sprintf("%sLIMIT %v OFFSET %v", pagination, f.pagination.Limit, f.pagination.Offset)
	}

	for _, value := range f.expression {
		expressionFilters = append(expressionFilters, value.Expression)
	}

	if len(expressionFilters) > 0 {
		conditionFilters = append(conditionFilters, expressionFilters...)
	}

	return fmt.Sprintf("%s %s %s",
		strings.Join(conditionFilters, " AND "),
		group,
		pagination)
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
	filter := &SqlFilter{
		pagination: sqlPaginationFilter{Limit: 100, Offset: 0},
	}
	return filter
}
