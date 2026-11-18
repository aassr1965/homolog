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

	var wf, rev string
	err = pool.QueryRow(context.Background(), "SELECT workflow_status, current_revision FROM dossier_checklist_items WHERE item_code = '1.P.8' AND dossier_id = '50b85414-6b29-4833-a5e5-ddf2e65d9a96'").Scan(&wf, &rev)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Item 1.P.8 -> workflow_status: %s, current_revision: %s\n", wf, rev)
}
