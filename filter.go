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

// SQL Pagination limit offset
type sqlPagination struct {
	Limit  int
	Offset int
}

// Filter struct
type SqlFilter struct {
	where      Condition
	orders     []sqlOrder
	group      []string
	having     Condition
	pagination sqlPagination
}

// Add or filter
// Deprecated: use where Condition merge
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

// Where conditions
func (f *SqlFilter) Where() *Condition {
	return &f.where
}

// Where conditions
func (f *SqlFilter) Having() *Condition {
	return &f.having
}

// Add filed to filter
// Deprecated: use where Condition merge
func (f *SqlFilter) AddFiledFilter(field string, condition string, value interface{}) *SqlFilter {
	f.where.AddExpression(field + " " + condition + " ?", value)
	return f
}

// Add in Filter
// Deprecated: use where Condition merge
func (f *SqlFilter) AddInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	f.where.AddExpression(fmt.Sprintf("%s IN (%s)", field, strings.Join(condition, ",")), values...)
	return f
}

// Add not in filter
// Deprecated: use where Condition merge
func (f *SqlFilter) AddNotInFilter(field string, values []interface{}) *SqlFilter {
	condition := make([]string, len(values))
	for i := range condition {
		condition[i] = "?"
	}
	f.where.AddExpression(fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(condition, ",")), values...)
	return f
}

// Add filter expression
// Deprecated: use where or having Condition merge
func (f *SqlFilter) AddExpression(expression string, values []interface{}) *SqlFilter {
	f.where.AddExpression(expression, values...)
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
	return append(f.where.GetArguments(), f.having.GetArguments()...)
}

// Make SQL query
func (f SqlFilter) String() string {
	var orders []string
	var result = make([]string, 0, 4)

	// where conditions
	if len(f.where.expression) > 0 || f.where.merge != nil {
		result = append(result, f.where.String())
	}

	// Prepare groups
	if len(f.group) > 0 {
		result = append(result, "GROUP BY "+strings.Join(f.group, ", "))
	}

	// having expression
	if len(f.having.expression) > 0 || f.having.merge != nil {
		result = append(result, "HAVING "+f.having.String())
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
	if len(f.where.expression) > 0 || f.where.merge != nil {
		return "WHERE " + f.String()
	}

	return f.String()
}

// New SQL Filter with pagination
func NewSqlFilter() *SqlFilter {
	return &SqlFilter{where: Condition{operator: ConditionOperatorAnd}}
}
