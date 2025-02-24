package db

import "strings"

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
