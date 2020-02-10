package godb

import (
	"fmt"
	"github.com/lib/pq"
	"testing"
)

func TestQB_String(t *testing.T) {
	mr := NewQB()
	mr.From("mv_right")
	mr.Columns("id", "contract_id", "object_id")
	mr.Where().AddExpression("object_id = ANY(?)", pq.Array([]string{"84f3ba22-5b7f-4967-80e2-451a123deff6", "81297e48-869e-49d6-876e-11078d982aff"}))
	mr.AddOrder("terrirtory_name")
	mr.SetPagination(10, 0)

	c := NewQB()
	c.From("mv_contracts")
	c.Columns("id", "contract_name")
	c.Where().AddExpression("contract_sum > ?", 23.45)
	c.SetPagination(5, 0)

	qb := NewQB()
	qb.With("mv_right_items", mr)
	qb.With("mv_contracts_items", c)
	qb.From("mv_object mo")
	qb.Columns("mo.id", "mo.title", "mo.rightholder_ids", "mr.id", "mr.contract_id")
	qb.Relate("JOIN mv_right_items AS mr ON mr.object_id = mo.id")
	qb.Relate("LEFT JOIN mv_contracts_items AS ci ON ci.id = mr.contract_id")
	qb.Where().AddExpression("mr.object_id IS NOT NULL")

	fmt.Println(qb.String())
}
