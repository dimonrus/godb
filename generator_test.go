package godb

import (
	"testing"
)

func TestMakeModel2(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	err = MakeModel(db, "models", "public", "login", "11modelx.tmpl", DefaultSystemColumnsSoft)
	if err != nil {
		t.Fatal(err)
	}
	err = MakeModel(db, "models", "public", "reset_password", "11modelx.tmpl", DefaultSystemColumnsSoft)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateDictionaryMapping(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	err = GenerateDictionaryMapping("models/dictionary_mapping.go", db)
	if err != nil {
		t.Fatal(err)
	}
}
