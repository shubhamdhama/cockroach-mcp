package db

import (
	"context"
	"fmt"
	"strings"
)

func ListTables() (string, error) {
	rows, err := GetDB().Query(
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`,
	)

	if err != nil {
		return "", err
	}

	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return "", err
		}
		tables = append(tables, table)
	}

	return strings.Join(tables, ", "), nil
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
	rows, err := GetDB().QueryContext(ctx, query)
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
			var v any
			b, ok := col.([]byte)
			if ok {
				v = string(b)
			} else {
				v = col
			}
			rowValues = append(rowValues, fmt.Sprintf("%v", v))
		}
		sb.WriteString("| " + strings.Join(rowValues, " | ") + " |\n")
	}

	return sb.String()
}
