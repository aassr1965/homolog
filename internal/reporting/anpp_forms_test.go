package reporting

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"homolog/internal/domain"
)

func readZipEntry(t *testing.T, zipBytes []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("archive zip invalide (%s): %v", name, err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("ouverture entrée %s: %v", name, err)
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("lecture entrée %s: %v", name, err)
			}
			return string(b)
		}
	}
	t.Fatalf("entrée %s introuvable dans l'archive", name)
	return ""
}

func TestListANPPTemplates(t *testing.T) {
	list := ListANPPTemplates()
	if len(list) < 13 {
		t.Fatalf("attendu au moins 13 gabarits ANPP embarqués, trouvé %d: %v", len(list), list)
	}
}

func TestGenerateDDMHomologationDOCX(t *testing.T) {
	org := &domain.Organization{
		LegalName:             "Laboratoire Biocare Algérie",
		MIPHLicenseNumber:     "MIPH-2026-ALG-001",
		TechnicalDirectorName: "Dr. Yacine Benkhalid",
	}
	dm := &domain.MedicalDevice{
		CommercialNameFR:         "Seringue Stérile Sanad 5mL",
		IntendedUseFR:            "Injection parentérale à usage unique",
		DMClass:                  "IIa",
		IsSterile:                true,
		LegalManufacturerName:    "MedTech GmbH",
		LegalManufacturerCountry: "Allemagne",
		GMDNCode:                 "47017",
	}

	out, err := GenerateDDMHomologationDOCX(org, dm)
	if err != nil {
		t.Fatalf("GenerateDDMHomologationDOCX erreur: %v", err)
	}
	docXML := readZipEntry(t, out, "word/document.xml")
	for _, want := range []string{"Seringue Stérile Sanad 5mL", "Laboratoire Biocare Algérie", "☑ Classe IIa", "☑ OUI", "47017"} {
		if !strings.Contains(docXML, want) {
			t.Errorf("valeur attendue absente de word/document.xml: %q", want)
		}
	}
}

func TestGenerateDDMRecepisseDOCX(t *testing.T) {
	org := &domain.Organization{
		LegalName:             "Laboratoire Biocare Algérie",
		Address:               "Zone Industrielle Rouiba, Alger",
		Phone:                 "+213 23 85 10 00",
		Email:                 "reg@biocare.dz",
		TechnicalDirectorName: "Dr. Yacine Benkhalid",
	}
	dm := &domain.MedicalDevice{
		CommercialNameFR:      "Cathéter Veineux 20G",
		IntendedUseFR:         "Abord veineux périphérique",
		DMClass:               "IIa",
		ManufacturerReference: "REF-CATH-20G",
	}
	eval := &domain.DMRecevabiliteEvaluation{
		DepositOrderNumber: "DDM-2026-042",
		ANPPAgentName:      "Mme. Amina K.",
		Item1Status:        "oui",
		Item2Status:        "oui",
		Item3Status:        "na",
		Item4Status:        "non",
	}

	out, err := GenerateDDMRecepisseDOCX(org, dm, eval)
	if err != nil {
		t.Fatalf("GenerateDDMRecepisseDOCX erreur: %v", err)
	}
	docXML := readZipEntry(t, out, "word/document.xml")
	for _, want := range []string{"Cathéter Veineux 20G", "Laboratoire Biocare Algérie", "reg@biocare.dz", "☑"} {
		if !strings.Contains(docXML, want) {
			t.Errorf("valeur attendue absente de ddm_recepisse_depot.docx: %q", want)
		}
	}
}

func TestGenerateDMCanevasNonHomologuesXLSX(t *testing.T) {
	org := &domain.Organization{
		LegalName:         "Biocare Import",
		Address:           "Alger Centre",
		MIPHLicenseNumber: "IMP-2025-99",
	}
	regs := []domain.DMRegularization{
		{
			RegularizationCode:       "REG-DM-001",
			Status:                   "depose_ddm",
			DDMAcknowledgementNumber: "AR-DDM-2026-11",
			QuarterlyVolumeImported:  15000,
			LastDeclarationDate:      time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
			MedicalDevice: &domain.MedicalDevice{
				CommercialNameFR: "Gants Chirurgicaux Stériles",
				IntendedUseFR:    "Protection chirurgicale",
				DMClass:          "IIa",
			},
		},
	}

	out, err := GenerateDMCanevasNonHomologuesXLSX(org, regs)
	if err != nil {
		t.Fatalf("GenerateDMCanevasNonHomologuesXLSX erreur: %v", err)
	}
	sheetXML := readZipEntry(t, out, "xl/worksheets/sheet1.xml")
	for _, want := range []string{"Biocare Import", "Gants Chirurgicaux Stériles", "AR-DDM-2026-11", "15000"} {
		if !strings.Contains(sheetXML, want) {
			t.Errorf("valeur attendue absente de sheet1.xml: %q", want)
		}
	}
}

func TestGenerateDDMBordereauAndDEPPDOCX(t *testing.T) {
	org := &domain.Organization{
		LegalName:             "Sanad Pharma Lab",
		TechnicalDirectorName: "Dr. Yacine Benkhalid",
		Phone:                 "021000000",
		Email:                 "pdt@pharma.dz",
	}
	dm := &domain.MedicalDevice{
		CommercialNameFR: "Stent Coronaire",
		IntendedUseFR:    "Cardiologie interventionnelle",
		DMClass:          "III",
	}

	for _, kind := range []string{"modification", "renouvellement", "transfert"} {
		b, err := GenerateDDMBordereauDOCX(kind, org, dm, "BV-2026-009", "Ajout site conditionnement")
		if err != nil {
			t.Fatalf("GenerateDDMBordereauDOCX(%s) erreur: %v", kind, err)
		}
		xmlStr := readZipEntry(t, b, "word/document.xml")
		if !strings.Contains(xmlStr, "BV-2026-009") {
			t.Errorf("BV %s: référence absente", kind)
		}
	}

	dossier := &domain.Dossier{
		DossierCode:   "DOS-2026-PARAC-NF",
		ProcedureCode: domain.ProcedureCode("NF"),
		Medicine: &domain.Medicine{
			BrandNameFR:          "Paracétamol SANAD",
			DCIFR:                "Paracétamol",
			DosageFR:             "500 mg",
			PharmaceuticalFormFR: "Comprimé",
		},
	}
	for _, kind := range []string{"enregistrement", "modification", "renouvellement", "transfert", "recepisse", "echantillons"} {
		b, err := GenerateDEPPFormDOCX(kind, dossier, org)
		if err != nil {
			t.Fatalf("GenerateDEPPFormDOCX(%s) erreur: %v", kind, err)
		}
		_ = readZipEntry(t, b, "word/document.xml")
	}
}
