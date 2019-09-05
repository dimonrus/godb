package godb

import "strings"

const (
	ConditionOperatorAnd = "AND"
	ConditionOperatorOr  = "OR"
	ConditionOperatorXor = "XOR"
)

// Merge with condition
type merge struct {
	operator  string
	condition []*condition
}

// Condition type
type condition struct {
	operator   string
	expression []string
	argument   []interface{}
	merge      *merge
}

// Init condition
func NewSqlCondition(operator string) *condition {
	return &condition{operator: operator}
}

// Get string of conditions
func (c *condition) String() string {
	if c.merge != nil {
		var slaves []string
		for i := range (*c.merge).condition {
			slaves = append(slaves, (*c.merge).condition[i].String())
		}
		slaves = append(slaves, "(" + strings.Join(c.expression, " "+c.operator+" ") + ")")
		return "(" + strings.Join(slaves, " "+c.merge.operator+" ") + ")"
	} else {
		return "(" + strings.Join(c.expression, " "+c.operator+" ") + ")"
	}
}

// Get arguments
func (c *condition) GetArguments() []interface{} {
	var arguments = make([]interface{}, 0)
	if c.merge != nil {
		for i := range (*c.merge).condition {
			arguments = append(arguments, (*c.merge).condition[i].GetArguments()...)
		}
	}
	return append(arguments, c.argument...)

}

// Add expression
func (c *condition) AddExpression(expression string, values ...interface{}) *condition {
	c.expression = append(c.expression, expression)
	c.argument = append(c.argument, values...)
	return c
}

// Merge with conditions
func (c *condition) Merge(operator string, conditions ...*condition) *condition {
	if len(conditions) > 0 {
		if c.merge == nil {
			c.merge = &merge{operator: operator, condition: conditions}
		} else {
			c.merge.condition = append(c.merge.condition, conditions...)
		}
	}
	return c
}
