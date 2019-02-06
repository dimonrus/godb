package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	_ "github.com/lib/pq"
)

type connection struct {}
type wrongConnection struct{connection}
type testData struct {
	Id int
	Code string
}

type testDatas []testData

func (c *wrongConnection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5431, "postgres", "root", "goreav")
}
func (c *wrongConnection) GetDbType() string {
	return "unknown"
}
func (c *connection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "root", "goreav")
}
func (c *connection) GetDbType() string {
	return "postgres"
}
func (c *connection) GetMaxConnection() int {
	return 50
}
func (c *connection) GetConnectionIdleLifetime() int {
	return 15
}

func TestDBO_InitError(t *testing.T) {
	_, err := DBO{
		Options: Options{
			Debug:  true,
			Logger: log.New(os.Stdout, "test:", log.Ldate|log.Ltime|log.Lshortfile),
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
			Logger: log.New(os.Stdout, "test:", log.Ldate|log.Ltime|log.Lshortfile),
		},
		Connection: &connection{},
	}.Init()
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
	db, _ := initDb()
	rows, err := db.Query("select id, code from apple_attribute")
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
