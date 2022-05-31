package godb

import (
	"database/sql"
	"fmt"
	"github.com/dimonrus/gocli"
	_ "github.com/lib/pq"
	"testing"
)

type testData struct {
	Id   int
	Code string
}

type testDatas []testData

type connection struct{}

func (c *connection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"10.10.10.100", 5433, "mpauth", "mpauth", "mpauth")
}
func (c *connection) GetDbType() string {
	return "postgres"
}
func (c *connection) GetMaxConnection() int {
	return 200
}
func (c *connection) GetMaxIdleConns() int {
	return 15
}

func (c *connection) GetConnMaxLifetime() int {
	return 50
}

type wrongConnection struct{ connection }

func (c *wrongConnection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5431, "postgres", "root", "goreav")
}
func (c *wrongConnection) GetDbType() string {
	return "unknown"
}

func TestDBO_InitError(t *testing.T) {
	_, err := DBO{
		Options: Options{
			Debug:  true,
			Logger: gocli.NewLogger(gocli.LoggerConfig{}),
		},
		Connection: &wrongConnection{},
	}.Init()
	if err == nil {
		t.Fatal("must be an error case")
	}
}

func initDb() (*DBO, error) {
	return DBO{
		Options: Options{
			Debug:  true,
			Logger: gocli.NewLogger(gocli.LoggerConfig{}),
		},
		Connection: &connection{},
	}.Init()
}

func createTable(db *DBO) (*DBO, error) {
	_, err := db.Exec("create table if not exists apple_attribute (id serial not null primary key, code text not null);")
	if err != nil {
		return db, err
	}
	_, err = db.Exec("insert into apple_attribute (code) values ('one'), ('two')")
	return db, err
}

func deleteTable(db *DBO) error {
	_, err := db.Exec("DROP TABLE IF EXISTS apple_attribute")
	return err
}

func createForeignTable(db *DBO) (*DBO, error) {
	_, err := db.Exec("create table if not exists apple_property (id serial not null primary key, name text not null, attribute_id int not null references apple_attribute(id));")
	if err != nil {
		return db, err
	}
	_, err = db.Exec("insert into apple_property (name, attribute_id) values ('one apple', 1)")
	return db, err
}

func deleteForeignTable(db *DBO) error {
	_, err := db.Exec("DROP TABLE IF EXISTS apple_property")
	return err
}

func TestSqlStmt_Exec(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	db, err = createTable(db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBO_Init(t *testing.T) {
	_, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
}

func testParseAll(rows *sql.Rows) (*testDatas, error) {
	var tds testDatas

	for rows.Next() {
		data := testData{}
		err := rows.Scan(&data.Id, &data.Code)
		if err != nil {
			return nil, err
		}
		tds = append(tds, data)
	}

	return &tds, nil
}

func TestDBO_Query(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("select id, code from apple_attribute where id in ($1, $2)", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	tds, err := testParseAll(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(*tds) == 0 {
		t.Fatal("no items in collection check records in db")
	}
}

func TestDBO_Exec(t *testing.T) {
	db, _ := initDb()
	_, err := db.Exec("update apple_attribute set code = 'name_test_update' where id = ?", 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("update apple_attribute set code = 'name' where id = ?", 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBO_QueryRow(t *testing.T) {
	db, _ := initDb()
	var code string
	var id int
	err := db.QueryRow("select id, code from apple_attribute where id = ?", 1).Scan(&id, &code)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 || code != "name" {
		t.Fatal("wrong query performed")
	}
}

func TestDBO_Begin(t *testing.T) {
	db, _ := initDb()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	tx.Rollback()
}

func TestSqlTx_Prepare(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := tx.Prepare("update apple_attribute set code = 'name_test_update' where id = ?")
	if err != nil {
		t.Fatal(err)
	}
	_, err = stmt.Exec(1)
	if err != nil {
		t.Fatal(err)
	}
	stmt, err = tx.Prepare("update apple_attribute set code = 'total_test_update' where id = ?")
	if err != nil {
		t.Fatal(err)
	}
	_, err = stmt.Exec(2)
	if err != nil {
		t.Fatal(err)
	}
	data := testData{}
	err = tx.QueryRow("select id, code from apple_attribute where id = ?", 2).Scan(&data.Id, &data.Code)
	if err != nil {
		t.Fatal(err)
	}
	if data.Code != "total_test_update" {
		t.Fatal("transaction update failed")
	}
	_, err = tx.Exec("update apple_attribute set code = 'total_test_update_new' where id = ?", 2)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := tx.Query("select id, code from apple_attribute where id = ?", 2)
	if err != nil {
		t.Fatal(err)
	}
	datas, err := testParseAll(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(*datas) == 0 {
		t.Fatal("no records found")
	}

	stmt, err = tx.Prepare("select id, code from apple_attribute where id = ?")
	if err != nil {
		t.Fatal(err)
	}
	err = stmt.QueryRow(2).Scan(&data.Id, &data.Code)
	if err != nil {
		t.Fatal(err)
	}
	if data.Code != "total_test_update_new" {
		t.Fatal("wrong code")
	}
	stmt, err = tx.Prepare("select id, code from apple_attribute where id in ($1, $2)")
	if err != nil {
		t.Fatal(err)
	}
	rows, err = stmt.Query(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	datas, err = testParseAll(rows)
	for _, value := range *datas {
		if value.Id == 2 && value.Code != "total_test_update_new" {
			t.Fatal("wrong code")
		}
	}
	tx.Rollback()
}

func TestDBO_ExecDelete(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}

	err = deleteTable(db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBO_Prepare(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	db, err = createTable(db)
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("update apple_attribute set code = 'name_test_update_prp' where id = ?")
	if err != nil {
		t.Fatal(err)
	}
	_, err = stmt.Exec(1)
	if err != nil {
		t.Fatal(err)
	}
	data := testData{}
	err = db.QueryRow("select id, code from apple_attribute where id = ?", 1).Scan(&data.Id, &data.Code)
	if err != nil {
		t.Fatal(err)
	}
	if data.Code != "name_test_update_prp" {
		t.Fatal("wrong code")
	}
	err = deleteTable(db)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkPreparePos(b *testing.B) {
	b.Run("normal", func(b *testing.B) {
		q := "update apple_attribute set code = 'name_test_update' where id = ? AND ab = ? OR ad = ? AND aa = ANY(?)"
		for i := 0; i < b.N; i++ {
			preparePositionalArgsQuery(q)
		}
		b.ReportAllocs()
	})
	b.Run("no_need", func(b *testing.B) {
		q := "update apple_attribute set code = 'name_test_update' where id = 1 AND ab = '2' OR ad = 'adad' AND aa = ANY(ARRAY[1,2,3])"
		for i := 0; i < b.N; i++ {
			preparePositionalArgsQuery(q)
		}
		b.ReportAllocs()
	})
}

func TestPreparePositionalArgsQuery(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		q := "update apple_attribute set code = 'name_test_update' where id = ? AND ab = ? OR ad = ? AND aa = ANY(ARRAY[1,2,3])"
		r := preparePositionalArgsQuery(q)
		if r != "update apple_attribute set code = 'name_test_update' where id = $1 AND ab = $2 OR ad = $3 AND aa = ANY(ARRAY[1,2,3])" {
			t.Fatal("wrong normal")
		}
	})
	t.Run("serial", func(t *testing.T) {
		q := "?????"
		r := preparePositionalArgsQuery(q)
		if r != "$1$2$3$4$5" {
			t.Fatal("wrong serial")
		}
	})
	t.Run("no_need_transform", func(t *testing.T) {
		q := "update apple_attribute set code = 'name_test_update' where id = 1 AND ab = '2' OR ad = 'adad' AND aa = ANY(ARRAY[1,2,3])"
		r := preparePositionalArgsQuery(q)
		if r != q {
			t.Fatal("wrong no_need_transform")
		}
	})
	t.Run("max_args", func(t *testing.T) {
		q := "INSERT INTO test_table (id, name, date, value, count, type_id, created_at, updated_at) VALUES "
		for i := 0; i < 255*255; {
			for j := 0; j < 8; j++ {
				if i == 0 {
					q += "(?, ?, ?, ?, ?, ?, ?, ?)"
				} else {
					q += ", (?, ?, ?, ?, ?, ?, ?, ?)"
				}
				i++
			}
		}
		r := preparePositionalArgsQuery(q)
		if r[len(r)-1] != ')' {
			t.Fatal("wrong max_args")
		}
	})
}

func BenchmarkSimpleQuery(b *testing.B) {
	db, _ := initDb()
	db.Debug = true
	var tableName string
	for i := 0; i < b.N; i++ {
		err := db.QueryRow("select table_name from information_schema.tables limit 1").Scan(&tableName)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}
