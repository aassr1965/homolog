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

func TestGenerateDCELotAndDVTIVDOCX(t *testing.T) {
	org := &domain.Organization{
		LegalName:             "Biocare Importation",
		Address:               "Rouiba, Alger",
		Email:                 "import@biocare.dz",
		Phone:                 "+213 23 00 11 22",
		TechnicalDirectorName: "Dr. Yacine Benkhalid",
	}
	ddf := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	ddp := time.Date(2029, 1, 10, 0, 0, 0, 0, time.UTC)
	lot := &domain.ImportLotRelease{
		ProductKind:           "dm",
		CommercialName:        "Cathéter Veineux 20G",
		Designation:           "Abord veineux périphérique",
		FormAspect:            "Boîte stérile rose 20G",
		LotNumber:             "LOT-2026-881",
		DDF:                   &ddf,
		DDP:                   &ddp,
		SalesUnits:            "5000 unités",
		HomologationOrDERef:   "DH-2025-114",
		DMClass:               "IIa",
		ImportProgramAttached: true,
		ProductionSite:        "MedTech GmbH",
		QCSite:                "Labo QC GmbH",
		ReleaseSite:           "Biocare Alger",
		CustomsDeclNumber:     "D10-2026-441",
		OriginCertAttached:    true,
		TechFormAttached:      true,
		SamplingPlanAttached:  true,
	}

	dmOut, err := GenerateDCEDMLotReceptionDOCX(org, lot)
	if err != nil {
		t.Fatalf("GenerateDCEDMLotReceptionDOCX erreur: %v", err)
	}
	dmXML := readZipEntry(t, dmOut, "word/document.xml")
	for _, want := range []string{"Biocare Importation", "Cathéter Veineux 20G", "LOT-2026-881", "IIa  ☑", "Oui ☑"} {
		if !strings.Contains(dmXML, want) {
			t.Errorf("dce_reception_dm_importes.docx: attendu %q", want)
		}
	}

	ppOut, err := GenerateDCEPPLotReceptionDOCX(org, lot)
	if err != nil {
		t.Fatalf("GenerateDCEPPLotReceptionDOCX erreur: %v", err)
	}
	_ = readZipEntry(t, ppOut, "word/document.xml")

	certOut, err := GenerateDCECertificatLiberationPPDOCX(lot)
	if err != nil {
		t.Fatalf("GenerateDCECertificatLiberationPPDOCX erreur: %v", err)
	}
	if !strings.Contains(readZipEntry(t, certOut, "word/document.xml"), "LOT-2026-881") {
		t.Errorf("dce_certificat_liberation_pp.docx: LOT-2026-881 absent")
	}

	pdt := &domain.User{FirstName: "Dr. Yacine", LastName: "Benkhalid", Email: "pdt@pharma.dz"}
	asst := &domain.User{FirstName: "Dr. Lynda", LastName: "Amrani"}
	dvtivOut, err := GenerateDVTIVBasePDTQPPVDOCX(org, pdt, asst, nil, 4, true, false, "fabricant_local", "importateur")
	if err != nil {
		t.Fatalf("GenerateDVTIVBasePDTQPPVDOCX erreur: %v", err)
	}
	dvtivXML := readZipEntry(t, dvtivOut, "word/document.xml")
	dvtivText := extractXMLText(dvtivXML)
	for _, want := range []string{"<w:strike/>", "<w:keepNext/>"} {
		if !strings.Contains(dvtivXML, want) {
			t.Errorf("dvtiv_base_pdt_qppv.docx XML: attendu %q", want)
		}
	}
	for _, want := range []string{
		"Biocare Importation",
		"Rouiba, Alger",
		"Dr. Yacine Benkhalid",
		"Dr. Lynda Amrani",
		"Détenteur de DE",
		"Exploitant de DE",
		"Fabricant",
		"Importateur",
	} {
		if !strings.Contains(dvtivText, want) {
			t.Errorf("dvtiv_base_pdt_qppv.docx texte: attendu %q", want)
		}
	}
}

func TestGenerateAllAdditionalDCEForms(t *testing.T) {
	org := &domain.Organization{
		LegalName:         "Biocare Algérie",
		Address:           "Rouiba, Alger",
		Phone:             "+213 21 00 11 22",
		MIPHLicenseNumber: "MIPH-2026-01",
	}
	ddf := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	ddp := time.Date(2029, 2, 1, 0, 0, 0, 0, time.UTC)
	lots := []*domain.ImportLotRelease{
		{
			ProductKind:         "pp",
			CommercialName:      "Biomox 1g",
			Designation:         "Amoxicilline",
			FormAspect:          "Comprimé pelliculé",
			LotNumber:           "LOT-PP-01",
			DDF:                 &ddf,
			DDP:                 &ddp,
			SalesUnits:          "10000 boites",
			HomologationOrDERef: "DE-2026-001",
			ProductionSite:      "Site Prod Alger",
			QCSite:              "LCQ Alger",
			ReleaseSite:         "Site Libération Rouiba",
			CreatedAt:           ddf,
		},
		{
			ProductKind:         "dm",
			CommercialName:      "StentBio",
			Designation:         "Stent coronaire",
			FormAspect:          "Boite stérile",
			LotNumber:           "LOT-DM-02",
			DDF:                 &ddf,
			DDP:                 &ddp,
			SalesUnits:          "500 unités",
			HomologationOrDERef: "DH-2026-002",
			DMClass:             "III",
			ProductionSite:      "Site DM Oran",
			QCSite:              "LCQ Oran",
			ReleaseSite:         "Site Libération Oran",
			CreatedAt:           ddf,
		},
	}

	compOut, err := GenerateDCECompositionDossierDOCX()
	if err != nil || !strings.Contains(readZipEntry(t, compOut, "word/document.xml"), "<w:keepNext/>") {
		t.Fatalf("GenerateDCECompositionDossierDOCX invalide: err=%v", err)
	}

	canOut, err := GenerateDCECanevasCertificatLiberationDOCX(org, lots[0])
	if err != nil || !strings.Contains(extractXMLText(readZipEntry(t, canOut, "word/document.xml")), "LOT-PP-01") {
		t.Fatalf("GenerateDCECanevasCertificatLiberationDOCX invalide: err=%v", err)
	}

	for _, kind := range []string{"pp", "dm"} {
		dqOut, err := GenerateDCEDeclarationQuotidienneDOCX(kind, org, lots)
		if err != nil {
			t.Fatalf("GenerateDCEDeclarationQuotidienneDOCX(%s) erreur: %v", kind, err)
		}
		txt := extractXMLText(readZipEntry(t, dqOut, "word/document.xml"))
		if !strings.Contains(txt, "Biocare Algérie") {
			t.Errorf("decl quotidienne (%s): org absent", kind)
		}
	}

	xlsxPP, err := GenerateDCEDeclarationJournaliereXLSX("pp", org, lots)
	if err != nil || !strings.Contains(readZipEntry(t, xlsxPP, "xl/worksheets/sheet1.xml"), "LOT-PP-01") {
		t.Fatalf("GenerateDCEDeclarationJournaliereXLSX(pp) invalide: err=%v", err)
	}

	xlsxDM, err := GenerateDCEDeclarationJournaliereXLSX("dm", org, lots)
	if err != nil || !strings.Contains(readZipEntry(t, xlsxDM, "xl/worksheets/sheet1.xml"), "LOT-DM-02") {
		t.Fatalf("GenerateDCEDeclarationJournaliereXLSX(dm) invalide: err=%v", err)
	}

	bvHom, err := GenerateDDMBordereauDOCX("homologation", org, &domain.MedicalDevice{
		CommercialNameFR: "StentBio",
		IntendedUseFR:    "Stent coronaire",
		DMClass:          "III",
	}, "BV-2026-HOM", "Homologation DM")
	if err != nil || !strings.Contains(extractXMLText(readZipEntry(t, bvHom, "word/document.xml")), "StentBio") {
		t.Fatalf("GenerateDDMBordereauDOCX(homologation) invalide: err=%v", err)
	}

	xlsxBoth, err := GenerateDCEDeclarationJournaliereXLSX("pp_dm", org, lots)
	if err != nil {
		t.Fatalf("GenerateDCEDeclarationJournaliereXLSX(pp_dm) erreur: %v", err)
	}
	if !strings.Contains(readZipEntry(t, xlsxBoth, "xl/worksheets/sheet1.xml"), "LOT-PP-01") {
		t.Errorf("decl journaliere pp_dm sheet1: LOT-PP-01 absent")
	}
	if !strings.Contains(readZipEntry(t, xlsxBoth, "xl/worksheets/sheet2.xml"), "LOT-DM-02") {
		t.Errorf("decl journaliere pp_dm sheet2: LOT-DM-02 absent")
	}

	srOut, err := GenerateDCETemplateCASRSecondaireDOCX(org)
	if err != nil || !strings.Contains(extractXMLText(readZipEntry(t, srOut, "word/document.xml")), "Biocare Algérie") {
		t.Fatalf("GenerateDCETemplateCASRSecondaireDOCX invalide: err=%v", err)
	}

	dossier := &domain.Dossier{
		DossierCode:   "DOS-2026-DEPP",
		ProcedureCode: domain.ProcedureCode("NI"),
		Medicine: &domain.Medicine{
			BrandNameFR:          "Biomox 1g",
			DCIFR:                "Amoxicilline",
			DosageFR:             "1g",
			PharmaceuticalFormFR: "Comprimé pelliculé",
		},
	}
	for _, kind := range []string{"pre-soumission", "structure-prix", "recepisse-pp"} {
		out, err := GenerateDEPPFormDOCX(kind, dossier, org)
		if err != nil {
			t.Fatalf("GenerateDEPPFormDOCX(%s) erreur: %v", kind, err)
		}
		txt := extractXMLText(readZipEntry(t, out, "word/document.xml"))
		if kind != "structure-prix" && !strings.Contains(txt, "Biomox 1g") {
			t.Errorf("GenerateDEPPFormDOCX(%s): Biomox 1g absent", kind)
		}
		if kind == "recepisse-pp" && !strings.Contains(readZipEntry(t, out, "word/document.xml"), "<w:strike/>") {
			t.Errorf("GenerateDEPPFormDOCX(recepisse-pp): <w:strike/> attendu sur Fabrication locale")
		}
	}
}
