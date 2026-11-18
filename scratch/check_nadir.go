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

	var role, email string
	err = pool.QueryRow(context.Background(), "SELECT email, role FROM users WHERE first_name = 'Nadir'").Scan(&email, &role)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Nadir (%s) -> role: %s\n", email, role)
}
