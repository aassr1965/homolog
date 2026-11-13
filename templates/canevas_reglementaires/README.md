# Bibliothèque des Canevas Réglementaires et Techniques Maîtres

> **Plateforme**: Homolog - Système d'Information Réglementaire et d'Enregistrement Pharmaceutique  
> **Référentiels**: Guide CTD ANPP (Juillet 2025) | Ligne Directrice Variations LD-DEPP07/V01 (Avril 2026) | Décrets 20-324 & 20-325 | Arrêté du 10 mai 2021 (DM) | Arrêté du 07 nov 2021 (Regroupement) | Zone Climatique IVb ($30^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$) | 21 CFR Part 11 & ALCOA+

---

## 1. Vue d'Ensemble du Référentiel Documentaire

Cette bibliothèque constitue le référentiel des gabarits réglementaires maîtres (*Regulatory Master Templates*) intégrés au studio d'édition collaborative de la plateforme **Homolog**. Chaque canevas est nativement outillé pour :
1. **L'interpolation dynamique de données** : Liaison bidirectionnelle avec le schéma relationnel PostgreSQL (données produit, établissements, sites de fabrication, études de stabilité, variations).
2. **Le respect strict de la Ségrégation des Tâches (SoD)** : Workflow séquentiel rigide à trois signatures électroniques horodatées (*Maker* $\rightarrow$ *Checker* $\rightarrow$ *Approver*) garantissant l'intégrité BPF / ALCOA+.
3. **Le verrouillage des jalons critiques PERT** : Conditionnement de la recevabilité administrative et technique (dépôt BV1/BV2, falaise de stabilité 6 mois Zone IVb, validation sur 3 lots industriels).

---

## 2. Matrice Globale des 16 Canevas Réglementaires

| N° | Code Document | Intitulé Officiel | Domaine / Dossier Cible | Rôle Rédacteur (*Maker*) | Rôle Vérificateur (*Checker*) | Rôle Approbateur (*Approver*) | Réf. Légale & Réglementaire | Dépendance PERT |
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
| **01** | `1.A.3` | Demande d'Enregistrement Médicament | NF, NI, RF, RI, VF, VI | Chargé Affaires Réglementaires | Pharmacien Affaires Réglementaires | Pharmacien Directeur Technique | Décret 20-325, Art. 5; Guide CTD 1.A.3 | Jalon Dépôt Initial (BV1) |
| **02** | `1.A.4` | Bordereau de Versement Redevances BV1/BV2 | Tous dossiers | Assistant Réglementaire / Compta | Pharmacien Affaires Réglementaires | Directeur Administratif & Financier | Décret 20-325, Art. 12; Barèmes ANPP | Étape Recevabilité (BV1) & Commission (BV2) |
| **03** | `1.P.8` | Engagement Validation 3 Lots & Non-Commercialisation | NF, MF (Fabrication Locale) | Responsable Assurance Qualité | Responsable Production / AQ | Pharmacien Directeur Technique | Guide CTD 1.P.8; BPF Algériennes | Jalon Pré-requis Libération Commerciale |
| **04** | `1.P.9` | Engagement Suivi Stabilité Continue Zone IVb | Tous médicaments | Responsable Contrôle Qualité | Responsable Assurance Qualité | Pharmacien Directeur Technique | Guide CTD 1.P.9; Zone IVb ANPP | Validation Péremption & Suivi Annuel |
| **05** | `1.H.8` | Liste Exhaustive des Sites de Fabrication | Tous médicaments | Chargé Affaires Réglementaires | Responsable Assurance Qualité | Pharmacien Directeur Technique | Guide CTD 1.H.8; Décret 21-82 | Qualification des Fournisseurs |
| **06** | `1.H.10` | Attestation Non-Modification Renouvellement | RF, RI (Renouvellement) | Pharmacien Affaires Réglementaires | Responsable Assurance Qualité | Pharmacien Directeur Technique | Décret 20-325, Art. 18; Guide CTD 1.H.10 | Dépôt Renouvellement Quinquennal |
| **07** | `1.P.1` | Résumé des Caractéristiques du Produit (RCP) | Tous médicaments | Pharmacien Affaires Réglementaires | Médecin / Pharmacien Conseil | Pharmacien Directeur Technique | Guide CTD Module 1.P.1 (10 rubriques) | Approbation Commission Enregistrement |
| **08** | `1.P.2` | Notice Patient Bilingue (Français / Arabe) | Tous médicaments | Chargé Affaires Réglementaires | Pharmacien Information Médicale | Pharmacien Directeur Technique | Guide CTD Module 1.P.2 (6 rubriques) | Validation Packaging & Étiquetage |
| **09** | `2.3` | Quality Overall Summary (QOS - Résumé Qualité) | NF, NI, RF, RI | Expert Qualité / Galéniste | Responsable AQ / Affaires Réglem. | Expert Qualité Agréé & PDT | ICH M4Q; Guide CTD Module 2.3 | Évaluation Technique ANPP |
| **10** | `3.2.P.5` | Spécifications Produit Fini (Libération/Péremption) | Tous médicaments | Responsable Contrôle Qualité | Responsable Assurance Qualité | Pharmacien Directeur Technique | ICH Q6A; Pharmacopée Européenne | Contrôle Libératoire des Lots |
| **11** | `3.2.P.8` | Protocole & Rapport Stabilité Zone IVb | Tous médicaments | Responsable Stabilité / CQ | Responsable Assurance Qualité | Pharmacien Directeur Technique | ICH Q1A(R2), Q1F; ANPP Zone IVb | Falaise Critique 6 mois ($ES=EF, MT=0$) |
| **12** | `3.2.P.3.5`| Validation Procédé Fabrication (3 Lots Industriels) | NF, MF (Fabrication Locale) | Responsable Validation / Prod | Responsable Assurance Qualité | Pharmacien Directeur Technique | ICH Q8, Q9, Q10; BPF Algériennes | Clôture Dossier Technique Post-AMM |
| **13** | `LD-DEPP07-TAB` | Tableau Comparatif des Modifications (Approuvé vs Proposé) | VMin, VMaj, VAdmin | Pharmacien Affaires Réglementaires | Responsable Assurance Qualité | Pharmacien Directeur Technique | Ligne Directrice LD-DEPP07/V01 (2026) | Instruction des Variations (34 rubriques) |
| **14** | `LD-DEPP07-ATT` | Attestation de Non-Modification des Autres Sections | VMin, VMaj, VAdmin | Pharmacien Affaires Réglementaires | Responsable Assurance Qualité | Pharmacien Directeur Technique | LD-DEPP07/V01, Art. 5 & 7 | Dépôt Dossier Variation |
| **15** | `DM-HOMOLOG-01` | Demande d'Homologation Dispositif Médical | Dispositifs Médicaux (I à III) | Chargé Réglementaire DM | Pharmacien Responsable DM | Pharmacien Directeur Technique | Décret 20-324; Arrêté du 10 mai 2021 | Dépôt Homologation DM |
| **16** | `DM-TECH-02` | Fiche Technique Réglementaire Dispositif Médical | Dispositifs Médicaux (I à III) | Ingénieur Biomédical / Qualité | Pharmacien Responsable DM | Pharmacien Directeur Technique | Arrêté du 10 mai 2021 (Annexe technique) | Expertise Technique DM |

---

## 3. Dictionnaire des Balises d'Interpolation Dynamique (PostgreSQL)

Les canevas utilisent le moteur de balisage Mustache (`{{...}}`) connecté aux vues sécurisées de la base PostgreSQL :

| Préfixe Variable | Table Source PostgreSQL | Description & Attributs Types | Exemple de Résolution |
|:---|:---|:---|:---|
| `{{dossier.*}}` | `dossiers` | Métadonnées dossier (`reference_anpp`, `type_procedure`, `current_state`, `submission_date`) | `ALG-2026-NF-0042` |
| `{{product.*}}` | `medicines` | Identité produit (`brand_name`, `dci`, `dosage`, `pharmaceutical_form`, `atc_code`, `shelf_life_months`) | `AMOXIPEN 500 mg, Gélule` |
| `{{establishment.*}}` | `establishments` | Exploitant (`name`, `legal_status`, `license_number`, `tax_id`, `address`, `phone`, `email`) | `LABORATOIRES PHARMA SUD SPA` |
| `{{pdt.*}}` | `users` (Role PDT) | Pharmacien Directeur Technique (`full_name`, `ordinal_number`, `qualification`, `email`) | `Dr. Yacine BENKHALED (Ordre n° 16/4582)` |
| `{{sites.*}}` | `manufacturing_sites` | Sites intervenants (`manufacturer_name`, `country`, `activity_type`, `gmp_cert_number`) | `Site Blida Ouest (Zone Industrielle)` |
| `{{stability.*}}` | `stability_studies` | Conditions d'essai (`storage_condition`, `temperature`, `humidity`, `sampling_interval`) | `Zone IVb : 30°C ± 2°C / 75% ± 5% RH` |
| `{{variation.*}}` | `variations` | Détails variation (`rubric_number`, `classification`, `approved_text`, `proposed_text`) | `Rubrique 14.a (VMaj) - Changement de site` |
| `{{dm.*}}` | `medical_devices` | Métadonnées DM (`commercial_name`, `risk_class`, `intended_use`, `gmdn_code`, `ce_certificate`) | `Seringue Stérile 3 Pièces 5 ml (Classe IIa)` |

---

## 4. Règles de Sécurité, Traçabilité et Audit Trail (21 CFR Part 11)

Conformément à la politique de sécurité du projet :
1. **Verrouillage de Modification** : Un canevas en état `SUBMITTED` ou `APPROVED` est rendu strictement immuable via déclencheurs d'audit PostgreSQL (`BEFORE UPDATE OR DELETE RAISE EXCEPTION`).
2. **Signature Électronique Qualifiée** : Chaque visa (*Maker*, *Checker*, *Approver*) enregistre l'empreinte cryptographique SHA-256 du document, l'identifiant unique de l'utilisateur, son rôle ordinale, l'horodatage UTC et l'adresse IP.
3. **Pénalité Anti-Auto-Approbation** : Le système empêche cryptographiquement et logiquement qu'un même utilisateur cumule les statuts de *Maker* et d'*Approver* sur un même document.
