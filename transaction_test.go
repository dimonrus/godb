package godb

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
	"testing"
	"time"
)

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
	fmt.Print(tx.String())
}

func initNativePQ() (*sql.DB, error) {
	cs := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"192.168.1.110", 5433, "gasap", "gasap", "gasap")
	dbo, err := sql.Open("postgres", cs)
	if err != nil {
		return nil, err
	}
	// Ping db
	err = dbo.Ping()
	if err != nil {
		return nil, err
	}
	// Set connection options
	dbo.SetMaxIdleConns(15)
	dbo.SetConnMaxLifetime(time.Second * 15)
	dbo.SetMaxOpenConns(50)
	return dbo, nil
}

//  18	 611418573 ns/op	    2159 B/op	      28 allocs/op
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
// 18	 622394080 ns/op	    8180 B/op	      38 allocs/op
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
