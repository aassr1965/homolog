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
		SELECT c.id, c.item_code, c.title_fr, c.assigned_department, COALESCE(c.assigned_user_id::text, 'NULL'),
		       c.due_date, c.workflow_status, c.updated_at
		FROM dossier_checklist_items c
		JOIN dossiers d ON d.id = c.dossier_id
		WHERE d.dossier_code = 'DOS-2026-PARAC-NF'
		ORDER BY c.updated_at DESC NULLS LAST
		LIMIT 10
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, code, title, dept, uid, wf string
		var dueDate, updatedAt interface{}
		rows.Scan(&id, &code, &title, &dept, &uid, &dueDate, &wf, &updatedAt)
		fmt.Printf("[%s] ID=%s Dept='%s' UserID='%s' Due=%v WF='%s' UpdatedAt=%v\n", code, id, dept, uid, dueDate, wf, updatedAt)
	}
}
