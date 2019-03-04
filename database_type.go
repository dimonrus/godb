package godb

import "database/sql"

// Logger
type Logger interface {
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Database Object Connection Interface
type Connection interface {
	String() string
	GetDbType() string
	GetMaxConnection() int
	GetConnectionIdleLifetime() int
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
