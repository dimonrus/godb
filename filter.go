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

// Filter Field
type sqlFieldsFilter struct {
	Field     string
	Condition string
	Value     interface{}
}

// In Filter
type sqlInFilter struct {
	Expression string
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
	fields     []sqlFieldsFilter
	inExpr     []sqlInFilter
	expression []sqlExpression
	orders     []sqlOrderFilter
	pagination sqlPaginationFilter
	arguments  []interface{}
}

// Add fileld to filter
func (f *SqlFilter) AddFiledFilter(field string, condition string, value interface{}) *SqlFilter {
	f.fields = append(f.fields, sqlFieldsFilter{Field: field, Condition: condition, Value: "?"})
	f.arguments = append(f.arguments, value)
	return f
}

// Add in Filter
func (f *SqlFilter) AddInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	inFilter := sqlInFilter{
		fmt.Sprintf("%s IN (%s)", field, strings.Join(condition, ",")),
	}
	f.inExpr = append(f.inExpr, inFilter)
	f.arguments = append(f.arguments, values...)
	return f
}

// Add not in filter
func (f *SqlFilter) AddNotInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	inFilter := sqlInFilter{
		fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(condition, ",")),
	}
	f.inExpr = append(f.inExpr, inFilter)
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
	var filters []string
	var inFilters []string
	var conditionFilters []string
	var expressionFilters []string
	var orders []string
	var pagination string

	for _, value := range f.orders {
		orders = append(orders, value.Field+" "+value.Direction)
	}

	if len(orders) > 0 {
		pagination = "ORDER BY " + strings.Join(orders, ", ") + " "
	}

	if f.pagination.Limit > 0 {
		pagination = fmt.Sprintf("%sLIMIT %v OFFSET %v", pagination, f.pagination.Limit, f.pagination.Offset)
	}

	for _, value := range f.fields {
		filters = append(filters, fmt.Sprintf("%s %s %s", value.Field, value.Condition, value.Value))
	}
	for _, value := range f.inExpr {
		inFilters = append(inFilters, value.Expression)
	}
	for _, value := range f.expression {
		expressionFilters = append(expressionFilters, value.Expression)
	}

	if len(filters) > 0 {
		conditionFilters = append(conditionFilters, filters...)
	}

	if len(inFilters) > 0 {
		conditionFilters = append(conditionFilters, inFilters...)
	}

	if len(expressionFilters) > 0 {
		conditionFilters = append(conditionFilters, expressionFilters...)
	}

	return fmt.Sprintf("%s %s",
		strings.Join(conditionFilters, " AND "),
		pagination)
}

// Get query with WHERE
func (f SqlFilter) GetWithWhere() string {
	if len(f.fields) > 0 || len(f.inExpr) > 0 || len(f.expression) > 0 {
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
