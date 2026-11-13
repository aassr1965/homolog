---
document_metadata:
  code_ctd_anpp: "LD-DEPP07-TAB"
  title: "Tableau Comparatif Officiel des Modifications (Approuvé vs Proposé) selon la Ligne Directrice LD-DEPP07/V01"
  legal_reference: "Ligne Directrice ANPP LD-DEPP07/V01 (Avril 2026) ; Décret exécutif n° 20-325, Article 17 ; Arrêté du 07 novembre 2021 (Regroupement)"
  target_procedures: ["VF", "VI", "VMin", "VMaj", "VAdmin"]
  security_roles:
    maker: "Pharmacien Affaires Réglementaires / Variations"
    checker: "Manager Assurance Qualité / Contrôle Changement"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.5 - Évaluation et Décision de Variation par l'ANPP"
  alcoa_compliance: "21 CFR Part 11 Compliant - Traçabilité Différentielle Rigide"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT - SOUS-DIRECTION DES VARIATIONS
### DOSSIER DE VARIATION : TABLEAU COMPARATIF DES MODIFICATIONS (LD-DEPP07/V01)

---

## 1. DONNÉES DU MÉDICAMENT ET RÉFÉRENCES DE LA DÉCISION D'ENREGISTREMENT

- **Dénomination Commerciale :** `{{product.brand_name}} {{product.dosage}}`
- **DCI de la Substance Active :** `{{product.dci}}`
- **Forme Pharmaceutique et Présentation :** `{{product.pharmaceutical_form}}` (Boîte de `{{product.pack_sizes}}`)
- **Numéro de la Décision d'Enregistrement Initiale :** `{{dossier.anpp_registration_number}}`
- **Date d'Octroi de la Décision :** `{{dossier.first_authorization_date}}`
- **Établissement Titulaire / Demandeur :** `{{establishment.name}}`
- **Numéro Dossier Homolog Variation :** `{{variation.dossier_reference}}`

---

## 2. CLASSIFICATION RÉGLEMENTAIRE ET RUBRIQUE OFFICIELLE (PARMI LES 34 RUBRIQUES)

- **Rubrique Officielle LD-DEPP07/V01 :** **Rubrique `{{variation.rubric_code}}`** - `{{variation.rubric_title}}`
- **Classification Réglementaire :**
  - [ ] **VAdmin** : Variation Purement Administrative (Délai d'instruction : 30 jours)
  - [ ] **VMin (Type IA/IB)** : Variation Mineure ne nécessitant pas d'expertise clinique complexe (Délai : 60 jours)
  - [ ] **VMaj (Type II)** : Variation Majeure avec impact potentiel sur la qualité, sécurité ou efficacité (Délai : 90 à 120 jours)
  - [ ] **Procédure de Regroupement (*Grouping*)** : Conformément à l'Arrêté du 07 novembre 2021 (plusieurs rubriques interdépendantes).

*Classification validée :* **`{{variation.classification}}`**

---

## 3. TABLEAU COMPARATIF STRUCTURE : SITUATION APPROUVÉE VS SITUATION PROPOSÉE

| Champ Analysé | Situation Actuellement Approuvée (Dossier AMM en Vigueur) | Situation Nouvelle Proposée (Soumise à Approbation ANPP) | Justification Technique & Scientifique |
|:---|:---|:---|:---|
| **Section CTD Concernée** | Module `{{variation.ctd_section_code}}` (Page `{{variation.old_page}}`) | Module `{{variation.ctd_section_code}}` (Page `{{variation.new_page}}`) | Mise à niveau BPF / Amélioration de procédé |
| **Description Précise de la Modification** | `{{variation.current_approved_text}}` | `{{variation.proposed_new_text}}` | `{{variation.technical_rationale}}` |
| **Impact Spécifications / Contrôles** | Ancienne limite : `{{variation.old_spec_limit}}` | Nouvelle limite : `{{variation.new_spec_limit}}` | Alignement sur la monographie Ph. Eur. révisée |
| **Site / Équipement / Matière** | `{{variation.old_site_or_material}}` | `{{variation.new_site_or_material}}` | Renforcement de la sécurité d'approvisionnement |

---

## 4. CONTRÔLE DES CONDITIONS D'APPLICATION (CHECKLIST LD-DEPP07/V01)

*Toutes les conditions cumulatives prévues par la Ligne Directrice pour la rubrique `{{variation.rubric_code}}` doivent être respectées. Si une seule condition fait défaut, la variation bascule automatiquement en Variation Majeure (VMaj).*

| Code Condition | Description de la Condition Réglementaire Requise | Statut Déclaré | Justification / Preuve Associée |
|:---|:---|:---|:---|
| **Condition C1** | `{{variation.condition_c1_description}}` | **SATISFAITE** | Confirmé par certificat analytique n° `{{variation.cert_ref_1}}` |
| **Condition C2** | `{{variation.condition_c2_description}}` | **SATISFAITE** | Absence de modification du profil de dissolution *in vitro* |
| **Condition C3** | `{{variation.condition_c3_description}}` | **SATISFAITE** | Aucune nouvelle impureté détectée au-delà du seuil ICH Q3A/B |
| **Condition C4** | `{{variation.condition_c4_description}}` | **SATISFAITE** | Les spécifications du produit fini demeurent strictement conformes |

---

## 5. BORDEREAU DES PIÈCES JUSTIFICATIVES FOURNIES (DOCUMENTS EXIGÉS)

| Réf. Document | Intitulé de la Pièce Jointe | Référence Interne / Date | Statut Documentaire |
|:---|:---|:---|:---|
| **DOC 1** | Formulaire officiel de demande de variation visé par le PDT | FORM-VAR-`{{variation.id}}` | Présent et signé |
| **DOC 2** | Tableau comparatif détaillé des modifications | LD-DEPP07-TAB-`{{variation.id}}` | Présent (ce document) |
| **DOC 3** | Rapport de validation de procédé / méthode d'analyse modifiée | VAL-REP-`{{variation.id}}` | Présent (Rapport validé) |
| **DOC 4** | Données de stabilité comparative (3 mois ou 6 mois Zone IVb) | STAB-IVB-`{{variation.id}}` | Présent (Données conformes) |
| **DOC 5** | Maquette d'étiquetage / Notice modifiée bilingue (si impact) | PACK-BAT-`{{variation.id}}` | Présent (Bilingue FR/AR) |
| **DOC 6** | Bordereau de paiement de la redevance réglementaire (BV1/BV2) | BV-VAR-`{{variation.id}}` | Quittance acquittée |

---

## 6. DÉCLARATION D'ÉVALUATION DU RISQUE QUALITÉ ET SÉCURITÉ

Le Pharmacien Directeur Technique et le Responsable Assurance Qualité certifient que :
1. Une analyse formelle des risques selon l'ICH Q9 a été réalisée préalablement au dépôt de la variation.
2. La modification proposée n'altère en aucune manière l'innocuité, la tolérance clinique, la biodisponibilité et l'efficacité thérapeutique du médicament.
3. Les lots commerciaux intégrant la modification feront l'objet d'un suivi de stabilité renforcé sous conditions Zone IVb ($30^\circ\text{C}/75\%\ \text{RH}$).

---

## 7. VISAS ET SIGNATURES ÉLECTRONIQUES (21 CFR PART 11)

| Rôle BPF | Nom & Prénom | Titre Professionnel | Date & Heure UTC | Empreinte Numérique SHA-256 |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Pharmacien Affaires Réglem. | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Responsable Assurance Qualité | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
