package domain

import (
	"testing"
)

func TestVariation_ClassificationHierarchy(t *testing.T) {
	if !ClassMmaj.IsHigherThan(ClassMmod) {
		t.Errorf("Mmaj doit être strictement plus sévère que Mmod")
	}
	if !ClassMmod.IsHigherThan(ClassMmin) {
		t.Errorf("Mmod doit être strictement plus sévère que Mmin")
	}
	if !ClassMmaj.IsHigherThan(ClassMmin) {
		t.Errorf("Mmaj doit être strictement plus sévère que Mmin")
	}
	if ClassMmin.IsHigherThan(ClassMmaj) {
		t.Errorf("Mmin ne doit pas être supérieur à Mmaj")
	}
}

func TestVariation_RubriquesRegistry(t *testing.T) {
	if len(MasterRubriquesRegistry) < 25 {
		t.Fatalf("Le catalogue de rubriques doit comporter au moins 25 rubriques, obtenu: %d", len(MasterRubriquesRegistry))
	}

	// Test recherche rubrique 1
	r1 := FindRubriqueDefinition("1")
	if r1 == nil {
		t.Fatalf("Rubrique 1 non trouvée")
	}
	if r1.DefaultClassification != ClassMmin {
		t.Errorf("Rubrique 1 doit être Mmin par défaut, obtenu: %s", r1.DefaultClassification)
	}

	// Test recherche rubrique 7.d.1 (API complet)
	r7d1 := FindRubriqueDefinition("7.d.1")
	if r7d1 == nil {
		t.Fatalf("Rubrique 7.d.1 non trouvée")
	}
	if r7d1.DefaultClassification != ClassMmod {
		t.Errorf("Rubrique 7.d.1 nominale attendue Mmod, obtenu: %s", r7d1.DefaultClassification)
	}
	if r7d1.EscalatedClassification != ClassMmaj {
		t.Errorf("Rubrique 7.d.1 escalade attendue Mmaj, obtenu: %s", r7d1.EscalatedClassification)
	}

	// Test recherche rubrique 30.b (Stérile)
	r30b := FindRubriqueDefinition("30.b")
	if r30b == nil {
		t.Fatalf("Rubrique 30.b non trouvée")
	}
	if r30b.DefaultClassification != ClassMmaj {
		t.Errorf("Rubrique 30.b nominale attendue Mmaj, obtenu: %s", r30b.DefaultClassification)
	}
}

func TestVariation_EscalationLogic(t *testing.T) {
	// Cas 1 : Rubrique 19.a (Excipients solides) - nominal Mmod, escalade vers Mmaj si conditions non remplies
	// Conditions : C1 (fonction), C2 (dissolution f2 >= 50), C3 (stabilité Zone IVb 3 lots)

	// Toutes conditions remplies -> Reste Mmod
	nom, eff, isEsc, _ := EvaluateRubriqueItem("19.a", map[string]bool{
		"C1": true,
		"C2": true,
		"C3": true,
	})
	if isEsc || eff != ClassMmod {
		t.Errorf("Avec toutes conditions C1..C3 remplies, 19.a doit rester Mmod, obtenu: %s (escaladé=%v)", eff, isEsc)
	}

	// Condition C2 (dissolution) manquante -> Doit escalader en Mmaj
	nom, eff, isEsc, reason := EvaluateRubriqueItem("19.a", map[string]bool{
		"C1": true,
		"C2": false,
		"C3": true,
	})
	if !isEsc {
		t.Errorf("Avec C2 manquante, 19.a doit être escaladée")
	}
	if eff != ClassMmaj {
		t.Errorf("19.a escaladée doit être Mmaj, obtenu: %s", eff)
	}
	if nom != ClassMmod {
		t.Errorf("19.a nominale doit rester Mmod")
	}
	if reason == "" {
		t.Errorf("Une raison d'escalade explicite doit être fournie")
	}

	// Cas 2 : Rubrique 7.b.1 (Starting material) - nominal Mmin, escalade Mmaj
	nom7, eff7, isEsc7, _ := EvaluateRubriqueItem("7.b.1", map[string]bool{
		"C1": false,
	})
	if !isEsc7 || eff7 != ClassMmaj {
		t.Errorf("Rubrique 7.b.1 sans C1 doit escalader en Mmaj, obtenu: %s (escaladé=%v)", eff7, isEsc7)
	}
	_ = nom7
}

func TestVariation_GroupingCalculation(t *testing.T) {
	// Grouping contenant :
	// - Rubrique 1 (Mmin)
	// - Rubrique 4 (Mmin)
	// - Rubrique 20.c (Mmod)
	items := []VariationItem{
		{RubriqueCode: "1", EffectiveClassification: ClassMmin, IsEscalated: false},
		{RubriqueCode: "4", EffectiveClassification: ClassMmin, IsEscalated: false},
		{RubriqueCode: "20.c", EffectiveClassification: ClassMmod, IsEscalated: false},
	}

	overall, hasEsc, highest := CalculateOverallClassification(items)
	if overall != ClassMmod {
		t.Errorf("Le grouping doit s'aligner sur Mmod (la plus sévère), obtenu: %s", overall)
	}
	if highest != "20.c" {
		t.Errorf("L'élément le plus sévère doit être 20.c, obtenu: %s", highest)
	}
	if hasEsc {
		t.Errorf("Aucune escalade ne devrait être détectée")
	}

	// Ajout d'une variation escaladée vers Mmaj
	items = append(items, VariationItem{
		RubriqueCode:            "19.a",
		EffectiveClassification: ClassMmaj,
		IsEscalated:             true,
	})

	overall2, hasEsc2, highest2 := CalculateOverallClassification(items)
	if overall2 != ClassMmaj {
		t.Errorf("Le grouping avec 19.a escaladée doit s'aligner sur Mmaj, obtenu: %s", overall2)
	}
	if !hasEsc2 {
		t.Errorf("L'escalade doit être signalée au niveau du grouping")
	}
	if highest2 != "19.a" {
		t.Errorf("L'élément déterminant doit être 19.a")
	}
}

func TestVariation_ChangeImpactTree(t *testing.T) {
	// Impact de Rubrique 1 (Dénomination) + Rubrique 20.c (Site PF)
	tree := GenerateChangeImpactTree([]string{"1", "20.c"})

	if len(tree.DirectRed) == 0 {
		t.Fatalf("DirectRed ne doit pas être vide")
	}
	if len(tree.PreservedGreen) == 0 {
		t.Fatalf("PreservedGreen ne doit pas être vide")
	}

	// Vérifier présence de 1.A.3 dans DirectRed
	has1A3 := false
	for _, it := range tree.DirectRed {
		if it.SectionCode == "1.A.3" {
			has1A3 = true
			break
		}
	}
	if !has1A3 {
		t.Errorf("1.A.3 doit être en impact direct rouge pour un changement de dénomination")
	}

	// Vérifier qu'une section non impactée comme 1.A.6 (Agrément MIPH) est en vert
	has1A6Green := false
	for _, it := range tree.PreservedGreen {
		if it.SectionCode == "1.A.6" {
			has1A6Green = true
			break
		}
	}
	if !has1A6Green {
		t.Errorf("1.A.6 doit être préservée (vert) pour ces modifications")
	}
}

func TestVariation_FeesCalculation(t *testing.T) {
	if fee := CalculateVariationFee(ClassMmaj); fee != 115000.00 {
		t.Errorf("Redevance Mmaj attendue 115 000 DA, obtenu: %f", fee)
	}
	if fee := CalculateVariationFee(ClassMmod); fee != 80000.00 {
		t.Errorf("Redevance Mmod attendue 80 000 DA, obtenu: %f", fee)
	}
	if fee := CalculateVariationFee(ClassMmin); fee != 40000.00 {
		t.Errorf("Redevance Mmin attendue 40 000 DA, obtenu: %f", fee)
	}
}
