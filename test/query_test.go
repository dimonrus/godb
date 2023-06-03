package test

import (
	"fmt"
	"github.com/dimonrus/godb/v2"
)

const testTable = "users"

type QueryTestCase struct {
	// Create table
	create string
	// drop table
	drop string
	// insert query
	insert []string
	// prepare statement
	prepare string
	// items for prepared statement
	items [][]any
	// select query
	fetch string
	// update query
	update string
	// delete query
	delete string
	// scan valued
	scan []any
}

// CreateTable
func (c QueryTestCase) CreateTable(q godb.Queryer) error {
	_, err := q.Exec(c.create)
	return err
}

// DropTable
func (c QueryTestCase) DropTable(q godb.Queryer) error {
	_, err := q.Exec(c.drop)
	return err
}

// Insert
func (c QueryTestCase) Insert(q godb.Queryer) error {
	for _, s := range c.insert {
		_, err := q.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

// Prepare
func (c QueryTestCase) Prepare(db *godb.SqlTx) (*godb.SqlStmt, error) {
	return db.Prepare(c.prepare)
}

// ExecStatement
func (c QueryTestCase) ExecStatement(stmt *godb.SqlStmt) error {
	for i := range c.items {
		_, err := stmt.Exec(c.items[i]...)
		if err != nil {
			return err
		}
	}
	return stmt.Close()
}

// Update
func (c QueryTestCase) Update(q godb.Queryer) error {
	_, err := q.Exec(c.update)
	return err
}

// Select
func (c QueryTestCase) Select(q godb.Queryer) error {
	rows, err := q.Query(c.fetch)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(c.scan...)
		if err != nil {
			return err
		}
		fmt.Println(c.scan...)
	}
	return nil
}

// Delete
func (c QueryTestCase) Delete(q godb.Queryer) error {
	_, err := q.Exec(c.delete)
	return err
}
