package godb

import (
	"database/sql"
	"github.com/dimonrus/gocli"
	"sync"
	"time"
)

// Queryer interface
type Queryer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*SqlStmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Connection Database Object Connection Interface
type Connection interface {
	String() string
	GetDbType() string
	GetMaxConnection() int
	GetConnMaxLifetime() int
	GetMaxIdleConns() int
}

// Options database object options
type Options struct {
	// Debug mode shows logs
	Debug bool
	// Logger
	Logger gocli.Logger
	// log message data
	logMessage chan string
	// TTL for transaction
	TransactionTTL time.Duration `yaml:"transactionTTL"`
}

// DBO Main Database Object
type DBO struct {
	*sql.DB
	Options
	Connection Connection
}

// SqlTx Transaction object
type SqlTx struct {
	m sync.Mutex
	*sql.Tx
	Options
	transaction *Transaction
}

// SqlStmt Statement object
type SqlStmt struct {
	m sync.Mutex
	*sql.Stmt
	Options
	query string
}
