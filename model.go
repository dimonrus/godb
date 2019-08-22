package godb

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dimonrus/gohelp"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var dbo *DBO

// Column information
type Column struct {
	Name              string  // DB column name
	ModelName         string  // Model name
	Default           *string // DB default value
	IsNullable        bool    // DB is nullable
	DataType          string  // DB column type
	ModelType         string  // Model type
	Schema            string  // DB Schema
	Table             string  // DB table
	Sequence          *string // DB sequence
	ForeignSchema     *string // DB foreign schema name
	ForeignTable      *string // DB foreign table name
	ForeignColumnName *string // DB foreign column name
	Description       *string // DB column description
	IsPrimaryKey      bool    // DB is primary key
	Json              string  // Model Json name
	Import            string  // Model Import custom lib
	IsArray           bool    // Array column
}

// Array of columns
type Columns []Column

// Parse Row
func parseColumnRow(rows *sql.Rows) (*Column, error) {
	column := Column{}
	err := rows.Scan(
		&column.Name,
		&column.DataType,
		&column.IsNullable,
		&column.Schema,
		&column.Table,
		&column.IsPrimaryKey,
		&column.Default,
		&column.Sequence,
		&column.ForeignSchema,
		&column.ForeignTable,
		&column.ForeignColumnName,
		&column.Description,
	)

	if err != nil {
		return nil, err
	}

	return &column, nil
}

// Get table columns from db
func GetTableColumns(schema string, table string) (*Columns, error) {
	query := fmt.Sprintf(`
SELECT a.attname                                                                       AS column_name,
       format_type(a.atttypid, a.atttypmod)                                            AS data_type,
       CASE WHEN a.attnotnull THEN FALSE ELSE TRUE END                                 AS is_nullable,
       s.nspname                                                                       AS schema,
       t.relname                                                                       AS table,
       CASE WHEN max(i.indisprimary::int)::BOOLEAN THEN TRUE ELSE FALSE END            AS is_primary,
       ic.column_default,
       pg_get_serial_sequence(ic.table_schema || '.' || ic.table_name, ic.column_name) AS sequence,
       max(ccu.table_schema)                                                           AS foreign_schema,
       max(ccu.table_name)                                                             AS foreign_table,
       max(ccu.column_name)                                                            AS foreign_column_name,
       col_description(t.oid, ic.ordinal_position)                                     AS description
FROM pg_attribute a
         JOIN pg_class t ON a.attrelid = t.oid
         JOIN pg_namespace s ON t.relnamespace = s.oid
         LEFT JOIN pg_index i ON i.indrelid = a.attrelid AND a.attnum = ANY (i.indkey)
         LEFT JOIN information_schema.columns AS ic
                   ON ic.column_name = a.attname AND ic.table_name = t.relname AND ic.table_schema = s.nspname
         LEFT JOIN information_schema.key_column_usage AS kcu
                   ON kcu.table_name = t.relname AND kcu.column_name = a.attname
         LEFT JOIN information_schema.table_constraints AS tc
                   ON tc.constraint_name = kcu.constraint_name AND tc.constraint_type = 'FOREIGN KEY'
         LEFT JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
WHERE a.attnum > 0
  AND NOT a.attisdropped
  AND s.nspname = '%s'
  AND t.relname = '%s'
GROUP BY a.attname, a.atttypid, a.atttypmod, a.attnotnull, s.nspname, t.relname, ic.column_default,
         ic.table_schema, ic.table_name, ic.column_name, a.attnum, t.oid, ic.ordinal_position
ORDER BY a.attnum;`, schema, table)

	rows, err := dbo.Query(query)
	if err != nil {
		return nil, err
	}

	var columns Columns
	var hasPrimary bool

	for rows.Next() {
		column, err := parseColumnRow(rows)
		if err != nil {
			return nil, err
		}

		name, err := gohelp.ToCamelCase(column.Name, true)
		if err != nil {
			return nil, err
		}

		json, err := gohelp.ToCamelCase(column.Name, false)
		if err != nil {
			return nil, err
		}

		column.ModelName = name
		column.Json = fmt.Sprintf(`%cjson:"%s"%c`, '`', json, '`')

		switch {
		case column.DataType == "bigint":
			column.ModelType = "int64"
		case column.DataType == "integer":
			column.ModelType = "int"
		case column.DataType == "text":
			column.ModelType = "string"
		case column.DataType == "double precision":
			column.ModelType = "float64"
		case column.DataType == "boolean":
			column.ModelType = "bool"
		case column.DataType == "ARRAY":
			column.ModelType = "[]interface{}"
		case column.DataType == "json":
			column.ModelType = "json.RawMessage"
			column.Import = `"encoding/json"`
		case column.DataType == "smallint":
			column.ModelType = "int16"
		case column.DataType == "date":
			column.ModelType = "time.Time"
			column.Import = `"time"`
		case strings.Contains(column.DataType, "character varying"):
			column.ModelType = "string"
		case strings.Contains(column.DataType, "numeric"):
			column.ModelType = "float32"
		case column.DataType == "uuid":
			column.ModelType = "string"
		case column.DataType == "jsonb":
			column.ModelType = "json.RawMessage"
			column.Import = `"encoding/json"`
		case column.DataType == "uuid[]":
			column.ModelType = "[]string"
			column.IsArray = true
			column.Import = `"github.com/lib/pq"`
		case column.DataType == "integer[]":
			column.ModelType = "[]int"
			column.IsArray = true
			column.Import = `"github.com/lib/pq"`
		case column.DataType == "bigint[]":
			column.ModelType = "[]int64"
			column.IsArray = true
			column.Import = `"github.com/lib/pq"`
		case column.DataType == "text[]":
			column.ModelType = "[]string"
			column.IsArray = true
			column.Import = `"github.com/lib/pq"`
		case strings.Contains(column.DataType, "timestamp"):
			column.ModelType = "time.Time"
			column.Import = `"time"`
		default:
			return nil, errors.New(fmt.Sprintf("unknown column type: %s", column.DataType))
		}

		if column.IsNullable && !column.IsArray {
			column.ModelType = "*" + column.ModelType
		}

		if column.IsPrimaryKey == true {
			hasPrimary = true
		}

		columns = append(columns, *column)
	}

	// column named id will be primary if no primary key
	if !hasPrimary {
		for key, column := range columns {
			if column.Name == "id" {
				columns[key].IsPrimaryKey = true
				break
			}
		}
	}

	return &columns, nil
}

// Start script
func MakeModel(db *DBO, path string, schema string, table string) error {
	if table == "" {
		return errors.New("table name is empty")
	}
	dbo = db
	return CreateModel(schema, table, path)
}

// Create file in os
func CreateModelFile(schema string, table string, path string) (*os.File, string, error) {
	fileName := fmt.Sprintf("%s", table)
	var filePath string
	if path != "" {
		folderPath := fmt.Sprintf(path)
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return nil, "", err
		}
		filePath = fmt.Sprintf("%s/%s.go", folderPath, fileName)
	} else {
		filePath = fmt.Sprintf("%s.go", fileName)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return nil, "", err
	}

	return f, filePath, nil
}

// Get model file header
func getModelHeader(imports []string) (bytes.Buffer, error) {
	baseImports := []string{`"strings"`, `"database/sql"`, `"errors"`, `"fmt"`, `"github.com/dimonrus/godb"`}
	imports = append(imports, baseImports...)
	t := `package models

import ({{ range $key, $import := .Imports }}{{ $import }}
	{{ end }}
)
`
	var buf bytes.Buffer

	tml := template.Must(template.New("").Parse(t))
	err := tml.Execute(&buf, struct {
		Imports []string
	}{
		Imports: imports,
	})

	return buf, err
}

// Get model struct
func getModelStruct(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `type {{ .Model }} struct { {{ range $key, $column := .Columns }}
	{{ $column.ModelName }} {{ $column.ModelType }} {{ $column.Json }} {{ if $column.Description }}// {{ $column.Description }}{{ end }}{{ end }}
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Get model parser
func getModelParser(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// Parse model column
func (m *{{ .Model }}) parse(rows *sql.Rows) (*{{ .Model }}, error) {
	err := rows.Scan(m.Values()...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Columns
func getColumns(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// Model columns
func (m *{{ .Model }}) Columns() []string {
	return []string{ {{ range $key, $column := .Columns }}{{ if $key }}, {{ end }}"{{ $column.Name }}"{{ end }} }	
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Column map
func getValues(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// Model values
func (m *{{ .Model }}) Values() (values []interface{}) {
	return append(values, {{ range $key, $column := .Columns }}{{ if $key }}, {{ end }}{{ if $column.IsArray }}pq.Array({{ end }}&m.{{ $column.ModelName }}{{ if $column.IsArray  }}){{ end }}{{ end }})
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Parse Template
func ParseCrudMethodTemplate(t string, model string, table string, columns Columns) (bytes.Buffer, error) {
	var buf bytes.Buffer
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"system": func(column Column) bool {
			return gohelp.ExistsInArrayString(column.Name, []string{"updated_at", "created_at", "deleted_at"}) ||
				(column.IsPrimaryKey && column.Sequence != nil)
		},
		"cameled": func(name string) string {
			cameled, err := gohelp.ToCamelCase(name, true)
			if err != nil {
				panic(err)
			}
			return cameled
		},
	}

	tml := template.Must(template.New("").Funcs(funcMap).Parse(t))
	err := tml.Execute(&buf, struct {
		Model   string
		Table   string
		Columns Columns
	}{
		Model:   model,
		Table:   table,
		Columns: columns,
	})

	return buf, err
}

// Get model loader
func getModelLoader(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// SQL load Query
func (m *{{ .Model }}) GetLoadQuery() string {
	return "SELECT " + strings.Join(m.Columns(), ",") + " FROM {{ .Table }} WHERE {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} AND {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{ $index }}{{ end }}{{ end }};"
}
// Load method
func (m *{{ .Model }}) Load(q godb.Queryer) (*{{ .Model }}, error) {
	if {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} && {{ end }}{{ $index = inc $index }} m.{{ $column.ModelName }} != {{ if eq $column.ModelType "string" }}""{{ else }}0{{ end }}{{ end }}{{ end }} {
		iterator, err := q.Query(m.GetLoadQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}m.{{ $column.ModelName }}{{ end }}{{ end }})
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if iterator == nil {
			return nil, errors.New("iterator is nil, nothing for load")
		}
		defer iterator.Close()
		if iterator.Next() == false {
			return nil, nil
		}
		return m.parse(iterator)
	}
	return nil, errors.New("no primary key specified, nothing for load")
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Get model loader
func getForeignModels(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `
{{ range $key, $column := .Columns }}{{ if $column.ForeignTable }}
// Load Foreign Model {{ cameled $column.ForeignTable }}
func (m *{{ $.Model }}) Get{{ cameled $column.ForeignTable }}(q godb.Queryer) (*{{ cameled $column.ForeignTable }}, error) {
	return (&{{ cameled $column.ForeignTable }} { {{ cameled $column.ForeignColumnName}}:m.{{ $column.ModelName }} }).Load(q)
}
{{ end }}{{ end }}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Get model deleter
func getModelDeleter(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// SQL delete Query
func (m *{{ .Model }}) GetDeleteQuery() string {
	return "DELETE FROM {{ .Table }} WHERE {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} AND {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{$index}}{{ end }}{{ end }};"
}
// Delete method
func (m *{{ .Model }}) Delete(q godb.Queryer) error {
	_, err := q.Exec(m.GetDeleteQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}m.{{ $column.ModelName }}{{ end }}{{ end }})
	return err
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Check if model exists
func getModelExister(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// SQL exists Query
func (m *{{ .Model }}) GetExistsQuery() string {
	return "SELECT EXISTS (SELECT 1 FROM {{ .Table }}) WHERE {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} AND {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{$index}}{{ end }}{{ end }};"
}
// Exists method
func (m *{{ .Model }}) IsExists(q godb.Queryer) (exists bool, err error) {
	err = q.QueryRow(m.GetExistsQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}m.{{ $column.ModelName }}{{ end }}{{ end }}).Scan(&exists)
	return
}
`
	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Get model saver
func getModelSaver(model string, table string, columns Columns) (bytes.Buffer, error) {
	var hasSequence bool
	// check for sequence and primary key
	for _, column := range columns {
		if column.IsPrimaryKey && column.Sequence != nil {
			hasSequence = true
			break
		}
	}
	var t string

	saveMethod := `//Model saver method
func (m *{{ .Model }}) Save(q godb.Queryer) (*{{ .Model }}, error) {
	var err error
	if {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} && {{ end }}{{ $index = inc $index }} m.{{ $column.ModelName }} != {{ if eq $column.ModelType "string" }}""{{ else }}0{{ end }}{{ end }}{{ end }} {
		err = q.QueryRow(m.GetSaveQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if or (not (system $column)) $column.IsPrimaryKey }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}}).
			Scan({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}})
	} else {
		err = q.QueryRow(m.GetSaveQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}}).
			Scan({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}})
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}
`
	if !hasSequence {
		saveMethod := `//Model saver method
func (m *{{ .Model }}) Save(q godb.Queryer) (*{{ .Model }}, error) {
	err := q.QueryRow(m.GetSaveQuery(), {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if or (not (system $column)) $column.IsPrimaryKey }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}}).
	    Scan({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if $index }}, {{ end }}&m.{{ $column.ModelName }}{{ $index = inc $index }}{{ end}}{{ end}})
	if err != nil {
		return nil, err
	}
	return m, nil
}
`
		t = fmt.Sprintf(`// SQL upsert Query
func (m *{{ .Model }}) GetSaveQuery() string {
	return %cINSERT INTO {{ .Table }} ({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $column.Name }}{{ $index = inc $index }}{{ end}}{{ end}}) VALUES ({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $index = inc $index }}${{ $index }}{{ end }}{{ end }}) 
	ON CONFLICT ({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}{{ $column.Name }}{{ end }}{{ end }}) DO UPDATE SET {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{ $index }}{{ end }}{{ end }}{{ range $key, $column := .Columns }}{{ if eq $column.Name "updated_at" }}, updated_at = NOW(){{ end }}{{ end }}
	{{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if eq $index 0 }}RETURNING {{ end }}{{ if $index }}, {{ end }}{{ $column.Name }}{{ $index = inc $index }}{{ end}}{{ end}};%c
}
%s`, '`', '`', saveMethod)
	} else {
		t = fmt.Sprintf(`// SQL upsert Query
func (m *{{ .Model }}) GetSaveQuery() string {
	if {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} && {{ end }}{{ $index = inc $index }} m.{{ $column.ModelName }} != {{ if eq $column.ModelType "string" }}""{{ else }}0{{ end }}{{ end }}{{ end }} {
		return %cUPDATE {{ .Table }} SET {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{ inc $key }}{{ end }}{{ end }}{{ range $key, $column := .Columns }}{{ if eq $column.Name "updated_at" }}, updated_at = NOW(){{ end }}{{ end }}
		WHERE {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }} AND {{ end }}{{ $index = inc $index }}{{ $column.Name }} = ${{ $index }}{{ end }}{{ end }}
		{{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if eq $index 0 }}RETURNING {{ end }}{{ if $index }}, {{ end }}{{ $column.Name }}{{ $index = inc $index }}{{ end}}{{ end}};%c
	} 
	return %cINSERT INTO {{ .Table }} ({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $index = inc $index }}{{ $column.Name }}{{ end}}{{ end}}) VALUES ({{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if not (system $column) }}{{ if $index }}, {{ end }}{{ $index = inc $index }}${{ $index }}{{ end }}{{ end }})
    {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if (system $column) }}{{ if eq $index 0 }}RETURNING {{ end }}{{ if $index }}, {{ end }}{{ $index = inc $index }}{{ $column.Name }}{{ end}}{{ end}};%c
}
%s`, '`', '`', '`', '`', saveMethod)
	}

	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Get model searcher
func getModelSearcher(model string, table string, columns Columns) (bytes.Buffer, error) {
	t := `// Search by filer
func Search{{ .Model }}(q godb.Queryer, filter *godb.SqlFilter) (*[]{{ .Model }}, {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}[]{{ $column.ModelType }}{{ end }}{{ end }}, error) {
	query := fmt.Sprintf("SELECT " + strings.Join((&{{ .Model }}{}).Columns(), ",") + " FROM {{ .Table }} %s", filter.GetWithWhere())
	rows, err := q.Query(query, filter.GetArguments()...)
	{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}
	entity{{ $column.ModelName }}s := make([]{{ $column.ModelType }}, 0){{ end }}{{ end }}
	if err != nil {
		return nil, {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}entity{{ $column.ModelName }}s{{ end }}{{ end }}, err
	}
	defer rows.Close()
	var result []{{ .Model }}
	for rows.Next() {
		row, err := (&{{ .Model }}{}).parse(rows)
		if err != nil {
			return &result, {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}entity{{ $column.ModelName }}s{{ end }}{{ end }}, err
		}
		{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}
		entity{{ $column.ModelName }}s = append(entity{{ $column.ModelName }}s, row.{{ $column.ModelName}}){{ end }}{{ end }}
		result = append(result, *row)
	}
	return &result, {{ $index := 0 }}{{ range $key, $column := .Columns }}{{ if $column.IsPrimaryKey }}{{ if $index }}, {{ end }}{{ $index = inc $index }}entity{{ $column.ModelName }}s{{ end }}{{ end }}, nil
}
`

	return ParseCrudMethodTemplate(t, model, table, columns)
}

// Create Model File
func CreateModel(schema string, table string, path string) error {
	var tableExists bool
	var imports []string

	columns, err := GetTableColumns(schema, table)
	if err != nil {
		return err
	}

	for _, column := range *columns {
		tableExists = true
		imports = gohelp.AppendUniqueString(imports, column.Import)
	}

	// Check if table not exist or no columns
	if !tableExists {
		return errors.New(fmt.Sprintf("table (%s) is not exists", table))
	}

	// Name of the model
	name := table
	if schema != "public" {
		name = schema + "_" + table
	}

	modelName, err := gohelp.ToCamelCase(name, true)
	if err != nil {
		return err
	}

	//Table name with schema
	tableName := fmt.Sprintf("%s.%s", schema, table)

	//create file
	file, path, err := CreateModelFile(schema, name, path)
	if err != nil {
		return err
	}

	defer file.Close()

	// Get header
	header, err := getModelHeader(imports)
	if err != nil {
		return err
	}

	model, err := getModelStruct(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	cols, err := getColumns(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	vals, err := getValues(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	parser, err := getModelParser(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	loader, err := getModelLoader(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	foreign, err := getForeignModels(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	deleter, err := getModelDeleter(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	existster, err := getModelExister(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	saver, err := getModelSaver(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	searcher, err := getModelSearcher(modelName, tableName, *columns)
	if err != nil {
		return err
	}

	_, err = file.Write(header.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(model.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(cols.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(vals.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(parser.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(loader.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(foreign.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(deleter.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(existster.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(saver.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(searcher.Bytes())
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "fmt", path)
	err = cmd.Run()
	if err != nil {
		return err
	}

	dbo.Logger.Printf("Model file created: %s", path)

	return nil
}
