package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"homolog/internal/domain"
)

func TestDMModification_ValidationAndFees(t *testing.T) {
	orgID := uuid.New()
	devID := uuid.New()

	// 1. Test validation champs obligatoires
	mod := &domain.DMModification{}
	err := mod.Validate()
	if err == nil {
		t.Fatal("attendu erreur pour organisation_id vide")
	}

	mod.OrganizationID = orgID
	mod.MedicalDeviceID = devID
	mod.ModificationCode = "MOD-DM-2026-001"
	mod.ModificationType = "INVALID"
	err = mod.Validate()
	if err == nil {
		t.Fatal("attendu erreur pour modification_type invalide")
	}

	// 2. Test calcul redevance substantielle
	mod.ModificationType = domain.DMModTypeSubstantielle
	mod.Title = "Changement de procédé de stérilisation"
	mod.Description = "Passage de l'ETO au rayonnement Gamma 25 kGy"
	err = mod.Validate()
	if err != nil {
		t.Fatalf("validation inattendue en échec: %v", err)
	}
	if mod.FeeAmount != domain.DMFeeSubstantielle {
		t.Errorf("attendu redevance %.2f DZD, reçu %.2f", domain.DMFeeSubstantielle, mod.FeeAmount)
	}

	// 3. Test calcul redevance non substantielle
	modMin := &domain.DMModification{
		OrganizationID:   orgID,
		MedicalDeviceID:  devID,
		ModificationCode: "MOD-DM-2026-002",
		ModificationType: domain.DMModTypeNonSubstantielle,
		Title:            "Mise à jour notice bilingue",
		Description:      "Clarification du texte arabe de la notice",
	}
	if err := modMin.Validate(); err != nil {
		t.Fatalf("validation modMin échouée: %v", err)
	}
	if modMin.FeeAmount != domain.DMFeeNonSubstantielle {
		t.Errorf("attendu redevance %.2f DZD, reçu %.2f", domain.DMFeeNonSubstantielle, modMin.FeeAmount)
	}

	// 4. Test génération de sceau 21 CFR Part 11
	mod.ID = uuid.New()
	seal := mod.GenerateSeal("Dr. Yacine Benkhalid", "DEC-MIPH-PDT-2022-7714")
	if !strings.HasPrefix(seal, "DM-AVN-SEAL-") {
		t.Errorf("préfixe sceau invalide: %s", seal)
	}
	if len(seal) < 20 {
		t.Errorf("longueur sceau insuffisante: %s", seal)
	}

	// 5. Test catalogue standard
	if len(domain.StandardDMCategories) < 10 {
		t.Errorf("catalogue des catégories incomplet: %d catégories", len(domain.StandardDMCategories))
	}
}
