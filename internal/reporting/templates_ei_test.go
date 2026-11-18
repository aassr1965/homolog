package reporting

import (
	"strings"
	"testing"
	"time"
)

func TestRenderAll8IntermediateReports(t *testing.T) {
	d := makeTestDossier()
	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// 1. EI-AVT-01 : Fiche Navette Hebdomadaire
	ei1 := RenderEIAVT01HTML(d, "LABORATOIRES SAIDAL", "MIPH-089", "Dr. BENALI", hash, now)
	if !strings.Contains(ei1, "EI-AVT-01") {
		t.Errorf("Code EI-AVT-01 manquant")
	}
	if !strings.Contains(ei1, "Fiche Navette d'Avancement") {
		t.Errorf("Titre EI-AVT-01 manquant")
	}
	if !strings.Contains(ei1, "DOCUMENT INTERNE DE TRAVAIL") {
		t.Errorf("Filigrane interne manquant dans EI-AVT-01")
	}

	// 2. EI-AVT-02 : Gap Analysis Matrix
	ei2 := RenderEIAVT02HTML(d, "LABORATOIRES SAIDAL", "MIPH-089", "Dr. BENALI", hash, now)
	if !strings.Contains(ei2, "EI-AVT-02") {
		t.Errorf("Code EI-AVT-02 manquant")
	}
	if !strings.Contains(ei2, "Relevé des Écarts & Pièces Manquantes") {
		t.Errorf("Titre EI-AVT-02 manquant")
	}

	// 3. EI-AVT-03 : Fiche de Revue Checker
	ei3 := RenderEIAVT03HTML(d, "3.2.P.8", "Stabilité Zone IVb", "Rédacteur A", "Checker B", "Données brutes manquantes", "Réintégrer chromatogrammes", "SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei3, "EI-AVT-03") {
		t.Errorf("Code EI-AVT-03 manquant")
	}
	if !strings.Contains(ei3, "Fiche de Revue & d'Annotation Checker") {
		t.Errorf("Titre EI-AVT-03 manquant")
	}
	if !strings.Contains(ei3, "Données brutes manquantes") {
		t.Errorf("Motif de rejet manquant")
	}

	// 4. EI-CHR-01 : Suivi des Stabilités Zone IVb
	ei4 := RenderEICHR01HTML(d, "LABORATOIRES SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei4, "EI-CHR-01") {
		t.Errorf("Code EI-CHR-01 manquant")
	}
	if !strings.Contains(ei4, "Zone IVb (30°C / 75% HR)") && !strings.Contains(ei4, "Zone IVb") {
		t.Errorf("Mention Zone IVb manquante")
	}
	if !strings.Contains(ei4, "M6 (Seuil)") {
		t.Errorf("Seuil légal M6 manquant")
	}

	// 5. EI-CHR-02 : Bioéquivalence CRO Tracker
	ei5 := RenderEICHR02HTML(d, "ClinCRO Algiers", "CHU Mustapha", "Labo Bioanalytique", 99.2, 100.5, "SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei5, "EI-CHR-02") {
		t.Errorf("Code EI-CHR-02 manquant")
	}
	if !strings.Contains(ei5, "Bioéquivalence") {
		t.Errorf("Mention bioéquivalence manquante")
	}
	if !strings.Contains(ei5, "BIOÉQUIVALENT") {
		t.Errorf("Verdict BPF manquant")
	}

	// 6. EI-VIG-01 : Plan d'action lettre de carence
	ei6 := RenderEIVIG01HTML(d, "ANPP-LET-2026", "15/01/2026", 60, 4, "SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei6, "EI-VIG-01") {
		t.Errorf("Code EI-VIG-01 manquant")
	}
	if !strings.Contains(ei6, "Instruction des Compléments") {
		t.Errorf("Titre instruction compléments manquant")
	}
	if !strings.Contains(ei6, "60 jours calendaires") {
		t.Errorf("Délai alloué manquant")
	}

	// 7. EI-CHG-01 : Évaluation d'impact Change Control
	ei7 := RenderEICHG01HTML(d, "CC-2026-001", "Changement presse comprimé", "Rubrique 23.c", 60000.0, 60, "SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei7, "EI-CHG-01") {
		t.Errorf("Code EI-CHG-01 manquant")
	}
	if !strings.Contains(ei7, "Change Control") {
		t.Errorf("Mention Change Control manquante")
	}
	if !strings.Contains(ei7, "Rubrique 23.c") {
		t.Errorf("Classification LD manquante")
	}

	// 8. EI-PIP-01 : Weekly Pipeline Digest CoDir
	ei8 := RenderEIPIP01HTML(d, "LABORATOIRES SAIDAL", "Dr. BENALI", hash, now)
	if !strings.Contains(ei8, "EI-PIP-01") {
		t.Errorf("Code EI-PIP-01 manquant")
	}
	if !strings.Contains(ei8, "Pipeline") {
		t.Errorf("Mention Pipeline manquante")
	}
	if !strings.Contains(ei8, "Indicateurs Consolidés") {
		t.Errorf("Indicateurs CoDir manquants")
	}
}
