package godb

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// Init Database Object
func (dbo DBO) Init() (*DBO, error) {
	db, err := getDb(dbo.Connection)
	if err != nil {
		return &dbo, err
	}
	dbo.DB = db

	return &dbo, nil
}

// Get Db Instance
func getDb(connection Connection) (*sql.DB, error) {
	//Open connection
	dbo, err := sql.Open(connection.GetDbType(), connection.String())
	if err != nil {
		return nil, err
	}
	// Ping db
	err = dbo.Ping()
	if err != nil {
		return nil, err
	}
	// Set connection options
	dbo.SetMaxIdleConns(connection.GetMaxIdleConns())
	dbo.SetConnMaxLifetime(time.Second * time.Duration(connection.GetConnMaxLifetime()))
	dbo.SetMaxOpenConns(connection.GetMaxConnection())

	return dbo, nil
}

// SQL Query
func (dbo *DBO) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if dbo.Debug == true {
		go logDbQuery(dbo.Logger, query, args...)
	}
	return dbo.DB.QueryContext(context.Background(), query, args...)
}

// SQL Exec
func (dbo *DBO) Exec(query string, args ...interface{}) (sql.Result, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if dbo.Debug == true {
		go logDbQuery(dbo.Logger, query, args...)
	}
	return dbo.DB.ExecContext(context.Background(), query, args...)
}

// SQL Query Row
func (dbo *DBO) QueryRow(query string, args ...interface{}) *sql.Row {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if dbo.Debug == true {
		go logDbQuery(dbo.Logger, query, args...)
	}
	return dbo.DB.QueryRowContext(context.Background(), query, args...)
}

// Begin transaction
func (dbo *DBO) Begin() (*SqlTx, error) {
	tx, err := dbo.DB.BeginTx(context.Background(), nil)
	return &SqlTx{tx, dbo.Options}, err
}

// Prepare Stmt
func (tx *SqlTx) Prepare(query string) (*SqlStmt, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	stmt, err := tx.PrepareContext(context.Background(), query)
	return &SqlStmt{stmt, tx.Options, query}, err
}

// Get Stmt
func (tx *SqlTx) Stmt(stmt *SqlStmt) *SqlStmt {
	stm := tx.StmtContext(context.Background(), stmt.Stmt)
	return &SqlStmt{stm, tx.Options, stmt.query}
}

// Exec Transaction
func (tx *SqlTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if tx.Debug == true {
		go logDbQuery(tx.Logger, query, args...)
	}
	return tx.Tx.ExecContext(context.Background(), query, args...)
}

// Query Transaction
func (tx *SqlTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if tx.Debug == true {
		go logDbQuery(tx.Logger, query, args...)
	}
	return tx.Tx.QueryContext(context.Background(), query, args...)
}

// Query Row Transaction
func (tx *SqlTx) QueryRow(query string, args ...interface{}) *sql.Row {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	if tx.Debug == true {
		go logDbQuery(tx.Logger, query, args...)
	}
	return tx.Tx.QueryRowContext(context.Background(), query, args...)
}

// Stmt Exec
func (st *SqlStmt) Exec(args ...interface{}) (sql.Result, error) {
	if strings.Contains(st.query, "?") {
		st.query = preparePositionalArgsQuery(st.query)
	}
	if st.Debug == true {
		go logDbQuery(st.Logger, st.query, args...)
	}

	return st.Stmt.ExecContext(context.Background(), args...)
}

// Stmt Query
func (st *SqlStmt) Query(args ...interface{}) (*sql.Rows, error) {
	if strings.Contains(st.query, "?") {
		st.query = preparePositionalArgsQuery(st.query)
	}
	if st.Debug == true {
		go logDbQuery(st.Logger, st.query, args...)
	}
	return st.Stmt.QueryContext(context.Background(), args...)
}

// Stmt Query Row
func (st *SqlStmt) QueryRow(args ...interface{}) *sql.Row {
	if strings.Contains(st.query, "?") {
		st.query = preparePositionalArgsQuery(st.query)
	}
	if st.Debug == true {
		go logDbQuery(st.Logger, st.query, args...)
	}
	return st.Stmt.QueryRowContext(context.Background(), args...)
}

// Position argument
func preparePositionalArgsQuery(query string) string {
	parts := strings.Split(query, "?")
	length := len(parts) - 1
	for index, _ := range parts {
		if index < length {
			parts[index] += "$" + strconv.Itoa(index+1)
		}
	}
	return strings.Join(parts, "")
}

// Logging query
func logDbQuery(logger Logger, query string, args ...interface{}) {
	queryString := strings.Join(strings.Fields(query), " ")
	logger.Printf("\n %s", queryString)
}
