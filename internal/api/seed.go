package api

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/auth"
	"homolog/internal/domain"
)

// SeedInitialData vérifie et insère l'organisation et les comptes BPF essentiels au 1er lancement
func SeedInitialData(ctx context.Context, pool *pgxpool.Pool, orgSlug, orgName string) error {
	// 1. Vérifier si l'organisation existe
	var orgID string
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		// Création de l'organisation
		query := `
			INSERT INTO organizations (slug, legal_name, miph_license_number, technical_director_name)
			VALUES ($1, $2, 'MIPH-DZ-2024-0089', 'Dr. Amine K.')
			RETURNING id
		`
		if err := pool.QueryRow(ctx, query, orgSlug, orgName).Scan(&orgID); err != nil {
			return err
		}
		log.Printf("[Init] Organisation créée : %s (%s)", orgName, orgSlug)
	}

	// 2. Utilisateurs par défaut
	initialUsers := []struct {
		Email     string
		Password  string
		Role      domain.UserRole
		FirstName string
		LastName  string
	}{
		{
			Email:     "dt@homolog.dz",
			Password:  "P@ssw0rdPharma2026!",
			Role:      domain.RoleDirecteurTechnique,
			FirstName: "Amine",
			LastName:  "K.",
		},
		{
			Email:     "admin@homolog.dz",
			Password:  "Admin@Pharma2026!",
			Role:      domain.RoleAdminLaboratoire,
			FirstName: "Karim",
			LastName:  "M.",
		},
		{
			Email:     "reg@homolog.dz",
			Password:  "Reg@Pharma2026!",
			Role:      domain.RoleResponsableAffairesReg,
			FirstName: "Sarah",
			LastName:  "B.",
		},
	}

	for _, u := range initialUsers {
		var exists bool
		_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE organization_id = $1 AND email = $2)", orgID, u.Email).Scan(&exists)
		if !exists {
			pwdHash, err := auth.HashPassword(u.Password, nil)
			if err != nil {
				return err
			}

			insertUserQuery := `
				INSERT INTO users (organization_id, email, password_hash, first_name, last_name, role)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = pool.Exec(ctx, insertUserQuery, orgID, u.Email, pwdHash, u.FirstName, u.LastName, string(u.Role))
			if err != nil {
				return err
			}
			log.Printf("[Init] Compte créé : %s (%s) - Rôle: %s", u.Email, u.FirstName+" "+u.LastName, u.Role)
		}
	}

	return nil
}
