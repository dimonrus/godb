package godb

import (
	"database/sql"
	"github.com/dimonrus/porterr"
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

// DB model interface
type IModel interface {
	Table() string
	Columns() []string
	Values() []interface{}
	Load(q Queryer) porterr.IError
	Save(q Queryer) porterr.IError
	Delete(q Queryer) porterr.IError
}

// DB model interface
type ISoftModel interface {
	IModel
	SoftLoad(q Queryer) porterr.IError
	SoftDelete(q Queryer) porterr.IError
	SoftRecover(q Queryer) porterr.IError
}
