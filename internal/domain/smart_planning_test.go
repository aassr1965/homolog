package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDetermineTargetDepositDate(t *testing.T) {
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	// Cas 1 : target_deposit_date explicite
	explicitDate := time.Date(2027, 8, 15, 0, 0, 0, 0, time.UTC)
	d1 := &Dossier{
		ProcedureCode:     ProcedureNI,
		TargetDepositDate: &explicitDate,
	}
	target1, origin1, isEst1 := DetermineTargetDepositDate(d1, now)
	if !target1.Equal(explicitDate) || origin1 != "dossier_target_date" || isEst1 {
		t.Fatalf("Cas 1 échoué : obtenu %v, origin=%s, isEst=%v", target1, origin1, isEst1)
	}

	// Cas 2 : Renouvellement avec date de fin de validité AMM (J-180 légal décret 20-325)
	validityEnd := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	d2 := &Dossier{
		ProcedureCode:       ProcedureRI,
		ANPPValidityEndDate: &validityEnd,
	}
	target2, origin2, isEst2 := DetermineTargetDepositDate(d2, now)
	expectedRenewalTarget := validityEnd.AddDate(0, 0, -180)
	if !target2.Equal(expectedRenewalTarget) || origin2 != "anpp_validity_j180" || isEst2 {
		t.Fatalf("Cas 2 échoué : attendu %v, obtenu %v, origin=%s", expectedRenewalTarget, target2, origin2)
	}

	// Cas 3 : Date estimée par défaut (+270 jours)
	d3 := &Dossier{
		ProcedureCode: ProcedureNI,
		CreatedAt:     now,
	}
	target3, origin3, isEst3 := DetermineTargetDepositDate(d3, now)
	expectedEstimated := now.AddDate(0, 0, 270)
	if !target3.Equal(expectedEstimated) || origin3 != "default_estimated" || !isEst3 {
		t.Fatalf("Cas 3 échoué : attendu %v, obtenu %v, origin=%s, isEst=%v", expectedEstimated, target3, origin3, isEst3)
	}
}

func TestEvaluateSmartPlanning_OnTrack(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	targetDeposit := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC) // ~304 jours dans le futur (> 240j)

	d := &Dossier{
		ID:                uuid.New(),
		DossierCode:       "DOS-2026-TEST-NI",
		ProcedureCode:     ProcedureNI,
		TargetDepositDate: &targetDeposit,
	}

	items := []ChecklistItem{
		{ItemCode: "3.2.P.8", TitleFR: "Stabilités", WorkflowStatus: WorkflowStatusDraftPending},
		{ItemCode: "2.3", TitleFR: "QOS", WorkflowStatus: WorkflowStatusDraftPending},
		{ItemCode: "1.P.2.1", TitleFR: "Notice FR", WorkflowStatus: WorkflowStatusDraftPending},
		{ItemCode: "1.P.2.2", TitleFR: "Notice AR", WorkflowStatus: WorkflowStatusDraftPending},
	}

	analysis := EvaluateSmartPlanning(d, items, now)

	if analysis.OverallHealth != HealthOnTrack {
		t.Fatalf("Attendu HealthOnTrack, obtenu %s", analysis.OverallHealth)
	}
	if len(analysis.Milestones) != 6 {
		t.Fatalf("Attendu 6 jalons, obtenu %d", len(analysis.Milestones))
	}

	// Vérification du J-240
	mJ240 := analysis.Milestones[0]
	if mJ240.Code != MilestoneJ240StabilitiesZoneIVb {
		t.Fatalf("Attendu jalon J-240, obtenu %s", mJ240.Code)
	}
	expectedJ240Date := targetDeposit.AddDate(0, 0, -240)
	if !mJ240.TargetDate.Equal(expectedJ240Date) {
		t.Fatalf("Attendu date J-240 %v, obtenu %v", expectedJ240Date, mJ240.TargetDate)
	}
	if mJ240.DaysRemaining <= 30 {
		t.Fatalf("J-240 devrait avoir plus de 30 jours restants, obtenu %d", mJ240.DaysRemaining)
	}
}

func TestEvaluateSmartPlanning_CriticalDelayAndFatalBlock(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	// Date de dépôt dans 60 jours : J-240 et J-150 sont largement dépassés !
	targetDeposit := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	d := &Dossier{
		ID:                uuid.New(),
		DossierCode:       "DOS-2026-RETARD-NI",
		ProcedureCode:     ProcedureNI,
		TargetDepositDate: &targetDeposit,
	}

	// Stabilités pas encore lancées (draft_pending)
	items := []ChecklistItem{
		{ItemCode: "3.2.P.8", TitleFR: "Stabilités", WorkflowStatus: WorkflowStatusDraftPending},
		{ItemCode: "2.3", TitleFR: "QOS", WorkflowStatus: WorkflowStatusDraftPending},
	}

	analysis := EvaluateSmartPlanning(d, items, now)

	// Doit être en FatalBlock car stabilités M6 non lancées alors qu'on est à J-60 du dépôt !
	if analysis.OverallHealth != HealthFatalBlock {
		t.Fatalf("Attendu HealthFatalBlock à cause des stabilités M6 impossibles à conclure, obtenu %s", analysis.OverallHealth)
	}
	if len(analysis.CriticalWarnings) == 0 {
		t.Fatalf("Attendu des alertes critiques dans l'analyse")
	}

	// Vérifier que le jalon J-240 est en FATAL_BLOCK
	mJ240 := analysis.Milestones[0]
	if mJ240.Health != HealthFatalBlock {
		t.Fatalf("J-240 devrait être en FATAL_BLOCK, obtenu %s", mJ240.Health)
	}
}

func TestEvaluateSmartPlanning_CompletedMilestones(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	targetDeposit := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	trackingNum := "ETASJIL-2026-9941"
	pdtSigned := now.AddDate(0, 0, -2)
	d := &Dossier{
		ID:                 uuid.New(),
		DossierCode:        "DOS-2026-ADVANCED-NI",
		ProcedureCode:      ProcedureNI,
		TargetDepositDate:  &targetDeposit,
		TrackingNumber:     &trackingNum,
		PDTSignedAt:        &pdtSigned,
		IsLockedForDeposit: true,
	}

	items := []ChecklistItem{
		{ItemCode: "3.2.P.8", TitleFR: "Stabilités", WorkflowStatus: WorkflowStatusApprovedSealed},
		{ItemCode: "2.3", TitleFR: "QOS", WorkflowStatus: WorkflowStatusApprovedSealed},
		{ItemCode: "1.P.2.1", TitleFR: "Notice FR", WorkflowStatus: WorkflowStatusApprovedSealed},
		{ItemCode: "1.P.2.2", TitleFR: "Notice AR", WorkflowStatus: WorkflowStatusApprovedSealed},
	}

	analysis := EvaluateSmartPlanning(d, items, now)

	// Vérifier que les jalons passés sont marqués COMPLETED
	for _, m := range analysis.Milestones {
		if m.Code != MilestoneJ0ANPPDepositDeadline {
			if !m.IsCompleted || m.Health != HealthCompleted {
				t.Fatalf("Le jalon %s devrait être COMPLETED, obtenu isCompleted=%v, health=%s", m.Code, m.IsCompleted, m.Health)
			}
		}
	}
}

func TestCalculateSmartDueDates(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	targetDeposit := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	d := &Dossier{
		TargetDepositDate: &targetDeposit,
	}

	items := []ChecklistItem{
		{ItemCode: "3.2.P.8", TitleFR: "Stabilités"},
		{ItemCode: "2.3", TitleFR: "QOS"},
		{ItemCode: "1.A.1", TitleFR: "Formulaire"},
		{ItemCode: "1.P.1", TitleFR: "RCP"},
		{ItemCode: "1.A.7", TitleFR: "Visa PDT"},
	}

	analysis := EvaluateSmartPlanning(d, items, now)
	dueDates := CalculateSmartDueDates(analysis, items)

	// 3.2.P.8 -> J-240
	if !dueDates["3.2.P.8"].Equal(targetDeposit.AddDate(0, 0, -240)) {
		t.Fatalf("Échéance erronée pour 3.2.P.8 : attendu %v, obtenu %v", targetDeposit.AddDate(0, 0, -240), dueDates["3.2.P.8"])
	}
	// 2.3 -> J-150
	if !dueDates["2.3"].Equal(targetDeposit.AddDate(0, 0, -150)) {
		t.Fatalf("Échéance erronée pour 2.3 : attendu %v, obtenu %v", targetDeposit.AddDate(0, 0, -150), dueDates["2.3"])
	}
	// 1.A.1 -> J-60
	if !dueDates["1.A.1"].Equal(targetDeposit.AddDate(0, 0, -60)) {
		t.Fatalf("Échéance erronée pour 1.A.1 : attendu %v, obtenu %v", targetDeposit.AddDate(0, 0, -60), dueDates["1.A.1"])
	}
	// 1.P.1 -> J-30
	if !dueDates["1.P.1"].Equal(targetDeposit.AddDate(0, 0, -30)) {
		t.Fatalf("Échéance erronée pour 1.P.1 : attendu %v, obtenu %v", targetDeposit.AddDate(0, 0, -30), dueDates["1.P.1"])
	}
	// 1.A.7 -> J-15
	if !dueDates["1.A.7"].Equal(targetDeposit.AddDate(0, 0, -15)) {
		t.Fatalf("Échéance erronée pour 1.A.7 : attendu %v, obtenu %v", targetDeposit.AddDate(0, 0, -15), dueDates["1.A.7"])
	}
}
