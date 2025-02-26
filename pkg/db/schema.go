package db

import (
	"context"
	"fmt"
	"strings"
)

func ListCRDBInternalTables(ctx context.Context) (string, error) {
	return queryInternal(
		ctx, `SELECT table_name, comment FROM [SHOW TABLES FROM crdb_internal WITH COMMENT]`)
}

// TODO (shubham): Revisit the design of these endpoints.
// Consider whether to merge ListCRDBInternalTables and ListSystemTables into one unified endpoint,
// maintain a static mapping of table descriptions, or simply document the table details in a text file
// to be provided to the LLM during initialization.
func ListSystemTables(ctx context.Context) (string, error) {
	return queryInternal(ctx,
		`SELECT table_name, estimated_row_count FROM [SHOW TABLES FROM system]`)
}

func ListTables(ctx context.Context, databaseName string) (string, error) {
	if databaseName == "" {
		return queryInternal(ctx,
			`SELECT database_name, schema_name, name FROM crdb_internal.tables WHERE database_name != 'system'`)
	}
	return queryInternal(ctx,
		`SELECT schema_name, name FROM crdb_internal.tables WHERE database_name = $1`, databaseName)
}

func ListDatabases(ctx context.Context) (string, error) {
	return queryInternal(ctx, `SHOW DATABASES`)
}

func ListClusterSettings(ctx context.Context) (string, error) {
	return queryInternal(ctx, `SHOW CLUSTER SETTINGS`)
}

func Execute(ctx context.Context, query string) (string, error) {
	result, err := GetDB().ExecContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to execute: %v", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to fetch affected rows: %d", err)
	}
	return fmt.Sprintf("Execution successful. Rows affected: %d", count), nil
}

func Query(ctx context.Context, query string) (string, error) {
	return queryInternal(ctx, query)
}

func queryInternal(ctx context.Context, query string, args ...any) (string, error) {
	rows, err := GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to query: %v", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve columns: %v", err)
	}

	var results [][]any
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return "", fmt.Errorf("failed to scan row: %v", err)
		}
		results = append(results, columns)
	}

	return formatAsMarkdown(cols, results), nil
}

func formatAsMarkdown(cols []string, results [][]any) string {
	var sb strings.Builder
	sb.WriteString("| " + strings.Join(cols, " | ") + " |\n")
	separator := make([]string, len(cols))
	for i := range separator {
		separator[i] = "---"
	}
	sb.WriteString("| " + strings.Join(separator, " | ") + " |\n")

	for _, row := range results {
		var rowValues []string
		for _, col := range row {
			var s string
			b, ok := col.([]byte)
			if ok {
				s = string(b)
			} else {
				s = fmt.Sprintf("%v", col)
			}
			s = strings.ReplaceAll(s, "\n", "\\n")
			rowValues = append(rowValues, s)
		}
		sb.WriteString("| " + strings.Join(rowValues, " | ") + " |\n")
	}

	return sb.String()
}
