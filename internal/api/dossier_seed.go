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

// SeedInitialDossiers injecte les 3 dossiers réglementaires de démonstration (Guide ANPP Juillet 2025 & DM)
func SeedInitialDossiers(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Idempotence : si des dossiers existent déjà, ne rien faire
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM dossiers WHERE organization_id = $1", orgID).Scan(&count)
	if count > 0 {
		return
	}

	// Récupération de l'utilisateur Directeur Technique
	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&dtUserID)

	// Récupération des médicaments et dispositifs médicaux créés par le seed pharma
	var amoxID, paracID, seringueID *uuid.UUID

	var id1 uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM medicines WHERE organization_id = $1 AND brand_name_fr ILIKE '%Amoxicilline%' LIMIT 1", orgID).Scan(&id1); err == nil {
		amoxID = &id1
	}

	var id2 uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM medicines WHERE organization_id = $1 AND brand_name_fr ILIKE '%Paracétamol%' LIMIT 1", orgID).Scan(&id2); err == nil {
		paracID = &id2
	}

	var id3 uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM medical_devices WHERE organization_id = $1 LIMIT 1", orgID).Scan(&id3); err == nil {
		seringueID = &id3
	}

	// Chargement des documents GED disponibles pour pré-rattachement automatique
	gedDocs := make(map[string]uuid.UUID)
	rows, err := pool.Query(ctx, "SELECT document_code, id FROM ged_documents WHERE organization_id = $1 AND document_code IS NOT NULL", orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code string
			var docID uuid.UUID
			if err := rows.Scan(&code, &docID); err == nil {
				gedDocs[code] = docID
			}
		}
	}

	dossierRepo := repository.NewDossierRepository(pool)

	// 1. Dossier 1 : DOS-2026-AMOX-NI (Importation Générique - Procédure NI)
	if amoxID != nil {
		now := time.Now().UTC()
		targetDate := now.AddDate(0, 3, 0)
		trackingRef := "ANPP/DEP/2026/0412"

		d1 := &domain.Dossier{
			ID:                uuid.New(),
			OrganizationID:    orgID,
			MedicineID:        amoxID,
			ProductType:       domain.ProductTypeMedicament,
			DossierCode:       "DOS-2026-AMOX-NI",
			ProcedureCode:     domain.ProcedureNI,
			InternalReference: "REF-SANAD-AMOX-01",
			TargetDepositDate: &targetDate,
			TrackingNumber:    &trackingRef,
			Status:            domain.DossierStatusInPreparation,
			ResponsibleID:     &dtUserID,
		}

		items1 := domain.GenerateCTDChecklist(domain.ProcedureNI, domain.OriginGeneriqueImporte, false)

		// Pré-rattacher des documents réels de la GED
		attachMatch := map[string]string{
			"1.A.1":   "ANPP-FORM-AMM-2025",
			"1.A.6":   "MIPH-AGR-2024-0089",
			"1.A.7":   "ANPP-DEC-PDT-2023",
			"1.H.4":   "FOB-CC-EUR-2025",
			"1.M.P.1": "BPF-CERT-2024-012",
			"1.M.S.1": "BPF-CERT-2024-012",
		}
		for i := range items1 {
			targetDocCode, exists := attachMatch[items1[i].ItemCode]
			if exists {
				if gID, ok := gedDocs[targetDocCode]; ok {
					items1[i].GEDDocumentID = &gID
					items1[i].Status = domain.ItemStatusAttached
					notes := "Pièce réglementaire officielle issue du coffre-fort BPF"
					items1[i].ValidationNotes = &notes
				}
			}
		}

		if _, err := dossierRepo.CreateDossier(ctx, d1, items1); err != nil {
			log.Printf("[Seed Dossier] Erreur création dossier AMOX-NI : %v", err)
		} else {
			log.Printf("[Seed Dossier] Dossier créé : DOS-2026-AMOX-NI (%d items checklist)", len(items1))
		}
	}

	// 2. Dossier 2 : DOS-2026-PARAC-NF (Fabrication Locale Générique - Procédure NF)
	if paracID != nil {
		now := time.Now().UTC()
		targetDate := now.AddDate(0, 2, 0)
		trackingRef := "ANPP/DEP/2026/0589"

		d2 := &domain.Dossier{
			ID:                uuid.New(),
			OrganizationID:    orgID,
			MedicineID:        paracID,
			ProductType:       domain.ProductTypeMedicament,
			DossierCode:       "DOS-2026-PARAC-NF",
			ProcedureCode:     domain.ProcedureNF,
			InternalReference: "REF-SANAD-PARAC-02",
			TargetDepositDate: &targetDate,
			TrackingNumber:    &trackingRef,
			Status:            domain.DossierStatusInPreparation,
			ResponsibleID:     &dtUserID,
		}

		items2 := domain.GenerateCTDChecklist(domain.ProcedureNF, domain.OriginGeneriqueLocal, false)

		attachMatch := map[string]string{
			"1.A.1":   "ANPP-FORM-AMM-2025",
			"1.A.6":   "MIPH-AGR-2024-0089",
			"1.A.7":   "ANPP-DEC-PDT-2023",
			"1.M.P.1": "BPF-CERT-2024-012",
		}
		for i := range items2 {
			targetDocCode, exists := attachMatch[items2[i].ItemCode]
			if exists {
				if gID, ok := gedDocs[targetDocCode]; ok {
					items2[i].GEDDocumentID = &gID
					items2[i].Status = domain.ItemStatusAttached
					notes := "Attestation conforme de l'établissement exploitant"
					items2[i].ValidationNotes = &notes
				}
			}
		}

		if _, err := dossierRepo.CreateDossier(ctx, d2, items2); err != nil {
			log.Printf("[Seed Dossier] Erreur création dossier PARAC-NF : %v", err)
		} else {
			log.Printf("[Seed Dossier] Dossier créé : DOS-2026-PARAC-NF (%d items checklist)", len(items2))
		}
	}

	// 3. Dossier 3 : DOS-2026-SERINGUE-DM (Homologation Dispositif Médical - Classe IIa stérile)
	if seringueID != nil {
		now := time.Now().UTC()
		targetDate := now.AddDate(0, 4, 0)
		trackingRef := "ANPP/DM/2026/0104"

		d3 := &domain.Dossier{
			ID:                uuid.New(),
			OrganizationID:    orgID,
			MedicalDeviceID:   seringueID,
			ProductType:       domain.ProductTypeDispositifMedical,
			DossierCode:       "DOS-2026-SERINGUE-DM",
			ProcedureCode:     domain.ProcedureDM,
			InternalReference: "REF-SANAD-DM-03",
			TargetDepositDate: &targetDate,
			TrackingNumber:    &trackingRef,
			Status:            domain.DossierStatusInPreparation,
			ResponsibleID:     &dtUserID,
		}

		items3 := domain.GenerateDMChecklist("IIa", true)

		attachMatch := map[string]string{
			"DM.A.2": "MIPH-AGR-2024-0089",
			"DM.A.3": "ANPP-DEC-PDT-2023",
			"DM.Q.1": "ISO-13485-CERT",
			"DM.Q.2": "CE-CERT-2024-DM",
		}
		for i := range items3 {
			targetDocCode, exists := attachMatch[items3[i].ItemCode]
			if exists {
				if gID, ok := gedDocs[targetDocCode]; ok {
					items3[i].GEDDocumentID = &gID
					items3[i].Status = domain.ItemStatusAttached
					notes := "Certification ISO / CE conforme pour DM stérile"
					items3[i].ValidationNotes = &notes
				}
			}
		}

		if _, err := dossierRepo.CreateDossier(ctx, d3, items3); err != nil {
			log.Printf("[Seed Dossier] Erreur création dossier SERINGUE-DM : %v", err)
		} else {
			log.Printf("[Seed Dossier] Dossier créé : DOS-2026-SERINGUE-DM (%d items checklist)", len(items3))
		}
	}

	log.Println("[Seed Dossier] Dossiers réglementaires modèles ANPP initialisés avec succès.")
}

// SeedWorkflowPendingApprovals initialise des pièces de démonstration au statut 'pending_approval' (Prêt pour Visa DT)
// pour alimenter le Parapheur Électronique et le Cockpit DT dès l'ouverture de l'application
func SeedWorkflowPendingApprovals(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Vérifier si des pièces sont déjà en attente d'approbation
	var pendingCount int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM dossier_checklist_items i 
		JOIN dossiers d ON d.id = i.dossier_id 
		WHERE d.organization_id = $1 AND i.workflow_status = 'pending_approval'
	`, orgID).Scan(&pendingCount)

	if pendingCount > 0 {
		return
	}

	// Récupérer le Maker (RA Lead) et le Checker (QA Manager)
	var makerID, checkerID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND email = 'ra.lead@pharma.dz'", orgID).Scan(&makerID)
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND email = 'qa.lead@pharma.dz'", orgID).Scan(&checkerID)

	if makerID == uuid.Nil {
		_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'responsable_affaires_reg' LIMIT 1", orgID).Scan(&makerID)
	}
	if checkerID == uuid.Nil {
		_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'contributeur_qualite_bpf' LIMIT 1", orgID).Scan(&checkerID)
	}

	now := time.Now().UTC()
	subTime := now.Add(-4 * time.Hour)
	revTime := now.Add(-1 * time.Hour)

	type pendingSeedItem struct {
		dossierCode  string
		itemCode     string
		currentRev   string
		checkerNotes string
	}

	seeds := []pendingSeedItem{
		{
			dossierCode:  "DOS-2026-AMOX-NI",
			itemCode:     "1.P.8",
			currentRev:   "v0.2",
			checkerNotes: "Revue qualité favorable. Engagement sur les 3 premiers lots industriels conforme aux exigences de l'ANPP.",
		},
		{
			dossierCode:  "DOS-2026-AMOX-NI",
			itemCode:     "1.P.9",
			currentRev:   "v0.2",
			checkerNotes: "Protocole de pharmacovigilance et suivi de stabilité en Zone IVb validé. Prêt pour scellement légal du Pharmacien DT.",
		},
		{
			dossierCode:  "DOS-2026-PARAC-NF",
			itemCode:     "1.H.10",
			currentRev:   "v0.2",
			checkerNotes: "Déclaration solennelle d'évaluation vérifiée par l'Assurance Qualité. Conforme BPF.",
		},
		{
			dossierCode:  "DOS-2026-PARAC-NF",
			itemCode:     "3.2.P.5",
			currentRev:   "v0.3",
			checkerNotes: "Spécifications de libération et de péremption du produit fini conformes aux monographies Ph. Eur. et validation analytique ICH Q2(R1).",
		},
		{
			dossierCode:  "DOS-2026-SERINGUE-DM",
			itemCode:     "DM.T.1",
			currentRev:   "v0.2",
			checkerNotes: "Dossier technique DM vérifié selon l'arrêté du 10 mai 2021. Nomenclature et biocompatibilité conformes.",
		},
	}

	for _, s := range seeds {
		var itemID uuid.UUID
		err := pool.QueryRow(ctx, `
			SELECT i.id 
			FROM dossier_checklist_items i 
			JOIN dossiers d ON d.id = i.dossier_id 
			WHERE d.organization_id = $1 AND d.dossier_code = $2 AND i.item_code = $3
		`, orgID, s.dossierCode, s.itemCode).Scan(&itemID)

		if err != nil {
			continue
		}

		updateQuery := `
			UPDATE dossier_checklist_items
			SET workflow_status = 'pending_approval',
			    current_revision = $1,
			    status = 'valide_qualite',
			    maker_user_id = $2,
			    maker_submitted_at = $3,
			    checker_user_id = $4,
			    checker_reviewed_at = $5,
			    checker_decision = 'approved',
			    checker_notes = $6,
			    updated_at = $5
			WHERE id = $7
		`
		_, err = pool.Exec(ctx, updateQuery, s.currentRev, makerID, subTime, checkerID, revTime, s.checkerNotes, itemID)
		if err == nil {
			// Inscription dans l'audit trail dossier_item_reviews
			reviewQuery := `
				INSERT INTO dossier_item_reviews (dossier_item_id, author_id, author_role, review_type, comment, revision_version)
				VALUES ($1, $2, 'contributeur_qualite_bpf', 'quality_passed', $3, $4)
			`
			_, _ = pool.Exec(ctx, reviewQuery, itemID, checkerID, s.checkerNotes, s.currentRev)
		}
	}

	log.Println("[Seed Parapheur] Pièces de démonstration prêtes pour visa DT initialisées dans le parapheur.")
}
