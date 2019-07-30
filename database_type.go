package godb

import "database/sql"

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
	Debug  bool
	Logger Logger
}

// Main Database Object
type DBO struct {
	*sql.DB
	Options
	Connection Connection
}

// Transaction
type SqlTx struct {
	*sql.Tx
	Options
}

// Stmt
type SqlStmt struct {
	*sql.Stmt
	Options
	query string
}

// Migration struct
type Migration struct {
	RegistryPath  string
	DBO           *DBO
	Config        ConnectionConfig
	Registry      *map[string][]IMigrationFile
	RegistryXPath string
}
