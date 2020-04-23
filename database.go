package godb

import (
	"context"
	"database/sql"
	"fmt"
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

// Prepare statement
func (dbo *DBO) Prepare(query string) (*SqlStmt, error) {
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	stmt, err := dbo.DB.PrepareContext(context.Background(), query)
	return &SqlStmt{Stmt: stmt, Options: dbo.Options, query: query}, err
}

// Begin transaction
func (dbo *DBO) Begin() (*SqlTx, error) {
	transaction := &Transaction{
		Id:        GenTransactionId(),
		TTL:       int(dbo.Options.TransactionTTL),
		BeginTime: time.Now(),
		done:      make(chan bool),
	}
	tx, err := dbo.DB.BeginTx(context.Background(), nil)
	stx := &SqlTx{Tx: tx, Options: dbo.Options, transaction: transaction}
	stx.delayedRollback()
	return stx, err
}

// Get Transaction id
func (tx *SqlTx) delayedRollback() {
	if tx.transaction != nil && tx.transaction.TTL > 0 {
		go func() {
			timer := time.After(time.Duration(tx.transaction.TTL) * time.Second)
			for {
				select {
				case <-tx.transaction.done:
					close(tx.transaction.done)
					return
				case <-timer:
					err := tx.rollback()
					if err != nil {
						tx.Logger.Println(err)
					}
					return
				}
			}
		}()
	}
	return
}

// Get Transaction id
func (tx *SqlTx) GetId() TransactionId {
	return tx.transaction.Id
}

// Get Transaction Information
func (tx *SqlTx) String() string {
	return fmt.Sprintf("Transaction: %+v", tx.transaction)
}

// Set Transaction id
func (tx *SqlTx) SetId(id TransactionId) {
	tx.transaction.Id = id
	return
}

// Commit
func (tx *SqlTx) commit() error {
	tx.transaction.CommitTime = time.Now()
	// Commit
	err := tx.Tx.Commit()
	if err != nil {
		tx.transaction.Message = new(string)
		*tx.transaction.Message = "Commit error: " + err.Error()
	}
	return err
}

// Commit
func (tx *SqlTx) Commit() error {
	// Stop timer
	tx.transaction.done <- true
	// Commit
	return tx.commit()
}

func (tx *SqlTx) rollback() error {
	tx.transaction.RollbackTime = time.Now()
	// Rollback
	err := tx.Tx.Rollback()
	if err != nil {
		tx.transaction.Message = new(string)
		*tx.transaction.Message = "Rollback error: " + err.Error()
	}
	return err
}

// Rollback
func (tx *SqlTx) Rollback() error {
	// Stop timer
	tx.transaction.done <- true
	// rollback
	return tx.rollback()
}

// Prepare Stmt
func (tx *SqlTx) Prepare(query string) (*SqlStmt, error) {
	tx.m.Lock()
	defer tx.m.Unlock()
	if strings.Contains(query, "?") {
		query = preparePositionalArgsQuery(query)
	}
	stmt, err := tx.PrepareContext(context.Background(), query)
	return &SqlStmt{Stmt: stmt, Options: tx.Options, query: query}, err
}

// Get Stmt
func (tx *SqlTx) Stmt(stmt *SqlStmt) *SqlStmt {
	tx.m.Lock()
	defer tx.m.Unlock()
	stm := tx.StmtContext(context.Background(), stmt.Stmt)
	return &SqlStmt{Stmt: stm, Options: tx.Options, query: stmt.query}
}

// Exec Transaction
func (tx *SqlTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx.m.Lock()
	defer tx.m.Unlock()
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
	tx.m.Lock()
	defer tx.m.Unlock()
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
	tx.m.Lock()
	defer tx.m.Unlock()
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
	st.m.Lock()
	defer st.m.Unlock()
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
	st.m.Lock()
	defer st.m.Unlock()
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
	st.m.Lock()
	defer st.m.Unlock()
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
	for index := range parts {
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
