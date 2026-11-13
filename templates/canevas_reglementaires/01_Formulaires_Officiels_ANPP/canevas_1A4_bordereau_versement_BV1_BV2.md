---
document_metadata:
  code_ctd_anpp: "1.A.4"
  title: "Bordereau de Versement des Redevances Réglementaires (BV1 et BV2)"
  legal_reference: "Décret exécutif n° 20-325 du 22 novembre 2020, Article 12 ; Arrêté interministériel fixant les redevances dues à l'ANPP"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI", "DM"]
  security_roles:
    maker: "Comptabilité / Assistant Réglementaire"
    checker: "Pharmacien Affaires Réglementaires"
    approver: "Directeur Administratif & Financier / PDT"
  gating_dependency: "Jalon 2.1 - Déblocage Recevabilité (BV1) & Commission d'Enregistrement (BV2)"
  alcoa_compliance: "21 CFR Part 11 Compliant - Traçabilité Bancaire et Fiscale"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ADMINISTRATION ET DES FINANCES - SERVICE RECOUVREMENT
### MODULE 1 - SECTION 1.A.4 : BORDEREAU DE PAIEMENT DES REDEVANCES RÉGLEMENTAIRES

---

## 1. IDENTIFICATION DU DOSSIER ET DE L'ASSUJETTI

- **Référence Dossier Homolog / ANPP :** `{{dossier.reference_anpp}}`
- **Dénomination Commerciale du Produit :** `{{product.brand_name}} {{product.dosage}}`
- **Forme et Présentation :** `{{product.pharmaceutical_form}}` (Boîte de `{{product.pack_sizes}}`)
- **Établissement Pharmaceutique Déposant :** `{{establishment.name}}`
- **Numéro d'Agrément Établissement :** `{{establishment.license_number}}`
- **Identifiant Fiscal Unique (NIF) :** `{{establishment.nif}}`
- **Code Client / Tiers ANPP :** `{{establishment.anpp_client_code}}`

---

## 2. DÉCOMPTE OFFICIEL DES REDEVANCES RÉGLEMENTAIRES

### 2.1 Barème Applicable (selon Décret exécutif 20-325 & Arrêtés d'application)

| Nature de l'Acte Réglementaire | Montant BV1 (Recevabilité Administrative & Dépôt) | Montant BV2 (Expertise Technique & Évaluation CTD) | Total Redevance Exigible |
|:---|:---|:---|:---|
| **Nouvel Enregistrement Fabrication Locale (NF)** | 100 000,00 DZD | 300 000,00 DZD | **400 000,00 DZD** |
| **Nouvel Enregistrement Importation (NI)** | 250 000,00 DZD | 750 000,00 DZD | **1 000 000,00 DZD** |
| **Renouvellement Fabrication Locale (RF)** | 50 000,00 DZD | 150 000,00 DZD | **200 000,00 DZD** |
| **Renouvellement Importation (RI)** | 100 000,00 DZD | 300 000,00 DZD | **400 000,00 DZD** |
| **Variation Majeure (VMaj - Type II)** | 50 000,00 DZD | 150 000,00 DZD | **200 000,00 DZD** |
| **Variation Mineure (VMin - Type IB/IA)** | 30 000,00 DZD | 70 000,00 DZD | **100 000,00 DZD** |
| **Homologation Dispositif Médical (Classe IIb / III)**| 80 000,00 DZD | 170 000,00 DZD | **250 000,00 DZD** |

*Ligne sélectionnée pour le présent dossier :*  
**Acte Réglementaire :** `{{redevance.act_label}}`  
**Montant Total Attribué :** `{{redevance.total_amount_dzd}}` Dinars Algériens (DZD).

---

## 3. VOLET 1 : BORDEREAU DE VERSEMENT BV1 (PHASE RECEVABILITÉ)

> **Règle de Gestion Homolog :** Le téléversement de la quittance BV1 certifiée est un **pré-requis bloquant (Gatekeeper)** pour la recevabilité administrative et l'assignation des experts rapporteurs à l'ANPP.

| Champ de Déclaration Bancaire | Détail de la Transaction |
|:---|:---|
| **Bénéficiaire du Paiement** | Agence Nationale des Produits Pharmaceutiques (ANPP) |
| **Banque / Compte du Trésor** | Trésor Public d'Alger - Compte Spécial n° `{{anpp_bank.treasury_rib}}` |
| **Banque Émettrice (Établissement)** | `{{establishment.bank_name}}` (Agence : `{{establishment.bank_branch}}`) |
| **Numéro de Virement / Chèque / Avis d'Opéré** | `{{redevance.bv1_transaction_reference}}` |
| **Montant Versé en Chiffres** | `{{redevance.bv1_amount}}` DZD |
| **Montant Versé en Lettres** | `{{redevance.bv1_amount_words}}` Dinars Algériens |
| **Date Effective de Débit / Valeur** | `{{redevance.bv1_payment_date}}` |
| **Numéro de Quittance Délivrée par l'ANPP** | `{{redevance.bv1_receipt_number}}` |

---

## 4. VOLET 2 : BORDEREAU DE VERSEMENT BV2 (PHASE ÉVALUATION TECHNIQUE)

> **Règle de Gestion Homolog :** L'émission du BV2 intervient après notification formelle de la recevabilité administrative conforme par l'ANPP. Le paiement conditionne le passage du dossier devant la Commission d'Enregistrement.

| Champ de Déclaration Bancaire | Détail de la Transaction |
|:---|:---|
| **Bénéficiaire du Paiement** | Agence Nationale des Produits Pharmaceutiques (ANPP) |
| **Numéro de Notification de Recevabilité ANPP** | `{{dossier.anpp_admissibility_letter_number}}` en date du `{{dossier.anpp_admissibility_date}}` |
| **Numéro de Virement / Quittance BV2** | `{{redevance.bv2_transaction_reference}}` |
| **Montant Versé en Chiffres** | `{{redevance.bv2_amount}}` DZD |
| **Montant Versé en Lettres** | `{{redevance.bv2_amount_words}}` Dinars Algériens |
| **Date Effective de Versement** | `{{redevance.bv2_payment_date}}` |
| **Numéro de Quittance Trésor ANPP (BV2)** | `{{redevance.bv2_receipt_number}}` |

---

## 5. BORDEREAU DES PIÈCES COMPTABLES ANNEXÉES

- [x] Copie certifiée conforme de l'ordre de virement bancaire irrévocable revêtu du cachet de la banque émettrice.
- [x] Avis d'opéré bancaire confirmant le crédit au compte de l'ANPP.
- [x] Quittance officielle originale délivrée par la régie des recettes de l'ANPP (document numérisé haute résolution et archivé sur stockage immuable SHA-256).

---

## 6. VISAS DE CONTRÔLE COMPTABLE ET VALIDATION ORDINAL

| Rôle du Signataire | Nom & Prénom | Titre / Qualité | Date du Visa | Signature Électronique / Hash |
|:---|:---|:---|:---|:---|
| **Émetteur Comptable (*Maker*)** | `{{finance.maker_name}}` | Comptable Trésorerie | `{{finance.maker_timestamp}}` | `{{finance.maker_hash}}` |
| **Vérificateur Réglementaire (*Checker*)**| `{{regulatory.checker_name}}`| Pharmacien Affaires Réglem.| `{{regulatory.checker_timestamp}}`| `{{regulatory.checker_hash}}`|
| **Approbateur Financier & Légal (*Approver*)**| `{{approver.daf_or_pdt_name}}`| Dir. Administratif & Financier / PDT | `{{approver.timestamp}}` | `{{approver.hash}}` |
