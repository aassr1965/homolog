---
document_metadata:
  code_ctd_anpp: "1.P.8"
  title: "Engagement du Pharmacien Directeur Technique pour la Validation du Procédé de Fabrication sur Trois Lots Industriels et Non-Commercialisation Préalable"
  legal_reference: "Décret exécutif n° 20-325 du 22 novembre 2020 ; Guide CTD ANPP (Juillet 2025) Module 1.P.8 ; Bonnes Pratiques de Fabrication (BPF) algériennes"
  target_procedures: ["NF", "MF"]
  security_roles:
    maker: "Responsable Assurance Qualité / Production"
    checker: "Manager Assurance Qualité"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.6 - Pré-requis Bloquant Libération des Lots Commerciaux"
  alcoa_compliance: "21 CFR Part 11 Compliant - Engagement Juridique et Ordinal"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT DES PRODUITS PHARMACEUTIQUES
### MODULE 1 - SECTION 1.P.8 : ENGAGEMENT SOLENNEL DU PHARMACIEN DIRECTEUR TECHNIQUE (FABRICATION LOCALE)

---

## 1. IDENTIFICATION DU MÉDICAMENT ET DE L'ÉTABLISSEMENT PRODUCTEUR

- **Dénomination Commerciale :** `{{product.brand_name}} {{product.dosage}}`
- **Dénomination Commune Internationale (DCI) :** `{{product.dci}}`
- **Forme Galénique et Présentation :** `{{product.pharmaceutical_form}}` (Boîte de `{{product.pack_sizes}}`)
- **Numéro de Dossier ANPP :** `{{dossier.reference_anpp}}`
- **Établissement Fabricant Demandeur :** `{{establishment.name}}`
- **Agrément Établissement Pharmaceutique n° :** `{{establishment.license_number}}`
- **Adresse de l'Unité de Production Industrielle :** `{{manufacturing_site.address}}`

---

## 2. ENGAGEMENT SOLENNEL DU PHARMACIEN DIRECTEUR TECHNIQUE

Je soussigné, **`Dr. {{pdt.full_name}}`**, titulaire du diplôme de Docteur en Pharmacie, inscrit à la Section Ordinale Régionale des Pharmaciens sous le numéro **`{{pdt.ordinal_number}}`**, exerçant les fonctions de **Pharmacien Directeur Technique** au sein de l'établissement pharmaceutique **`{{establishment.name}}`** ;

Conformément aux dispositions impératives du **Décret exécutif n° 20-325 du 22 novembre 2020** relatif à l'enregistrement des produits pharmaceutiques et aux exigences techniques du **Guide de soumission CTD de l'ANPP (Version Juillet 2025)** :

### 2.1 Engagement de Réalisation de la Validation Industrielle
Je m'engage formellement et solennellement à :
1. Réaliser la **validation prospective complète du procédé de fabrication** sur **trois (03) lots consécutifs de taille industrielle commerciale**, conformément au protocole de validation validé ci-joint (Module 3.2.P.3.5) et aux règles des Bonnes Pratiques de Fabrication (BPF) en vigueur.
2. Effectuer l'ensemble des contrôles en cours de procédé (IPC) et les contrôles analytiques libératoires sur les matières premières, articles de conditionnement et produits finis, selon les spécifications et méthodes analytiques validées déposées dans le dossier d'enregistrement.
3. Consigner l'ensemble des données brutes, déviations éventuelles, investigations OOS/OOT et résultats de capabilité procédé ($C_{pk} \ge 1.33$) dans des rapports formels de validation de procédé et de nettoyage.

### 2.2 Engagement Absolu de Non-Commercialisation Avant Accord
Je m'engage formellement et sous ma responsabilité ordinale, civile et pénale :
> **À NE PROCÉDER À AUCUNE COMMERCIALISATION, DISTRIBUTION, DISPENSATION OU MISE SUR LE MARCHÉ NATIONAL** de l'un quelconque des lots fabriqués, avant d'avoir transmis à l'Agence Nationale des Produits Pharmaceutiques (ANPP) les rapports complets de validation des trois (03) lots industriels consécutifs, accompagnés des bulletins d'analyses libératoires conformes, et d'avoir obtenu l'accord formel écrit de levée de réserve émis par l'ANPP.

### 2.3 Suivi Stabilité des Lots de Validation
Je m'engage à placer immédiatement des échantillons représentatifs de ces trois (03) lots industriels en études de stabilité en temps réel sous les conditions climatiques réglementaires de la **Zone IVb ($30^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$)** et à en transmettre annuellement les résultats d'étape à l'ANPP.

---

## 3. PLANIFICATION ET TAILLES DES TROIS LOTS DE VALIDATION

| N° Lot Prévu | Taille Unitaire du Lot (Unités / Boîtes) | Équipements Critiques Qualifiés | Période Prévisionnelle de Production | Statut Réglementaire |
|:---|:---|:---|:---|:---|
| **Lot Industriel V01** | `{{validation.batch_size_1}}` | Ligne de fabrication n° `{{validation.line_id}}` | Semaine `{{validation.week_1}}` | Soumis à validation |
| **Lot Industriel V02** | `{{validation.batch_size_2}}` | Ligne de fabrication n° `{{validation.line_id}}` | Semaine `{{validation.week_2}}` | Soumis à validation |
| **Lot Industriel V03** | `{{validation.batch_size_3}}` | Ligne de fabrication n° `{{validation.line_id}}` | Semaine `{{validation.week_3}}` | Soumis à validation |

---

## 4. SIGNATURES ÉLECTRONIQUES QUALIFIÉES ET HORODATAGE SÉCURISÉ

| Niveau de Responsabilité | Nom, Prénom & Titre | Rôle Système Homolog | Horodatage Certifié UTC | Empreinte Numérique (SHA-256) |
|:---|:---|:---|:---|:---|
| **Rédacteur Qualité (*Maker*)** | `{{audit.maker_name}}` | Responsable AQ / Validation | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Vérificateur Production (*Checker*)** | `{{audit.checker_name}}` | Responsable de Production | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approbateur Légal (*Approver*)** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
