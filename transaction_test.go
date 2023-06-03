package godb

import (
	"database/sql"
	"fmt"
	"github.com/dimonrus/gocli"
	// _ "github.com/lib/pq"
	"sync"
	"testing"
	"time"
)

type postgresDockerConnection struct{}

func (c *postgresDockerConnection) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"0.0.0.0", 5432, "migrate", "migrate", "migrate")
}

func (c *postgresDockerConnection) GetDbType() string {
	return "postgres"
}

func (c *postgresDockerConnection) GetMaxConnection() int {
	return 200
}

func (c *postgresDockerConnection) GetMaxIdleConns() int {
	return 15
}

func (c *postgresDockerConnection) GetConnMaxLifetime() int {
	return 50
}

func initDb() (*DBO, error) {
	return DBO{
		Options: Options{
			Debug:          true,
			QueryProcessor: PreparePositionalArgsQuery,
			Logger:         gocli.NewLogger(gocli.LoggerConfig{}),
		},
		Connection: &postgresDockerConnection{},
	}.Init()
}

func TestAsyncTransaction(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	db.TransactionTTL = 4
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	var ver string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = tx.QueryRow("select version();").Scan(&ver)
		if err != nil {
			t.Log(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stm, err := tx.Prepare("select current_date;")
		if err != nil {
			t.Log(err)
			return
		}
		_, err = stm.Exec()
		if err != nil {
			t.Log(err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		stm, err := tx.Prepare("select * FROM pg_stat_activity;")
		if err != nil {
			t.Log(err)
			return
		}
		_, err = stm.Exec()
		if err != nil {
			t.Log(err)
			return
		}
	}()
	err = tx.QueryRow("select version();").Scan(&ver)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
}

func initNativePQ() (*sql.DB, error) {
	conn := &postgresDockerConnection{}
	dbo, err := sql.Open(conn.GetDbType(), conn.String())
	if err != nil {
		return nil, err
	}
	// Ping db
	err = dbo.Ping()
	if err != nil {
		return nil, err
	}
	// Set connection options
	dbo.SetMaxIdleConns(conn.GetMaxIdleConns())
	dbo.SetConnMaxLifetime(time.Second * time.Duration(conn.GetConnMaxLifetime()))
	dbo.SetMaxOpenConns(conn.GetMaxConnection())
	return dbo, nil
}

// BenchmarkNativePq-4   	      92	  11482190 ns/op	    1346 B/op	      20 allocs/op
func BenchmarkNativePq(b *testing.B) {
	db, err := initNativePQ()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		tx, err := db.Begin()
		if err != nil {
			b.Fatal(err)
		}
		_, err = tx.Exec("select * from pg_stat_activity;")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tx.Exec("select version();")
		if err != nil {
			b.Fatal(err)
		}
		err = tx.Rollback()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}

// oos: darwin
// goarch: arm64
// pkg: github.com/dimonrus/godb/v2
// BenchmarkTransaction
// BenchmarkTransaction-12    	     696	   1691994 ns/op	    1632 B/op	      24 allocs/op
func BenchmarkTransaction(b *testing.B) {
	db, err := initDb()
	if err != nil {
		b.Fatal(err)
	}
	db.Options.Debug = false
	db.TransactionTTL = 4
	for i := 0; i < b.N; i++ {
		tx, err := db.Begin()
		if err != nil {
			b.Fatal(err)
		}
		_, err = tx.Exec("select * from pg_stat_activity;")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tx.Exec("select version();")
		if err != nil {
			b.Fatal(err)
		}
		err = tx.Rollback()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}

func TestTransactionPool(t *testing.T) {
	pool := NewTransactionPool()
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	db.TransactionTTL = 3
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	tx1 := GenTransactionId()
	pool.Set(tx1, tx)
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	tx2 := GenTransactionId()
	pool.Set(tx2, tx)

	if pool.Count() != 2 {
		t.Fatal("pool have to contain 2 transaction")
	}

	time.Sleep(time.Second * 4)

	tx = pool.Get(tx2)
	if tx != nil {
		_, err = tx.Exec("select version();")
		if err == sql.ErrTxDone {
			pool.UnSet(tx2)
		} else {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 3)
	if pool.Count() != 1 {
		t.Fatal("pool have to contain 1 transaction")
	}
}
