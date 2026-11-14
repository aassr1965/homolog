package api

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// SeedInitialDeficiencies injecte des données de démonstration pour le Studio de Compléments ANPP (Jalon 1.7)
func SeedInitialDeficiencies(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Idempotence : si des lettres de compléments existent déjà, ne rien faire
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM anpp_deficiency_letters WHERE organization_id = $1", orgID).Scan(&count)
	if count > 0 {
		return
	}

	// Récupération d'un dossier actif (de préférence le dossier princeps ou générique amoxicilline)
	var dossierID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM dossiers 
		WHERE organization_id = $1 
		ORDER BY created_at ASC 
		LIMIT 1
	`, orgID).Scan(&dossierID)
	if err != nil {
		return
	}

	// Récupération de l'utilisateur Pharmacien Directeur Technique pour les visas 1.A.7
	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&dtUserID)

	// Récupération d'un utilisateur Affaires Réglementaires pour l'attribution
	var raUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'responsable_affaires_reglementaires' LIMIT 1", orgID).Scan(&raUserID)

	now := time.Now().UTC()
	// Lettre reçue il y a 76 jours avec un délai légal de 90 jours -> 14 jours restants (Alerte CRITIQUE J-14)
	receivedDate := now.AddDate(0, 0, -76)
	dueDate := receivedDate.AddDate(0, 0, 90)

	var docID *string
	var existingDocID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM ged_documents WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1", orgID).Scan(&existingDocID); err == nil {
		s := existingDocID.String()
		docID = &s
	}

	letter := domain.ANPPDeficiencyLetter{
		ID:                uuid.New().String(),
		OrganizationID:    orgID.String(),
		DossierID:         dossierID.String(),
		ReferenceNumber:   "ANPP/DEPP/N°1428/2026",
		ReceivedDate:      receivedDate,
		LegalDeadlineDays: 90,
		DueDate:           dueDate,
		Status:            domain.DeficiencyLetterStatusPartiallyAnswered,
		DocumentID:        docID,
		Notes:             "Commission technique d'enregistrement du médicament — Séance plénière du 12 Septembre 2026. Notification impérative avec mise en demeure sous peine de rejet définitif du dossier.",
	}

	raUserIDStr := raUserID.String()
	dtUserIDStr := dtUserID.String()
	validatedAt := now.AddDate(0, 0, -2)

	tasks := []domain.ANPPDeficiencyTask{
		{
			QuestionNumber:  1,
			QuestionText:    "Fournir les profils d'impuretés organiques et de solvants résiduels selon la section 3.2.S.3.2 pour trois lots industriels consécutifs conformément à la monographie Ph. Eur.",
			ModuleTarget:    "Module 3",
			Priority:        "haute",
			AssignedTo:      &raUserIDStr,
			DueDate:         dueDate,
			Status:          domain.DeficiencyTaskStatusValidatedDT,
			ResponseSummary: "Rapport analytique chromatographique (HPLC/GC-MS) validant que les teneurs en solvants résiduels (classe 2 et 3) demeurent strictement sous les seuils ICH Q3C. Certificats d'analyses des 3 lots industriels validés.",
			ValidatedByDTID: &dtUserIDStr,
			ValidatedAt:     &validatedAt,
		},
		{
			QuestionNumber:  2,
			QuestionText:    "Transmettre la maquette définitive du conditionnement secondaire et notice patient intégrant le caducée national algérien, le code GTIN et le tarif public homologué en DZD (1.M.1).",
			ModuleTarget:    "Module 1",
			Priority:        "urgente",
			AssignedTo:      &raUserIDStr,
			DueDate:         dueDate,
			Status:          domain.DeficiencyTaskStatusAnswered,
			ResponseSummary: "Maquettes d'emballage secondaire révisées intégrant le logo officiel Sanad Pharma et le caducée algérien haute visibilité. Fichiers vectoriels PDF conformes aux spécifications graphiques ANPP 2025 prêts pour visa final.",
		},
		{
			QuestionNumber: 3,
			QuestionText:   "Justifier la méthode de dosage bioanalytique validée (LC-MS/MS) et fournir les courbes d'étalonnage pour l'étude de bioéquivalence comparative sous conditions de jeûne (5.3.1.4).",
			ModuleTarget:   "Module 5",
			Priority:       "urgente",
			AssignedTo:     &raUserIDStr,
			DueDate:        dueDate,
			Status:         domain.DeficiencyTaskStatusPending,
		},
	}

	repo := repository.NewDeficiencyRepository(pool)
	_, err = repo.CreateLetterWithTasks(ctx, &letter, tasks)
	if err != nil {
		log.Printf("[AVERTISSEMENT] Erreur seed compléments ANPP: %v", err)
		return
	}

	log.Printf("[SEED] Lettre officielle de compléments ANPP %s injectée avec succès (Jalon 1.7).", letter.ReferenceNumber)
}
