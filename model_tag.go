package godb

import "strings"

// ModelFiledTag All possible model field tag properties
// tag must have 3 symbol length
type ModelFiledTag struct {
	// DB column name
	Column string `tag:"col"`
	// Foreign key definition
	ForeignKey string `tag:"frk"`
	// Has sequence
	IsSequence bool `tag:"seq"`
	// Is primary key
	IsPrimaryKey bool `tag:"prk"`
	// Is not null
	IsRequired bool `tag:"req"`
	// Is system like as created_at, updated_at, deleted_at
	IsSystem bool `tag:"sys"`
	// Is unique
	IsUnique bool `tag:"unq"`
}

// Prepare string tag
func (t ModelFiledTag) String() string {
	b := strings.Builder{}
	if t.Column != "" {
		b.WriteString("col~" + t.Column + ";")
	}
	if t.ForeignKey != "" {
		b.WriteString("frk~" + t.ForeignKey + ";")
	}
	if t.IsSequence {
		b.WriteString("seq;")
	}
	if t.IsPrimaryKey {
		b.WriteString("prk;")
	}
	if t.IsRequired {
		b.WriteString("req;")
	}
	if t.IsSystem {
		b.WriteString("sys;")
	}
	if t.IsUnique {
		b.WriteString("unq;")
	}
	return b.String()
}

// ParseModelFiledTag parse validation tag for rule and arguments
// Example
// db:"col~created_at;seq;sys;prk;frk~master.table(id,name);req;unq'"
func ParseModelFiledTag(tag string) (field ModelFiledTag) {
	if tag == "" || len(tag) < 3 {
		return
	}
	var indexStart, i int
	for i < len(tag) {
		if tag[i] == ';' || i == len(tag)-1 {
			switch tag[indexStart : indexStart+3] {
			case "seq":
				field.IsSequence = true
				i++
				indexStart = i
			case "prk":
				field.IsPrimaryKey = true
				i++
				indexStart = i
			case "req":
				field.IsRequired = true
				i++
				indexStart = i
			case "sys":
				field.IsSystem = true
				i++
				indexStart = i
			case "unq":
				field.IsUnique = true
				i++
				indexStart = i
			case "col":
				// Must be ~ according to format
				if tag[indexStart+3] != '~' {
					break
				}
				field.Column = tag[indexStart+4 : i]
				i++
				indexStart = i
			case "frk":
				// Must be ~ according to format
				if tag[indexStart+3] != '~' {
					break
				}
				field.ForeignKey = tag[indexStart+4 : i]
				i++
				indexStart = i
			}
		}
		i++
	}
	return
}
