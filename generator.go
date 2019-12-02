package godb

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dimonrus/gohelp"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var DefaultSystemColumnsSoft = SystemColumns{Created: "created_at", Updated: "updated_at", Deleted: "deleted_at"}
var DefaultSystemColumns = SystemColumns{Created: "created_at", Updated: "updated_at"}

type SystemColumns struct {
	Created string
	Updated string
	Deleted string
}

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
	IsCreated         bool    // Is created at column
	IsUpdated         bool    // Is updated at column
	IsDeleted         bool    // Is deleted at column

}

// Array of columns
type Columns []Column

// Get imports
func (c Columns) GetImports() []string {
	// imports in model file
	var imports []string

	for i := range c {
		imports = gohelp.AppendUniqueString(imports, c[i].Import)
	}

	return imports
}

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
func GetTableColumns(dbo Queryer, schema string, table string, sysCols SystemColumns) (*Columns, error) {
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

		if column.Name == sysCols.Created {
			column.IsCreated = true
		}
		if column.Name == sysCols.Updated {
			column.IsUpdated = true
		}
		if column.Name == sysCols.Deleted {
			column.IsDeleted = true
		}

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
			column.ModelType = "[]int64"
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

// Template helper functions
func getHelperFunc(systemColumns SystemColumns) template.FuncMap {
	return template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"system": func(column Column) bool {
			return gohelp.ExistsInArrayString(column.Name, []string{systemColumns.Created, systemColumns.Updated, systemColumns.Deleted}) ||
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

// Create model
func MakeModel(db Queryer, path string, schema string, table string, templatePath string, systemColumns SystemColumns) error {
	// Imports in model file
	var imports = []string{
		`"strings"`,
		`"database/sql"`,
		`"fmt"`,
		`"net/http"`,
		`"github.com/dimonrus/godb"`,
		`"github.com/dimonrus/porterr"`,
	}

	// Name of model
	var name = table

	if table == "" {
		return errors.New("table name is empty")
	}

	// New Template
	tmpl := template.New("model").Funcs(getHelperFunc(systemColumns))

	templateFile, err := os.Open(templatePath)
	if err != nil {
		return err
	}

	// Read template
	data, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return err
	}

	// Open model template
	tmpl = template.Must(tmpl.Parse(string(data)))

	// Columns
	columns, err := GetTableColumns(db, schema, table, systemColumns)
	if err != nil {
		return err
	}

	// Create all foreign models if not exists
	for i := range *columns {
		c := (*columns)[i]
		if c.ForeignTable != nil {
			var found bool
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if info == nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				data, err := ioutil.ReadAll(file)
				if err != nil {
					return err
				}
				cammeled, err := gohelp.ToCamelCase(*c.ForeignTable, true)
				if err != nil {
					return err
				}
				if strings.Contains(string(data), fmt.Sprintf("type %s struct {", cammeled)) {
					found = true
				}
				return nil
			})
			if err != nil {
				return err
			}
			if !found {
				err = MakeModel(db, path, schema, *c.ForeignTable, templatePath, systemColumns)
				if err != nil {
					return err
				}
			}
		}
	}

	// Collect imports
	for _, column := range *columns {
		imports = gohelp.AppendUniqueString(imports, column.Import)
	}

	// Name of the model
	if schema != "public" && schema != "" {
		name = schema + "_" + table
	}

	// To camel case
	modelName, err := gohelp.ToCamelCase(name, true)
	if err != nil {
		return err
	}

	var hasSequence bool
	// Check for sequence and primary key
	for _, column := range *columns {
		if column.IsPrimaryKey && column.Sequence != nil {
			hasSequence = true
			break
		}
	}

	// Create file
	file, path, err := CreateModelFile(schema, table, path)
	if err != nil {
		return err
	}

	// Parse template to file
	err = tmpl.Execute(file, struct {
		Model       string
		Table       string
		Columns     Columns
		HasSequence bool
		Imports     []string
	}{
		Model:       modelName,
		Table:       schema + "." + table,
		Columns:     *columns,
		HasSequence: hasSequence,
		Imports:     imports,
	})

	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	// Format code
	cmd := exec.Command("go", "fmt", path)
	err = cmd.Run()
	if err != nil {
		return err
	}

	if dbo, ok := db.(*DBO); ok {
		dbo.Logger.Printf("Model file created: %s", path)
	}

	return nil
}
