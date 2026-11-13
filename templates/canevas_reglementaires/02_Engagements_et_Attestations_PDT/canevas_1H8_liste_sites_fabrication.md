---
document_metadata:
  code_ctd_anpp: "1.H.8"
  title: "Liste Exhaustive et Déclaration des Sites de Fabrication, Conditionnement, Contrôle et Libération des Lots"
  legal_reference: "Décret exécutif n° 21-82 du 23 février 2021 ; Guide CTD ANPP (Juillet 2025) Module 1.H.8 ; Bonnes Pratiques de Fabrication"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI"]
  security_roles:
    maker: "Chargé Affaires Réglementaires"
    checker: "Responsable Assurance Qualité / Fournisseurs"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.2 - Qualification des Acteurs Industriels & Inspections BPF"
  alcoa_compliance: "21 CFR Part 11 Compliant - Traçabilité Géographique et Réglementaire"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT DES PRODUITS PHARMACEUTIQUES
### MODULE 1 - SECTION 1.H.8 : LISTE OFFICIELLE DES SITES INTERVENANTS

---

## 1. IDENTIFICATION DU MÉDICAMENT

- **Dénomination Commerciale :** `{{product.brand_name}}`
- **Dosage et Forme :** `{{product.dosage}}`, `{{product.pharmaceutical_form}}`
- **Numéro Dossier ANPP :** `{{dossier.reference_anpp}}`
- **Titulaire / Exploitant Déposant :** `{{establishment.name}}`

---

## 2. RÉPERTOIRE EXHAUSTIF DES SITES INDUSTRIELS INTERVENANTS

### 2.1 Site(s) de Fabrication de la Substance Active (Principe Actif)
- **Raison Sociale du Fabricant :** `{{site_api.name}}`
- **Adresse Physique Complète :** `{{site_api.street}}, {{site_api.postal_code}} {{site_api.city}}`
- **Pays :** `{{site_api.country}}`
- **Coordonnées GPS :** Latitude `{{site_api.gps_lat}}`, Longitude `{{site_api.gps_long}}`
- **Référence CEP / DMF / ASMF :** `{{site_api.cep_or_dmf_number}}`
- **Numéro du Certificat BPF (GMP) :** `{{site_api.gmp_certificate_number}}` délivré par `{{site_api.gmp_authority}}` le `{{site_api.gmp_issue_date}}`.

### 2.2 Site(s) de Fabrication de la Forme Vrac (Produit Intermédiaire)
- **Raison Sociale :** `{{site_bulk.name}}`
- **Adresse Physique Complète :** `{{site_bulk.street}}, {{site_bulk.city}}, {{site_bulk.country}}`
- **Numéro d'Autorisation d'Exploitation / Agrément :** `{{site_bulk.license_number}}`
- **Périmètre des Activités Réalisées :** Pesée, granulation, séchage, compression/remplissage de la forme nue vrac.

### 2.3 Site(s) de Conditionnement Primaire (Contact Direct)
- **Raison Sociale :** `{{site_primary.name}}`
- **Adresse Physique Complète :** `{{site_primary.street}}, {{site_primary.city}}, {{site_primary.country}}`
- **Type de Conditionnement Réalisé :** Mise sous blister thermoformé Alu/PVC ou flaconnage.

### 2.4 Site(s) de Conditionnement Secondaire (Emballage Extérieur)
- **Raison Sociale :** `{{site_secondary.name}}`
- **Adresse Physique Complète :** `{{site_secondary.street}}, {{site_secondary.city}}, {{site_secondary.country}}`
- **Opérations Réalisées :** Mise en étui, insertion de la notice patient bilingue FR/AR, application du numéro de lot, de la date de péremption et du code DataMatrix.

### 2.5 Site(s) de Contrôle Qualité Libératoire (Physico-Chimique et Microbiologique)
- **Raison Sociale du Laboratoire :** `{{site_qc.name}}`
- **Adresse Physique :** `{{site_qc.street}}, {{site_qc.city}}, {{site_qc.country}}`
- **Accréditation / Reconnaissance BPF :** `{{site_qc.accreditation_reference}}`
- **Nature des Contrôles :** Contrôle libératoire complet sur produit fini conditionné.

### 2.6 Établissement Responsable de la Libération des Lots en Algérie
- **Raison Sociale :** `{{establishment.name}}`
- **Adresse de l'Établissement Pharmaceutique :** `{{establishment.address}}`
- **Pharmacien Directeur Technique Responsable de la Libération :** `Dr. {{pdt.full_name}}` (Ordre n° `{{pdt.ordinal_number}}`)

---

## 3. DÉCLARATION SUR L'HONNEUR DU TITULAIRE

Le Pharmacien Directeur Technique certifie sous sa responsabilité que :
1. Aucun autre site que ceux expressément listés ci-dessus n'intervient dans le cycle de production ou d'analyse du produit fini.
2. Tous les sites mentionnés disposent d'un agrément ou d'une autorisation de fabrication valide délivrée par les autorités sanitaires compétentes de leur pays d'implantation.
3. Des contrats techniques et d'assurance qualité (*Quality Technical Agreements*) conformes aux BPF ont été formellement signés entre l'exploitant et chacun des sites sous-traitants.

---

## 4. VALIDATION ET SIGNATURES ÉLECTRONIQUES

| Étape de Signature | Signataire | Titre Professionnel | Date & Heure UTC | Empreinte SHA-256 |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Spécialiste Réglementaire | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Responsable Audit Qualité | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
