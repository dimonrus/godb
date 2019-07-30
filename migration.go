package godb

import (
	"database/sql"
	"fmt"
	"os"
	"text/template"
	"time"
)

// Upgrade
func (m *Migration) Upgrade(class string) error {
	for _, migration := range (*m.Registry)[class] {
		tx, err := m.DBO.Begin()
		if err != nil {
			return err
		}
		var applyTime uint64
		// Check migration that already applied
		query := fmt.Sprintf("SELECT apply_time FROM migration_%s WHERE version = $1", class)
		err = m.DBO.QueryRow(query, migration.GetVersion()).Scan(&applyTime)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		// If already applied continue
		if applyTime != 0 {
			continue
		}
		// Apply new migration
		err = migration.Up(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		// Update migration table
		query = fmt.Sprintf("INSERT INTO migration_%s (version, apply_time) VALUES ($1, $2);", class)
		_, err = tx.Exec(query, migration.GetVersion(), time.Now().Unix())
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

// Downgrade
func (m *Migration) Downgrade(class string, version string) error {
	var applyTime uint64
	for _, migration := range (*m.Registry)[class] {
		if migration.GetVersion() != version {
			continue
		}
		// Check migration that already applied
		query := fmt.Sprintf("SELECT apply_time FROM migration_%s WHERE version = $1", class)
		err := m.DBO.QueryRow(query, migration.GetVersion()).Scan(&applyTime)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		// If no migration found
		if applyTime == 0 {
			m.DBO.Logger.Print("No migration for downgrade")
			return nil
		}
		// Begin
		tx, err := m.DBO.Begin()
		if err != nil {
			return err
		}
		// Downgrade migration
		err = migration.Down(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		// Delete from migration table
		query = fmt.Sprintf("DELETE FROM migration_%s WHERE version = $1;", class)
		_, err = tx.Exec(query, version)
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// Init migration
func (m *Migration) InitMigration(class string) error {
	tableName := fmt.Sprintf("migration_%s", class)
	query := `SELECT to_regclass('%s.public.%s');`

	// Check if table exists
	var regClass sql.NullString
	err := m.DBO.QueryRow(fmt.Sprintf(query, m.Config.Name, tableName)).Scan(&regClass)

	if err != nil {
		return err
	}

	if regClass.Valid == false {
		query = `CREATE TABLE migration_%s (
			version TEXT NOT NULL,
			apply_time BIGINT NOT NULL);
			CREATE UNIQUE INDEX ON migration_%s (version);`

		_, err = m.DBO.Exec(fmt.Sprintf(query, class, class))
		if err != nil {
			return err
		}
	}

	return nil
}

// Create migration file
func (m *Migration) CreateMigrationFile(class string, name string) error {
	fileName := fmt.Sprintf("m_%v_%s", time.Now().Unix(), name)
	folderPath := fmt.Sprintf("%s/%s", m.MigrationPath, class)
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s.go", folderPath, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	migrationTemplate := m.GetTemplate()

	err = migrationTemplate.Execute(f, struct {
		Class             string
		MigrationTypeName string
		RegistryPath      string
		RegistryXPath     string
	}{
		Class:             class,
		MigrationTypeName: fileName,
		RegistryPath:      m.RegistryPath,
		RegistryXPath:     m.RegistryXPath,
	})
	if err != nil {
		return err
	}

	m.DBO.Logger.Printf("Migration created: %s", filePath)

	return nil
}

func (m *Migration) GetTemplate() *template.Template {
	var migrationTemplate = template.Must(template.New("").
		Parse(`// {{ .Class }} Migration file
package {{ .Class }}

import (
	"github.com/dimonrus/godb"
	"{{ .RegistryPath }}"
)

type {{ .MigrationTypeName }} struct {}

func init() {
	{{ .RegistryXPath }}["{{ .Class }}"] = append({{ .RegistryXPath }}["{{ .Class }}"], {{ .MigrationTypeName }}{})
}

func (m {{ .MigrationTypeName }}) GetVersion () string {
	return "{{ .MigrationTypeName }}"
}

func (m {{ .MigrationTypeName }}) Up (tx *godb.SqlTx) error {
	// write code here

	return nil
}

func (m {{ .MigrationTypeName }}) Down (tx *godb.SqlTx) error {
	// write code here

	return nil
}
`))
	return migrationTemplate
}
