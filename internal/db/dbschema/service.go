package dbschema

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) GetCreateTableStatements() (string, error) {
	tables, err := s.getTables()
	if err != nil {
		return "", err
	}

	var output []string
	for _, table := range tables {
		createStmt, err := s.buildCreateTable(table)
		if err != nil {
			return "", err
		}
		output = append(output, createStmt)
	}

	return strings.Join(output, "\n\n"), nil
}

func (s *Service) GetAllAsString() (string, error) {
	createStatements, err := s.GetCreateTableStatements()
	if err != nil {
		return "", fmt.Errorf("erro ao obter declarações de criação: %w", err)
	}
	if createStatements == "" {
		return "Nenhuma tabela encontrada no banco de dados.", nil
	}
	return createStatements, nil
}

func (s *Service) ExtractTableNames() []string {
	re := regexp.MustCompile(`(?i)create table "?(\w+)"?`)
	allStr, err := s.GetAllAsString()
	if err != nil {
		return nil
	}

	matches := re.FindAllStringSubmatch(allStr, -1)

	var names []string
	seen := map[string]bool{}
	for _, m := range matches {
		name := strings.ToLower(m[1])
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}

func (s *Service) GetColumnsTable(table string) ([]string, error) {
	columns, err := s.getColumns(table)
	if err != nil {
		return nil, err
	}
	return columns, nil
}

func (s *Service) getTables() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func (s *Service) buildCreateTable(table string) (string, error) {
	columns, err := s.getColumns(table)
	if err != nil {
		return "", err
	}

	pk, err := s.getPrimaryKey(table)
	if err != nil {
		return "", err
	}

	fks, err := s.getForeignKeys(table)
	if err != nil {
		return "", err
	}

	stmt := fmt.Sprintf("CREATE TABLE %s (\n", table)
	stmt += strings.Join(columns, ",\n")

	if pk != "" {
		stmt += fmt.Sprintf(",\nPRIMARY KEY (%s)", pk)
	}

	for _, fk := range fks {
		stmt += fmt.Sprintf(",\nFOREIGN KEY (%s) REFERENCES %s(%s)", fk.Column, fk.RefTable, fk.RefColumn)
	}

	stmt += "\n);"
	return stmt, nil
}

func (s *Service) getColumns(table string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = $1
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var name, dataType, nullable string
		if err := rows.Scan(&name, &dataType, &nullable); err != nil {
			return nil, err
		}
		nullStr := ""
		if nullable == "NO" {
			nullStr = " NOT NULL"
		}
		cols = append(cols, fmt.Sprintf("  %q %s%s", name, dataType, nullStr))
	}
	return cols, nil
}

func (s *Service) getPrimaryKey(table string) (string, error) {
	var pk sql.NullString
	err := s.db.QueryRow(`
	SELECT string_agg(a.attname, ', ')
	FROM pg_index i
	JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
	WHERE i.indrelid = $1::regclass AND i.indisprimary
`, table).Scan(&pk)

	if err != nil {
		return "", err
	}

	if !pk.Valid {
		return "", nil
	}

	return pk.String, nil

}

type foreignKey struct {
	Column    string
	RefTable  string
	RefColumn string
}

func (s *Service) getForeignKeys(table string) ([]foreignKey, error) {
	rows, err := s.db.Query(`
		SELECT
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM 
			information_schema.table_constraints AS tc 
			JOIN information_schema.key_column_usage AS kcu
			  ON tc.constraint_name = kcu.constraint_name
			JOIN information_schema.constraint_column_usage AS ccu
			  ON ccu.constraint_name = tc.constraint_name
		WHERE constraint_type = 'FOREIGN KEY' AND tc.table_name=$1;
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []foreignKey
	for rows.Next() {
		var fk foreignKey
		if err := rows.Scan(&fk.Column, &fk.RefTable, &fk.RefColumn); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, nil
}
