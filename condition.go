package godb

import "strings"

const (
	ConditionOperatorAnd = "AND"
	ConditionOperatorOr  = "OR"
	ConditionOperatorXor = "XOR"
)

// Merge with Condition
type merge struct {
	operator  string
	condition []*Condition
}

// Condition type
type Condition struct {
	operator   string
	expression []string
	argument   []interface{}
	merge      *merge
}

// Init Condition
func NewSqlCondition(operator string) *Condition {
	return &Condition{operator: operator}
}

// Get string of conditions
func (c *Condition) String() string {
	if c.merge != nil {
		var slaves []string
		for i := range (*c.merge).condition {
			slaves = append(slaves, (*c.merge).condition[i].String())
		}
		if c.expression != nil {
			slaves = append(slaves, "(" + strings.Join(c.expression, " "+c.operator+" ") + ")")
		}
		return "(" + strings.Join(slaves, " "+c.merge.operator+" ") + ")"
	} else {
		if c.expression != nil {
			return "(" + strings.Join(c.expression, " "+c.operator+" ") + ")"
		}
		return ""
	}
}

// Get arguments
func (c *Condition) GetArguments() []interface{} {
	var arguments = make([]interface{}, 0)
	if c.merge != nil {
		for i := range (*c.merge).condition {
			arguments = append(arguments, (*c.merge).condition[i].GetArguments()...)
		}
	}
	return append(arguments, c.argument...)

}

// Add expression
func (c *Condition) AddExpression(expression string, values ...interface{}) *Condition {
	c.expression = append(c.expression, expression)
	c.argument = append(c.argument, values...)
	return c
}

// Merge with conditions
func (c *Condition) Merge(operator string, conditions ...*Condition) *Condition {
	if len(conditions) > 0 {
		for i := range conditions {
			if conditions[i] == nil {
				continue
			}
			if c.merge == nil {
				c.merge = &merge{operator: operator, condition: []*Condition{}}
			}
			c.merge.condition = append(c.merge.condition, conditions[i])
		}
	}
	return c
}
