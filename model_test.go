package godb

import (
	"github.com/dimonrus/porterr"
	"testing"
	"time"
)

type TestModel struct {
	Id        int       `json:"id"columns:"id"`
	Name      string    `json:"name"columns:"name"`
	CreatedAt time.Time `json:"createdAt"columns:"created_at"`
}

// Model table name
func (m *TestModel) Table() string {
	return "public.test_model"
}

// Model columns
func (m *TestModel) Columns() []string {
	return []string{"id", "name", "created_at",}
}

// Model values
func (m *TestModel) Values() (values []interface{}) {
	return append(values, &m.Id, &m.Name, &m.Name, &m.CreatedAt, )
}
func (m *TestModel) Load(q Queryer) porterr.IError {return nil}
func (m *TestModel) Save(q Queryer) porterr.IError {return nil}
func (m *TestModel) Delete(q Queryer) porterr.IError {return nil}

func TestModelDeleteQuery(t *testing.T) {
	m := &TestModel{}
	c := NewSqlCondition(ConditionOperatorAnd)
	c.AddExpression("created_at >= NOW()")
	sql, e := ModelDeleteQuery(m, c)
	if e != nil {
		t.Fatal(e)
	}
	if sql != "DELETE FROM public.test_model WHERE (created_at >= NOW());" {
		t.Fatal("Wrong sql prepared")
	}

	sql, e = ModelDeleteQuery(m, nil)
	if e != nil {
		t.Fatal(e)
	}
	if sql != "DELETE FROM public.test_model;" {
		t.Fatal("Wrong sql prepared")
	}
}
