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

	_, err = pool.Exec(context.Background(), "ALTER TABLE dossier_checklist_items ADD COLUMN IF NOT EXISTS assigned_at TIMESTAMPTZ;")
	if err != nil {
		log.Fatal(err)
	}
	// For already assigned item 1.A.3, set assigned_at = updated_at
	_, err = pool.Exec(context.Background(), "UPDATE dossier_checklist_items SET assigned_at = updated_at WHERE item_code = '1.A.3' AND due_date IS NOT NULL;")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Col assigned_at created and 1.A.3 updated!")
}
