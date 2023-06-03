package godb

import "testing"

// BenchmarkPreparePos/normal-8             5087773               230.5 ns/op           288 B/op          2 allocs/op
func BenchmarkPreparePos(b *testing.B) {
	b.Run("normal", func(b *testing.B) {
		q := "update apple_attribute set code = 'name_test_update' where id = ? AND ab = ? OR ad = ? AND aa = ANY(?)"
		for i := 0; i < b.N; i++ {
			PreparePositionalArgsQuery(q)
		}
		b.ReportAllocs()
	})
}

func TestPreparePositionalArgsQuery(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		q := "update apple_attribute set code = 'name_test_update' where id = ? AND ab = ? OR ad = ? AND aa = ANY(ARRAY[1,2,3])"
		r := PreparePositionalArgsQuery(q)
		t.Log(r)
		if r != "update apple_attribute set code = 'name_test_update' where id = $1 AND ab = $2 OR ad = $3 AND aa = ANY(ARRAY[1,2,3])" {
			t.Fatal("wrong normal")
		}
	})
	t.Run("serial", func(t *testing.T) {
		q := "?????"
		r := PreparePositionalArgsQuery(q)
		if r != "$1$2$3$4$5" {
			t.Fatal("wrong serial")
		}
	})
	t.Run("no_need_transform", func(t *testing.T) {
		q := "update apple_attribute set code = 'name_test_update' where id = 1 AND ab = '2' OR ad = 'adad' AND aa = ANY(ARRAY[1,2,3])"
		r := PreparePositionalArgsQuery(q)
		if r != q {
			t.Fatal("wrong no_need_transform")
		}
	})
	t.Run("max_args", func(t *testing.T) {
		q := "INSERT INTO test_table (id, name, date, value, count, type_id, created_at, updated_at) VALUES "
		for i := 0; i < 255*255; {
			for j := 0; j < 8; j++ {
				if i == 0 {
					q += "(?, ?, ?, ?, ?, ?, ?, ?)"
				} else {
					q += ", (?, ?, ?, ?, ?, ?, ?, ?)"
				}
				i++
			}
		}
		r := PreparePositionalArgsQuery(q)
		if r[len(r)-1] != ')' {
			t.Fatal("wrong max_args")
		}
	})
}
