---
document_metadata:
  code_ctd_anpp: "DM-HOMOLOG-01"
  title: "Formulaire de Demande d'Homologation de Dispositif Médical à Usage Humain"
  legal_reference: "Décret exécutif n° 20-324 du 22 novembre 2020 ; Arrêté du 10 mai 2021 fixant les modalités d'homologation des dispositifs médicaux"
  target_procedures: ["DM-INIT", "DM-RENOUV", "DM-MODIF"]
  security_roles:
    maker: "Chargé Réglementaire Dispositifs Médicaux"
    checker: "Pharmacien Responsable Dispositifs Médicaux"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.1 - Dépôt de Dossier d'Homologation DM"
  alcoa_compliance: "21 CFR Part 11 Compliant - Traçabilité Réglementaire Dispositifs Médicaux"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'HOMOLOGATION DES DISPOSITIFS MÉDICAUX
### FORMULAIRE OFFICIEL DE DEMANDE D'HOMOLOGATION (DÉCRET EXÉCUTIF 20-324)

---

## 1. NATURE ET OBJET DE LA DEMANDE

- [ ] **Homologation Initiale** d'un nouveau Dispositif Médical
- [ ] **Renouvellement quinquennal** d'une décision d'homologation existante
- [ ] **Modification substantielle** des termes d'une décision d'homologation

*Référence Interne Dossier Homolog :* **`{{dm.dossier_reference}}`**  
*Date de Soumission :* **`{{dm.submission_date}}`**

---

## 2. IDENTIFICATION ET CLASSIFICATION DU DISPOSITIF MÉDICAL

### 2.1 Informations Produit
- **Dénomination Commerciale :** `{{dm.commercial_name}}`
- **Modèle(s) / Référence(s) Catalogue :** `{{dm.model_references_list}}`
- **Nomenclature Internationale (Code GMDN / EMDN) :** `{{dm.gmdn_code}}` - `{{dm.gmdn_term}}`
- **Statut de Stérilité :**
  - [ ] Dispositif Médical Stérile (Procédé : `{{dm.sterilization_method}}`)
  - [ ] Dispositif Médical Non Stérile

### 2.2 Classification de Risque (Décret 20-324 & Arrêté du 10 mai 2021)
- [ ] **Classe I** (Faible risque - Non stérile, sans mesurage)
- [ ] **Classe I Stérile ($Is$)** ou avec **Fonction de Mesurage ($Im$)**
- [ ] **Classe IIa** (Risque potentiel modéré)
- [ ] **Classe IIb** (Risque potentiel élevé)
- [ ] **Classe III** (Très haut risque - Dispositifs critiques, implantables à long terme)
- [ ] **DMIA** (Dispositifs Médicaux Implantables Actifs)

*Règle de classification appliquée (Règles 1 à 22) :* **Règle `{{dm.classification_rule_number}}`**  
*Justification de la classe :* `{{dm.classification_justification}}`

### 2.3 Destination Médicale et Revendications Cliniques
- **Usage Prévu :** `{{dm.intended_use_description}}`
- **Indications Thérapeutiques ou Diagnostiques :** `{{dm.clinical_indications}}`
- **Population Cible et Utilisateurs :** `{{dm.target_population_and_users}}` (Professionnels de santé en milieu hospitalier ou grand public).

---

## 3. IDENTIFICATION DU DEMANDEUR ET DES SITES DE PRODUCTION

### 3.1 Établissement Demandeur / Exploitant Déposant (Algérie)
- **Raison Sociale :** `{{establishment.name}}`
- **Statut Juridique :** Fabricant Local / Importateur Distributeur Agréé
- **Numéro d'Agrément DM délivré par le MSPRH / ANPP :** `{{establishment.dm_license_number}}`
- **Adresse Complète :** `{{establishment.address}}`
- **Pharmacien Directeur Technique / Responsable DM :** `Dr. {{pdt.full_name}}` (Ordre n° `{{pdt.ordinal_number}}`)

### 3.2 Fabricant Légal (Legal Manufacturer)
- **Raison Sociale :** `{{dm.legal_manufacturer_name}}`
- **Adresse du Siège :** `{{dm.legal_manufacturer_address}}`, Pays : `{{dm.legal_manufacturer_country}}`
- **Certificat Système de Gestion de la Qualité :** **ISO 13485:2016** n° `{{dm.iso13485_cert_number}}` délivré par `{{dm.notified_body_name}}`, valide jusqu'au `{{dm.iso13485_expiry_date}}`.
- **Déclaration de Conformité CE / Marquage :** Certificat n° `{{dm.ce_certificate_number}}` émis par l'Organisme Notifié n° `{{dm.notified_body_id}}`.

### 3.3 Sites de Fabrication et de Stérilisation
- **Site de Production Physique :** `{{dm.manufacturing_site_name}}`, `{{dm.manufacturing_site_address}}`
- **Site de Stérilisation (si sous-traitée) :** `{{dm.sterilization_site_name}}`, `{{dm.sterilization_site_address}}` (Certifié ISO 11135 / ISO 11137).

---

## 4. BORDEREAU DES PIÈCES DU DOSSIER TECHNIQUE D'HOMOLOGATION

- [x] Copie de l'agrément de l'établissement demandeur.
- [x] Mandat officiel de représentation exclusive dûment légalisé liant le fabricant au demandeur local.
- [x] Certificat de conformité ISO 13485 du fabricant légal en cours de validité.
- [x] Certificat de marquage CE ou certificat de libre vente (CLV) émis par l'autorité sanitaire du pays d'origine.
- [x] Fiche technique réglementaire détaillée (Arrêté du 10 mai 2021).
- [x] Rapport d'évaluation clinique et données d'investigation clinique (ISO 14155).
- [x] Rapport d'analyse et de gestion des risques (ISO 14971).
- [x] Rapports de biocompatibilité (ISO 10993).
- [x] Validation de stérilisation et tests de vieillissement packaging.
- [x] Maquettes d'étiquetage et notices d'utilisation bilingues (Arabe / Français).
- [x] Quittance de versement de la redevance réglementaire BV1 (Décret 20-325).

---

## 5. ENGAGEMENT DU PHARMACIEN DIRECTEUR TECHNIQUE

Je soussigné, **`Dr. {{pdt.full_name}}`**, Pharmacien Directeur Technique :
1. Atteste de l'exactitude absolue des données figurant dans la présente demande et ses pièces jointes.
2. M'engage à assurer la matériovigilance continue conformément à la réglementation algérienne et à notifier sans délai tout incident ou risque d'incident grave mettant en jeu la vie d'un patient.
3. M'engage à garantir la traçabilité des lots et numéros de série mis sur le marché algérien.

---

## 6. VALIDATION ET SIGNATURES ÉLECTRONIQUES

| Rôle BPF / DM | Nom & Prénom | Titre Professionnel | Date & Heure UTC | Empreinte Numérique SHA-256 |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.dm_specialist_name}}` | Chargé Affaires Réglem. DM | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.dm_pharmacist_name}}`| Pharmacien Responsable DM | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
