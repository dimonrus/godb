package godb

import (
	"strconv"
	"strings"
)

const (
	// ConflictActionNothing On conflict action do nothing
	ConflictActionNothing = "NOTHING"

	// ConflictActionUpdate On conflict action do update nothing
	ConflictActionUpdate = "UPDATE"
)

// Insert query builder
type Insert struct {
	// with query
	with sqlWith
	// into table name
	into string
	// from insert
	from string
	// list of columns
	columns []string
	// list of values
	values []any
	// conflict expression
	conflict conflict
	// returning
	returning []string
}

// On conflict query part
type conflict struct {
	// object of conflict
	object string
	// action on conflict
	action string
	// set of changes
	set []string
	// condition
	condition Condition
	// constraint
	constraint string
}

// String conflict expression
func (c *conflict) String() string {
	if c.IsEmpty() {
		return ""
	}
	b := strings.Builder{}
	b.WriteString("ON CONFLICT")
	if c.object != "" {
		b.WriteString(" (" + c.object + ")")
	}
	if c.constraint != "" {
		b.WriteString(" ON CONSTRAINT " + c.constraint)
	}
	if len(c.set) > 0 {
		if c.action != "" {
			b.WriteString(" DO " + c.action)
		}
		b.WriteString(" SET " + strings.Join(c.set, ","))
		if !c.condition.IsEmpty() {
			b.WriteString(" WHERE " + c.condition.String())
		}
	} else {
		if !c.condition.IsEmpty() {
			b.WriteString(" WHERE " + c.condition.String())
		}
		if c.action != "" {
			b.WriteString(" DO " + c.action)
		}
	}
	return b.String()
}

// IsEmpty Is conflict empty
func (c *conflict) IsEmpty() bool {
	return c.object == "" && c.action == "" && len(c.set) == 0 && c.condition.IsEmpty() && c.constraint == ""
}

// Object of conflict
func (c *conflict) Object(object string) *conflict {
	c.object = object
	return c
}

// ResetObject reset
func (c *conflict) ResetObject() *conflict {
	c.object = ""
	return c
}

// Action of conflict
func (c *conflict) Action(action string) *conflict {
	c.action = action
	return c
}

// ResetAction reset action
func (c *conflict) ResetAction() *conflict {
	c.action = ""
	return c
}

// Set of expressions on conflict
func (c *conflict) Set(expr ...string) *conflict {
	c.set = append(c.set, expr...)
	return c
}

// ResetSet of expressions on conflict
func (c *conflict) ResetSet() *conflict {
	c.set = make([]string, 0)
	return c
}

// Condition set conflict condition
func (c *conflict) Condition(cond Condition) *conflict {
	c.condition = cond
	return c
}

// ResetCondition reset condition
func (c *conflict) ResetCondition() *conflict {
	c.condition = Condition{}
	return c
}

// Constraint set constraint
func (c *conflict) Constraint(constraint string) *conflict {
	c.constraint = constraint
	return c
}

// ResetConstraint reset constraint
func (c *conflict) ResetConstraint() *conflict {
	c.constraint = ""
	return c
}

// Get sql insert query
func (i *Insert) String() string {
	b := strings.Builder{}
	if i.with.Len() > 0 {
		b.WriteString("WITH")
		for index, w := range i.with.queries {
			if index < len(i.with.queries)-1 {
				b.WriteString(" " + i.with.keys[index] + " AS (" + w.String() + "),")
			} else {
				b.WriteString(" " + i.with.keys[index] + " AS (" + w.String() + ")")
			}
		}
		b.WriteString(" ")
	}
	if i.into != "" {
		b.WriteString("INSERT INTO " + i.into)
	}
	if len(i.columns) > 0 {
		b.WriteString(" (" + strings.Join(i.columns, ", ") + ")")
	}
	if i.from != "" {
		b.WriteString(" " + i.from)
	} else if len(i.values) > 0 {
		b.WriteString(" VALUES ")
		var k = 1
		for j := 0; j <= len(i.values)-1; j++ {
			b.WriteString("(")
			for u := 0; u < len(i.columns); u++ {
				b.WriteString("$" + strconv.Itoa(k))
				k++
				if u == len(i.columns)-1 {
					break
				}
				b.WriteString(", ")
			}
			b.WriteString(")")
			if j == len(i.values)-1 {
				break
			}
			b.WriteString(", ")
		}
	}
	if !i.conflict.IsEmpty() {
		b.WriteString(" " + i.Conflict().String())
	}
	if len(i.returning) > 0 {
		b.WriteString(" RETURNING " + strings.Join(i.returning, ", "))
	}
	b.WriteString(";")
	return b.String()
}

// SQL return query with params
func (i *Insert) SQL() (query string, params []any) {
	return i.String(), i.values
}

// Conflict get conflict expression
func (i *Insert) Conflict() *conflict {
	return &i.conflict
}

// AddReturning Add returning expression
func (i *Insert) AddReturning(returning ...string) *Insert {
	i.returning = append(i.returning, returning...)
	return i
}

// ResetReturning Reset returning expressions
func (i *Insert) ResetReturning() *Insert {
	i.returning = make([]string, 0)
	return i
}

// From insert from
func (i *Insert) From(from string) *Insert {
	i.from = from
	return i
}

// ResetFrom clear from
func (i *Insert) ResetFrom() *Insert {
	i.from = ""
	return i
}

// AddValues Add row values
func (i *Insert) AddValues(values []any) *Insert {
	i.values = append(i.values, values)
	return i
}

// SetValues Set row values by index
func (i *Insert) SetValues(index int, values []any) *Insert {
	i.values[index] = values
	return i
}

// ResetValues Reset all values
func (i *Insert) ResetValues() *Insert {
	i.values = make([]any, 0)
	return i
}

// GetValues get values by indexes
func (i *Insert) GetValues(start, end int) []any {
	return i.values[start:end]
}

// GetAllValues get all values
func (i *Insert) GetAllValues() []any {
	return i.values
}

// Into Set into value
func (i *Insert) Into(into string) *Insert {
	i.into = into
	return i
}

// ResetInto Set into empty string
func (i *Insert) ResetInto() *Insert {
	i.into = ""
	return i
}

// Columns Set columns
func (i *Insert) Columns(column ...string) *Insert {
	i.columns = append(i.columns, column...)
	return i
}

// Columns Set columns
func (i *Insert) ResetColumns() *Insert {
	i.columns = make([]string, 0)
	return i
}

// Get With
func (i *Insert) GetWith(name string) *QB {
	for j, key := range i.with.keys {
		if key == name {
			return i.with.queries[j]
		}
	}
	return nil
}

// Add With
func (i *Insert) With(name string, qb *QB) *Insert {
	if name != "" {
		i.with.queries = append(i.with.queries, qb)
		i.with.keys[len(i.with.queries)-1] = name
	}
	return i
}

// ResetWith Reset With query
func (i *Insert) ResetWith() *Insert {
	i.with.queries = make([]*QB, 0)
	i.with.keys = make(map[int]string, 0)
	return i
}

// New Insert Query Builder
func NewInsert() *Insert {
	return &Insert{
		with: sqlWith{
			keys:    make(map[int]string),
			queries: make([]*QB, 0),
		},
	}
}
