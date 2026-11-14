package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvaluatePreSubmission_IncompleteDossier(t *testing.T) {
	dossierID := uuid.New()
	dossier := &Dossier{
		ID:            dossierID,
		DossierCode:   "DOS-2026-TEST-NF",
		ProcedureCode: ProcedureNF,
		ProductType:   ProductTypeMedicament,
		Medicine: &Medicine{
			BrandNameFR: "DOLIPRANE 500mg",
			DCIFR:       "PARACETAMOL",
			DosageFR:    "500 mg",
		},
		Items: []ChecklistItem{
			{
				ItemCode:         "1.A.1",
				TitleFR:          "Formulaire officiel de demande d'AMM",
				RequirementLevel: RequirementExiged,
				IsMandatory:      true,
				Status:           ItemStatusMissing,
			},
		},
	}

	report := EvaluatePreSubmission(dossier, nil, nil, "dt@pharma.dz")

	if report.IsEligible {
		t.Errorf("Le dossier incomplet ne doit pas être éligible au dépôt")
	}
	if report.CriticalBlockersCount == 0 {
		t.Errorf("Des bloqueurs critiques devraient être détectés")
	}
	if report.CompletenessScore != 0 {
		t.Errorf("Attendu 0%% de complétude, reçu %d%%", report.CompletenessScore)
	}
}

func TestEvaluatePreSubmission_CompleteAndEligible(t *testing.T) {
	dossierID := uuid.New()
	pdtUserID := uuid.New()
	pdtRole := "Directeur Technique (1.A.7)"
	seal := "a1b2c3d4e5f67890abcdef1234567890"
	now := time.Now().UTC()

	docID := uuid.New()
	dossier := &Dossier{
		ID:            dossierID,
		DossierCode:   "DOS-2026-COMPLETE",
		ProcedureCode: ProcedureNF,
		ProductType:   ProductTypeMedicament,
		Medicine: &Medicine{
			BrandNameFR: "AMOXICILLINE 500mg",
			DCIFR:       "AMOXICILLINE",
			DosageFR:    "500 mg",
		},
		PDTSignedAt:      &now,
		PDTSignedBy:      &pdtUserID,
		PDTSignatureRole: &pdtRole,
		PDTSignatureSeal: &seal,
		Items: []ChecklistItem{
			{
				ItemCode:         "1.A.1",
				TitleFR:          "Formulaire officiel de demande",
				RequirementLevel: RequirementExiged,
				IsMandatory:      true,
				Status:           ItemStatusApprovedDT,
				GEDDocumentID:    &docID,
				GEDDocument: &GEDDocument{
					ID:             docID,
					DocumentCode:   "DOC-1A1-2026",
					CurrentVersion: "v1.0",
				},
			},
		},
	}

	est := &Establishment{
		NameFR:                       "Laboratoires Sanad Pharma Algérie",
		MIPHLicenseNumber:            "MIPH-02-2023-0042",
		TechnicalDirectorName:        "Dr. Karima T.",
		TechnicalDirectorDecisionRef: "DEC-MIPH-DT-2023/0286",
		Wilaya:                       "Alger",
	}

	receiptRef := "QUIT-TRESOR-84920"
	slips := []ANPPPaymentSlip{
		{
			ID:               uuid.New().String(),
			SlipNumber:       "BV1-2026-0001",
			SlipType:         "BV1",
			Status:           "paye",
			AmountDZD:        150000.0,
			AmountWordsFR:    "Cent cinquante mille dinars algériens",
			AmountWordsAR:    "مائة وخمسون ألف دينار جزائري",
			PaymentReference: receiptRef,
			PaidAt:           &now,
		},
	}

	report := EvaluatePreSubmission(dossier, slips, est, "dt@pharma.dz")

	if !report.IsEligible {
		t.Fatalf("Le dossier complet devrait être éligible, bloqueurs : %v", report.Checks)
	}
	if report.CriticalBlockersCount != 0 {
		t.Errorf("Attendu 0 bloqueurs critiques, reçu %d", report.CriticalBlockersCount)
	}
	if report.CompletenessScore != 100 {
		t.Errorf("Attendu 100%% de complétude, reçu %d%%", report.CompletenessScore)
	}

	// Test génération Manifeste E-Services
	manifest, err := BuildANPPEservicesManifest(dossier, slips, est, "dt@pharma.dz")
	if err != nil {
		t.Fatalf("Échec génération manifeste E-Services : %v", err)
	}
	if manifest.Applicant.AgreementNumber != est.MIPHLicenseNumber {
		t.Errorf("Numéro d'agrément mismatch")
	}
	if manifest.FinancialClearance.PaymentReference != receiptRef {
		t.Errorf("Référence de quittance mismatch")
	}
	if manifest.IntegritySeal.ManifestDigestHex == "" {
		t.Errorf("Le Digest SHA-256 du manifeste ne doit pas être vide")
	}

	jsonBytes, err := manifest.ToJSON()
	if err != nil || len(jsonBytes) == 0 {
		t.Errorf("Échec sérialisation JSON manifeste: %v", err)
	}

	xmlBytes, err := manifest.ToXML()
	if err != nil || len(xmlBytes) == 0 {
		t.Errorf("Échec sérialisation XML manifeste: %v", err)
	}
}
