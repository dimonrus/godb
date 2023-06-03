package test

import (
	"fmt"
	"github.com/dimonrus/gocli"
	"github.com/dimonrus/godb/v2"
	// _ "github.com/mattn/go-sqlite3"
	"testing"
)

type sqliteDockerConnection struct{}

func (c *sqliteDockerConnection) String() string {
	return fmt.Sprintf("./migrate.db")
}

func (c *sqliteDockerConnection) GetDbType() string {
	return "sqlite3"
}

func (c *sqliteDockerConnection) GetMaxConnection() int {
	return 200
}

func (c *sqliteDockerConnection) GetMaxIdleConns() int {
	return 15
}

func (c *sqliteDockerConnection) GetConnMaxLifetime() int {
	return 50
}

var sqliteTestCase = QueryTestCase{
	create: "CREATE TABLE IF NOT EXISTS users (id INTEGER NOT NULL PRIMARY KEY, name TEXT, age INTEGER)",
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

func getSqliteDockerConnection() (*godb.DBO, error) {
	return godb.DBO{
		Options: godb.Options{
			Debug:  true,
			Logger: gocli.NewLogger(gocli.LoggerConfig{}),
		},
		Connection: &sqliteDockerConnection{},
	}.Init()
}

func TestIsTableExistsSqlite(t *testing.T) {
	q, err := getSqliteDockerConnection()
	if err != nil {
		t.Fatal(err)
	}
	testCase := sqliteTestCase
	t.Run("test_table_sqlite_ok", func(t *testing.T) {
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
	t.Run("test_table_sqlite_fail", func(t *testing.T) {
		if godb.IsTableExists(q, testTable+"nok", "") {
			t.Fatal("table must not exists")
		}
	})
}

func TestSqliteCase(t *testing.T) {
	q, err := getSqliteDockerConnection()
	if err != nil {
		t.Fatal(err)
	}
	testCase := sqliteTestCase
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
		if err != nil {
			t.Fatal(err)
		}
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
