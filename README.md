# GODB

*The database wrapper for manage postgres db*

## Init connection example

```
import _ "github.com/lib/pq"

connectionConfig := godb.PostgresConnectionConfig{...}

dbo, err := godb.DBO{
    Options: godb.Options{
        Debug:  true,
        QueryProcessor: godb.PreparePositionalArgsQuery,
        Logger: App.GetLogger(),
    },
    Connection: &connectionConfig,
}.Init()

```

## Transaction functions

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

#### If you find this project useful or want to support the author, you can send tokens to any of these wallets
- Bitcoin: bc1qgx5c3n7q26qv0tngculjz0g78u6mzavy2vg3tf
- Ethereum: 0x62812cb089E0df31347ca32A1610019537bbFe0D
- Dogecoin: DET7fbNzZftp4sGRrBehfVRoi97RiPKajV