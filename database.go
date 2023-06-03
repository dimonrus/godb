package godb

import (
	"context"
	"database/sql"
	"time"
)

// Init Database Object
func (dbo DBO) Init() (*DBO, error) {
	db, err := getDb(dbo.Connection)
	if err != nil {
		return &dbo, err
	}
	dbo.DB = db
	dbo.logMessage = make(chan string, dbo.Connection.GetMaxConnection())
	go logger(dbo.Logger, dbo.logMessage)
	return &dbo, nil
}

// ConnType get connection type
func (dbo *DBO) ConnType() string {
	return dbo.Connection.GetDbType()
}

// Query SQL exec query
func (dbo *DBO) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if dbo.Options.QueryProcessor != nil {
		query = dbo.Options.QueryProcessor(query)
	}
	if dbo.Debug == true {
		dbo.logMessage <- query
	}
	return dbo.DB.QueryContext(context.Background(), query, args...)
}

// Exec SQL run query
func (dbo *DBO) Exec(query string, args ...interface{}) (sql.Result, error) {
	if dbo.Options.QueryProcessor != nil {
		query = dbo.Options.QueryProcessor(query)
	}
	if dbo.Debug == true {
		dbo.logMessage <- query
	}
	return dbo.DB.ExecContext(context.Background(), query, args...)
}

// QueryRow SQL query row
func (dbo *DBO) QueryRow(query string, args ...interface{}) *sql.Row {
	if dbo.Options.QueryProcessor != nil {
		query = dbo.Options.QueryProcessor(query)
	}
	if dbo.Debug == true {
		dbo.logMessage <- query
	}
	return dbo.DB.QueryRowContext(context.Background(), query, args...)
}

// Prepare statement
func (dbo *DBO) Prepare(query string) (*SqlStmt, error) {
	if dbo.Options.QueryProcessor != nil {
		query = dbo.Options.QueryProcessor(query)
	}
	stmt, err := dbo.DB.PrepareContext(context.Background(), query)
	return &SqlStmt{Stmt: stmt, Options: dbo.Options, query: query}, err
}

// Begin transaction
func (dbo *DBO) Begin() (*SqlTx, error) {
	tx, err := dbo.DB.BeginTx(context.Background(), nil)
	stx := &SqlTx{
		Tx:      tx,
		Options: dbo.Options,
		transaction: &Transaction{
			TTL: int(dbo.Options.TransactionTTL),
		},
		Connection: dbo.Connection,
	}
	stx.delayedRollback()
	return stx, err
}

// Delayed rollback
func (tx *SqlTx) delayedRollback() {
	if tx.transaction != nil && tx.transaction.TTL > 0 {
		go func() {
			timer := time.After(time.Duration(tx.transaction.TTL) * time.Second)
			tx.transaction.done = make(chan struct{})
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

// Commit
func (tx *SqlTx) commit() error {
	// Commit
	return tx.Tx.Commit()
}

// Commit transaction
func (tx *SqlTx) Commit() error {
	// Stop timer
	if tx.transaction.done != nil {
		tx.transaction.done <- struct{}{}
	}
	// Commit
	return tx.commit()
}

func (tx *SqlTx) rollback() error {
	// Rollback
	return tx.Tx.Rollback()
}

// Rollback transaction
func (tx *SqlTx) Rollback() error {
	// Stop timer
	if tx.transaction.done != nil {
		tx.transaction.done <- struct{}{}
	}
	// rollback
	return tx.rollback()
}

// Prepare Stmt
func (tx *SqlTx) Prepare(query string) (*SqlStmt, error) {
	tx.m.Lock()
	defer tx.m.Unlock()
	if tx.Options.QueryProcessor != nil {
		query = tx.Options.QueryProcessor(query)
	}
	stmt, err := tx.PrepareContext(context.Background(), query)
	return &SqlStmt{Stmt: stmt, Options: tx.Options, query: query}, err
}

// Stmt Get Stmt
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
	if tx.Options.QueryProcessor != nil {
		query = tx.Options.QueryProcessor(query)
	}
	if tx.Debug == true {
		tx.logMessage <- query
	}
	return tx.Tx.ExecContext(context.Background(), query, args...)
}

// Query Transaction
func (tx *SqlTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx.m.Lock()
	defer tx.m.Unlock()
	if tx.Options.QueryProcessor != nil {
		query = tx.Options.QueryProcessor(query)
	}
	if tx.Debug == true {
		tx.logMessage <- query
	}
	return tx.Tx.QueryContext(context.Background(), query, args...)
}

// QueryRow Query Row transaction
func (tx *SqlTx) QueryRow(query string, args ...interface{}) *sql.Row {
	tx.m.Lock()
	defer tx.m.Unlock()
	if tx.Options.QueryProcessor != nil {
		query = tx.Options.QueryProcessor(query)
	}
	if tx.Debug == true {
		tx.logMessage <- query
	}
	return tx.Tx.QueryRowContext(context.Background(), query, args...)
}

// ConnType get connection type
func (tx *SqlTx) ConnType() string {
	return tx.Connection.GetDbType()
}

// Exec Stmt Exec
func (st *SqlStmt) Exec(args ...interface{}) (sql.Result, error) {
	st.m.Lock()
	defer st.m.Unlock()
	if st.Options.QueryProcessor != nil {
		st.query = st.Options.QueryProcessor(st.query)
	}
	if st.Debug == true {
		st.logMessage <- st.query
	}
	return st.Stmt.ExecContext(context.Background(), args...)
}

// Query Stmt Query
func (st *SqlStmt) Query(args ...interface{}) (*sql.Rows, error) {
	st.m.Lock()
	defer st.m.Unlock()
	if st.Options.QueryProcessor != nil {
		st.query = st.Options.QueryProcessor(st.query)
	}
	if st.Debug == true {
		st.logMessage <- st.query
	}
	return st.Stmt.QueryContext(context.Background(), args...)
}

// QueryRow Stmt Query Row
func (st *SqlStmt) QueryRow(args ...interface{}) *sql.Row {
	st.m.Lock()
	defer st.m.Unlock()
	if st.Options.QueryProcessor != nil {
		st.query = st.Options.QueryProcessor(st.query)
	}
	if st.Debug == true {
		st.logMessage <- st.query
	}
	return st.Stmt.QueryRowContext(context.Background(), args...)
}
