package godb

import (
	"fmt"
	"github.com/dimonrus/gohelp"
	"github.com/dimonrus/porterr"
	"github.com/lib/pq"
	"testing"
	"time"
)

type ModelIntegration struct {
	schema string
}

type TestModel struct {
	Id        *int       `json:"id" db:"col~id;req;seq;"`
	Name      *string    `json:"name" db:"col~name;req;"`
	Pages     []string   `json:"pages" db:"col~pages;"`
	SomeInt   *int       `json:"someInt" db:"col~some_int;"`
	CreatedAt *time.Time `json:"createdAt" db:"col~created_at;cat;"`
}

// Model table name
func (m *TestModel) Table() string {
	return "public.test_model"
}

// Model columns
func (m *TestModel) Columns() []string {
	return []string{"id", "name", "pages", "some_int", "created_at"}
}

// Model values
func (m *TestModel) Values() []interface{} {
	return []interface{}{&m.Id, &m.Name, pq.Array(&m.Pages), &m.SomeInt, &m.CreatedAt}
}
func (m *TestModel) Load(q Queryer) porterr.IError   { return nil }
func (m *TestModel) Save(q Queryer) porterr.IError   { return nil }
func (m *TestModel) Delete(q Queryer) porterr.IError { return nil }

func NewTestModel() *TestModel {
	id := gohelp.GetRndId()
	someInt := gohelp.GetRndNumber(10, 3000)
	name := gohelp.RandString(10)
	pages := []string{"one", "two"}
	return &TestModel{
		Id:      &id,
		Name:    &name,
		Pages:   pages,
		SomeInt: &someInt,
	}
}

func TestModelDeleteQuery(t *testing.T) {
	m := NewTestModel()
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
	m := NewTestModel()
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

//classic BenchmarkModelInsertQuery-8   	  689485	      1741 ns/op	     224 B/op	      11 allocs/op
func BenchmarkModelInsertQuery(b *testing.B) {
	m := NewTestModel()
	for i := 0; i < b.N; i++ {
		_, _, e := ModelInsertQuery(m, &m.Id, &m.Name, &m.Pages)
		if e != nil {
			b.Fatal(e)
		}
	}
	b.ReportAllocs()
}

func TestModelValues(t *testing.T) {
	m := NewTestModel()
	vals := ModelValues(m, "id", "pages", "some_int")
	fmt.Printf("%+v\n", vals)
	vals[0] = new(int)
	fmt.Println(m)
}

func BenchmarkModelValues(b *testing.B) {
	m := NewTestModel()
	for i := 0; i < b.N; i++ {
		ModelValues(m, "id", "pages", "some_int")
	}
	b.ReportAllocs()
}

func TestModelColumn(t *testing.T) {
	m := NewTestModel()
	cond := NewSqlCondition(ConditionOperatorAnd)
	cond.AddExpression("id = ?", 1)
	q, _, e := ModelUpdateQuery(m, cond, &m.Name, &m.SomeInt)
	if e != nil {
		t.Fatal(e)
	}
	fmt.Print(q)
}

func TestModelColumns(t *testing.T) {
	m := NewTestModel()
	columns := ModelColumns(m, &m.Name, &m.SomeInt)
	if len(columns) != 2 {
		t.Fatal("wrong")
	}
	if columns[0] != "name" {
		t.Fatal("wrong")
	}
	if columns[1] != "some_int" {
		t.Fatal("wrong")
	}
}

func BenchmarkTestModel(b *testing.B) {
	var c = make([]TestModel, b.N)
	//var m = &TestModel{}
	for i := 0; i < b.N; i++ {
		c[i].Id = &i
	}
	fmt.Println(len(c), c[len(c)-1].Id)
	b.ReportAllocs()
}
