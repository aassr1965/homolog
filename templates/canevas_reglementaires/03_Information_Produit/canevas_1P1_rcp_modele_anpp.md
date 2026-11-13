---
document_metadata:
  code_ctd_anpp: "1.P.1"
  title: "Résumé des Caractéristiques du Produit (RCP) - Modèle Harmonisé ANPP"
  legal_reference: "Guide CTD ANPP (Juillet 2025) Module 1.P.1 ; Décret exécutif n° 20-325, Article 5 ; Recommandations de l'OMS"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI"]
  security_roles:
    maker: "Pharmacien Affaires Réglementaires / Médical"
    checker: "Médecin / Pharmacien Conseil d'Information Médicale"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.4 - Validation des Textes Légaux par la Commission d'Enregistrement"
  alcoa_compliance: "21 CFR Part 11 Compliant - Texte Légal Contractuel"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## RÉSUMÉ DES CARACTÉRISTIQUES DU PRODUIT (RCP)
### MODÈLE OFFICIEL EN DIX (10) RUBRIQUES RÉGLEMENTAIRES

---

### 1. DÉNOMINATION DU MÉDICAMENT
**`{{product.brand_name}} {{product.dosage}}, {{product.pharmaceutical_form}}`**

---

### 2. COMPOSITION QUALITATIVE ET QUANTITATIVE

#### 2.1 Substance Active
Chaque unité de prise (gélule, comprimé, cuillère-mesure de $5\,\text{ml}$) contient :
- **`{{product.dci}}` :** `{{active_ingredient.quantity}}` (sous forme de `{{active_ingredient.salt_form}}`).

#### 2.2 Excipients à Effet Notoire
Chaque unité contient :
- `{{excipient_notoire.name}}` : `{{excipient_notoire.quantity}}`.

*Pour la liste complète des excipients, voir rubrique 6.1.*

---

### 3. FORME PHARMACEUTIQUE
`{{product.pharmaceutical_form}}`  
*Description visuelle :* `{{product.visual_appearance_description}}` (Ex : Comprimé pelliculé blanc, oblong, sécable sur une face, mesurant $14\,\text{mm} \times 6\,\text{mm}$).

---

### 4. DONNÉES CLINIQUES

#### 4.1 Indications Thérapeutiques
`{{product.brand_name}}` est indiqué dans le traitement de :
- `{{clinical.indication_1}}`
- `{{clinical.indication_2}}`

#### 4.2 Posologie et Mode d'Administration
- **Adultes et Adolescents (> 12 ans) :** `{{clinical.posology_adult}}`
- **Population Pédiatrique :** `{{clinical.posology_pediatric}}`
- **Insuffisance Rénale ou Hépatique :** `{{clinical.posology_adjustment}}`
- **Mode d'administration :** Voie orale. À prendre de préférence `{{clinical.administration_instructions}}`.

#### 4.3 Contre-Indications
- Hypersensibilité avérée à la substance active `{{product.dci}}` ou à l'un des excipients listés en 6.1.
- `{{clinical.contraindications_list}}`

#### 4.4 Mises en Garde Spéciales et Précautions d'Emploi
- `{{clinical.warnings_and_precautions}}`
- *Présence d'excipients à effet notoire :* Les patients présentant une intolérance héréditaire rare à `{{excipient_notoire.name}}` ne doivent pas prendre ce médicament.

#### 4.5 Interactions avec d'Autres Médicaments et Autres Formes d'Interactions
- `{{clinical.interactions_list}}`

#### 4.6 Fertilité, Grossesse et Allaitement
- **Grossesse :** `{{clinical.pregnancy_data}}`
- **Allaitement :** `{{clinical.lactation_data}}`

#### 4.7 Effets sur l'Aptitude à Conduire des Véhicules et à Utiliser des Machines
`{{clinical.driving_effects}}`

#### 4.8 Effets Indésirables
Les effets indésirables sont classés selon la terminologie MedDRA par système organe et par fréquence :

| Classe de Système d'Organes (SOC) | Très Fréquent ($\ge 1/10$) | Fréquent ($\ge 1/100, < 1/10$) | Rare ($\ge 1/10\,000, < 1/1\,000$) |
|:---|:---|:---|:---|
| **Affections Gastro-intestinales** | Nausées, diarrhée | Douleurs abdominales | Vomissements sévères |
| **Affections du Système Nerveux** | Céphalées | Vertiges légers | Convulsions rares |
| **Affections Cutanées** | Rash cutané bénin | Prurit | Syndrome de Stevens-Johnson |

*Déclaration des effets indésirables suspectés :* La déclaration après enregistrement est essentielle. Elle permet une surveillance continue du rapport bénéfice/risque. Les professionnels de santé déclarent tout effet indésirable au Centre National de Pharmacovigilance et Matériovigilance (CNPM).

#### 4.9 Surdosage
- **Symptômes :** `{{clinical.overdose_symptoms}}`
- **Prise en charge :** `{{clinical.overdose_treatment}}`

---

### 5. PROPRIÉTÉS PHARMACOLOGIQUES

#### 5.1 Propriétés Pharmacodynamiques
- **Classe Pharmaco-thérapeutique :** `{{product.atc_description}}`, **Code ATC :** `{{product.atc_code}}`.
- **Mécanisme d'action :** `{{pharmacology.mechanism_of_action}}`

#### 5.2 Propriétés Pharmacocinétiques
- **Absorption :** `{{pharmacology.absorption}}`
- **Distribution :** `{{pharmacology.distribution}}` (Liaison aux protéines plasmatiques : `{{pharmacology.protein_binding}}` %).
- **Biotransformation :** `{{pharmacology.metabolism}}`
- **Élimination :** Demi-vie d'élimination : `{{pharmacology.half_life}}` heures ; voie principale d'excrétion : `{{pharmacology.elimination_route}}`.

#### 5.3 Données de Sécurité Préclinique
`{{pharmacology.preclinical_safety}}`

---

### 6. DONNÉES PHARMACEUTIQUES

#### 6.1 Liste des Excipients
`{{product.excipients_full_list}}`

#### 6.2 Incompatibilités Majeures
Sans objet (ou spécifier si liquide injectable).

#### 6.3 Durée de Conservation
`{{product.shelf_life_months}}` mois.

#### 6.4 Précautions Particulières de Conservation (Zone Climatique IVb)
Conserver à une température ne dépassant pas $30^\circ\text{C}$, dans son emballage d'origine, à l'abri de l'humidité et de la lumière.

#### 6.5 Nature et Contenu de l'Emballage Extérieur
`{{product.secondary_packaging_description}}`. Présentations : boîte de `{{product.pack_sizes}}` unités.

#### 6.6 Précautions Particulières d'Élimination et de Manipulation
Tout médicament non utilisé ou déchet doit être éliminé conformément à la réglementation locale relative aux déchets pharmaceutiques.

---

### 7. TITULAIRE DE LA DÉCISION D'ENREGISTREMENT
**`{{establishment.name}}`**  
`{{establishment.address}}` - Algérie  
Téléphone : `{{establishment.phone}}` | E-mail : `{{establishment.email}}`

---

### 8. NUMÉRO(S) DE L'ENREGISTREMENT (DÉCISION ANPP)
**N° d'Enregistrement ANPP :** `{{dossier.anpp_registration_number}}`

---

### 9. DATE DE PREMIÈRE AUTORISATION / DE RENOUVELLEMENT
- **Date de première autorisation :** `{{dossier.first_authorization_date}}`
- **Date de dernier renouvellement :** `{{dossier.renewal_date}}`

---

### 10. DATE DE MISE À JOUR DU TEXTE
**`{{dossier.rcp_last_update_date}}`**

---

## SIGNATURES DU TEXTE OFFICIEL (RCP)

| Étape | Signataire | Titre | Date & Heure | Signature SHA-256 |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Pharmacien Affaires Réglem. | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Médecin / Pharmacien Conseil | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
