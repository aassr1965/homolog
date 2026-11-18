//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://homolog:homolog_secret_2026@localhost:5434/homolog?sslmode=disable"
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Connexion échouée : %v", err)
	}
	defer pool.Close()

	query := `
		SELECT i.id, i.item_code, i.title_fr, i.is_locked, COALESCE(i.locked_by::text, 'NULL'), 
		       COALESCE(i.locked_at::text, 'NULL'), i.current_revision, i.workflow_status,
		       COALESCE(u.first_name, 'NULL'), COALESCE(u.last_name, 'NULL'),
		       COALESCE(i.ged_document_id::text, 'NULL')
		FROM dossier_checklist_items i
		LEFT JOIN users u ON i.locked_by = u.id
		WHERE i.item_code = '1.P.8'
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Requête échouée : %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, code, title, lockedBy, lockedAt, rev, wf, fn, ln, gedID string
		var lockedBool bool
		err := rows.Scan(&id, &code, &title, &lockedBool, &lockedBy, &lockedAt, &rev, &wf, &fn, &ln, &gedID)
		if err != nil {
			log.Fatalf("Scan échoué : %v", err)
		}
		fmt.Printf("Item [%s] %s\n", code, title)
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  IsLocked: %v\n", lockedBool)
		fmt.Printf("  LockedBy: %s (%s %s)\n", lockedBy, fn, ln)
		fmt.Printf("  LockedAt: %s\n", lockedAt)
		fmt.Printf("  Revision: %s\n", rev)
		fmt.Printf("  WorkflowStatus: %s\n", wf)
		fmt.Printf("  GEDDocID: %s\n", gedID)
	}
}
