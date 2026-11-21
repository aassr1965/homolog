package api

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// SeedInitialVariations provisionne des demandes de variations de démonstration (Jalon 2.4 - LD-DEPP07/V01)
func SeedInitialVariations(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Idempotence : si des variations existent déjà, ne rien faire
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM anpp_variation_requests WHERE organization_id = $1", orgID).Scan(&count)
	if count > 0 {
		return
	}

	// Récupérer un médicament enregistré
	var medID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM medicines WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1", orgID).Scan(&medID)
	if err != nil {
		return
	}

	// Récupérer le Pharmacien Directeur Technique
	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&dtUserID)

	varRepo := repository.NewVariationRepository(pool)

	// -------------------------------------------------------------------------
	// Variation Démo 1 : Grouping avec escalade Mmaj (Rubrique 7.d.1 + 10.a.1)
	// -------------------------------------------------------------------------
	var1ID := uuid.New()
	var1 := domain.VariationRequest{
		ID:                       var1ID,
		OrganizationID:           orgID,
		MedicineID:               medID,
		TrackingNumber:           "VAR-2026-AMOX-01",
		Title:                    "Nouveau site de synthèse API et mise à l'échelle de lot (LD-DEPP07/V01)",
		Description:              "Ajout d'un nouveau site de fabrication pour la substance active Amoxicilline trihydratée avec optimisation du procédé de cristallisation.",
		IsGrouping:               true,
		GroupingJustification:    "Regroupement en vertu de l'article 09 de l'arrêté du 07/11/2021 : le changement de site de fabrication API entraîne corollairement une augmentation de la taille de lot de synthèse.",
		OverallClassification:   domain.ClassMmaj,
		NominalClassification:   domain.ClassMmod,
		HasEscalation:            true,
		EscalationSummary:        "Reclassement automatique de la rubrique 7.d.1 de Mmod vers Mmaj en raison de données de stabilités Zone IVb 3 lots encore en cours.",
		CurrentApprovedSituation: "Synthèse de la substance active sur le site historique approuvé (Site A, Hyderabad, Inde). Taille de lot pilote approuvée : 50 kg.",
		ProposedSituation:         "Synthèse sur le nouveau site B (Gujarat, Inde) avec équipement automatisé et taille de lot industrielle de 250 kg (facteur x5).",
		TechnicalJustification:    "Sécurisation de la chaîne d'approvisionnement nationale en API, diminution substantielle des solvants résiduels et conformité ICH Q11.",
		NonModificationAttested:  true,
		NonModificationDetails:   "Le Pharmacien Directeur Technique atteste qu'aucune autre section en dehors des modules 1.H.8, 1.M.S.1, 3.2.S.2 et 3.2.P.5 n'a été modifiée.",
		FeeAmountDZD:             115000.00,
		Status:                   "pret_visa_dt",
		CreatedBy:                &dtUserID,
	}

	if err := varRepo.CreateVariationRequest(ctx, &var1); err != nil {
		log.Printf("[SeedVariations] Erreur création variation 1: %v", err)
		return
	}

	// Ajouter l'item 7.d.1 (escaladé Mmaj)
	item1 := domain.VariationItem{
		VariationRequestID: var1ID,
		RubriqueCode:       "7.d.1",
		ConditionsEvaluated: map[string]bool{
			"C1": true,  // Spécifications de libération équivalentes
			"C2": false, // Stabilités Zone IVb 3 lots non conclues -> ESCALADE Mmaj !
			"C3": true,  // Certificat BPF valide
		},
		CurrentState: "Site historique approuvé A (Hyderabad, Inde)",
		ProposedState: "Nouveau site de synthèse B (Gujarat, Inde)",
		Notes:         "Protocole de stabilité Zone IVb engagé à M3",
	}
	_ = varRepo.AddVariationItem(ctx, &item1)

	// Ajouter l'item 10.a.1 (Mmin nominale)
	item2 := domain.VariationItem{
		VariationRequestID: var1ID,
		RubriqueCode:       "10.a.1",
		ConditionsEvaluated: map[string]bool{
			"C1": true, // Facteur <= 10
			"C2": true, // Paramètres critiques constants
		},
		CurrentState: "Taille de lot SA : 50 kg",
		ProposedState: "Taille de lot SA : 250 kg (facteur x5)",
	}
	_ = varRepo.AddVariationItem(ctx, &item2)

	// Lier un document GED existant si disponible
	var sampleGEDDocID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM ged_documents WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1", orgID).Scan(&sampleGEDDocID)
	if err == nil {
		// Rattacher à la pièce 1 (Lettre de demande) et pièce 2 (Tableau comparatif)
		docs, _ := varRepo.GetVariationDocuments(ctx, var1ID)
		if len(docs) >= 2 {
			_ = varRepo.AttachDocument(ctx, docs[0].ID, sampleGEDDocID)
			_ = varRepo.AttachDocument(ctx, docs[1].ID, sampleGEDDocID)
		}
	}

	// -------------------------------------------------------------------------
	// Variation Démo 2 : Variation Mineure Mmin (Rubrique 27 - Pharmacopée)
	// -------------------------------------------------------------------------
	var2ID := uuid.New()
	var2 := domain.VariationRequest{
		ID:                       var2ID,
		OrganizationID:           orgID,
		MedicineID:               medID,
		TrackingNumber:           "VAR-2026-AMOX-02",
		Title:                    "Mise à jour des spécifications PF selon dernière édition Pharmacopée Européenne",
		Description:              "Alignement des méthodes de dissolution et d'impuretés organiques sur la 11ème édition de la Ph. Eur.",
		IsGrouping:               false,
		OverallClassification:   domain.ClassMmin,
		NominalClassification:   domain.ClassMmin,
		HasEscalation:            false,
		CurrentApprovedSituation: "Spécifications analytiques selon la 10ème édition de la Pharmacopée Européenne.",
		ProposedSituation:         "Alignement strict sur la 11ème édition révisée de la monographie officielle.",
		TechnicalJustification:    "Conformité réglementaire obligatoire aux monographies actualisées de la Pharmacopée officielle reconnue en Algérie.",
		NonModificationAttested:  true,
		FeeAmountDZD:             40000.00,
		Status:                   "brouillon",
		CreatedBy:                &dtUserID,
	}

	if err := varRepo.CreateVariationRequest(ctx, &var2); err == nil {
		item2A := domain.VariationItem{
			VariationRequestID: var2ID,
			RubriqueCode:       "27",
			ConditionsEvaluated: map[string]bool{
				"C1": true,
			},
			CurrentState: "Monographie Ph. Eur. 10.0",
			ProposedState: "Monographie Ph. Eur. 11.4",
		}
		_ = varRepo.AddVariationItem(ctx, &item2A)
	}

	log.Printf("[SeedVariations] Variations post-AMM démo LD-DEPP07 initialisées avec succès (VAR-2026-AMOX-01 & 02).")
}
