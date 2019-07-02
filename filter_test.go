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
	if filter.GetWithWhere() != "WHERE name IN (?,?) GROUP BY name, entity ORDER BY name desc " {
		t.Fatal("wrong work")
	}

	filter.ResetGroupBy()
	fmt.Println(filter.GetWithWhere())
	if filter.GetWithWhere() != "WHERE name IN (?,?)  ORDER BY name desc " {
		t.Fatal("wrong work")
	}
}
