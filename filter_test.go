package godb

import (
	"fmt"
	"github.com/lib/pq"
	"testing"
)

func TestSqlFilter_AddOrder(t *testing.T) {
	filter := SqlFilter{}
	var values = []int64{11,12}
	filter.Where().AddExpression("name = ANY(?)", pq.Array(values))
	filter.AddOrder("name DESC")
	filter.GroupBy("name")
	filter.GroupBy("entity")

	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE (name = ANY(?)) GROUP BY name, entity ORDER BY name DESC" {
		t.Fatal("wrong work")
	}

	filter.ResetGroupBy()
	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE (name = ANY(?)) ORDER BY name DESC" {
		t.Fatal("wrong work")
	}
}

func TestSqlFilter_AddOrFilters(t *testing.T) {
	filter := NewSqlFilter()
	filter.Where().Merge(ConditionOperatorAnd,
		NewSqlCondition(ConditionOperatorOr).
		AddExpression("one = ?", 1).
		AddExpression("one = ?", 2),
	)
	filter.Where().AddExpression("two != ?",10)

	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE ((one = ? OR one = ?) AND (two != ?))" {
		t.Fatal("wrong work")
	}
}

func TestSqlFilter_Having(t *testing.T) {
	filter := NewSqlFilter()
	filter.Where().Merge(ConditionOperatorAnd,
		NewSqlCondition(ConditionOperatorOr).
			AddExpression("one = ?", 1).
			AddExpression("one = ?"),
	)
	filter.Where().AddExpression("two != ?",10)
	filter.Having().AddExpression("three = ?", 3)
	filter.AddOrder("created_at DESC")
	filter.GroupBy("one")
	filter.SetPagination(10, 20)

	fmt.Println(filter.GetWithWhere())
	fmt.Println(filter.GetArguments())

	if filter.GetWithWhere() != "WHERE ((one = ? OR one = ?) AND (two != ?)) GROUP BY one HAVING (three = ?) ORDER BY created_at DESC LIMIT 10 OFFSET 20" {
		t.Fatal("Wrong filter works")
	}
}

func TestSqlFilter_Where(t *testing.T) {
	filter := NewSqlFilter()
	filter.Where().
		AddExpression("two != ?",10).
		AddExpression("three = ?", 3)

	filter.Where().Merge(ConditionOperatorOr, NewSqlCondition(ConditionOperatorAnd).AddExpression("one = ?", 1))
	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE ((one = ?) OR (two != ? AND three = ?))" {
		t.Fatal("Wrong filter works")
	}
}
