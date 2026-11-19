package domain

import (
	"fmt"
	"strings"
)

// GatingSeverity définit le niveau de criticité d'un verrou réglementaire
type GatingSeverity string

const (
	GatingSeverityBlocking GatingSeverity = "BLOCKING" // Bloquant impératif (interdit la transition)
	GatingSeverityWarning  GatingSeverity = "WARNING"  // Avertissement informatif BPF
)

// GatingAction représente l'action de workflow contrôlée par le gating
type GatingAction string

const (
	GatingActionSubmitReview GatingAction = "submit_review" // Passage en revue Checker (under_review)
	GatingActionApprovePDT   GatingAction = "approve_pdt"   // Visa légal Pharmacien DT (approved_sealed)
	GatingActionInspection   GatingAction = "inspection"    // Consultation de l'état des verrous (UI)
)

// PrerequisiteCondition spécifie l'exigence minimale d'état pour lever le verrou
type PrerequisiteCondition string

const (
	ConditionApprovedSealed        PrerequisiteCondition = "approved_sealed"          // Doit être scellé v1.0 par PDT
	ConditionAtLeastUnderReview    PrerequisiteCondition = "at_least_under_review"    // Au moins en cours de revue Checker
	ConditionAtLeastPendingApprove PrerequisiteCondition = "at_least_pending_approve" // Au moins validé par Checker (prêt PDT)
	ConditionRevisionMatch         PrerequisiteCondition = "revision_match"           // Versions synchronisées (parité)
	ConditionAllModule3Approved    PrerequisiteCondition = "all_module3_approved"     // 100% des [E] du Module 3 scellés
)

// GatingRule modélise une règle du DAG de dépendances CTD / DM
type GatingRule struct {
	RuleID             string                `json:"rule_id"`
	TargetPrefix       string                `json:"target_prefix"`        // Préfixe de l'item cible (ex: "1.P.1", "1.P.2", "2.3")
	TargetAction       GatingAction          `json:"target_action"`        // Action contrôlée
	RequiredItemPrefix string                `json:"required_item_prefix"` // Préfixe de la pièce prérequise
	RequiredCondition  PrerequisiteCondition `json:"required_condition"`
	TargetPhase        int                   `json:"target_phase"` // Phase DAG (1 à 5)
	DescriptionFR      string                `json:"description_fr"`
	LegalRef           string                `json:"legal_ref"`
	Severity           GatingSeverity        `json:"severity"`
}

// UnmetPrerequisite détaille un point de blocage actif
type UnmetPrerequisite struct {
	RuleID            string                `json:"rule_id"`
	RequiredItemCode  string                `json:"required_item_code"`
	RequiredTitle     string                `json:"required_title"`
	CurrentStatus     string                `json:"current_status"`
	ExpectedCondition PrerequisiteCondition `json:"expected_condition"`
	BlockingReason    string                `json:"blocking_reason"`
	LegalRef          string                `json:"legal_ref"`
}

// GatingEvaluationResult contient le diagnostic complet des verrous pour un item
type GatingEvaluationResult struct {
	TargetItemCode     string              `json:"target_item_code"`
	TargetItemTitle    string              `json:"target_item_title"`
	IsBlocked          bool                `json:"is_blocked"`
	ActivePhase        int                 `json:"active_phase"`
	UnmetPrerequisites []UnmetPrerequisite `json:"unmet_prerequisites"`
	Warnings           []string            `json:"warnings"`
	SummaryMessage     string              `json:"summary_message"`
}

// StandardCTDGatingRules définit le catalogue officiel des règles de dépendance ANPP & BPF
var StandardCTDGatingRules = []GatingRule{
	// -------------------------------------------------------------------------
	// Règle 1 : Stabilités Zone IVb (3.2.P.8) ➔ RCP (1.P.1) & Notice (1.P.2 / 1.P.2.1 / 1.P.2.2)
	// -------------------------------------------------------------------------
	{
		RuleID:             "GAT-STAB-TO-RCP",
		TargetPrefix:       "1.P.1",
		TargetAction:       GatingActionApprovePDT,
		RequiredItemPrefix: "3.2.P.8",
		RequiredCondition:  ConditionApprovedSealed,
		TargetPhase:        4,
		DescriptionFR:      "Le Résumé des Caractéristiques du Produit (RCP 1.P.1) ne peut être scellé tant que les stabilités Zone IVb (3.2.P.8) ne sont pas validées.",
		LegalRef:           "Guide CTD ANPP Juillet 2025 Module 1.P & 3.2.P.8 (Seuil légal M6)",
		Severity:           GatingSeverityBlocking,
	},
	{
		RuleID:             "GAT-STAB-TO-NOTICE",
		TargetPrefix:       "1.P.2",
		TargetAction:       GatingActionApprovePDT,
		RequiredItemPrefix: "3.2.P.8",
		RequiredCondition:  ConditionApprovedSealed,
		TargetPhase:        4,
		DescriptionFR:      "La Notice Patient (1.P.2) ne peut être scellée tant que les données de stabilité Zone IVb (3.2.P.8) ne sont pas formellement conclues et scellées.",
		LegalRef:           "Guide CTD ANPP Juillet 2025 Module 1.P.2 & 3.2.P.8",
		Severity:           GatingSeverityBlocking,
	},

	// -------------------------------------------------------------------------
	// Règle 2 : Parité Linguistique Synchrone Notice FR (1.P.2.1) ➔ Notice AR (1.P.2.2)
	// -------------------------------------------------------------------------
	{
		RuleID:             "GAT-PARITE-LINGUISTIQUE-REVIEW",
		TargetPrefix:       "1.P.2.2",
		TargetAction:       GatingActionSubmitReview,
		RequiredItemPrefix: "1.P.2.1",
		RequiredCondition:  ConditionAtLeastUnderReview,
		TargetPhase:        4,
		DescriptionFR:      "La Notice en langue arabe (1.P.2.2) ne peut être soumise à revue tant que la notice française (1.P.2.1) n'est pas rédigée et transmise au Checker.",
		LegalRef:           "Décret exécutif n° 20-325 Art. 12 & Guide CTD ANPP 1.P.2 (Parité synchrone FR/AR)",
		Severity:           GatingSeverityBlocking,
	},
	{
		RuleID:             "GAT-PARITE-LINGUISTIQUE-APPROVE",
		TargetPrefix:       "1.P.2.2",
		TargetAction:       GatingActionApprovePDT,
		RequiredItemPrefix: "1.P.2.1",
		RequiredCondition:  ConditionApprovedSealed,
		TargetPhase:        4,
		DescriptionFR:      "La Notice en langue arabe (1.P.2.2) ne peut être approuvée par le PDT que si la notice française (1.P.2.1) est préalablement scellée.",
		LegalRef:           "Décret exécutif n° 20-325 Art. 12 & Règle bloquante GxP",
		Severity:           GatingSeverityBlocking,
	},

	// -------------------------------------------------------------------------
	// Règle 3 : Module 3 Qualité ➔ Résumé Global de la Qualité QOS (Module 2.3)
	// -------------------------------------------------------------------------
	{
		RuleID:             "GAT-MODULE3-TO-QOS",
		TargetPrefix:       "2.3",
		TargetAction:       GatingActionApprovePDT,
		RequiredItemPrefix: "3.",
		RequiredCondition:  ConditionAllModule3Approved,
		TargetPhase:        4,
		DescriptionFR:      "Le Résumé Global de la Qualité (QOS Module 2.3) ne peut être scellé tant que toutes les rubriques obligatoires du Module 3 ne sont pas validées.",
		LegalRef:           "Guide CTD ANPP Juillet 2025 Module 2.3 & ICH M4Q (Synthèse des données figées)",
		Severity:           GatingSeverityBlocking,
	},

	// -------------------------------------------------------------------------
	// Règle 4 : Validation Procédé & Lots Pilotes ➔ Stabilités Zone IVb (3.2.P.8)
	// -------------------------------------------------------------------------
	{
		RuleID:             "GAT-PROCED-TO-STAB",
		TargetPrefix:       "3.2.P.8",
		TargetAction:       GatingActionSubmitReview,
		RequiredItemPrefix: "3.2.P.3.5",
		RequiredCondition:  ConditionApprovedSealed,
		TargetPhase:        3,
		DescriptionFR:      "Les études de stabilité (3.2.P.8) ne peuvent être validées sans le rapport de validation de procédé sur les lots pilotes (3.2.P.3.5).",
		LegalRef:           "BPF Algériennes & Guide CTD 3.2.P.3.5 / 3.2.P.8",
		Severity:           GatingSeverityWarning,
	},
}

// EvaluateItemGating analyse si une section peut franchir l'étape demandée
func EvaluateItemGating(allItems []ChecklistItem, targetItem *ChecklistItem, action GatingAction) GatingEvaluationResult {
	result := GatingEvaluationResult{
		TargetItemCode:  targetItem.ItemCode,
		TargetItemTitle: targetItem.TitleFR,
		IsBlocked:       false,
		ActivePhase:     CalculateDossierDAGPhase(allItems),
	}

	for _, rule := range StandardCTDGatingRules {
		// Vérifier si la règle concerne cet item cible
		if !matchesItemCodePrefix(targetItem.ItemCode, rule.TargetPrefix) {
			continue
		}

		// Si l'action ciblée par la règle ne correspond pas à l'action en cours :
		// - Si l'action est Inspection (UI), on évalue les conditions d'approbation pour informer
		if action != GatingActionInspection && rule.TargetAction != action {
			continue
		}

		// Évaluation de la condition requise
		switch rule.RequiredCondition {
		case ConditionApprovedSealed:
			unmet := checkItemApprovedSealed(allItems, rule)
			if unmet != nil {
				if rule.Severity == GatingSeverityBlocking {
					result.IsBlocked = true
					result.UnmetPrerequisites = append(result.UnmetPrerequisites, *unmet)
				} else {
					result.Warnings = append(result.Warnings, unmet.BlockingReason)
				}
			}

		case ConditionAtLeastUnderReview:
			unmet := checkItemAtLeastStatus(allItems, rule, WorkflowStatusUnderReview)
			if unmet != nil {
				if rule.Severity == GatingSeverityBlocking {
					result.IsBlocked = true
					result.UnmetPrerequisites = append(result.UnmetPrerequisites, *unmet)
				} else {
					result.Warnings = append(result.Warnings, unmet.BlockingReason)
				}
			}

		case ConditionAllModule3Approved:
			unmetList := checkAllModule3Approved(allItems, rule)
			if len(unmetList) > 0 {
				result.IsBlocked = true
				result.UnmetPrerequisites = append(result.UnmetPrerequisites, unmetList...)
			}
		}
	}

	if result.IsBlocked {
		var msgs []string
		for _, u := range result.UnmetPrerequisites {
			msgs = append(msgs, fmt.Sprintf("• %s (Réf: %s)", u.BlockingReason, u.LegalRef))
		}
		result.SummaryMessage = fmt.Sprintf("Verrou GxP actif sur %s : %s", targetItem.ItemCode, strings.Join(msgs, " | "))
	} else {
		result.SummaryMessage = fmt.Sprintf("Aucun verrou actif sur %s : éligible pour l'action demandée.", targetItem.ItemCode)
	}

	return result
}

// CalculateDossierDAGPhase calcule la phase d'avancement active (1 à 5)
func CalculateDossierDAGPhase(items []ChecklistItem) int {
	// Phase 1 : Méthodes & Matières (3.2.S & 3.2.P.5)
	phase1Complete := true
	phase2Complete := true
	phase3Complete := true
	phase4Complete := true

	for _, it := range items {
		if !it.IsMandatory || it.RequirementLevel != RequirementExiged {
			continue
		}

		// Phase 1 : 3.2.S ou 3.2.P.5
		if strings.HasPrefix(it.ItemCode, "3.2.S") || strings.HasPrefix(it.ItemCode, "3.2.P.5") {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				phase1Complete = false
			}
		}

		// Phase 2 : Procédé 3.2.P.3 & CoA 1.P.4
		if strings.HasPrefix(it.ItemCode, "3.2.P.3") || strings.HasPrefix(it.ItemCode, "1.P.4") {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				phase2Complete = false
			}
		}

		// Phase 3 : Stabilités 3.2.P.8 & Bioéquivalence 5.3
		if strings.HasPrefix(it.ItemCode, "3.2.P.8") || strings.HasPrefix(it.ItemCode, "5.3") {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				phase3Complete = false
			}
		}

		// Phase 4 : Notice 1.P.2, RCP 1.P.1 & QOS 2.3
		if strings.HasPrefix(it.ItemCode, "1.P.1") || strings.HasPrefix(it.ItemCode, "1.P.2") || strings.HasPrefix(it.ItemCode, "2.3") {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				phase4Complete = false
			}
		}
	}

	if !phase1Complete {
		return 1 // Phase 1 : Méthodes & Matières
	}
	if !phase2Complete {
		return 2 // Phase 2 : Procédés & Lots Pilotes
	}
	if !phase3Complete {
		return 3 // Phase 3 : Stabilités Zone IVb & Bioéquivalence
	}
	if !phase4Complete {
		return 4 // Phase 4 : QOS, Notice & RCP
	}
	return 5 // Phase 5 : Engagements & Dépôt Légal PDT
}

// Helpers internes de vérification

func matchesItemCodePrefix(code, prefix string) bool {
	if code == prefix || strings.HasPrefix(code, prefix+".") {
		return true
	}
	return false
}

func checkItemApprovedSealed(items []ChecklistItem, rule GatingRule) *UnmetPrerequisite {
	for _, it := range items {
		if matchesItemCodePrefix(it.ItemCode, rule.RequiredItemPrefix) {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				return &UnmetPrerequisite{
					RuleID:            rule.RuleID,
					RequiredItemCode:  it.ItemCode,
					RequiredTitle:     it.TitleFR,
					CurrentStatus:     string(it.WorkflowStatus),
					ExpectedCondition: rule.RequiredCondition,
					BlockingReason:    rule.DescriptionFR,
					LegalRef:          rule.LegalRef,
				}
			}
			return nil // Trouvé et conforme
		}
	}
	// Si la pièce requise n'existe pas du tout dans la checklist
	return &UnmetPrerequisite{
		RuleID:            rule.RuleID,
		RequiredItemCode:  rule.RequiredItemPrefix,
		RequiredTitle:     "Pièce prérequise",
		CurrentStatus:     "non_instanciee",
		ExpectedCondition: rule.RequiredCondition,
		BlockingReason:    rule.DescriptionFR,
		LegalRef:          rule.LegalRef,
	}
}

func checkItemAtLeastStatus(items []ChecklistItem, rule GatingRule, minStatus WorkflowStatus) *UnmetPrerequisite {
	for _, it := range items {
		if matchesItemCodePrefix(it.ItemCode, rule.RequiredItemPrefix) {
			if !isWorkflowStatusAtLeast(it.WorkflowStatus, minStatus) {
				return &UnmetPrerequisite{
					RuleID:            rule.RuleID,
					RequiredItemCode:  it.ItemCode,
					RequiredTitle:     it.TitleFR,
					CurrentStatus:     string(it.WorkflowStatus),
					ExpectedCondition: rule.RequiredCondition,
					BlockingReason:    rule.DescriptionFR,
					LegalRef:          rule.LegalRef,
				}
			}
			return nil
		}
	}
	return nil
}

func isWorkflowStatusAtLeast(actual, minStatus WorkflowStatus) bool {
	order := map[WorkflowStatus]int{
		WorkflowStatusDraftPending:    1,
		WorkflowStatusDrafting:        2,
		WorkflowStatusUnderReview:     3,
		WorkflowStatusPendingApproval: 4,
		WorkflowStatusApprovedSealed:  5,
	}
	return order[actual] >= order[minStatus]
}

func checkAllModule3Approved(items []ChecklistItem, rule GatingRule) []UnmetPrerequisite {
	var unmets []UnmetPrerequisite
	for _, it := range items {
		if it.ModuleNumber == 3 && it.IsMandatory && it.RequirementLevel == RequirementExiged {
			if it.WorkflowStatus != WorkflowStatusApprovedSealed {
				unmets = append(unmets, UnmetPrerequisite{
					RuleID:            rule.RuleID,
					RequiredItemCode:  it.ItemCode,
					RequiredTitle:     it.TitleFR,
					CurrentStatus:     string(it.WorkflowStatus),
					ExpectedCondition: ConditionApprovedSealed,
					BlockingReason:    fmt.Sprintf("La section Module 3 '%s' doit être scellée avant de valider le QOS (2.3).", it.ItemCode),
					LegalRef:          rule.LegalRef,
				})
			}
		}
	}
	return unmets
}
