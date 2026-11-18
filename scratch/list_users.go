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

	rows, err := pool.Query(context.Background(), "SELECT email, first_name, last_name, role FROM users ORDER BY role")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var email, fn, ln, role string
		rows.Scan(&email, &fn, &ln, &role)
		fmt.Printf("%s %s (%s) -> %s\n", fn, ln, email, role)
	}
}
