package repository

import (
	"testing"
)

func TestDetermineDefaultDepartment_MatrixMapping(t *testing.T) {
	tests := []struct {
		itemCode  string
		subModule string
		moduleNum int
		expected  string
		comment   string
	}{
		// Module 1 standard
		{"1.A.1", "1.A", 1, "RA", "Récépissé de dépôt -> RA"},
		{"1.A.3", "1.A", 1, "RA", "Formulaire 1.A.3 -> RA"},
		{"1.H.1", "1.H", 1, "RA", "Courrier Holder -> RA"},
		{"1.P.1", "1.P", 1, "RA", "Notices et RCP -> RA"},

		// Module 1 spécificités décentralisées (CDC §7.4.1)
		{"1.P.8", "1.P", 1, "PROD", "Engagement 3 lots industriels -> Production"},
		{"1.P.4", "1.P", 1, "QC", "Certificat d'analyse CoA SA -> Contrôle Qualité"},
		{"1.P.5", "1.P", 1, "QC", "Certificat d'analyse CoA PF -> Contrôle Qualité"},
		{"1.M.P.3", "1.M", 1, "RD", "Bioéquivalence / Biowaiver dissolution -> R&D"},
		{"1.M.S.1", "1.M", 1, "EXTERNAL", "Agrément fabricant SA -> Partenaire Externe / API Holder"},
		{"1.M.S.3", "1.M", 1, "EXTERNAL", "Lettre d'accès CEP / ASMF -> Partenaire Externe / API Holder"},

		// Module 2 (Synthèse Qualité QOS)
		{"2.3.S", "2.3", 2, "RD", "Résumé global de la qualité QOS -> R&D"},

		// Module 3
		{"3.2.S.1", "3.2.S", 3, "EXTERNAL", "Substance active ASMF / DMF -> Partenaire Externe / API Holder"},
		{"3.2.S.4", "3.2.S", 3, "QC", "Contrôle de la substance active -> QC"},
		{"3.2.P.2", "3.2.P", 3, "RD", "Développement pharmaceutique -> R&D"},
		{"3.2.P.3.1", "3.2.P", 3, "PROD", "Formule de fabrication -> Production"},
		{"3.2.P.3.5", "3.2.P", 3, "PROD", "Validation de procédé 3 lots -> Production"},
		{"3.2.P.5.1", "3.2.P", 3, "QC", "Spécifications produit fini -> QC"},
		{"3.2.P.8.1", "3.2.P", 3, "QC", "Stabilités Zone IVb -> QC"},
		{"3.2.A.1", "3.2.A", 3, "QC", "Installations et équipements -> Fallback Module 3 QC"},

		// Module 4 & Module 5
		{"4.2.1", "4.2", 4, "RD", "Rapports non cliniques -> Fallback Module 4 RD"},
		{"5.3.1.2", "5.3", 5, "EXTERNAL", "Étude clinique bioéquivalence in vivo -> CRO / EXTERNAL"},

		// Dispositifs Médicaux (DM - Décret 20-324 / Arrêté 10 mai 2021)
		{"DM.A.1", "DM_TECH", 0, "RA", "Formulaire d'homologation DM -> RA"},
		{"DM.A.4", "DM_TECH", 0, "RA", "Contrat fabricant / demandeur -> RA"},
		{"DM.I.1", "DM_TECH", 0, "RA", "Notice d'utilisation bilingue -> RA"},
		{"DM.Q.1", "DM_TECH", 0, "QA", "Certificat ISO 13485 -> Assurance Qualité"},
		{"DM.Q.2", "DM_TECH", 0, "QA", "Certificat de conformité CE -> Assurance Qualité"},
		{"DM.T.1", "DM_TECH", 0, "RD", "Descriptif technique DM -> R&D"},
		{"DM.T.2", "DM_TECH", 0, "RD", "Gestion des risques ISO 14971 -> R&D"},
		{"DM.T.3", "DM_TECH", 0, "PROD", "Validation procédé fabrication DM -> Production"},
		{"DM.S.1", "DM_TECH", 0, "PROD", "Procédé de stérilisation ISO 11135 -> Production"},
		{"DM.S.2", "DM_TECH", 0, "PROD", "Barrière stérile ISO 11607 -> Production"},
		{"DM.C.1", "DM_TECH", 0, "RD", "Évaluation clinique DM -> R&D"},
	}

	for _, tt := range tests {
		t.Run(tt.itemCode, func(t *testing.T) {
			got := determineDefaultDepartment(tt.itemCode, tt.subModule, tt.moduleNum)
			if got != tt.expected {
				t.Errorf("determineDefaultDepartment(%s, %s, %d) = %s; attendu %s (%s)",
					tt.itemCode, tt.subModule, tt.moduleNum, got, tt.expected, tt.comment)
			}
		})
	}
}
