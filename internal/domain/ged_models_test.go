package domain_test

import (
	"testing"
	"time"

	"homolog/internal/domain"

	"github.com/google/uuid"
)

func TestStandardGEDTaxonomy(t *testing.T) {
	branches := domain.GetStandardGEDTaxonomy()
	if len(branches) != 5 {
		t.Fatalf("Attendu 5 branches maîtresses, obtenu %d", len(branches))
	}

	totalFolders := 0
	for _, b := range branches {
		totalFolders += len(b.Folders)
	}

	if totalFolders != 21 {
		t.Fatalf("Attendu 21 sous-dossiers standardisés selon Section 8 du CDC, obtenu %d", totalFolders)
	}

	// Test de validation de chemin
	validPath := "01_Affaires_Reglementaires/01_01_Agrements_Etablissement"
	if !domain.ValidateTaxonomyPath(validPath) {
		t.Errorf("Le chemin valide %s devrait être accepté", validPath)
	}

	invalidPath := "01_Affaires_Reglementaires/01_99_Inconnu"
	if domain.ValidateTaxonomyPath(invalidPath) {
		t.Errorf("Le chemin invalide %s devrait être rejeté", invalidPath)
	}
}

func TestVersionLabelGeneration(t *testing.T) {
	tests := []struct {
		current  string
		major    bool
		firstApp bool
		expected string
	}{
		{"", false, false, "v0.1"},
		{"v0.1", false, false, "v0.2"},
		{"v0.1", false, true, "v1.0"},
		{"v1.0", false, false, "v1.1"},
		{"v1.1", false, false, "v1.2"},
		{"v1.2", true, false, "v2.0"},
		{"v2.0", false, false, "v2.1"},
		{"v2.5", true, false, "v3.0"},
	}

	for _, tt := range tests {
		got := domain.GenerateNextVersionLabel(tt.current, tt.major, tt.firstApp)
		if got != tt.expected {
			t.Errorf("GenerateNextVersionLabel(%q, %v, %v) = %q, attendu %q",
				tt.current, tt.major, tt.firstApp, got, tt.expected)
		}
	}
}

func TestDocLifecycleTransitions(t *testing.T) {
	doc := &domain.GEDDocumentDetailed{
		Status: domain.GEDStatusBrouillon,
	}

	// Brouillon -> EnRevue (OK)
	allowed, reason := doc.CanTransitionTo(domain.GEDStatusEnRevue, domain.RoleContributeurQualiteBPF)
	if !allowed {
		t.Errorf("Brouillon -> EnRevue devrait être autorisé: %s", reason)
	}

	// Brouillon -> EnVigueur (Rejeté)
	allowed, _ = doc.CanTransitionTo(domain.GEDStatusEnVigueur, domain.RoleDirecteurTechnique)
	if allowed {
		t.Errorf("Brouillon -> EnVigueur direct devrait être interdit")
	}

	// EnRevue -> EnApprobationDT (OK)
	doc.Status = domain.GEDStatusEnRevue
	allowed, _ = doc.CanTransitionTo(domain.GEDStatusEnApprobationDT, domain.RoleResponsableAffairesReg)
	if !allowed {
		t.Errorf("EnRevue -> EnApprobationDT devrait être autorisé")
	}

	// EnApprobationDT -> EnVigueur par Contributeur (Refusé)
	doc.Status = domain.GEDStatusEnApprobationDT
	allowed, _ = doc.CanTransitionTo(domain.GEDStatusEnVigueur, domain.RoleContributeurQualiteBPF)
	if allowed {
		t.Errorf("Seul le PDT doit pouvoir mettre en vigueur")
	}

	// EnApprobationDT -> EnVigueur par PDT (Autorisé)
	allowed, _ = doc.CanTransitionTo(domain.GEDStatusEnVigueur, domain.RoleDirecteurTechnique)
	if !allowed {
		t.Errorf("Le PDT doit pouvoir approuver et mettre en vigueur")
	}
}

func TestSignatureSeal(t *testing.T) {
	uID := uuid.New()
	sha := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	now := time.Now()
	ip := "192.168.1.50"

	seal1 := domain.ComputeSignatureSeal(sha, uID, string(domain.RoleDirecteurTechnique), now, ip)
	seal2 := domain.ComputeSignatureSeal(sha, uID, string(domain.RoleDirecteurTechnique), now, ip)

	if seal1 != seal2 {
		t.Errorf("Le scellement doit être déterministe pour des paramètres identiques")
	}

	if len(seal1) != 64 {
		t.Errorf("Le scellement doit être un hash hexadécimal de 64 caractères SHA-256")
	}
}
