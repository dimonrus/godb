package godb

import (
	"fmt"
	"strings"
)

// Query Builder struct
type QB struct {
	with       map[string]*QB
	columns    []string
	from       []string
	join       []string
	where      Condition
	orders     []string
	group      []string
	union      []*QB
	having     Condition
	pagination sqlPagination
}

// Add With
func (f *QB) With(name string, qb *QB) *QB {
	if name != "" {
		f.with[name] = qb
	}
	return f
}

// Reset With
func (f *QB) ResetWith() *QB {
	f.with = make(map[string]*QB)
	return f
}

// Union
func (f *QB) Union(qb *QB) *QB {
	if qb != nil {
		f.union = append(f.union, qb)
	}
	return f
}

// Reset Union
func (f *QB) ResetUnion() *QB {
	f.union = make([]*QB, 0)
	return f
}

// Get With
func (f *QB) GetWith(name string) *QB {
	if v, ok := f.with[name]; ok {
		return v
	}
	return nil
}

// Add column
func (f *QB) Columns(column ...string) *QB {
	f.columns = append(f.columns, column...)
	return f
}

// Reset column
func (f *QB) ResetColumns() *QB {
	f.columns = []string{}
	return f
}

// Add from
func (f *QB) From(table ...string) *QB {
	f.from = append(f.from, table...)
	return f
}

// Add from model
func (f *QB) ModelFrom(model ...IModel) *QB {
	for _, m := range model {
		f.from = append(f.from, m.Table())
	}
	return f
}

// Add column from model
func (f *QB) ModelColumns(model ...IModel) *QB {
	for _, m := range model {
		f.columns = append(f.columns, m.Columns()...)
	}
	return f
}

// Reset column
func (f *QB) ResetFrom() *QB {
	f.from = []string{}
	return f
}

// Add join
func (f *QB) Relate(relation ...string) *QB {
	f.join = append(f.join, relation...)
	return f
}

// Reset join
func (f *QB) ResetRelations() *QB {
	f.join = []string{}
	return f
}

// Where conditions
func (f *QB) Where() *Condition {
	return &f.where
}

// Where conditions
func (f *QB) Having() *Condition {
	return &f.having
}

// Add Order
func (f *QB) AddOrder(expression ...string) *QB {
	f.orders = append(f.orders, expression...)
	return f
}

// Reset Order
func (f *QB) ResetOrder() *QB {
	f.orders = []string{}
	return f
}

// Add Group
func (f *QB) GroupBy(fields ...string) *QB {
	f.group = append(f.group, fields...)
	return f
}

// Reset Group
func (f *QB) ResetGroupBy() *QB {
	f.group = []string{}
	return f
}

// Set pagination
func (f *QB) SetPagination(limit int, offset int) *QB {
	f.pagination = sqlPagination{Limit: limit, Offset: offset}
	return f
}

// Get arguments
func (f *QB) GetArguments() []interface{} {
	arguments := make([]interface{}, 0)
	if len(f.with) > 0 {
		for _, w := range f.with {
			arguments = append(arguments, w.GetArguments()...)
		}
	}

	arguments = append(arguments, append(f.where.GetArguments(), f.having.GetArguments()...)...)

	if len(f.union) > 0 {
		for _, u := range f.union {
			arguments = append(arguments, u.GetArguments()...)
		}
	}
	return arguments
}

// Make SQL query
func (f QB) String() string {
	var result = make([]string, 0)
	var with = make([]string, 0)
	var union = make([]string, 0)

	// With render
	if len(f.with) > 0 {
		for name, w := range f.with {
			with = append(with, name+" AS ("+w.String()+")")
		}
		result = append(result, "WITH "+strings.Join(with, ", "))
	}

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

	// Union render
	if len(f.union) > 0 {
		for _, u := range f.union {
			union = append(union, u.String())
		}
		result = append(result, "UNION "+strings.Join(union, " UNION "))
	}

	return strings.Join(result, " ")
}

// New Query Builder
func NewQB() *QB {
	return &QB{
		with:   make(map[string]*QB),
		where:  Condition{operator: ConditionOperatorAnd},
		having: Condition{operator: ConditionOperatorAnd},
	}
}
