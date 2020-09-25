package godb

import (
	"github.com/dimonrus/gohelp"
	"strconv"
	"testing"
)

func testCreateAppleAttributeTable(q Queryer) error {
	_, err := q.Exec("CREATE TABLE IF NOT EXISTS apple_attribute (id serial not null primary key, code text not null, created_at TIMESTAMP NOT NULL DEFAULT localtimestamp);")
	return err
}

func testDropAppleAttributeTable(q Queryer) error {
	_, err := q.Exec("DROP TABLE IF EXISTS apple_attribute")
	return err
}

func testCreateApplePropertyTable(q Queryer) error {
	_, err := q.Exec("CREATE TABLE IF NOT EXISTS apple_property (id serial not null primary key, name text not null, attribute_id int not null references apple_attribute(id));")
	return err
}

func deleteApplePropertyTable(q Queryer) error {
	_, err := q.Exec("DROP TABLE IF EXISTS apple_property")
	return err
}

func testFillData(q Queryer) error {
	qa := "INSERT INTO apple_attribute (code) VALUES "
	var semi string
	for i := 0; i < 100_000; i++ {
		if i < 99999 {
			semi = ","
		} else {
			semi = ""
		}
		qa += "('" + gohelp.RandString(8) + "')" + semi
	}
	_, err := q.Exec(qa)
	if err != nil {
		return err
	}

	qa = "INSERT INTO apple_property (name, attribute_id) VALUES "
	for i := 0; i < 400_000; i++ {
		if i < 399999 {
			semi = ","
		} else {
			semi = ""
		}
		qa += "('" + gohelp.RandString(8) + "'," + strconv.Itoa(gohelp.GetRndNumber(1, 99999)) + ")" + semi
	}
	_, err = q.Exec(qa)
	return err
}

func testInitTestTables(q Queryer) error {
	db, err := initDb()
	if err != nil {
		return err
	}
	err = testCreateAppleAttributeTable(db)
	if err != nil {
		return err
	}
	err = testCreateApplePropertyTable(db)
	if err != nil {
		return err
	}
	//return testFillData(db)
	return nil
}


func NATestScanRows(t *testing.T) {
	_, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	//m1 := NewProfileData()
	//m2 := NewProfileAgent()
	//
	////q := `SELECT aa.*, ap.* FROM apple_attribute aa LEFT JOIN apple_property ap ON ap.attribute_id = aa.id LIMIT 100`
	//q := `SELECT pd.*, pa.* FROM profile_data pd LEFT JOIN profile_agent pa ON pd.id = pa.profile_data_id WHERE pa.id IS NOT NULL LIMIT 100`
	//rows, err := db.Query(q)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//vals := append(m1.Values(), m2.Values()...)
	//for rows.Next() {
	//	err := rows.Scan(vals...)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//}
	//
	//pd := NewProfileDataCollection()
	//pd.SetPagination(100, 100)
	//e := pd.Load(db)
	//if e != nil {
	//	t.Fatal(e)
	//}
	//
	//for pd.Next() {
	//	fmt.Println(*pd.Item().Id)
	//}
	//
	//_ = m1
	//_ = m2
}
