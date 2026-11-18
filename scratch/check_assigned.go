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
		SELECT c.item_code, c.title_fr, COALESCE(c.assigned_department, 'NULL'), COALESCE(c.assigned_user_id::text, 'NULL'),
		       c.workflow_status, c.is_locked
		FROM dossier_checklist_items c
		JOIN dossiers d ON d.id = c.dossier_id
		WHERE d.dossier_code = 'DOS-2026-PARAC-NF' AND (c.assigned_department IS NOT NULL OR c.is_locked = true)
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var code, title, dept, uid, wf string
		var locked bool
		rows.Scan(&code, &title, &dept, &uid, &wf, &locked)
		fmt.Printf("[%s] Dept='%s' UserID='%s' WF='%s' Locked=%v\n", code, dept, uid, wf, locked)
	}
}
