---
document_metadata:
  code_ctd_anpp: "2.3"
  title: "Quality Overall Summary (QOS) - Résumé Global de la Qualité (ICH M4Q)"
  legal_reference: "Guide CTD ANPP (Juillet 2025) Module 2.3 ; Directive ICH M4Q(R1) ; Pharmacopée Européenne en vigueur"
  target_procedures: ["NF", "NI", "RF", "RI"]
  security_roles:
    maker: "Expert Qualité / Galéniste Agréé"
    checker: "Manager Affaires Réglementaires / Contrôle Qualité"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.2 - Évaluation Technique du Dossier Chimico-Pharmaceutique"
  alcoa_compliance: "21 CFR Part 11 Compliant - Rapport d'Expertise Technique Signé"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT - DIVISION TECHNIQUE QUALITÉ
### MODULE 2.3 : QUALITY OVERALL SUMMARY (QOS) - RÉSUMÉ GLOBAL DE LA QUALITÉ

---

## 1. INTRODUCTION ET INFORMATIONS GÉNÉRALES SUR LE PRODUIT

- **Dénomination Commerciale :** `{{product.brand_name}}`
- **Substance Active (DCI) :** `{{product.dci}}`
- **Dosage et Forme Galénique :** `{{product.dosage}}`, `{{product.pharmaceutical_form}}`
- **Voie d'Administration :** `{{product.route_of_administration}}`
- **Demandeur / Exploitant :** `{{establishment.name}}`
- **Fabricant du Produit Fini :** `{{site_bulk.name}}` (`{{site_bulk.country}}`)
- **Numéro Dossier ANPP :** `{{dossier.reference_anpp}}`

---

## 2. SECTION 2.3.S : SUBSTANCE ACTIVE (DRUG SUBSTANCE)

### 2.3.S.1 Informations Générales
- **Nomenclature :** DCI : `{{product.dci}}` ; Nom chimique UICPA : `{{active_substance.iupac_name}}` ; N° CAS : `{{active_substance.cas_number}}`.
- **Formule Brute et Masse Moléculaire :** `{{active_substance.chemical_formula}}` (Masse : `{{active_substance.molecular_weight}}` g/mol).
- **Structure Chimique :** `{{active_substance.chemical_structure_description}}`.
- **Propriétés Physico-chimiques :** Aspect (poudre blanche ou presque blanche), solubilité aqueuse en fonction du pH (classification BCS : Classe `{{active_substance.bcs_class}}`), polymorphisme cristallin (Forme `{{active_substance.polymorphic_form}}`), hygroscopicité, point de fusion ($`{{active_substance.melting_point}}`^\circ\text{C}$).

### 2.3.S.2 Fabrication de la Substance Active
- **Fabricant(s) Qualifié(s) :** `{{site_api.name}}`, Adresse : `{{site_api.address}}`.
- **Procédure d'Accès au Dossier :** Certificat de Conformité à la Pharmacopée Européenne (CEP n° `{{site_api.cep_number}}`) ou Dossier Permanent de la Substance Active (ASMF/DMF n° `{{site_api.dmf_number}}`).
- **Schéma de Synthèse :** Voie de synthèse chimique multi-étapes résumée, solvants et réactifs utilisés, étapes critiques et contrôles en cours (IPC).

### 2.3.S.3 Caractérisation et Profil d'Impuretés
- **Impuretés Organiques Liées au Procédé :** Conformes aux seuils de qualification de l'ICH Q3A.
- **Solvants Résiduels :** Conformes aux limites de l'ICH Q3C (Solvants de classe 2 et 3).
- **Impuretés Élémentaires :** Évaluation des risques conforme à l'ICH Q3D démontrant l'absence de dépassement des limites d'exposition quotidienne permise (PDE).
- **Impuretés Génotoxiques / Nitrosamines :** Évaluation du risque de contamination par les nitrosamines (ICH M7) démontrant l'absence de risque avéré (ou teneur < seuil toxicologique acceptable $AI < 18\,\text{ng/jour}$).

### 2.3.S.4 Contrôle de la Substance Active
- **Spécifications :** Conformes à la monographie de la Pharmacopée Européenne n° `{{active_substance.monograph_number}}`.
- **Méthodes d'Analyse :** Validées selon l'ICH Q2(R1/R2) pour le dosage et les substances apparentées.

### 2.3.S.5 Système de Fermeture et Stabilité
- **Emballage :** Double sachet en polyéthylène vierge de qualité pharmaceutique placé dans un fût en polyéthylène haute densité.
- **Période de Recontrôle (*Retest Period*) :** `{{active_substance.retest_period_months}}` mois à température ambiante ($< 25^\circ\text{C}$ ou $< 30^\circ\text{C}$).

---

## 3. SECTION 2.3.P : PRODUIT FINI (DRUG PRODUCT)

### 2.3.P.1 Description et Composition du Médicament
- **Composition Qualitative et Quantitative Complète :** Tableau détaillé incluant la substance active et tous les excipients (liants, diluants, lubrifiants, délitants, pelliculage) avec justification de leurs rôles fonctionnels.
- **Type de Conditionnement :** Blister Alu/PVC thermoformé étanche à l'humidité ou flacon PEHD avec bouchon sécurité-enfant.

### 2.3.P.2 Développement Pharmaceutique (ICH Q8)
- **Compatibilité Substance Active / Excipients :** Démontrée par études binaires thermo-analytiques (DSC, FTIR) et de stabilité accélérée.
- **Attributs de Qualité Critiques (CQA) :** Dissolution *in vitro*, teneur, homogénéité d'unité de prise, profil de dégradation.
- **Démarche *Quality by Design* (QbD) :** Espace de conception (*Design Space*) et analyse des modes de défaillance (AMDEC/FMEA).
- **Bioéquivalence / Équivalence Thérapeutique (pour Génériques) :** Référence de l'étude clinique de bioéquivalence n° `{{clinical.be_study_reference}}` réalisée selon les lignes directrices internationales (intervalle de confiance à $90\%$ du ratio des moyennes géométriques $\text{AUC}_{0-t}$ et $C_{\max}$ compris strictement entre $80.00\%$ et $125.00\%$) ou justification de bio-exemption (*Biowaiver*) basée sur le système BCS.

### 2.3.P.3 Procédé de Fabrication
- **Type de Procédé :** Procédé standard (voie sèche / compression directe ou granulation humide).
- **Paramètres Critiques de Procédé (CPP) :** Temps de mélange, force de compression, température d'entrée d'air lors du pelliculage.
- **Validation du Procédé :** Protocole et résultats de validation sur 3 lots industriels consécutifs (Section 3.2.P.3.5).

### 2.3.P.4 Contrôle des Excipients
Tous les excipients utilisés répondent aux monographies de la Pharmacopée Européenne en vigueur. Aucun excipient nouveau ou d'origine animale présentant un risque d'ESB/TSE n'est incorporé.

### 2.3.P.5 Contrôle du Produit Fini (ICH Q6A)
Spécifications à libération et à péremption établies et justifiées. Méthodes d'analyse HPLC indicatrices de stabilité (*Stability-indicating methods*) dûment validées.

### 2.3.P.8 Stabilité du Produit Fini en Zone IVb
- **Études Accélérées ($40^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$) :** Données sur 6 mois démontrant l'absence de dégradation significative.
- **Études Long Terme Zone IVb ($30^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$) :** Données en temps réel sur 3 lots industriels couvrant la durée de conservation revendiquée (`{{product.shelf_life_months}}` mois).
- **Conclusion Stabilité :** Conservation à une température ne dépassant pas $30^\circ\text{C}$ dans l'emballage d'origine.

---

## 4. DÉCLARATION CRITIQUE DE L'EXPERT QUALITÉ AGRÉÉ

Je soussigné, **`{{expert.full_name}}`**, Docteur en Pharmacie / Expert Qualité agréé :
1. Déclare avoir examiné de manière critique l'ensemble des données relatives à la qualité chimique, pharmaceutique et biologique figurant dans le Module 3 du présent dossier.
2. Certifie que les données présentées démontrent la cohérence, la reproductibilité et la qualité constante du médicament `{{product.brand_name}}`.
3. Atteste que le procédé de fabrication et les contrôles analytiques garantissent la conformité du produit aux normes de qualité requises pour son utilisation clinique chez l'Homme.

---

## 5. VISAS ET SIGNATURES ÉLECTRONIQUES QUALIFIÉES

| Rôle BPF | Signataire | Titre Professionnel | Date & Heure UTC | Empreinte SHA-256 |
|:---|:---|:---|:---|:---|
| **Rédacteur (*Maker*)** | `{{audit.maker_name}}` | Expert Qualité / Galéniste | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Vérificateur (*Checker*)** | `{{audit.checker_name}}` | Manager Affaires Réglem. | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approbateur (*Approver*)** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
