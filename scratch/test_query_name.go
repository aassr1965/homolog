//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), "postgres://homolog:homolog_secret_2026@localhost:5434/homolog?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	query := `
		SELECT i.item_code, i.is_locked, 
		       COALESCE(u_lck.first_name || ' ' || u_lck.last_name, '') as locked_by_name,
		       COALESCE(i.locked_by::text, 'NULL')
		FROM dossier_checklist_items i
		LEFT JOIN users u_lck ON i.locked_by = u_lck.id
		WHERE i.item_code = '1.P.8' AND i.dossier_id = '50b85414-6b29-4833-a5e5-ddf2e65d9a96'
	`
	var code, lckName, lckID string
	var isLocked bool
	err = pool.QueryRow(context.Background(), query).Scan(&code, &isLocked, &lckName, &lckID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Item %s -> is_locked: %v, locked_by ID: %s, locked_by_name: '%s'\n", code, isLocked, lckID, lckName)
}
