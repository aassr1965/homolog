---
document_metadata:
  code_ctd_anpp: "LD-DEPP07-ATT"
  title: "Attestation Solennelle de Non-Modification des Autres Sections du Dossier d'Enregistrement lors d'une Demande de Variation"
  legal_reference: "Ligne Directrice ANPP LD-DEPP07/V01 (Avril 2026), Articles 5 et 7 ; Décret exécutif n° 20-325"
  target_procedures: ["VF", "VI", "VMin", "VMaj", "VAdmin"]
  security_roles:
    maker: "Pharmacien Affaires Réglementaires"
    checker: "Responsable Assurance Qualité"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.5 - Pièce Maîtresse Obligatoire Dépôt Dossier Variation"
  alcoa_compliance: "21 CFR Part 11 Compliant - Attestation Solennelle d'Intégrité"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT - SOUS-DIRECTION DES VARIATIONS
### ATTESTATION SOLENNELLE D'ENGAGEMENT ET DE NON-MODIFICATION COLLATÉRALE

---

## 1. IDENTIFICATION DU PRODUIT ET DE LA DEMANDE DE MODIFICATION

- **Dénomination du Médicament :** `{{product.brand_name}} {{product.dosage}}`
- **DCI :** `{{product.dci}}`
- **Forme Galénique :** `{{product.pharmaceutical_form}}`
- **N° Décision d'Enregistrement Actuelle :** `{{dossier.anpp_registration_number}}`
- **Titulaire de l'Enregistrement :** `{{establishment.name}}`
- **Objet de la Variation :** Rubrique `{{variation.rubric_code}}` - `{{variation.rubric_title}}`
- **Référence Interne Dossier Homolog :** `{{variation.dossier_reference}}`

---

## 2. ATTESTATION SUR L'HONNEUR DU PHARMACIEN DIRECTEUR TECHNIQUE

Je soussigné, **`Dr. {{pdt.full_name}}`**, en ma qualité de **Pharmacien Directeur Technique** de l'établissement pharmaceutique **`{{establishment.name}}`**, titulaire de la décision d'enregistrement du médicament mentionné ci-dessus ;

Vu les dispositions de la **Ligne Directrice LD-DEPP07/V01 (Version Avril 2026)** de l'Agence Nationale des Produits Pharmaceutiques relative aux modifications des termes de la décision d'enregistrement ;

### Déclare solennellement sur l'honneur et engage ma responsabilité pleine et entière :

1. **Absence de Toute Autre Modification :**  
   Qu'à l'exception expresse et stricte de la modification faisant l'objet de la présente demande de variation (précisée dans le tableau comparatif officiel ci-annexé sous la rubrique `{{variation.rubric_code}}`), **AUCUNE AUTRE MODIFICATION, DIRECTE OU INDIRECTE, N'A ÉTÉ APPORTÉE** aux éléments qualitatifs, quantitatifs, industriels, analytiques, précliniques ou cliniques approuvés dans le dossier d'enregistrement d'origine dudit médicament.

2. **Intégrité de la Formule et du Procédé :**  
   Que la formule galénique, les sources et spécifications des autres matières premières, les méthodes d'analyse de routine, le procédé de fabrication et les contrôles en cours demeurent rigoureusement conformes aux données initialement approuvées par l'ANPP.

3. **Maintien de la Qualité et Stabilité Zone IVb :**  
   Que la modification apportée n'entraîne aucune altération des spécifications du produit fini et ne compromet d'aucune manière la stabilité du produit sous les conditions climatiques de la **Zone IVb ($30^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$)**.

4. **Conformité BPF des Nouveaux Intervenants (si applicable) :**  
   Qu'en cas de changement ou d'adjonction de fournisseur de matière première, d'article de conditionnement ou de site d'analyse, l'entité concernée dispose d'une certification Bonnes Pratiques de Fabrication (BPF) en cours de validité délivrée par une autorité sanitaire compétente, et a fait l'objet d'un audit de qualification concluant par notre département Assurance Qualité.

---

## 3. CLAUSE PÉNALE ET ORDINAL

Je reconnais qu'en cas de fausse déclaration, de dissimulation d'une modification substantielle ou de non-respect des engagements formulés dans la présente attestation, l'établissement et moi-même nous exposons aux sanctions administratives prévues par la législation (suspension immédiate ou retrait de la décision d'enregistrement) ainsi qu'aux poursuites pénales et disciplinaires prévues par le Code de Déontologie Pharmaceutique et la Loi n° 18-11 relative à la santé.

---

## 4. SIGNATURE ÉLECTRONIQUE QUALIFIÉE ET HORODATAGE DE L'ATTESTATION

| Rôle BPF | Signataire | Titre Professionnel | Date & Heure UTC | Empreinte Numérique (SHA-256) |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Pharmacien Affaires Réglem. | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Directeur Assurance Qualité | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
