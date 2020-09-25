package godb

import (
	"testing"
)

func TestMakeModel2(t *testing.T) {
	db, err := initDb()
	if err != nil {
		t.Fatal(err)
	}
	err = MakeModel(db, "models", "public", "profile_agent", "modelx.tmpl", DefaultSystemColumnsSoft)
	if err != nil {
		t.Fatal(err)
	}
	err = MakeModel(db, "models", "public", "profile_data", "modelx.tmpl", DefaultSystemColumnsSoft)
	if err != nil {
		t.Fatal(err)
	}
}