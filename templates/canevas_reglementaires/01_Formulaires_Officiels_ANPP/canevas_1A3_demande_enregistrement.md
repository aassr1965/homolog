---
document_metadata:
  code_ctd_anpp: "1.A.3"
  title: "Formulaire de Demande d'Enregistrement de Médicament à Usage Humain"
  legal_reference: "Décret exécutif n° 20-325 du 22 novembre 2020 ; Arrêté du 16 janvier 2021 ; Guide CTD ANPP (Juillet 2025)"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI"]
  security_roles:
    maker: "Regulatory Affairs Associate / Lead"
    checker: "Regulatory Affairs Manager"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.1 - Dépôt de Dossier & Recevabilité Administrative"
  alcoa_compliance: "21 CFR Part 11 Compliant - Immutable upon Approval"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT DES PRODUITS PHARMACEUTIQUES
### MODULE 1 - SECTION 1.A.3 : FORMULAIRE DE DEMANDE D'ENREGISTREMENT

---

## 1. NATURE ET OBJET DE LA DEMANDE

### 1.1 Type de Procédure Réglementaire
- [ ] **NF** : Médicament issu d'une fabrication locale intégrale (Nouvelle Enregistrement Fabrication)
- [ ] **NI** : Médicament importé (Nouvel Enregistrement Importation)
- [ ] **RF** : Renouvellement quinquennal d'un produit fabriqué localement
- [ ] **RI** : Renouvellement quinquennal d'un produit importé
- [ ] **VF** : Variation d'un médicament de fabrication locale (LD-DEPP07/V01)
- [ ] **VI** : Variation d'un médicament importé (LD-DEPP07/V01)

*Valeur dynamique injectée :* **`{{dossier.type_procedure}}`**  
*Numéro d'identification dossier ANPP :* **`{{dossier.reference_anpp}}`**  
*Date de soumission :* **`{{dossier.submission_date}}`**

### 1.2 Justification du Type de Dossier (Base Juridique)
- [ ] Médicament générique (Article 8 du Décret 20-325) - Référence princeps : `{{dossier.reference_originator_name}}`
- [ ] Médicament princeps / nouvelle molécule
- [ ] Médicament à usage bien établi / traditionnellement utilisé
- [ ] Médicament hybride / composition modifiée
- [ ] Médicament biosimilaire

---

## 2. IDENTIFICATION DE L'ÉTABLISSEMENT DEMANDEUR

| Rubrique Réglementaire | Information Requise | Données Enregistrées (PostgreSQL) |
|:---|:---|:---|
| **Dénomination Sociale** | Nom légal de l'établissement exploitant | `{{establishment.name}}` |
| **Forme Juridique** | Statut (SPA, SARL, EURL, etc.) | `{{establishment.legal_status}}` |
| **Numéro d'Agrément MSPRH/ANPP**| Numéro de décision d'agrément de l'exploitant | `{{establishment.license_number}}` |
| **Date d'Octroi de l'Agrément** | Date de délivrance officielle | `{{establishment.license_grant_date}}` |
| **Adresse du Siège Social** | Adresse postale complète, Wilaya | `{{establishment.address}}, {{establishment.wilaya}}` |
| **Téléphone / Télécopie** | Lignes directes professionnelles | `{{establishment.phone}} / {{establishment.fax}}` |
| **Adresse Électronique Officielle**| Courriel institutionnel sécurisé | `{{establishment.email}}` |
| **Pharmacien Directeur Technique** | Nom, Prénom, N° d'inscription à l'Ordre | `Dr. {{pdt.full_name}}` (Ordre n° : `{{pdt.ordinal_number}}`) |

---

## 3. IDENTIFICATION DU MÉDICAMENT

### 3.1 Caractéristiques Galéniques et Administratives
- **Dénomination Commerciale (Nom de Marque) :** `{{product.brand_name}}`
- **Dénomination Commune Internationale (DCI) :** `{{product.dci}}`
- **Dosage en Principe Actif :** `{{product.dosage}}`
- **Forme Pharmaceutique :** `{{product.pharmaceutical_form}}`
- **Voie(s) d'Administration :** `{{product.route_of_administration}}`
- **Classe Pharmaco-Thérapeutique (Code ATC) :** `{{product.atc_code}}` - `{{product.atc_description}}`

### 3.2 Présentation et Conditionnement
- **Conditionnement Primaire (au contact immédiat) :** `{{product.primary_packaging_description}}` (Ex : Blister Alu/PVC, Flacon PEHD, Ampoule verre type I).
- **Conditionnement Secondaire (emballage extérieur) :** `{{product.secondary_packaging_description}}` (Ex : Boîte carton pliante contenant 30 unités avec notice).
- **Tailles de Conditionnement (Présentations soumises) :** `{{product.pack_sizes}}`
- **Durée de Conservation (Validité) :** `{{product.shelf_life_months}}` mois.
- **Conditions de Conservation Réglementaires :** Conserver à une température ne dépassant pas $30^\circ\text{C}$, à l'abri de l'humidité et de la lumière (**Conditions Climatiques Zone IVb algérienne**).

---

## 4. FORMULE DE COMPOSITION QUALITATIVE ET QUANTITATIVE

*Formule unitaire pour une unité de prise (ou pour $100\,\text{ml}$ / $100\,\text{g}$) :*

| Rôle du Composant | Nom de la Substance (DCI / Nom Chimique) | Teneur par Unité | Norme de Référence (Ph. Eur. / USP / Interne) | Fonction Galénique |
|:---|:---|:---|:---|:---|
| **Substance Active (PA)** | `{{active_ingredient.name}}` | `{{active_ingredient.quantity}}` | `{{active_ingredient.monograph}}` | Principe Actif |
| **Excipient 1** | `{{excipient_1.name}}` | `{{excipient_1.quantity}}` | `{{excipient_1.monograph}}` | Diluant / Liant |
| **Excipient 2** | `{{excipient_2.name}}` | `{{excipient_2.quantity}}` | `{{excipient_2.monograph}}` | Agent de désagrégation |
| **Excipient 3** | `{{excipient_3.name}}` | `{{excipient_3.quantity}}` | `{{excipient_3.monograph}}` | Lubrifiant |
| **Excipient à Effet Notoire** | `{{excipient_notoire.name}}` *(si applicable)* | `{{excipient_notoire.quantity}}` | `{{excipient_notoire.monograph}}` | Excipient à effet notoire |

---

## 5. LISTE DES SITES INTERVENANTS ET RESPONSABILITÉS

| Rôle Industriel | Raison Sociale de l'Établissement | Adresse Physique du Site | N° Agrément / Autorisation | Pays |
|:---|:---|:---|:---|:---|
| **Fabricant de la Substance Active** | `{{site_api.name}}` | `{{site_api.address}}` | DMF/CEP n° `{{site_api.dmf_number}}` | `{{site_api.country}}` |
| **Fabricant de la Forme Vrac** | `{{site_bulk.name}}` | `{{site_bulk.address}}` | Agrément n° `{{site_bulk.license_number}}` | `{{site_bulk.country}}` |
| **Conditionnement Primaire** | `{{site_primary_pack.name}}` | `{{site_primary_pack.address}}` | Agrément n° `{{site_primary_pack.license_number}}` | `{{site_primary_pack.country}}` |
| **Conditionnement Secondaire** | `{{site_secondary_pack.name}}` | `{{site_secondary_pack.address}}`| Agrément n° `{{site_secondary_pack.license_number}}`| `{{site_secondary_pack.country}}`|
| **Contrôle Qualité Libératoire**| `{{site_qc.name}}` | `{{site_qc.address}}` | Certificat BPF n° `{{site_qc.gmp_certificate}}` | `{{site_qc.country}}` |
| **Libération Finale des Lots** | `{{site_release.name}}` | `{{site_release.address}}` | Agrément n° `{{site_release.license_number}}` | Algérie |

---

## 6. SITUATION RÉGLEMENTAIRE INTERNATIONALE (Pour Médicaments Importés - NI / RI)

- **Pays de Fabrication / Provenance :** `{{origin.country}}`
- **Numéro d'Autorisation de Mise sur le Marché dans le pays d'origine :** `{{origin.marketing_auth_number}}`
- **Date de première autorisation dans le pays d'origine :** `{{origin.marketing_auth_date}}`
- **Statut du Certificat de Produit Pharmaceutique (CPP OMS) :** Fourni conformément au modèle recommandé par l'Organisation Mondiale de la Santé, datant de moins de deux ans et dûment légalisé.

---

## 7. ATTESTATION SUR L'HONNEUR ET ENGAGEMENT LÉGAL DU DEMANDEUR

Je soussigné, **`Dr. {{pdt.full_name}}`**, en ma qualité de **Pharmacien Directeur Technique** de l'établissement pharmaceutique **`{{establishment.name}}`** :
1. Certifie que toutes les informations fournies dans le présent formulaire de demande et dans l'ensemble des modules du dossier CTD ci-annexé sont exactes, véridiques et conformes à la réglementation algérienne en vigueur.
2. M'engage formellement à respecter l'ensemble des dispositions du Décret exécutif n° 20-325 du 22 novembre 2020 relatif à l'enregistrement des produits pharmaceutiques.
3. M'engage à n'apporter aucune modification de formule, de spécification, de procédé de fabrication ou d'élément d'emballage sans autorisation préalable expresse de l'ANPP (selon les directives de variations LD-DEPP07/V01).
4. Déclare sur l'honneur que le produit sera fabriqué et contrôlé en stricte conformité avec les Bonnes Pratiques de Fabrication (BPF) en vigueur en République Algérienne Démocratique et Populaire.

---

## 8. GRILLE DE VALIDATION ÉLECTRONIQUE ET SÉGRÉGATION DES TÂCHES (SoD)

```
[ÉTAPE 1 : RÉDACTION] ──────────► [ÉTAPE 2 : VÉRIFICATION] ──────────► [ÉTAPE 3 : APPROBATION]
   Chargé Affaires Réglem.           Manager Réglementaire               Pharmacien Dir. Technique
   (Signature Électronique)          (Contrôle Qualité/Conformité)       (Engagement Légal & Ordinal)
```

| Rôle BPF / ALCOA+ | Nom & Prénom | Fonction | Date & Heure UTC | Empreinte Numérique (SHA-256) | Statut / Décision |
|:---|:---|:---|:---|:---|:---|
| **Rédacteur (*Maker*)** | `{{audit.maker_name}}` | Chargé Affaires Réglementaires | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` | **VALIDÉ** |
| **Vérificateur (*Checker*)**| `{{audit.checker_name}}` | Manager Affaires Réglementaires| `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` | **VÉRIFIÉ ET CONFORME** |
| **Approbateur (*Approver*)**| `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` | **APPROUVÉ - BON POUR DÉPÔT** |
