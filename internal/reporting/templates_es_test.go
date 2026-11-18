package reporting

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"homolog/internal/domain"
)

func makeTestDossier() *domain.Dossier {
	now := time.Now().UTC()
	docID := uuid.New()
	ref := "AMM-DZ-2026-0891"
	return &domain.Dossier{
		ID:                    uuid.New(),
		OrganizationID:        uuid.New(),
		DossierCode:           "DOS-2026-PARAC-NF",
		ProcedureCode:         "NF",
		ProductType:           domain.ProductTypeMedicine,
		Status:                domain.DossierStatusDraft,
		CompletenessPercent:   75,
		ANPPOfficialReference: &ref,
		Medicine: &domain.Medicine{
			ID:                   uuid.New(),
			BrandNameFR:          "PARACÉTAMOL LPA 500 mg",
			DCIFR:                "Paracétamol",
			DosageFR:             "500 mg",
			PharmaceuticalFormFR: "Comprimé pelliculé",
			ATCCode:              "N02BE01",
		},
		Items: []domain.ChecklistItem{
			{
				ID:                 uuid.New(),
				ModuleNumber:       1,
				ItemCode:           "1.0",
				TitleFR:            "Lettre de candidature et bordereau de transmission",
				RequirementLevel:   domain.RequirementExiged,
				AssignedDepartment: "RA",
				GEDDocumentID:      &docID,
			},
			{
				ID:                 uuid.New(),
				ModuleNumber:       1,
				ItemCode:           "1.P.8",
				TitleFR:            "Engagement de disponibilité continue sur le marché national",
				RequirementLevel:   domain.RequirementExiged,
				AssignedDepartment: "RA",
				GEDDocumentID:      &docID,
			},
			{
				ID:                 uuid.New(),
				ModuleNumber:       3,
				ItemCode:           "3.2.P.1",
				TitleFR:            "Description et composition du produit fini",
				RequirementLevel:   domain.RequirementExiged,
				AssignedDepartment: "QC",
				GEDDocumentID:      &docID,
			},
			{
				ID:                 uuid.New(),
				ModuleNumber:       5,
				ItemCode:           "5.3.1.2",
				TitleFR:            "Étude comparative de bioéquivalence in vivo",
				RequirementLevel:   domain.RequirementPartial,
				AssignedDepartment: "RD",
				GEDDocumentID:      nil,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestRenderESREG03HTML(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	html := RenderESREG03HTML(d, "LABORATOIRES SAIDAL", "MIPH-AGR-089", "Alger", "Dr. BENALI", "MIPH/DT/089", domain.WatermarkOfficial, hash, qrURL, now)

	if !strings.Contains(html, "ES-REG-03") {
		t.Errorf("Code ES-REG-03 manquant dans le rendu")
	}
	if !strings.Contains(html, "PARACÉTAMOL LPA 500 mg") {
		t.Errorf("Dénomination produit manquante")
	}
	if !strings.Contains(html, "DOS-2026-PARAC-NF") {
		t.Errorf("Code dossier manquant")
	}
	if !strings.Contains(html, "M1 • 1.0") || !strings.Contains(html, "M1 • 1.P.8") {
		t.Errorf("Sections CTD manquantes dans le tableau")
	}
	if !strings.Contains(html, "Page 1 sur 1") {
		t.Errorf("Pagination légale manquante")
	}
}

func TestRenderESREG04HTML(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	html := RenderESREG04HTML(d, "Rubrique 12.a", "Mmod", "Durée 36 mois", "Prolongation 48 mois", "Stabilités IVb", "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkDraft, hash, qrURL, now)

	if !strings.Contains(html, "ES-REG-04") {
		t.Errorf("Code ES-REG-04 manquant")
	}
	if !strings.Contains(html, "LD-DEPP07/V01") {
		t.Errorf("Référence LD-DEPP07/V01 manquante")
	}
	if !strings.Contains(html, "Rubrique 12.a") {
		t.Errorf("Rubrique de variation manquante")
	}
	if !strings.Contains(html, "Situation Approuvée Antérieure") {
		t.Errorf("Tableau comparatif manquant")
	}
}

func TestRenderESREG05HTML(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	html := RenderESREG05HTML(d, "LOT-PARAC-099", "02/2026", "02/2029", "SCL-LNCPP-DZ-01", "LOG-IVB-01", "+15°C à +25°C", 300, "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkDraft, hash, qrURL, now)

	if !strings.Contains(html, "ES-REG-05") {
		t.Errorf("Code ES-REG-05 manquant")
	}
	if !strings.Contains(html, "Laboratoire National de Contrôle (LNCPP)") {
		t.Errorf("Mention LNCPP manquante")
	}
	if !strings.Contains(html, "SCL-LNCPP-DZ-01") {
		t.Errorf("Numéro de scellé manquant")
	}
	if !strings.Contains(html, "LOG-IVB-01") {
		t.Errorf("Identifiant datalogger manquant")
	}
}

func TestRenderESFIN02HTML(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	// Cas Renouvellement Médicament (205 000 DA)
	html := RenderESFIN02HTML(d, "RENOUVELLEMENT_QUINQUENNAL", 205000.0, "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkOfficial, hash, qrURL, now)

	if !strings.Contains(html, "ES-FIN-02") {
		t.Errorf("Code ES-FIN-02 manquant")
	}
	if !strings.Contains(html, "205000.00 DA") && !strings.Contains(html, "205 000") {
		t.Errorf("Montant 205 000 DA manquant dans le bordereau")
	}
	if !strings.Contains(strings.ToLower(html), "deux cent cinq mille dinars algériens") {
		t.Errorf("Montant en lettres incorrect: %s", html)
	}
	if !strings.Contains(html, "Volet 1 : Demandeur") || !strings.Contains(html, "Volet 3 : Régie ANPP") {
		t.Errorf("Structure tripartie manquante")
	}
}

func TestRenderCommitmentLetters(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	// 1. Lettre 1.P.8 (Disponibilité continue)
	l1P8 := RenderCommitmentLetter1P8(d, "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkDraft, hash, qrURL, now)
	if !strings.Contains(l1P8, "Module 1.P.8") {
		t.Errorf("Module 1.P.8 manquant")
	}
	if !strings.Contains(l1P8, "trois (03) mois de consommation nationale moyenne") {
		t.Errorf("Engagement stock 3 mois manquant")
	}
	if !strings.Contains(l1P8, "six (06) mois à l'avance") {
		t.Errorf("Délai de notification 6 mois manquant")
	}

	// 2. Lettre 1.P.9 (Pharmacovigilance)
	l1P9 := RenderCommitmentLetter1P9(d, "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkDraft, hash, qrURL, now)
	if !strings.Contains(l1P9, "Module 1.P.9") {
		t.Errorf("Module 1.P.9 manquant")
	}
	if !strings.Contains(l1P9, "Centre National de Pharmacovigilance") {
		t.Errorf("Mention CNPM manquante")
	}
	if !strings.Contains(l1P9, "quinze (15) jours calendaires") {
		t.Errorf("Délai 15 jours manquant")
	}

	// 3. Lettre 1.H.10 (Non-modification)
	l1H10 := RenderCommitmentLetter1H10(d, "SAIDAL", "AGR-089", "Alger", "Dr. BENALI", "DT/089", domain.WatermarkDraft, hash, qrURL, now)
	if !strings.Contains(l1H10, "Module 1.H.10") {
		t.Errorf("Module 1.H.10 manquant")
	}
	if !strings.Contains(l1H10, "Attestation d'Inaltérabilité") {
		t.Errorf("Titre déclaration inaltérabilité manquant")
	}
	if !strings.Contains(strings.ToLower(l1H10), "aucune modification d'aucune nature") && !strings.Contains(strings.ToLower(l1H10), "aucune modification d&#39;aucune nature") {
		t.Errorf("Engagement de non-modification manquant")
	}
}
