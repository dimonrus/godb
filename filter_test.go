package godb

import (
	"fmt"
	"testing"
)

func TestSqlFilter_AddOrder(t *testing.T) {
	filter := SqlFilter{}
	var values []interface{}
	values = append(values, 11)
	values = append(values, 12)
	filter.AddInFilter("name", values)
	filter.AddOrder("name", "desc")
	filter.GroupBy("name")
	filter.GroupBy("entity")

	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE (name IN (?,?)) GROUP BY name, entity ORDER BY name desc" {
		t.Fatal("wrong work")
	}

	filter.ResetGroupBy()
	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE (name IN (?,?)) ORDER BY name desc" {
		t.Fatal("wrong work")
	}
}

func TestSqlFilter_AddOrFilters(t *testing.T) {
	filter := NewSqlFilter()

	f := NewSqlFilter().AddFiledFilter("one", "=", 3)
	fmt.Println(f.String())

	filter.AddOrFilters(
		NewSqlFilter().AddFiledFilter("one", "=", 1),
		NewSqlFilter().AddFiledFilter("one", "=", 5),
	)
	filter.AddFiledFilter("two", "!=", 10)

	fmt.Println(filter.GetArguments())

	//filter.AddOrder("one", "DESC")
	//filter.AddOrder("two", "ASC")

	//filter.GroupBy("one, two")
	//filter.SetPagination(10, 20)
	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE (((one = ?) OR (one = ?)) AND two != ?)" {
		t.Fatal("wrong work")
	}
}

func TestSqlFilter_Having(t *testing.T) {
	filter := NewSqlFilter()
	filter.AddOrFilters(
		NewSqlFilter().AddFiledFilter("one", "=", 1),
		NewSqlFilter().AddFiledFilter("one", "=", 5),
	)
	filter.AddFiledFilter("two", "!=", 10)
	filter.Having().AddExpression("three = ?", 3)
	filter.AddOrder("created_at", "DESC")
	filter.GroupBy("one")
	filter.SetPagination(10, 20)

	fmt.Println(filter.GetWithWhere())
	fmt.Println(filter.GetArguments())

}
