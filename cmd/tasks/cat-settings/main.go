package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	appConnStr = "postgresql://root@localhost:26257/system?options=-ccluster%3Ddemoapp&sslmode=disable"
	sysConnStr = "postgresql://root@localhost:26257/system?sslmode=disable"
)

type clusterSettings []setting

type setting struct {
	name        string
	value       any
	description string
}

func main() {
	// Open db connections
	appDB, err := sql.Open("postgres", appConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to application tenant: %v", err)
	}
	defer appDB.Close()

	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to system tenant: %v", err)
	}
	defer sysDB.Close()

	// Get list of settings and their defaults
	settings, err := listSettings(sysDB)
	if err != nil {
		log.Fatalf("Failed to list cluster settings: %v", err)
	}

	operatorSettable := make([]setting, 0)
	tenantSettable := make([]setting, 0)
	uncategorized := make([]setting, 0)

	// Iterate over settings and set them using both system and application tenants
	fmt.Println("Setting cluster settings using application tenant:")
	for _, db := range []*sql.DB{appDB, sysDB} {
		for _, s := range settings {
			if strings.Contains(s.name, "unsafe") {
				uncategorized = append(uncategorized, s)
				continue
			}

			err := setSetting(db, s)
			if err != nil {
				if strings.Contains(err.Error(), "is only settable by the operator") {
					operatorSettable = append(operatorSettable, s)
				} else if strings.Contains(err.Error(), "changing the setting from a virtual cluster") {
					tenantSettable = append(tenantSettable, s)
				} else {
					log.Println(err)
					time.Sleep(time.Millisecond * 100)

					uncategorized = append(uncategorized, s)
				}
			}
		}
	}

	fmt.Println("Operator settable settings: ")
	for _, s := range operatorSettable {
		fmt.Println(s.name)
	}
	fmt.Println()

	fmt.Println("Tenant settable settings: ")
	for _, s := range tenantSettable {
		fmt.Println(s.name)
	}
	fmt.Println()

	fmt.Println("Uncategorized settings: ")
	for _, s := range uncategorized {
		fmt.Println(s.name)
	}
	fmt.Println()
}

func listSettings(db *sql.DB) (clusterSettings, error) {
	rows, err := db.Query("WITH cs AS (SHOW ALL CLUSTER SETTINGS) SELECT variable, value, description FROM cs;")
	if err != nil {
		return nil, fmt.Errorf("Failed to list cluster settings: %v", err)
	}
	defer rows.Close()

	settings := make([]setting, 0)
	for rows.Next() {
		var name, description string
		var value any
		err := rows.Scan(&name, &value, &description)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan row: %v", err)
		}
		settings = append(settings, setting{name, value, description})
	}

	return settings, nil
}

func setSetting(db *sql.DB, s setting) error {
	query := fmt.Sprintf("SET CLUSTER SETTING %s = %q;", s.name, s.value)
	log.Println(query)
	time.Sleep(time.Millisecond * 100)

	if _, err := db.ExecContext(context.Background(), query); err != nil {
		return err
	}
	return nil
}
