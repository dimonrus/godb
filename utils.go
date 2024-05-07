package godb

import (
	"database/sql"
	"fmt"
	"github.com/dimonrus/gocli"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// init logger callback
var logger = func(lg gocli.Logger, message chan string) {
	for s := range message {
		lg.Println(s)
	}
}

// cache for positional args
var positionalArgs = prepareAllPositions()

// Init position args from $1 to $65535
func prepareAllPositions() []string {
	result := make([]string, 1<<16)
	for i := int64(1); i < 1<<16; i++ {
		result[i] = "$" + strconv.FormatInt(i, 10)
	}
	return result
}

// IsTableExists check if table exists
func IsTableExists(q Queryer, table, schema string) bool {
	query := fmt.Sprintf(`SELECT table_name FROM information_schema.tables WHERE table_name = '%s'`, table)
	if schema != "" {
		query += fmt.Sprintf(" AND table_schema = '%s'", schema)
	}
	if c, ok := q.(IConnType); ok {
		if c.ConnType() == "sqlite3" {
			query = fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s'", table)
		}
	}
	var tableName *string
	err := q.QueryRow(query).Scan(&tableName)
	if err != nil {
		if o, ok := q.(IOptions); ok {
			o.GetLogger().Errorln(err.Error())
		}
	}
	return err == nil && tableName != nil && *tableName == table
}

// PreparePositionalArgsQuery Position argument
func PreparePositionalArgsQuery(query string) string {
	if !strings.Contains(query, "?") {
		return query
	}
	var ll = len(query)
	var b = make([]byte, ll*2)
	var j int64 = 1
	var i, k, l, s int
	for i < len(query) {
		if query[i] == '?' {
			if i+1 < ll && query[i+1] == '?' {
				i++
				l += len(query[s:i])
				if l > len(b) {
					ll = ll * 2
					b = append(b, make([]byte, ll)...)
				}
				copy(b[k:l], query[s:i])
			} else {
				l += len(query[s:i])
				if l > len(b) {
					ll = ll * 2
					b = append(b, make([]byte, ll)...)
				}
				copy(b[k:l], query[s:i])
				k = l
				l += len(positionalArgs[j])
				if l > len(b) {
					ll = ll * 2
					b = append(b, make([]byte, ll)...)
				}
				copy(b[k:l], positionalArgs[j])
				j++
			}
			s = i + 1
			k = l
		}
		i++
	}
	if i > s {
		p := query[s:]
		l += len(p)
		copy(b[k:l], p)
	}
	b = b[:l]
	return *(*string)(unsafe.Pointer(&b))
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
