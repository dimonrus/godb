GODB

The database wrapper for manage postgres db

Init connection example

```
import _ "github.com/lib/pq"

connectionConfig := godb.PostgresConnectionConfig{...}

dbo, err := godb.DBO{
    Options: godb.Options{
        Debug:  true,
        Logger: App.GetLogger(),
    },
    Connection: &connectionConfig,
}.Init()

```

Transaction functions

```
func StartTransaction() *godb.SqlTx {
	tx, err := dbo.Begin()
	if err != nil {
		panic(err)
		return nil
	}
	return tx
}

func EndTransaction(q *godb.SqlTx, e porterr.IError) {
	var err error
	if e != nil {
		err = q.Rollback()
	} else {
		err = q.Commit()
	}
	if err != nil {
		panic(err)
	}
	return
}

// usage

tx := StartTransaction()
defer func() { EndTransaction(tx, e) }()

```