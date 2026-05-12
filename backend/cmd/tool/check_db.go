package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "data/pve_manager.db")
	if err != nil {
		fmt.Printf("Open failed: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA table_info(app_deployments)")
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Columns in app_deployments:")
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		fmt.Printf("  %d: %s (%s, notnull=%d, pk=%d)\n", cid, name, ctype, notnull, pk)
	}
}
