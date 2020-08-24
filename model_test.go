package godb

import (
	"fmt"
	"github.com/dimonrus/porterr"
	"testing"
	"time"
)

type TestModel struct {
	Id        int       `json:"id" seq:"true" column:"id"`
	Name      string    `json:"name" column:"name"`
	SomeInt      int    `json:"someInt" column:"some_int"`
	CreatedAt time.Time `json:"createdAt" column:"created_at"`
	notAField string
}

// Model table name
func (m *TestModel) Table() string {
	return "public.test_model"
}

// Model columns
func (m *TestModel) Columns() []string {
	return []string{"id", "name", "created_at"}
}

// Model values
func (m *TestModel) Values() (values []interface{}) {
	return append(values, &m.Id, &m.Name, &m.Name, &m.CreatedAt)
}
func (m *TestModel) Load(q Queryer) porterr.IError   { return nil }
func (m *TestModel) Save(q Queryer) porterr.IError   { return nil }
func (m *TestModel) Delete(q Queryer) porterr.IError { return nil }

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

func TestModelInsertQuery(t *testing.T) {
	m := &TestModel{}
	q, cols, e := ModelInsertQuery(m)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(q, cols)

	q, cols, e = ModelInsertQuery(m, &m.Name, &m.SomeInt)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(q, cols)
}

func BenchmarkModelInsertQuery(b *testing.B) {
	m := &TestModel{}
	for i := 0; i < b.N; i++ {
		_, _, e := ModelInsertQuery(m)
		if e != nil {
			b.Fatal(e)
		}
	}
	b.ReportAllocs()
}

func TestModelValues(t *testing.T) {
	m := &TestModel{
		Id:        10,
		Name:      "scdscs",
		SomeInt:   12123,
		CreatedAt: time.Now(),
	}
	vals := ModelValues(m, "id", "some_int")
	fmt.Println(vals)
}

func TestModelColumn(t *testing.T) {
	m := &TestModel{
		Id:   0,
		Name: "asasf",
	}
	cond := NewSqlCondition(ConditionOperatorAnd)
	cond.AddExpression("id = ?", 1)
	q, _, e := ModelUpdateQuery(m, cond, &m.Name)
	if e != nil {
		t.Fatal(e)
	}

	fmt.Print(q)
}

func BenchmarkTestModel(b *testing.B) {
	var c = make([]TestModel, b.N)
	//var m = &TestModel{}
	for i := 0; i < b.N; i++ {
		c[i].Id = i
	}
	fmt.Println(len(c), c[len(c)-1].Id)
	b.ReportAllocs()
}
