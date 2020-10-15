package godb

import (
	"fmt"
	"github.com/dimonrus/porterr"
	"github.com/lib/pq"
	"testing"
	"time"
)

type ModelIntegration struct {
	schema string
}

type TestModel struct {
	Id        int       `json:"id" sequence:"+" column:"id"`
	Name      string    `json:"name" sequence:"-" column:"name"`
	Pages     []string  `json:"pages" sequence:"-" column:"pages"`
	SomeInt   int       `json:"someInt" sequence:"-" column:"some_int"`
	CreatedAt time.Time `json:"createdAt" sequence:"-" column:"created_at"`
	ModelIntegration
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
func (m *TestModel) Values() []interface{} {
	return[]interface{}{&m.Id, &m.Name, &m.Name, pq.Array(&m.Pages), &m.CreatedAt}
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
		Pages:     []string{"one", "two"},
		SomeInt:   12123,
		CreatedAt: time.Now(),
	}
	vals := ModelValues(m, "id", "pages", "some_int")
	fmt.Printf("%+v\n", vals)
	vals[0] = new(int)

	fmt.Println(m)
}

func BenchmarkModelValues(b *testing.B) {
	m := &TestModel{
		Id:        10,
		Name:      "scdscs",
		Pages:     []string{"one", "two"},
		SomeInt:   12123,
		CreatedAt: time.Now(),
	}
	for i := 0; i < b.N; i++ {
		ModelValues(m, "id", "pages", "some_int")
	}
	b.ReportAllocs()
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
