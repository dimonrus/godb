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

// Database Object Connection Interface
type Connection interface {
	String() string
	GetDbType() string
	GetMaxConnection() int
	GetConnMaxLifetime() int
	GetMaxIdleConns() int
}

// Migration file interface
type IMigrationFile interface {
	Up(tx *SqlTx) error
	Down(tx *SqlTx) error
	GetVersion() string
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

// Main Database Object
type DBO struct {
	*sql.DB
	Options
	Connection Connection
}

// Transaction
type SqlTx struct {
	m sync.Mutex
	*sql.Tx
	Options
	transaction *Transaction
}

// Stmt
type SqlStmt struct {
	m sync.Mutex
	*sql.Stmt
	Options
	query string
}

// MigrationRegistry migration registry
type MigrationRegistry map[string][]IMigrationFile

// Migration struct
type Migration struct {
	MigrationPath string
	DBO           *DBO
	Config        ConnectionConfig
	Registry      MigrationRegistry
	RegistryPath  string
	RegistryXPath string
}

// TransactionId transaction identifier
type TransactionId string

// TransactionPool transaction pool
type TransactionPool struct {
	transactions map[TransactionId]*SqlTx
	m            sync.RWMutex
}

// Transaction params
type Transaction struct {
	// Time to live in unix timestampt
	// 0 - no TTL for transaction
	TTL int
	// Event on transaction done
	done chan struct{}
}
