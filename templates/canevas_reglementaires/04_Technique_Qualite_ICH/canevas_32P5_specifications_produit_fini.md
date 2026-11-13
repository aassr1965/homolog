---
document_metadata:
  code_ctd_anpp: "3.2.P.5"
  title: "Spécifications du Produit Fini à Libération et à Péremption (ICH Q6A)"
  legal_reference: "Directive ICH Q6A ; Guide CTD ANPP (Juillet 2025) Module 3.2.P.5 ; Pharmacopée Européenne en vigueur"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI"]
  security_roles:
    maker: "Responsable Contrôle Qualité / Laboratoire Analytique"
    checker: "Manager Assurance Qualité"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.2 - Contrôles Libératoires et Conformité Réglementaire"
  alcoa_compliance: "21 CFR Part 11 Compliant - Normes Analytiques Immuables"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## DIRECTION DE L'ENREGISTREMENT - MODULE 3 : QUALITÉ CHIMIQUE & PHARMACEUTIQUE
### SECTION 3.2.P.5 : SPÉCIFICATIONS OFFICIELLES DU PRODUIT FINI

---

## 1. IDENTIFICATION DU PRODUIT ET DE LA PRÉSENTATION

- **Dénomination Commerciale :** `{{product.brand_name}} {{product.dosage}}`
- **Substance Active (DCI) :** `{{product.dci}}`
- **Forme Galénique :** `{{product.pharmaceutical_form}}`
- **Conditionnement Primaire :** `{{product.primary_packaging_description}}`
- **Numéro Interne Spécification Qualité :** `SPEC-PF-{{product.id}}-V{{product.spec_version}}`
- **Date d'Entrée en Vigueur :** `{{product.spec_effective_date}}`

---

## 2. TABLEAU MATRICIEL DES SPÉCIFICATIONS (LIBÉRATION VS PÉREMPTION)

| Paramètre de Contrôle | Méthode d'Essai (Ph. Eur. / Interne) | Limites à Libération Industrielle (*Release*) | Limites à Péremption (*Shelf-Life*) | Réf. Validation Analytique |
|:---|:---|:---|:---|:---|
| **Caractères Organoleptiques** | Examen visuel (Ph. Eur. 2.2.1) | Conforme au standard descriptif (aspect, couleur, absence de défauts) | Identique à la libération | Contrôle visuel qualifié |
| **Identification de la SA** | HPLC-UV (temps de rétention & spectre UV) + FTIR | Temps de rétention conforme à l'étalon ($\pm 2\%$) et spectre IR concordant | Identique à la libération | VAL-METH-`{{product.id}}`-ID |
| **Uniformité des Unités de Prise** | Ph. Eur. 2.9.40 (Uniformité de teneur ou variation de masse) | Valeur d'acceptation $AV \le 15.0$ (Niveau L1, $n=10$) | Non applicable en routine stabilité | Pharmacopée Européenne |
| **Dosage de la Substance Active** | HPLC-UV (Méthode indicatrice de stabilité) | **$95.0\% - 105.0\%$** de la valeur théorique étiquetée | **$90.0\% - 105.0\%$** de la valeur théorique | VAL-METH-`{{product.id}}`-ASSAY |
| **Substances Apparentées / Impuretés** | HPLC-UV à gradient (ICH Q3B) | - Impureté A : $\le 0.15\%$<br>- Impureté B : $\le 0.15\%$<br>- Toute impureté inconnue : $\le 0.10\%$<br>- Total impuretés : $\le 0.50\%$ | - Impureté A : $\le 0.30\%$<br>- Impureté B : $\le 0.30\%$<br>- Toute impureté inconnue : $\le 0.20\%$<br>- Total impuretés : $\le 1.50\%$ | VAL-METH-`{{product.id}}`-IMP |
| **Cinétique de Dissolution *In Vitro***| Ph. Eur. 2.9.3 (Appareil à palettes, milieu tampon pH 6.8, 50 rpm) | $Q \ge 80\%$ dissous en 30 minutes ($n=6$) | $Q \ge 75\%$ dissous en 45 minutes | VAL-METH-`{{product.id}}`-DISS |
| **Temps de Désagrégation** | Ph. Eur. 2.9.1 | $\le 15$ minutes dans l'eau à $37^\circ\text{C}$ | $\le 15$ minutes | Pharmacopée Européenne |
| **Friabilité** | Ph. Eur. 2.9.7 (Comprimés non pelliculés) | Perte de masse $\le 1.0\%$ sans bris | Non applicable | Pharmacopée Européenne |
| **Résistance à la Rupture (Dureté)**| Ph. Eur. 2.9.8 | $70 - 120\,\text{N}$ (Moyenne sur 10 comprimés) | $\ge 60\,\text{N}$ | Contrôle en cours IPC |
| **Teneur en Eau / Humidité** | Titrage coulométrique Karl Fischer (Ph. Eur. 2.5.12) | $\le 3.0\%\ \text{m/m}$ | $\le 5.0\%\ \text{m/m}$ | Ph. Eur. 2.5.12 |
| **Solvants Résiduels** | Chromatographie phase gazeuse HS-GC (Ph. Eur. 2.4.24) | Conformes aux limites de l'ICH Q3C | Non testé en routine stabilité | VAL-METH-`{{product.id}}`-GC |
| **Propreté Microbiologique** | Ph. Eur. 2.6.12 & 2.6.13 | - TAMC $\le 10^3\,\text{UFC/g}$<br>- TYMC $\le 10^2\,\text{UFC/g}$<br>- Absence d'*Escherichia coli* dans $1\,\text{g}$ | Identique à la libération | Ph. Eur. 2.6.12/13 |

---

## 3. JUSTIFICATION DES DIFFÉRENCES LIBÉRATION VS PÉREMPTION

Les limites élargies autorisées pour la péremption (teneur en PA abaissée à $90.0\%$, total des impuretés étendu à $1.50\%$) ont été rigoureusement justifiées sur la base :
1. Des résultats réels des études de stabilité accélérée et long terme sous conditions climatiques **Zone IVb ($30^\circ\text{C} \pm 2^\circ\text{C} / 75\% \pm 5\%\ \text{RH}$)**.
2. De la qualification toxicologique des produits de dégradation conformément à la ligne directrice ICH Q3B(R2).
3. De l'absence d'impact sur la sécurité et l'efficacité clinique du traitement.

---

## 4. VALIDATION ET SIGNATURES ORDINALES

| Échelon Qualité | Nom & Prénom | Titre Professionnel | Date & Heure UTC | Empreinte Numérique SHA-256 |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.qc_analyst_name}}` | Responsable Laboratoire CQ | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.qa_validator_name}}`| Responsable Assurance Qualité| `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
