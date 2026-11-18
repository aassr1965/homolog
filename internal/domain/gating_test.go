package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"homolog/internal/domain"
)

func TestGating_NoticeBlockedByStabilities(t *testing.T) {
	dossierID := uuid.New()

	items := []domain.ChecklistItem{
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     3,
			ItemCode:         "3.2.P.8",
			TitleFR:          "Stabilités Zone IVb",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusDrafting, // Pas encore scellé
		},
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     1,
			ItemCode:         "1.P.2",
			TitleFR:          "Notice Patient",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusPendingApproval,
		},
	}

	targetNotice := &items[1]

	// 1. Évaluation pour visa PDT : doit être BLOQUÉ
	res := domain.EvaluateItemGating(items, targetNotice, domain.GatingActionApprovePDT)
	if !res.IsBlocked {
		t.Fatalf("La notice 1.P.2 aurait dû être bloquée car 3.2.P.8 n'est pas scellée")
	}
	if len(res.UnmetPrerequisites) == 0 {
		t.Fatalf("Attendu au moins un prérequis manquant pour 1.P.2")
	}
	if res.UnmetPrerequisites[0].RequiredItemCode != "3.2.P.8" {
		t.Fatalf("Le prérequis attendu était 3.2.P.8, reçu %s", res.UnmetPrerequisites[0].RequiredItemCode)
	}

	// 2. Passage de 3.2.P.8 à approved_sealed : doit être DÉVERROUILLÉ
	items[0].WorkflowStatus = domain.WorkflowStatusApprovedSealed

	res2 := domain.EvaluateItemGating(items, targetNotice, domain.GatingActionApprovePDT)
	if res2.IsBlocked {
		t.Fatalf("La notice 1.P.2 devrait être déverrouillée maintenant que 3.2.P.8 est scellée : %v", res2.SummaryMessage)
	}
}

func TestGating_PariteLinguistiqueNoticeFRandAR(t *testing.T) {
	dossierID := uuid.New()

	items := []domain.ChecklistItem{
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     3,
			ItemCode:         "3.2.P.8",
			TitleFR:          "Stabilités Zone IVb",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusApprovedSealed, // Stabilités validées pour se focaliser sur la parité
		},
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     1,
			ItemCode:         "1.P.2.1",
			TitleFR:          "Notice en langue française",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusDrafting, // Encore en rédaction
		},
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     1,
			ItemCode:         "1.P.2.2",
			TitleFR:          "Notice en langue arabe",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusDrafting,
		},
	}

	targetNoticeAR := &items[2]

	// 1. Tenter de soumettre la notice AR en revue Checker alors que la notice FR est encore en draft
	res := domain.EvaluateItemGating(items, targetNoticeAR, domain.GatingActionSubmitReview)
	if !res.IsBlocked {
		t.Fatalf("La notice arabe aurait dû être bloquée pour la revue car la notice française n'est pas encore en revue")
	}

	// 2. Transmettre la notice FR au Checker
	items[1].WorkflowStatus = domain.WorkflowStatusUnderReview

	res2 := domain.EvaluateItemGating(items, targetNoticeAR, domain.GatingActionSubmitReview)
	if res2.IsBlocked {
		t.Fatalf("La notice arabe devrait être soumise à la revue une fois la notice française en revue : %v", res2.SummaryMessage)
	}

	// 3. Tenter d'approuver la notice AR par le PDT alors que la notice FR n'est pas encore approuvée
	res3 := domain.EvaluateItemGating(items, targetNoticeAR, domain.GatingActionApprovePDT)
	if !res3.IsBlocked {
		t.Fatalf("La notice arabe ne peut pas être scellée avant la notice française")
	}

	// 4. Scellement de la notice FR ➔ déverrouillage du scellement de la notice AR
	items[1].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	res4 := domain.EvaluateItemGating(items, targetNoticeAR, domain.GatingActionApprovePDT)
	if res4.IsBlocked {
		t.Fatalf("La notice arabe devrait pouvoir être scellée : %v", res4.SummaryMessage)
	}
}

func TestGating_QOSBlockedByModule3(t *testing.T) {
	dossierID := uuid.New()

	items := []domain.ChecklistItem{
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     3,
			ItemCode:         "3.2.S.1",
			TitleFR:          "Nomenclature Substance Active",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusApprovedSealed,
		},
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     3,
			ItemCode:         "3.2.P.1",
			TitleFR:          "Description Produit Fini",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusDrafting, // Pas encore scellé
		},
		{
			ID:               uuid.New(),
			DossierID:        dossierID,
			ModuleNumber:     2,
			ItemCode:         "2.3",
			TitleFR:          "Résumé Global de la Qualité (QOS)",
			RequirementLevel: domain.RequirementExiged,
			IsMandatory:      true,
			WorkflowStatus:   domain.WorkflowStatusDrafting,
		},
	}

	targetQOS := &items[2]

	// 1. Évaluation QOS : doit être bloqué par 3.2.P.1
	res := domain.EvaluateItemGating(items, targetQOS, domain.GatingActionApprovePDT)
	if !res.IsBlocked {
		t.Fatalf("Le QOS (2.3) aurait dû être bloqué car 3.2.P.1 n'est pas scellé")
	}

	// 2. Scellement de 3.2.P.1 : déverrouillage du QOS
	items[1].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	res2 := domain.EvaluateItemGating(items, targetQOS, domain.GatingActionApprovePDT)
	if res2.IsBlocked {
		t.Fatalf("Le QOS devrait être déverrouillé maintenant que tout le Module 3 est scellé : %v", res2.SummaryMessage)
	}
}

func TestGating_DAGPhaseProgression(t *testing.T) {
	dossierID := uuid.New()

	items := []domain.ChecklistItem{
		{ModuleNumber: 3, ItemCode: "3.2.S.4", RequirementLevel: domain.RequirementExiged, IsMandatory: true, WorkflowStatus: domain.WorkflowStatusDrafting, DossierID: dossierID},
		{ModuleNumber: 3, ItemCode: "3.2.P.3", RequirementLevel: domain.RequirementExiged, IsMandatory: true, WorkflowStatus: domain.WorkflowStatusDrafting, DossierID: dossierID},
		{ModuleNumber: 3, ItemCode: "3.2.P.8", RequirementLevel: domain.RequirementExiged, IsMandatory: true, WorkflowStatus: domain.WorkflowStatusDrafting, DossierID: dossierID},
		{ModuleNumber: 1, ItemCode: "1.P.2", RequirementLevel: domain.RequirementExiged, IsMandatory: true, WorkflowStatus: domain.WorkflowStatusDrafting, DossierID: dossierID},
	}

	// Initialement : Phase 1
	if phase := domain.CalculateDossierDAGPhase(items); phase != 1 {
		t.Fatalf("Phase attendue 1, reçu %d", phase)
	}

	// Complétion Phase 1 (3.2.S.4) -> Passe à Phase 2
	items[0].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	if phase := domain.CalculateDossierDAGPhase(items); phase != 2 {
		t.Fatalf("Phase attendue 2, reçu %d", phase)
	}

	// Complétion Phase 2 (3.2.P.3) -> Passe à Phase 3
	items[1].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	if phase := domain.CalculateDossierDAGPhase(items); phase != 3 {
		t.Fatalf("Phase attendue 3, reçu %d", phase)
	}

	// Complétion Phase 3 (3.2.P.8) -> Passe à Phase 4
	items[2].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	if phase := domain.CalculateDossierDAGPhase(items); phase != 4 {
		t.Fatalf("Phase attendue 4, reçu %d", phase)
	}

	// Complétion Phase 4 (1.P.2) -> Passe à Phase 5
	items[3].WorkflowStatus = domain.WorkflowStatusApprovedSealed
	if phase := domain.CalculateDossierDAGPhase(items); phase != 5 {
		t.Fatalf("Phase attendue 5, reçu %d", phase)
	}
}
