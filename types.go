package godb

import (
	"database/sql"
	"github.com/dimonrus/gocli"
	"sync"
	"time"
)

// Connection Database Object Connection Interface
type Connection interface {
	// String Get prepared URI
	String() string
	// GetDbType return database type
	GetDbType() string
	// GetMaxConnection return maximum connection
	GetMaxConnection() int
	// GetConnMaxLifetime return maximum lifetime
	GetConnMaxLifetime() int
	// GetMaxIdleConns return maximum idle connections count
	GetMaxIdleConns() int
}

// Queryer interface
type Queryer interface {
	// Exec query
	Exec(query string, args ...interface{}) (sql.Result, error)
	// Prepare statement
	Prepare(query string) (*SqlStmt, error)
	// Query rows
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow single row
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Options database object options
type Options struct {
	// Debug mode shows logs
	Debug bool
	// Logger
	Logger gocli.Logger
	// log message data
	logMessage chan string
	// Query preprocessing
	QueryProcessor func(query string) string
	// TTL for transaction
	TransactionTTL time.Duration `yaml:"transactionTTL"`
}

// IOptions interface helps to get logger
type IOptions interface {
	// GetLogger return logger
	GetLogger() gocli.Logger
}

// GetLogger return logger
func (o *Options) GetLogger() gocli.Logger {
	return o.Logger
}

// IConnType interface helps to get connection type
type IConnType interface {
	// ConnType return connection type
	ConnType() string
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
	Connection  Connection
}

// SqlStmt Statement object
type SqlStmt struct {
	m sync.Mutex
	*sql.Stmt
	Options
	query string
}
