package godb

// SQL Filter

import (
	"fmt"
	"strings"
)

// SQL Pagination limit offset
type sqlPagination struct {
	Limit  int
	Offset int
}

// Filter struct
type SqlFilter struct {
	columns    []string
	from       []string
	join       []string
	where      Condition
	orders     []string
	group      []string
	having     Condition
	pagination sqlPagination
}

// Add column
func (f *SqlFilter) Columns(column ...string) *SqlFilter {
	f.columns = append(f.columns, column...)
	return f
}

// Reset column
func (f *SqlFilter) ResetColumns() *SqlFilter {
	f.columns = []string{}
	return f
}

// Add from
func (f *SqlFilter) From(table ...string) *SqlFilter {
	f.from = append(f.from, table...)
	return f
}

// Reset column
func (f *SqlFilter) ResetFrom() *SqlFilter {
	f.from = []string{}
	return f
}

// Add join
func (f *SqlFilter) Relate(relation ...string) *SqlFilter {
	f.join = append(f.join, relation...)
	return f
}

// Reset join
func (f *SqlFilter) ResetRelations() *SqlFilter {
	f.join = []string{}
	return f
}

// Where conditions
func (f *SqlFilter) Where() *Condition {
	return &f.where
}

// Where conditions
func (f *SqlFilter) Having() *Condition {
	return &f.having
}

// Add Order
func (f *SqlFilter) AddOrder(expression ...string) *SqlFilter {
	f.orders = append(f.orders, expression...)
	return f
}

// Reset Order
func (f *SqlFilter) ResetOrder() *SqlFilter {
	f.orders = []string{}
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
	var result = make([]string, 0, 6)

	// Select columns
	if len(f.columns) > 0 {
		result = append(result, "SELECT "+strings.Join(f.columns, ", "))
	}

	// From table
	if len(f.from) > 0 {
		result = append(result, "FROM "+strings.Join(f.from, ", "))
	}

	// From table
	if len(f.join) > 0 {
		result = append(result, strings.Join(f.join, " "))
	}

	// Where conditions
	if len(f.where.expression) > 0 || f.where.merge != nil {
		result = append(result, "WHERE "+f.where.String())
	}

	// Prepare groups
	if len(f.group) > 0 {
		result = append(result, "GROUP BY "+strings.Join(f.group, ", "))
	}

	// Prepare having expression
	if len(f.having.expression) > 0 || f.having.merge != nil {
		result = append(result, "HAVING "+f.having.String())
	}

	// Prepare orders
	if len(f.orders) > 0 {
		result = append(result, "ORDER BY "+strings.Join(f.orders, ", "))
	}

	// Prepare pagination
	if f.pagination.Limit > 0 {
		result = append(result, fmt.Sprintf("LIMIT %v OFFSET %v", f.pagination.Limit, f.pagination.Offset))
	}

	return strings.Join(result, " ")
}

// New SQL Filter with pagination
func NewSqlFilter() *SqlFilter {
	return &SqlFilter{
		where:  Condition{operator: ConditionOperatorAnd},
		having: Condition{operator: ConditionOperatorAnd},
	}
}
