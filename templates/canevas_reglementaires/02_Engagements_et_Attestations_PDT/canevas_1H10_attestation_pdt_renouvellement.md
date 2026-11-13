---
document_metadata:
  code_ctd_anpp: "1.H.10"
  title: "Déclaration sur l'Honneur et Attestation du Pharmacien Directeur Technique pour le Renouvellement Quinquennal d'Enregistrement"
  legal_reference: "Décret exécutif n° 20-325 du 22 novembre 2020, Article 18 ; Guide CTD ANPP (Juillet 2025) Module 1.H.10"
  target_procedures: ["RF", "RI"]
  security_roles:
    maker: "Pharmacien Affaires Réglementaires"
    checker: "Responsable Assurance Qualité / Pharmacovigilance"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.5 - Dépôt du Dossier de Renouvellement Quinquennal"
  alcoa_compliance: "21 CFR Part 11 Compliant - Attestation Juridique Solennelle"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT DES PRODUITS PHARMACEUTIQUES
### MODULE 1 - SECTION 1.H.10 : ATTESTATION DE RENOUVELLEMENT QUINQUENNAL

---

## 1. IDENTIFICATION DU PRODUIT FAISANT L'OBJET DU RENOUVELLEMENT

- **Dénomination Commerciale :** `{{product.brand_name}} {{product.dosage}}`
- **Dénomination Commune Internationale (DCI) :** `{{product.dci}}`
- **Forme Pharmaceutique et Présentation :** `{{product.pharmaceutical_form}}` (Boîte de `{{product.pack_sizes}}`)
- **Numéro de la Décision d'Enregistrement Actuelle :** `{{dossier.current_amm_number}}`
- **Date d'Octroi de la Décision Initiale :** `{{dossier.current_amm_grant_date}}`
- **Date d'Échéance Quinquennale :** `{{dossier.current_amm_expiry_date}}`
- **Établissement Titulaire / Exploitant :** `{{establishment.name}}`

---

## 2. DÉCLARATION SOLENNELLE DE NON-MODIFICATION DE LA QUALITÉ

Je soussigné, **`Dr. {{pdt.full_name}}`**, Pharmacien Directeur Technique de l'établissement **`{{establishment.name}}`**, atteste sur l'honneur et engage ma responsabilité pleine et entière que :

1. **Non-Modification Qualitative et Quantitative :**  
   La formule qualitative et quantitative en substances actives et excipients, telle qu'approuvée dans le dossier d'enregistrement initial (ou modifiée par variations formellement autorisées par l'ANPP), est demeurée **strictement inchangée**.

2. **Spécifications Analytiques et Contrôles :**  
   Les spécifications à libération et à péremption ainsi que les méthodes de contrôle physico-chimique et microbiologique du produit fini n'ont subi aucune modification non autorisée, et demeurent conformes aux monographies actuelles des pharmacopées de référence (Pharmacopée Européenne en vigueur).

3. **Procédé de Fabrication et Sites Industriels :**  
   Le procédé de fabrication, les équipements industriels, les tailles de lot et les sites de fabrication, de conditionnement et de contrôle libératoire sont strictement identiques à ceux figurant au dossier actuellement enregistré.

---

## 3. HISTORIQUE DES VARIATIONS AUTORISÉES PENDANT LE QUINQUENNAT

| Date Décision ANPP | N° Décision Variation | Rubrique LD-DEPP07 | Type (VMaj / VMin / VAdmin) | Objet Succinct de la Modification | Statut de Déploiement |
|:---|:---|:---|:---|:---|:---|
| `{{var_hist.date_1}}` | `{{var_hist.ref_1}}` | `{{var_hist.rubric_1}}` | `{{var_hist.type_1}}` | `{{var_hist.desc_1}}` | Implémenté sur lot `{{var_hist.lot_1}}` |
| `{{var_hist.date_2}}` | `{{var_hist.ref_2}}` | `{{var_hist.rubric_2}}` | `{{var_hist.type_2}}` | `{{var_hist.desc_2}}` | Implémenté sur lot `{{var_hist.lot_2}}` |
| *Aucune autre variation* | - | - | - | Aucune modification non autorisée | - |

---

## 4. BILAN DE PHARMACOVIGILANCE ET RAPPORT BÉNÉFICE / RISQUE

- **Rapport Périodique Actualisé de Sécurité (PSUR / RPAS) :** Les rapports périodiques de sécurité ont été régulièrement transmis au Centre National de Pharmacovigilance et Matériovigilance (CNPM).
- **Rapport Bénéfice / Risque :** Le rapport entre les bénéfices cliniques et les risques associés à l'utilisation du médicament demeure **favorable et inchangé**.
- **Mesures de Restriction ou Retrait :** Le produit n'a fait l'objet d'aucun retrait d'autorisation de mise sur le marché, suspension ou interdiction de commercialisation sur le territoire national ou à l'étranger pour des raisons de sécurité ou de défaut qualité majeur.

---

## 5. VALIDATION DU FORMULAIRE DE RENOUVELLEMENT

| Rôle BPF | Nom & Prénom | Fonction Officielle | Date & Heure UTC | Signature Numérique (SHA-256) |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Pharmacien Affaires Réglem. | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Pharmacovigilance / AQ Lead | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
