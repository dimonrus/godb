package godb

import (
	"database/sql"
	"sync"
	"time"
)

// Logger
type Logger interface {
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

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

// Database Object Options
type Options struct {
	Debug          bool
	Logger         Logger
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

// Migration registry
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

// Transaction identifier
type TransactionId string

// Transaction pool
type TransactionPool struct {
	transactions map[TransactionId]*SqlTx
	m sync.RWMutex
}

// Transaction params
type Transaction struct {
	// Time to live in unix timestampt
	// 0 - no TTL for transaction
	TTL int
	// Event on transaction done
	done chan bool
}
