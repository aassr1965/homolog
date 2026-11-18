package reporting

import (
	"strings"
	"testing"
	"time"

	"homolog/internal/domain"
)

func TestMasterRegistry_CompletenessAndIntegrity(t *testing.T) {
	defs := GetAllReportDefinitions()
	if len(defs) < 28 {
		t.Fatalf("Le registre des états documentaires doit contenir au moins 28 définitions, reçu: %d", len(defs))
	}

	seenCodes := make(map[string]bool)
	esCount := 0
	eiCount := 0
	mtCount := 0

	for _, d := range defs {
		if strings.TrimSpace(d.Code) == "" {
			t.Errorf("Code d'état documentaire vide détecté: %+v", d)
		}
		if seenCodes[d.Code] {
			t.Errorf("Code en doublon détecté dans le registre: %s", d.Code)
		}
		seenCodes[d.Code] = true

		if strings.TrimSpace(d.Title) == "" {
			t.Errorf("Titre manquant pour l'état %s", d.Code)
		}
		if strings.TrimSpace(d.LegalReference) == "" {
			t.Errorf("Référence légale manquante pour l'état %s", d.Code)
		}
		if len(d.SupportedFormats) == 0 {
			t.Errorf("Aucun format supporté pour l'état %s", d.Code)
		}

		switch d.Kind {
		case KindEtatReglementaire:
			esCount++
			if !d.RequiresPDTSignature && d.Category != CategoryFinancesRedevances && d.Category != CategoryVigilanceComplements && d.Category != CategoryPilotageStrategique && d.Code != "ES-DCE-03" {
				t.Errorf("L'état réglementaire %s devrait requérir un visa légal", d.Code)
			}
		case KindEtatIntermediaire:
			eiCount++
			if d.RequiresPDTSignature {
				t.Errorf("L'état intermédiaire %s ne doit pas bloquer sur un visa régalien externe", d.Code)
			}
		case KindMasterTemplate:
			mtCount++
		}
	}

	if esCount < 19 {
		t.Errorf("Nombre insuffisant d'États Réglementaires (ES) trouvés: %d (attendu >= 19)", esCount)
	}
	if eiCount != 8 {
		t.Errorf("Nombre d'États Intermédiaires (EI) inattendu: %d (attendu exactement 8)", eiCount)
	}
}

func TestGetReportDefinition(t *testing.T) {
	// 1. Recherche insensible à la casse
	def, found := GetReportDefinition("es-reg-01")
	if !found {
		t.Fatalf("ES-REG-01 non trouvé avec minuscules")
	}
	if def.Code != "ES-REG-01" {
		t.Errorf("Attendu ES-REG-01, reçu %s", def.Code)
	}
	if def.Category != CategoryDossiersCTD {
		t.Errorf("Catégorie inattendue pour ES-REG-01: %s", def.Category)
	}

	// 2. Recherche état intermédiaire
	defEI, foundEI := GetReportDefinition("EI-AVT-01")
	if !foundEI {
		t.Fatalf("EI-AVT-01 introuvable")
	}
	if defEI.Kind != KindEtatIntermediaire {
		t.Errorf("Attendu KindEtatIntermediaire, reçu %s", defEI.Kind)
	}

	// 3. Recherche état inexistant
	_, foundNone := GetReportDefinition("ES-INCONNU-99")
	if foundNone {
		t.Errorf("Un code inexistant ne doit pas être trouvé")
	}
}

func TestGetReportsByCategoryAndKind(t *testing.T) {
	ctdReports := GetReportsByCategory(CategoryDossiersCTD)
	if len(ctdReports) == 0 {
		t.Errorf("Aucun rapport trouvé pour CategoryDossiersCTD")
	}

	eiReports := GetReportsByKind(KindEtatIntermediaire)
	if len(eiReports) != 8 {
		t.Errorf("Attendu 8 rapports pour KindEtatIntermediaire, reçu %d", len(eiReports))
	}
}

func TestGenerateQRCodeDataURI(t *testing.T) {
	// Cas normal
	dataURI, err := GenerateQRCodeDataURI("https://sanad.labo.dz/verify/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", 100)
	if err != nil {
		t.Fatalf("Erreur inattendue lors de la génération du QR Code: %v", err)
	}
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Fatalf("Format data URI invalide: %s", dataURI[:30])
	}

	// Cas chaîne vide
	_, errEmpty := GenerateQRCodeDataURI("", 100)
	if errEmpty == nil {
		t.Errorf("Une erreur était attendue pour un contenu QR vide")
	}
}

func TestRenderWatermark(t *testing.T) {
	wmDraft := RenderWatermark(domain.WatermarkDraft)
	if !strings.Contains(wmDraft, "watermark-draft") || !strings.Contains(wmDraft, "BROUILLON") {
		t.Errorf("Rendu watermark brouillon invalide: %s", wmDraft)
	}

	wmOfficial := RenderWatermark(domain.WatermarkOfficial)
	if !strings.Contains(wmOfficial, "watermark-official") || !strings.Contains(wmOfficial, "OFFICIEL") {
		t.Errorf("Rendu watermark officiel invalide: %s", wmOfficial)
	}

	wmArchived := RenderWatermark(domain.WatermarkArchived)
	if !strings.Contains(wmArchived, "watermark-archived") || !strings.Contains(wmArchived, "ARCHIVÉ") {
		t.Errorf("Rendu watermark archivé invalide: %s", wmArchived)
	}

	wmCustom := RenderWatermark(domain.WatermarkDraft, "TEXTE SPÉCIFIQUE")
	if !strings.Contains(wmCustom, "TEXTE SPÉCIFIQUE") {
		t.Errorf("Texte personnalisé non pris en compte: %s", wmCustom)
	}
}

func TestRenderInstitutionalHeader(t *testing.T) {
	htmlHeader := RenderInstitutionalHeader(
		"LABORATOIRE SAIDAL",
		"MIPH-AGR-2024-0089",
		"El Harrach, Alger",
		"Dr. Yacine BENALI",
		"MIPH/DT/2021/089",
		"Bordereau Officiel",
	)

	if !strings.Contains(htmlHeader, "République Algérienne Démocratique et Populaire") {
		t.Errorf("Mention républicaine manquante dans le header")
	}
	if !strings.Contains(htmlHeader, "Agence Nationale des Produits Pharmaceutiques") {
		t.Errorf("Mention ANPP manquante dans le header")
	}
	if !strings.Contains(htmlHeader, "LABORATOIRE SAIDAL") {
		t.Errorf("Nom d'établissement manquant")
	}
	if !strings.Contains(htmlHeader, "MIPH-AGR-2024-0089") {
		t.Errorf("Agrément manquant")
	}
}

func TestRenderLegalFooter(t *testing.T) {
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	qrURL := "https://sanad.labo.dz/verify/" + hash

	footer := RenderLegalFooter("ES-REG-01", domain.WatermarkOfficial, hash, qrURL, 1, 3, now)

	if !strings.Contains(footer, "data:image/png;base64,") {
		t.Errorf("QR-Code data URI manquant dans le footer")
	}
	if !strings.Contains(footer, "Page 1 sur 3") {
		t.Errorf("Pagination Page 1 sur 3 manquante dans le footer")
	}
	if !strings.Contains(footer, "OFFICIEL") {
		t.Errorf("Statut officiel manquant dans le footer")
	}
	if !strings.Contains(footer, "21 CFR Part 11") {
		t.Errorf("Mention de conformité 21 CFR Part 11 manquante")
	}
}

func TestRenderSignaturesBlock(t *testing.T) {
	// 1. Sans signature (en attente)
	blockUnsigned := RenderSignaturesBlock("Dr. BENALI", "MIPH/DT/089", nil, "")
	if !strings.Contains(blockUnsigned, "En attente de visa") {
		t.Errorf("Mention en attente de visa manquante")
	}

	// 2. Avec signature 2FA
	now := time.Now().UTC()
	blockSigned := RenderSignaturesBlock("Dr. BENALI", "MIPH/DT/089", &now, "Direction de l'Enregistrement")
	if !strings.Contains(blockSigned, "SCELLÉ & SIGNÉ ÉLECTRONIQUEMENT (2FA)") {
		t.Errorf("Mention scellé et signé 2FA manquante")
	}
	if !strings.Contains(blockSigned, "Direction de l&#39;Enregistrement") && !strings.Contains(blockSigned, "Direction de l'Enregistrement") {
		t.Errorf("Titre du destinataire manquant")
	}
}

func TestRenderDocumentWrap(t *testing.T) {
	doc := RenderDocumentWrap("Test Document", "<div class='content'>Contenu Test</div>")
	if !strings.Contains(doc, "<!DOCTYPE html>") {
		t.Errorf("Doctype manquant")
	}
	if !strings.Contains(doc, "@page {") {
		t.Errorf("Styles @page CSS manquants")
	}
	if !strings.Contains(doc, "Contenu Test") {
		t.Errorf("Corps du document manquant")
	}
}
