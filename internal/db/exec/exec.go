package exec

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Executor struct {
	db *sql.DB
}

func New(db *sql.DB) *Executor {
	return &Executor{db: db}
}

var forbiddenCommands = []string{
	"delete", "update", "insert", "drop", "alter", "create", "truncate",
}

func (e *Executor) Execute(sqlQuery string) ([]map[string]any, error) {
	cleanSQL := strings.ToLower(strings.TrimSpace(sqlQuery))

	if !strings.HasPrefix(cleanSQL, "select") {
		return nil, errors.New("apenas comandos SELECT são permitidos")
	}

	for _, cmd := range forbiddenCommands {
		matched, _ := regexp.MatchString(`\b`+cmd+`\b`, cleanSQL)
		if matched {
			return nil, fmt.Errorf("comando proibido detectado: %s", cmd)
		}
	}

	rows, err := e.db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range cols {
			val := values[i]

			if b, ok := val.([]byte); ok {
				strVal := string(b)

				if f, err := strconv.ParseFloat(strVal, 64); err == nil {
					row[col] = f
				} else {
					row[col] = strVal
				}
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}
