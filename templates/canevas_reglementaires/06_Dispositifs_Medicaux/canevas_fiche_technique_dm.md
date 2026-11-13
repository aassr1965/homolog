---
document_metadata:
  code_ctd_anpp: "DM-TECH-02"
  title: "Fiche Technique Réglementaire du Dispositif Médical (Arrêté du 10 mai 2021)"
  legal_reference: "Arrêté du 10 mai 2021 fixant les modalités d'homologation des DM ; Séries ISO 10993, ISO 13485, ISO 14971, ISO 15223-1"
  target_procedures: ["DM-INIT", "DM-RENOUV", "DM-MODIF"]
  security_roles:
    maker: "Ingénieur Biomédical / Qualité DM"
    checker: "Pharmacien Responsable Dispositifs Médicaux"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.2 - Expertise Technique et Clinique de la Commission DM"
  alcoa_compliance: "21 CFR Part 11 Compliant - Fiche Technique Conforme Arrêté 10/05/2021"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'HOMOLOGATION DES DISPOSITIFS MÉDICAUX
### FICHE TECHNIQUE RÉGLEMENTAIRE OFFICIELLE (ARRÊTÉ DU 10 MAI 2021)

---

## 1. IDENTIFICATION TECHNIQUE DU DISPOSITIF MÉDICAL

- **Dénomination Commerciale :** `{{dm.commercial_name}}`
- **Type de Dispositif Médical :** `{{dm.device_type}}`
- **Classe de Risque Réglementaire :** Classe `{{dm.risk_class}}`
- **Code Nomenclature Internationale :** GMDN `{{dm.gmdn_code}}`
- **Durée de Vie Utile / Validité :** `{{dm.shelf_life_years}}` ans à compter de la date de stérilisation / fabrication.
- **Conditions de Stockage :** Conserver entre $5^\circ\text{C}$ et $30^\circ\text{C}$, dans un endroit sec, à l'abri du soleil direct.

---

## 2. CARACTÉRISTIQUES PHYSIQUES ET DIMENSIONNELLES

| Caractéristique | Spécification Technique / Tolérance | Méthode d'Essai / Norme de Référence |
|:---|:---|:---|
| **Dimensions Globales** | Longueur : `{{dm.dim_length}}` mm ; Diamètre : `{{dm.dim_diameter}}` mm | Mesure optique / Jauge micrométrique |
| **Volume Nominal / Capacité**| `{{dm.nominal_volume}}` ml (Tolérance : $\pm 2.0\%$) | ISO 7886-1 |
| **Résistance Mécanique** | Résistance à la traction $\ge 15\,\text{N}$ ; Résistance rupture | ISO 11070 |
| **Étanchéité Hydraulique/Air**| Maintien sous pression de $300\,\text{kPa}$ pendant 30 s sans fuite | ISO 7886-1 Annexe B |
| **Force d'Actionnement** | Force de glissement du piston $\le 8.0\,\text{N}$ | Banc de traction dynamométrique |

---

## 3. MATÉRIAUX CONSTITUTIFS ET SÉCURITÉ CHIMIQUE

*Description des matériaux entrant en contact direct ou indirect avec le patient ou les fluides biologiques :*

| Pièce / Composant | Matériau Chimique Spécifié | Grade / Référence Fournisseur | Contact Patient (Direct / Indirect / Aucun) | Absence Phtalates / Latex |
|:---|:---|:---|:---|:---|
| **Corps Principal** | Polypropylène médical (PP) | Grade Médical USP Classe VI | Contact fluide injectable | Sans DEHP, Sans Latex |
| **Piston / Joint** | Élastomère thermoplastique (TPE) | Conforme Ph. Eur. 3.2.9 | Contact fluide injectable | Sans Latex naturel |
| **Aiguille / Canule** | Acier inoxydable AISI 304 | Conforme ISO 9626 | Contact invasif transitoire | Inerte, biocompatible |
| **Connecteur Luer** | Polycarbonate (PC) | Conforme ISO 80369-7 | Contact circuit de perfusion | Sans bisphénol A (BPA) |

---

## 4. BIOCOMPATIBILITÉ ET ÉVALUATION BIOLOGIQUE (SÉRIE ISO 10993)

*Conformément à la norme ISO 10993-1 pour les dispositifs en contact prolongé ou invasif :*

| Type d'Essai Biologique | Norme Spécifique | Résultat de l'Essai | Statut de Conformité |
|:---|:---|:---|:---|
| **Cytotoxicité In Vitro** | ISO 10993-5 | Viabilité cellulaire $> 85\%$ (Test MTT) | **CONFORME - NON CYTOTOXIQUE** |
| **Irritation / Réactivité Cutanée**| ISO 10993-10 | Indice d'irritation primaire $PII = 0.0$ | **CONFORME - NON IRRITANT** |
| **Sensibilisation Cutanée** | ISO 10993-10 | Test de maximisation chez le cobaye négatif | **CONFORME - NON SENSIBILISANT** |
| **Toxicité Systémique Aiguë** | ISO 10993-11 | Aucun signe clinique d'intoxication systémique | **CONFORME - NON TOXIQUE** |
| **Pyrogénicité** | ISO 10993-11 | Absence de réaction fébrile (Test pyrogènes) | **CONFORME - APYROGÈNE** |
| **Hémocompatibilité** | ISO 10993-4 | Taux d'hémolyse $< 2.0\%$, absence de thrombus | **CONFORME - HÉMOCOMPATIBLE** |

---

## 5. VALIDATION DU PROCÉDÉ DE STÉRILISATION (POUR DISPOSITIFS STÉRILES)

- **Méthode de Stérilisation :** Oxyde d'Éthylène (ETO) selon la norme **ISO 11135:2014**.
- **Niveau d'Assurance de Stérilité (SAL) :** $\mathbf{SAL = 10^{-6}}$ démontré par la méthode du demi-cycle (*Overkill approach*).
- **Résidus de Stérilisation (ISO 10993-7) :**
  - Teneur résiduelle en Oxyde d'Éthylène (EO) : $< 4.0\,\text{mg/dispositif}$ (Limite admissible : $< 10\,\text{mg}$).
  - Teneur en Chlorhydrine d'Éthylène (ECH) : $< 2.0\,\text{mg/dispositif}$ (Limite admissible : $< 5\,\text{mg}$).
- **Intégrité du Système de Barrière Stérile (ISO 11607-1/2) :**
  - Emballage sachet pelable papier médical / film plastique thermocollé.
  - Résistance de scellage $> 1.5\,\text{N/15 mm}$ et étanchéité validée par test aux bulles (ASTM F2096).

---

## 6. ÉTIQUETAGE, NOTICE D'INSTRUCTION ET PICTOGRAMMES (ISO 15223-1)

- **Notice d'Utilisation Bilingue :** Rédaction conjointe en langues **Arabe et Française** décrivant le mode opératoire, les précautions et les contre-indications.
- **Pictogrammes Normalisés Apposés sur le Conditionnement :**
  - Symbole Stérile : `STERILE|EO` (Stérilisé à l'oxyde d'éthylène)
  - Symbole Usage Unique : `②` (Ne pas réutiliser)
  - Symbole Date Limite d'Utilisation : `⌛` (AAAA-MM)
  - Symbole Numéro de Lot : `LOT`
  - Symbole Attention / Consulter la Notice : `⚠` / `📖`
  - Symbole Sans Latex : `LATEX FREE`

---

## 7. VISAS ET SIGNATURES TECHNIQUES

| Rôle BPF / DM | Nom & Prénom | Titre Professionnel | Date & Heure UTC | Empreinte Numérique (SHA-256) |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.biomedical_engineer_name}}`| Ingénieur Biomédical | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.qa_dm_director_name}}` | Responsable Assurance Qualité| `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
