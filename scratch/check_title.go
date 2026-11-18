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

	var title string
	err = pool.QueryRow(context.Background(), "SELECT title_fr FROM dossier_checklist_items WHERE item_code = '1.P.8' AND dossier_id = '50b85414-6b29-4833-a5e5-ddf2e65d9a96'").Scan(&title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("title_fr: %q\n", title)
}
