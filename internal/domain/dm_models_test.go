package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"homolog/internal/domain"
)

func TestDMRecevabiliteEvaluation_CalculateConformity(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*domain.DMRecevabiliteEvaluation)
		wantScore    int
		wantDecision string
	}{
		{
			name: "13 points conformes donne complet_et_recevable",
			setup: func(e *domain.DMRecevabiliteEvaluation) {
				e.Item1Status = domain.RecevabiliteStatusConforme
				e.Item2Status = domain.RecevabiliteStatusConforme
				e.Item3Status = domain.RecevabiliteStatusConforme
				e.Item4Status = domain.RecevabiliteStatusConforme
				e.Item5Status = domain.RecevabiliteStatusConforme
				e.Item6Status = domain.RecevabiliteStatusConforme
				e.Item7Status = domain.RecevabiliteStatusConforme
				e.Item8Status = domain.RecevabiliteStatusConforme
				e.Item9Status = domain.RecevabiliteStatusConforme
				e.Item10Status = domain.RecevabiliteStatusConforme
				e.Item11Status = domain.RecevabiliteStatusConforme
				e.Item12Status = domain.RecevabiliteStatusConforme
				e.Item13Status = domain.RecevabiliteStatusConforme
			},
			wantScore:    13,
			wantDecision: domain.RecevabiliteDecisionComplet,
		},
		{
			name: "1 réserve donne incomplet_avec_reserves",
			setup: func(e *domain.DMRecevabiliteEvaluation) {
				e.Item1Status = domain.RecevabiliteStatusConforme
				e.Item2Status = domain.RecevabiliteStatusConforme
				e.Item3Status = domain.RecevabiliteStatusConforme
				e.Item4Status = domain.RecevabiliteStatusConforme
				e.Item5Status = domain.RecevabiliteStatusConforme
				e.Item6Status = domain.RecevabiliteStatusConforme
				e.Item7Status = domain.RecevabiliteStatusConforme
				e.Item8Status = domain.RecevabiliteStatusReserve // réserve sur matière première
				e.Item9Status = domain.RecevabiliteStatusConforme
				e.Item10Status = domain.RecevabiliteStatusConforme
				e.Item11Status = domain.RecevabiliteStatusConforme
				e.Item12Status = domain.RecevabiliteStatusConforme
				e.Item13Status = domain.RecevabiliteStatusConforme
			},
			wantScore:    12,
			wantDecision: domain.RecevabiliteDecisionReserves,
		},
		{
			name: "1 non conforme donne irrecevable",
			setup: func(e *domain.DMRecevabiliteEvaluation) {
				e.Item1Status = domain.RecevabiliteStatusConforme
				e.Item2Status = domain.RecevabiliteStatusConforme
				e.Item3Status = domain.RecevabiliteStatusConforme
				e.Item4Status = domain.RecevabiliteStatusNonConforme // redevance non payée
				e.Item5Status = domain.RecevabiliteStatusConforme
				e.Item6Status = domain.RecevabiliteStatusConforme
				e.Item7Status = domain.RecevabiliteStatusConforme
				e.Item8Status = domain.RecevabiliteStatusConforme
				e.Item9Status = domain.RecevabiliteStatusConforme
				e.Item10Status = domain.RecevabiliteStatusConforme
				e.Item11Status = domain.RecevabiliteStatusConforme
				e.Item12Status = domain.RecevabiliteStatusConforme
				e.Item13Status = domain.RecevabiliteStatusConforme
			},
			wantScore:    12,
			wantDecision: domain.RecevabiliteDecisionIrrecevable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eval := &domain.DMRecevabiliteEvaluation{}
			tc.setup(eval)
			eval.CalculateConformity()

			if eval.ConformityScore != tc.wantScore {
				t.Errorf("ConformityScore = %d, want %d", eval.ConformityScore, tc.wantScore)
			}
			if eval.Decision != tc.wantDecision {
				t.Errorf("Decision = %s, want %s", eval.Decision, tc.wantDecision)
			}
		})
	}
}

func TestDMRecevabiliteEvaluation_GenerateSeal(t *testing.T) {
	eval := &domain.DMRecevabiliteEvaluation{
		ID:                 uuid.New(),
		DepositOrderNumber: "ANPP/DDM/REC/2026/0148",
		ConformityScore:    13,
		Decision:           domain.RecevabiliteDecisionComplet,
	}

	seal := eval.GenerateSeal("pdt@pharma.dz", "DEC-MIPH-2022-7714")
	if !strings.HasPrefix(seal, "DM-SEAL-REC-") {
		t.Errorf("Seal invalide : %s, attendu préfixe 'DM-SEAL-REC-'", seal)
	}
	if len(seal) < 20 {
		t.Errorf("Seal trop court : %s", seal)
	}
	if eval.PDTSignatureSeal != seal {
		t.Errorf("PDTSignatureSeal non assigné correctement : %s vs %s", eval.PDTSignatureSeal, seal)
	}
}

func TestDMRegularization_CalculateUrgency(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		reg         domain.DMRegularization
		wantUrgency string
	}{
		{
			name: "Homologué donne statut regularise",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusHomologueRegularise,
				LegalDeadline: now.Add(10 * 24 * time.Hour),
			},
			wantUrgency: "regularise",
		},
		{
			name: "Délai <= 30j donne critique",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusCommercialiseDerogatoire,
				LegalDeadline: now.Add(20 * 24 * time.Hour),
			},
			wantUrgency: "critique",
		},
		{
			name: "Délai <= 60j donne urgent",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusDossierDeposeDDM,
				LegalDeadline: now.Add(45 * 24 * time.Hour),
			},
			wantUrgency: "urgent",
		},
		{
			name: "Délai <= 90j donne alerte",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusEnInstruction,
				LegalDeadline: now.Add(75 * 24 * time.Hour),
			},
			wantUrgency: "alerte",
		},
		{
			name: "Délai > 90j donne dans_les_delais",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusCommercialiseDerogatoire,
				LegalDeadline: now.Add(120 * 24 * time.Hour),
			},
			wantUrgency: "dans_les_delais",
		},
		{
			name: "Délai dépassé donne expire",
			reg: domain.DMRegularization{
				Status:        domain.DMRegStatusCommercialiseDerogatoire,
				LegalDeadline: now.Add(-5 * 24 * time.Hour),
			},
			wantUrgency: "expire",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.reg
			r.CalculateUrgency()
			if r.UrgencyLevel != tc.wantUrgency {
				t.Errorf("UrgencyLevel = %s, want %s (days: %d)", r.UrgencyLevel, tc.wantUrgency, r.DaysRemaining)
			}
		})
	}
}

func TestDMRegularization_Validate(t *testing.T) {
	validReg := domain.DMRegularization{
		MedicalDeviceID:           uuid.New(),
		RegularizationCode:        "REG-DM-2026-001",
		Status:                    domain.DMRegStatusCommercialiseDerogatoire,
		HospitalDistributionShare: 70,
		RetailDistributionShare:   30,
	}

	if err := validReg.Validate(); err != nil {
		t.Errorf("Validation attendue sans erreur, reçu: %v", err)
	}

	// Code manquant
	badReg := validReg
	badReg.RegularizationCode = ""
	if err := badReg.Validate(); err == nil {
		t.Error("Attendu erreur pour code manquant")
	}

	// Device ID vide
	badReg = validReg
	badReg.MedicalDeviceID = uuid.Nil
	if err := badReg.Validate(); err == nil {
		t.Error("Attendu erreur pour MedicalDeviceID nil")
	}

	// Répartition != 100%
	badReg = validReg
	badReg.HospitalDistributionShare = 60
	badReg.RetailDistributionShare = 30 // somme 90
	if err := badReg.Validate(); err == nil {
		t.Error("Attendu erreur pour répartition != 100%")
	}

	// Statut invalide
	badReg = validReg
	badReg.Status = "statut_inconnu"
	if err := badReg.Validate(); err == nil {
		t.Error("Attendu erreur pour statut invalide")
	}
}
