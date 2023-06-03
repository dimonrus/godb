package test

import (
	"fmt"
	"github.com/dimonrus/gocli"
	"github.com/dimonrus/godb/v2"
	// _ "github.com/lib/pq"
	"testing"
)

var postgresTestCase = QueryTestCase{
	create: "CREATE TABLE IF NOT EXISTS users (id INT NOT NULL PRIMARY KEY, name TEXT, age INT)",
	drop:   "DROP TABLE IF EXISTS users",
	insert: []string{
		"INSERT INTO users (id, name, age) VALUES (1, 'John', 33)",
		"INSERT INTO users (id, name, age) VALUES (2, 'Ksenia', 26)",
		"INSERT INTO users (id, name, age) VALUES (3, 'Michael', 43)",
	},
	prepare: "INSERT INTO users (id, name, age) VALUES (?, ?, ?)",
	items: [][]any{
		{10, "Foo", 15},
		{11, "Bar", 16},
		{12, "Baz", 17},
	},
	fetch:  "SELECT * FROM users",
	update: "UPDATE users SET age = 27 WHERE id = 2",
	delete: "DELETE FROM users WHERE name = 'Michael'",
	scan:   []any{new(int), new(string), new(int)},
}

type postgresDockerConnection struct{}

func (c *postgresDockerConnection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"0.0.0.0", 5432, "migrate", "migrate", "migrate")
}

func (c *postgresDockerConnection) GetDbType() string {
	return "postgres"
}

func (c *postgresDockerConnection) GetMaxConnection() int {
	return 200
}

func (c *postgresDockerConnection) GetMaxIdleConns() int {
	return 15
}

func (c *postgresDockerConnection) GetConnMaxLifetime() int {
	return 50
}

func getPostgresDockerConnection() (*godb.DBO, error) {
	return godb.DBO{
		Options: godb.Options{
			Debug:          true,
			QueryProcessor: godb.PreparePositionalArgsQuery,
			Logger:         gocli.NewLogger(gocli.LoggerConfig{}),
		},
		Connection: &postgresDockerConnection{},
	}.Init()
}

func TestIsTableExistsPostgres(t *testing.T) {
	q, err := getPostgresDockerConnection()
	if err != nil {
		t.Fatal(err)
	}
	testCase := postgresTestCase
	t.Run("test_table_postgres_ok", func(t *testing.T) {
		err = testCase.CreateTable(q)
		if err != nil {
			t.Fatal(err)
		}
		if !godb.IsTableExists(q, testTable, "") {
			t.Fatal("table must exists")
		}
		err = testCase.DropTable(q)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("test_table_postgres_fail", func(t *testing.T) {
		if godb.IsTableExists(q, testTable+"nok", "") {
			t.Fatal("table must not exists")
		}
	})
}

func TestPostgresCase(t *testing.T) {
	q, err := getPostgresDockerConnection()
	if err != nil {
		t.Fatal(err)
	}
	testCase := postgresTestCase
	t.Run("simple_connection", func(t *testing.T) {
		if godb.IsTableExists(q, testTable, "") {
			err = testCase.DropTable(q)
			if err != nil {
				t.Fatal(err)
			}
		}
		err = testCase.CreateTable(q)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err = testCase.DropTable(q)
			if err != nil {
				t.Fatal(err)
			}
		}()
		err = testCase.Insert(q)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Select(q)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Update(q)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Delete(q)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("in_transaction", func(t *testing.T) {
		tx, err := q.Begin()
		defer func() {
			var tError error
			if err == nil {
				tError = tx.Commit()
			} else {
				tError = tx.Rollback()
			}
			if tError != nil {
				panic(tError)
			}
		}()
		if err != nil {
			t.Fatal(err)
		}
		if godb.IsTableExists(tx, testTable, "") {
			err = testCase.DropTable(q)
			if err != nil {
				t.Fatal(err)
			}
		}
		err = testCase.CreateTable(tx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err = testCase.DropTable(tx)
			if err != nil {
				t.Fatal(err)
			}
		}()
		err = testCase.Insert(tx)
		if err != nil {
			t.Fatal(err)
		}
		stmt, err := testCase.Prepare(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.ExecStatement(stmt)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Select(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Update(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.Delete(tx)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("in_transaction_rollback", func(t *testing.T) {
		tx, err := q.Begin()
		if err != nil {
			t.Fatal(err)
		}
		err = testCase.CreateTable(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = tx.Rollback()
		if err != nil {
			t.Fatal(err)
		}
		if godb.IsTableExists(q, testTable, "") {
			t.Fatal("table must not exists")
		}
	})
}
