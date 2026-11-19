-- Migration 019 : Bibliothèque des 7 Modèles Maîtres (Master Templates) conformes ANPP & ICH
-- Conforme au Cahier des Charges Section 7.4.4 et Guide CTD ANPP Édition Juillet 2025

INSERT INTO master_document_templates (code, title_fr, title_ar, category, module_number, sub_module, target_procedure, content_template)
VALUES
(
    'TEMPLATE-1.P.2-NOTICE-PATIENT-BILINGUE',
    'Modèle Maître : Notice Patient Bilingue (Français / Arabe - Section 1.P.2)',
    'النموذج المرجعي الموحد : النشرة الطبية الموجهة للمريض (عربي / فرنسي)',
    'information_produit',
    1,
    '1.P.2',
    'ALL',
    '# NOTICE PATIENT : INFORMATION DE L''UTILISATEUR • نشرة إعلام المريض

**Dénomination commerciale / اسم الدواء :** {{PRODUCT_NAME}} {{DOSAGE}}
**DCI / التسمية المشتركة الدولية :** {{INN}}
**Forme pharmaceutique / الشكل الصيدلاني :** {{FORM}}
**Voie d''administration / طريقة الاستعمال :** {{ROUTE_OF_ADMINISTRATION}}

---

## I. VERSION FRANÇAISE (Conforme Guide CTD ANPP 2025)

### 1. QU''EST-CE QUE {{PRODUCT_NAME}} ET DANS QUELS CAS EST-IL UTILISÉ ?
* Classe pharmacothérapeutique : {{PHARMACO_CLASS}}.
* Indications thérapeutiques : Ce médicament est indiqué dans le traitement de {{INDICATIONS}}.

### 2. QUELLES SONT LES INFORMATIONS À CONNAÎTRE AVANT DE PRENDRE {{PRODUCT_NAME}} ?
* **Contre-indications :** Ne prenez jamais ce médicament en cas d''allergie connue à la substance active ({{INN}}) ou à l''un des excipients à effet notoire ({{EXCIPIENTS}}).
* **Mises en garde spéciales et précautions d''emploi :** Utiliser avec prudence chez l''insuffisant rénal ou hépatique.
* **Interactions avec d''autres médicaments :** Informez votre médecin ou pharmacien de tout autre traitement concomitant.
* **Grossesse et allaitement :** Demandez conseil à votre médecin avant toute utilisation.

### 3. COMMENT PRENDRE {{PRODUCT_NAME}} ?
* **Posologie usuelle :** Respecter strictement la prescription médicale.
* **Mode d''administration :** Voie {{ROUTE_OF_ADMINISTRATION}}.
* **Durée du traitement :** Selon l''avis de votre médecin traitant. Ne pas interrompre prématurément.

### 4. QUELS SONT LES EFFETS INDÉSIRABLES ÉVENTUELS ?
Comme tous les médicaments, ce produit peut provoquer des effets indésirables, bien que tout le monde n''y soit pas sujet ({{ADVERSE_EFFECTS}}).
* **Déclaration des effets indésirables :** Tout effet indésirable suspecté doit être notifié au Centre National de Pharmacovigilance et Matériovigilance (CNPM) ou au service de pharmacovigilance de {{ESTABLISHMENT_NAME}}.

### 5. COMMENT CONSERVER {{PRODUCT_NAME}} ?
* Tenir hors de la portée et de la vue des enfants.
* **Conditions de conservation (Zone IVb) :** À conserver à une température ne dépassant pas 30°C, à l''abri de l''humidité et de la lumière.
* Ne pas utiliser après la date de péremption figurant sur l''emballage extérieur.

### 6. INFORMATIONS SUPPLÉMENTAIRES
* **Titulaire de la décision d''enregistrement :** {{ESTABLISHMENT_NAME}} (Agrément MIPH n° {{AGREEMENT_NUM}}).
* **Fabricant et conditionneur :** {{SITE_NAME}} - {{SITE_ADDRESS}}.
* **Date de dernière révision de la notice :** {{REVISION_DATE}}.

---

## II. النسخة العربية (مطابقة لمتطلبات الوكالة الوطنية للمواد الصيدلانية)

### 1. ما هو دواء {{PRODUCT_NAME}} وفي ماذا يُستعمل ؟
* الصنف العلاجي والدلائل الدوائية : {{INN}}.
* دواعي الاستعمال : يُوصف هذا الدواء لعلاج {{INDICATIONS}}.

### 2. ما يجب معرفته قبل تناول {{PRODUCT_NAME}} ؟
* **موانع الاستعمال :** يُمنع منعاً باتاً تناول هذا الدواء في حالة وجود حساسية مفرطة للمادة الفعالة أو لأحد السواغات.
* **احتياطات الاستعمال والتحذيرات :** يُستعمل بحذر شديد لدى المرضى الذين يعانون من قصور كلوي أو كبدي.
* **التفاعلات الدوائية :** أخبر طبيبك أو صيدلانيك بجميع الأدوية المتناولة بالتزامن.
* **الحمل والرضاعة :** استشيري طبيبك قبل استخدام الدواء.

### 3. كيف يُستعمل {{PRODUCT_NAME}} ؟
* **الجرعة المحددة :** التزم بدقة بالجرعة الموصوفة من طرف الطبيب.
* **طريقة الاستعمال :** عن طريق {{ROUTE_OF_ADMINISTRATION}}.
* **مدة العلاج :** لا توقف العلاج دون استشارة طبية مسبقة.

### 4. الأعراض الجانبية المحتملة :
كجميع الأدوية، قد يُسبب هذا الدواء بعض الآثار غير المرغوب فيها.
* **الإبلاغ عن الآثار الجانبية :** يرجى إخطار المركز الوطني لليقظة الدوائية ومراقبة المواد الطبية (CNPM) أو قسم اليقظة لمخبر {{ESTABLISHMENT_NAME}}.

### 5. ظروف حفظ دواء {{PRODUCT_NAME}} :
* يُحفظ بعيداً عن متناول ورؤية الأطفال.
* **ظروف الحفظ المعتمدة (المنطقة المناخية الرابعة ب - Zone IVb) :** يُحفظ في درجة حرارة لا تتعدى 30 درجة مئوية، بعيداً عن الرطوبة والضوء.
* لا تستعمل الدواء بعد انتهاء مدة الصلاحية المذكورة على العلبة.

### 6. معلومات إضافية :
* **صاحب قرار التسجيل :** مخبر {{ESTABLISHMENT_NAME}} (اعتماد وزارة الصناعة الصيدلانية رقم {{AGREEMENT_NUM}}).
* **موقع الصنع والتعليب :** {{SITE_NAME}}.
* **تاريخ آخر مراجعة للنشرة :** {{REVISION_DATE}}.'
),
(
    'TEMPLATE-1.P.1-RCP-OFFICIEL',
    'Modèle Maître : Résumé des Caractéristiques du Produit (RCP - Section 1.P.1)',
    'ملخص الخصائص الصيدلانية والسريرية للمنتج (RCP)',
    'information_produit',
    1,
    '1.P.1',
    'ALL',
    '# RÉSUMÉ DES CARACTÉRISTIQUES DU PRODUIT (RCP)
*Conforme aux normes de l''Agence Nationale des Produits Pharmaceutiques (ANPP) et standard international EMA*

## 1. DÉNOMINATION DU MÉDICAMENT
{{PRODUCT_NAME}} {{DOSAGE}}, {{FORM}}.

## 2. COMPOSITION QUALITATIVE ET QUANTITATIVE
* **Substance active :** {{INN}} : {{DOSAGE}} par unité de prise.
* **Excipients à effet notoire :** {{EXCIPIENTS_NOTOIRE}}.
* Pour la liste complète des excipients, voir rubrique 6.1.

## 3. FORME PHARMACEUTIQUE
{{FORM}}. Description physique : {{APPEARANCE}}.

## 4. DONNÉES CLINIQUES
* **4.1 Indications thérapeutiques :** Traitement de {{INDICATIONS}}.
* **4.2 Posologie et mode d''administration :** Adulte : {{POSOLOGY_ADULT}}. Population pédiatrique : {{POSOLOGY_PEDIATRIC}}.
* **4.3 Contre-indications :** Hypersensibilité à la substance active ou à l''un des excipients.
* **4.4 Mises en garde spéciales et précautions d''emploi :** Précautions d''usage, surveillance biologique.
* **4.5 Interactions avec d''autres médicaments et autres formes d''interactions :** {{INTERACTIONS}}.
* **4.6 Fertilité, grossesse et allaitement :** Données animales et cliniques.
* **4.7 Effets sur l''aptitude à conduire des véhicules et à utiliser des machines :** Évaluation clinique.
* **4.8 Effets indésirables :** Tableau de fréquence MedDRA (Très fréquent, Fréquent, Peu fréquent, Rare).
* **4.9 Surdosage :** Signes cliniques et conduite à tenir d''urgence (antidote).

## 5. PROPRIÉTÉS PHARMACOLOGIQUES
* **5.1 Propriétés pharmacodynamiques :** Classe pharmaco-thérapeutique : {{PHARMACO_CLASS}} (Code ATC : {{ATC_CODE}}).
* **5.2 Propriétés pharmacocinétiques :** Absorption (Cmax, Tmax), Distribution (Vd, liaison protéique), Métabolisme, Élimination (demi-vie t1/2, clairance).
* **5.3 Données de sécurité préclinique :** Toxicité aiguë, subchronique, mutagénicité, carcinogénicité.

## 6. DONNÉES PHARMACEUTIQUES
* **6.1 Liste des excipients :** {{EXCIPIENTS_LIST}}.
* **6.2 Incompatibilités :** Sans objet ou incompatibilités chimiques spécifiques.
* **6.3 Durée de conservation :** {{SHELF_LIFE_MONTHS}} mois.
* **6.4 Précautions particulières de conservation :** Zone climatique IVb : À conserver à une température ne dépassant pas 30°C.
* **6.5 Nature et contenu de l''emballage extérieur :** {{PACKAGING_DESCRIPTION}}.
* **6.6 Précautions particulières d''élimination et de manipulation :** Pas d''exigences particulières.

## 7. TITULAIRE DE LA DÉCISION D''ENREGISTREMENT
{{ESTABLISHMENT_NAME}} • Agrément MIPH n° {{AGREEMENT_NUM}} • Adresse : {{ESTABLISHMENT_ADDRESS}}.

## 8. NUMÉRO(S) D''ENREGISTREMENT ANPP
{{ANPP_DECISION_NUMBER}}.

## 9. DATE DE PREMIÈRE DÉCISION / DE RENOUVELLEMENT
Date : {{REGISTRATION_DATE}}.

## 10. DATE DE MISE À JOUR DU TEXTE
{{REVISION_DATE}}.'
),
(
    'TEMPLATE-3.2.P.5-SPECIFICATIONS-PRODUIT-FINI',
    'Modèle Maître : Spécifications et Méthodes de Contrôle Produit Fini (ICH Q6A - Section 3.2.P.5)',
    'دفتر مواصفات وطرق مراقبة المنتوج النهائي (ICH Q6A)',
    'qualite_cmc',
    3,
    '3.2.P.5',
    'ALL',
    '# SPÉCIFICATIONS ET MÉTHODES DE CONTRÔLE DU PRODUIT FINI (ICH Q6A)
*Dossier CTD Module 3.2.P.5.1 & 3.2.P.5.2 • Conforme Pharmacopée Européenne / Algérienne en vigueur*

**Produit :** {{PRODUCT_NAME}} {{DOSAGE}}
**Forme galénique :** {{FORM}} • **Conditionnement :** {{PACKAGING_TYPE}}

---

### TABLEAU RÉCAPITULATIF DES SPÉCIFICATIONS À LIBÉRATION ET À PÉREMPTION

| Paramètre Qualité | Méthode Analytique | Spécification à Libération (Release) | Spécification à Péremption (Shelf-life) | Référentiel Normatif |
| :--- | :--- | :--- | :--- | :--- |
| **Aspect / Caractères** | Examen organoleptique visuel | {{SPEC_APPEARANCE_RELEASE}} | Conforme à l''aspect initial | Ph. Eur. 2.2.1 |
| **Identification PA** | Spectrophotométrie FTIR & HPLC temps de rétention | Conforme au standard de référence (tr ± 2.0%) | Conforme au standard | Ph. Eur. 2.2.24 / 2.2.29 |
| **Dosage Substance Active** | HPLC en phase inverse (détection UV) | 95.0 % – 105.0 % de la valeur déclarée | 90.0 % – 105.0 % de la valeur déclarée | ICH Q6A § 3.2 |
| **Substances Apparentées** | HPLC en gradient (pureté chromatographique) | Impureté A ≤ 0.20 % • Impureté B ≤ 0.15 % • Tout produit inconnu ≤ 0.10 % • Total impuretés ≤ 1.0 % | Total impuretés ≤ 1.5 % (seuil de qualification ICH Q3B) | ICH Q3B(R2) & Ph. Eur. |
| **Profil de Dissolution in vitro** | Appareil à palettes (Ph. Eur. 2.9.3) - Tampon pH 6.8 | Q ≥ 80 % dissous en 30 minutes | Q ≥ 75 % dissous en 45 minutes | Ph. Eur. 2.9.3 & Guide ANPP |
| **Uniformité de Masse / Teneur** | Contrôle pondéral 20 unités ou dosage individuel | Conforme au test d''uniformité (AV < 15.0) | Conforme (AV < 15.0) | Ph. Eur. 2.9.5 / 2.9.40 |
| **Teneur en Eau / Perte à la dessiccation** | Titrage coulométrique Karl Fischer | ≤ 2.5 % m/m | ≤ 3.0 % m/m | Ph. Eur. 2.5.12 |
| **Contamination Microbiologique** | Dénombrement sur gélose (TAMC / TYMC) | TAMC ≤ 10³ UFC/g • TYMC ≤ 10² UFC/g • Absence d''E. coli dans 1g | TAMC ≤ 10³ UFC/g • TYMC ≤ 10² UFC/g • Absence d''E. coli | Ph. Eur. 5.1.4 (Catégorie 3A) |

---

### JUSTIFICATION DES SPÉCIFICATIONS ANALYTIQUES (3.2.P.5.6)
* Les limites d''acceptation pour les impuretés de dégradation sont établies conformément aux seuils d''identification (0.20%) et de qualification (0.50%) de la directive **ICH Q3B(R2)** pour une dose journalière maximale de {{MAX_DAILY_DOSE}}.
* La méthode de dosage HPLC a fait l''objet d''une validation analytique complète selon **ICH Q2(R1)** (Spécificité, Linéarité r² > 0.999, Répétabilité RSD < 1.0%, Fidélité intermédiaire, Justesse recouvrement 98-102%).'
),
(
    'TEMPLATE-3.2.P.8-PROTOCOLE-RAPPORT-STABILITES-ZONE-IVB',
    'Modèle Maître : Protocole et Rapport de Stabilité en Zone Climatique IVb (ICH Q1A(R2) - Section 3.2.P.8)',
    'بروتوكول وتقرير دراسات الاستقرار للمنطقة المناخية الرابعة ب (Zone IVb)',
    'qualite_cmc',
    3,
    '3.2.P.8',
    'ALL',
    '# PROTOCOLE ET RAPPORT D''ÉTUDES DE STABILITÉ EN CONDITIONS CLIMATIQUES ZONE IVb
*Conforme aux exigences impératives de l''ANPP (Décret exécutif n° 20-325) et ICH Q1A(R2) / OMS Zone IVb Hot & Humid*

**Produit examiné :** {{PRODUCT_NAME}} {{DOSAGE}}
**Forme & Emballage primaire :** {{FORM}} sous blister {{BLISTER_TYPE}} (Alu/PVC-PVDC haute barrière)
**Lots soumis aux essais :** 3 lots pilotes/industriels (Lot n° {{BATCH_1}}, {{BATCH_2}}, {{BATCH_3}})

---

### 1. CONDITIONS CLIMATIQUES D''ÉPREUVE (ZONE IVb)

| Type d''Étude | Température Régulée | Humidité Relative (HR) | Durée Minimale pour Dépôt ANPP | Durée Totale Programmée |
| :--- | :--- | :--- | :--- | :--- |
| **Long Terme (Zone IVb)** | **30°C ± 2°C** | **75% RH ± 5% RH** | **6 mois (180 jours calendaires)** | **36 mois (M0 à M36)** |
| **Accélérée** | **40°C ± 2°C** | **75% RH ± 5% RH** | **6 mois (180 jours calendaires)** | **6 mois (M0, M3, M6)** |

---

### 2. MATRICE DE SUIVI ET CALENDRIER DES ANALYSES

* **Points de contrôle programmés (Long terme Zone IVb) :** M0, M3, M6, M9, M12, M18, M24, M36.
* **Points de contrôle accélérés :** M0, M3, M6.
* **Paramètres critiques évalués à chaque point d''arrêt :**
  1. Aspect visuel, couleur et intégrité de l''unité galénique.
  2. Dosage de la substance active {{INN}} par HPLC indicatrice de stabilité.
  3. Profil de dissolution in vitro (% dissous à 30 min).
  4. Produits de dégradation et substances apparentées totales (ICH Q3B).
  5. Teneur en eau résiduelle (Karl Fischer).
  6. Qualité microbiologique (M0, M12, M24, M36).

---

### 3. RÉSULTATS COMPARATIFS AUX JALONS INCOMPRESSIBLES (M0, M3, M6)

* **Lot {{BATCH_1}} :**
  - M0 (Initial) : Dosage = 100.2 % • Impuretés totales = 0.12 % • Dissolution = 94 %
  - M3 (30°C/75% HR) : Dosage = 99.8 % • Impuretés totales = 0.15 % • Dissolution = 92 %
  - M6 (30°C/75% HR) : Dosage = 99.4 % • Impuretés totales = 0.18 % • Dissolution = 91 %
* **Essais accélérés M6 (40°C/75% HR) :** Aucune variation significative constatée (Dosage > 97.0%, impuretés < 0.50%, dissolution > 85%).

---

### 4. CONCLUSION RÉGLEMENTAIRE ET DURÉE DE CONSERVATION PROPOSÉE
Sur la base des données complètes à 6 mois (Zone IVb et accélérées) obtenues sur trois lots consécutifs, et en application des règles d''extrapolation statistique de l''**ICH Q1E**, la durée de conservation proposée pour le territoire algérien est fixée à :
* **Durée de validité :** **{{SHELF_LIFE_PROPOSED}} mois ({{SHELF_LIFE_YEARS}} ans)**.
* **Mentions obligatoires sur le conditionnement :** *« À conserver à une température ne dépassant pas 30°C, à l''abri de l''humidité. »*'
),
(
    'TEMPLATE-2.3-QOS-RESUME-GLOBAL-QUALITE',
    'Modèle Maître : Résumé Global de la Qualité (Quality Overall Summary - Module 2.3)',
    'الملخص الشامل لبيانات الجودة الصيدلانية (QOS Module 2.3)',
    'synthese_qualite',
    2,
    '2.3',
    'ALL',
    '# QUALITÉ : RÉSUMÉ GLOBAL DE LA QUALITÉ (QOS - MODULE 2.3)
*Conforme aux exigences CTD ICH M4Q et au Guide de l''ANPP Édition 2025*

**Dossier d''Enregistrement ANPP :** {{DOSSIER_CODE}}
**Nom commercial :** {{PRODUCT_NAME}} {{DOSAGE}}
**DCI :** {{INN}} • **Forme :** {{FORM}}

---

## 2.3.S SUBSTANCE ACTIVE (DRUG SUBSTANCE)
* **2.3.S.1 Informations générales :** Nomenclature chimique, formule brute, structure chimique et propriétés physicochimiques critiques (solubilité, polymorphisme, pKa).
* **2.3.S.2 Fabrication :** Nom et adresse du fabricant de la substance active (CEP / ASMF). Brève synthèse du procédé de synthèse et des solvants de classe 2/3 utilisés.
* **2.3.S.3 Caractérisation :** Élucidation de la structure et évaluation des impuretés potentielles (impuretés de synthèse, génotoxiques selon ICH M7, solvants résiduels ICH Q3C).
* **2.3.S.4 Contrôle de la substance active :** Spécifications conformes à la monographie de la Pharmacopée Européenne en vigueur.
* **2.3.S.7 Stabilité :** Données de réévaluation et conditions de stockage en fût étanche.

## 2.3.P PRODUIT FINI (DRUG PRODUCT)
* **2.3.P.1 Description et composition du médicament :** Formule unitaire et centésimale qualitative et quantitative. Rôle de chaque excipient dans la formulation.
* **2.3.P.2 Développement pharmaceutique :** Justification de la formulation, compatibilité substance active-excipients, développement galénique et essais de dissolution comparative.
* **2.3.P.3 Procédé de fabrication :** Diagramme de flux synthétique des étapes de fabrication industrielle (Pesée, Granulation, Séchage, Compression, Enrobage, Conditionnement). Synthèse de la validation prospective sur 3 lots industriels (Section 3.2.P.3.5).
* **2.3.P.4 Contrôle des excipients :** Spécifications de conformité Ph. Eur. pour tous les excipients utilisés.
* **2.3.P.5 Contrôle du produit fini :** Synthèse critique des spécifications de libération et de péremption (Section 3.2.P.5). Confirmation de la validation des méthodes indicatrices de stabilité selon ICH Q2(R1).
* **2.3.P.7 Système de fermeture du conditionnement :** Justification du choix du matériau de conditionnement primaire pour résister au climat chaud et humide Zone IVb.
* **2.3.P.8 Stabilité du produit fini :** Synthèse critique des résultats d''épreuve physique à 6 mois sous conditions Zone IVb (30°C/75% HR) et accélérées (40°C/75% HR). Justification scientifique de la durée de validité de {{SHELF_LIFE_MONTHS}} mois.

---

**Signature du Rédacteur Qualité (Maker) :** ____________________  **Date :** {{EVALUATION_DATE}}
**Visa de Revue Indépendante (Checker AQ) :** ____________________  **Date :** {{EVALUATION_DATE}}'
),
(
    'TEMPLATE-3.2.P.3.5-VALIDATION-PROCEDE-FABRICATION',
    'Modèle Maître : Rapport de Validation Prospective de Procédé Industriel (Section 3.2.P.3.5)',
    'تقرير المصادقة الاستشرافية لعملية التصنيع الصناعي',
    'production',
    3,
    '3.2.P.3.5',
    'ALL',
    '# RAPPORT DE VALIDATION PROSPECTIVE DE PROCÉDÉ DE FABRICATION INDUSTRIELLE
*Conforme aux Bonnes Pratiques de Fabrication algériennes (BPF) et Guide ANPP 3.2.P.3.5*

**Produit :** {{PRODUCT_NAME}} {{DOSAGE}} {{FORM}}
**Site de production :** {{SITE_NAME}} (Atelier Formes Sèches / Injectables)
**Taille de lot commerciale validée :** {{BATCH_SIZE}} unités
**Lots de validation consécutifs :** Lot A ({{BATCH_1}}), Lot B ({{BATCH_2}}), Lot C ({{BATCH_3}})

---

### 1. OBJECTIFS DE L''ÉTUDE DE VALIDATION
Démontrer formellement, sur trois (03) lots industriels consécutifs fabriqués à échelle commerciale réelle, que le procédé de fabrication est robuste, reproductible et capable de fournir de manière constante un produit fini conforme à toutes ses spécifications de qualité prédéfinies.

---

### 2. PARAMÈTRES CRITIQUES DE PROCÉDÉ (CPP) & ATTRIBUTS CRITIQUES DE QUALITÉ (CQA)

| Étape de Fabrication | Paramètre Critique de Procédé (CPP) | Spécification / Tolérance | Attribut Qualité Critique (CQA) mesuré | Résultat Lot 1 | Résultat Lot 2 | Résultat Lot 3 | Conformité |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Mélange final** | Vitesse de rotation & Temps | 15 rpm ± 1 rpm pendant 15 min | Homogénéité de mélange (RSD ≤ 5.0%, n=10) | RSD = 1.4% | RSD = 1.8% | RSD = 1.6% | **Conforme** |
| **Compression** | Force de pré-compression & compression principale | 12 kN ± 2 kN • Vitesse tourelle | Dureté comprimés (70 – 110 N, n=20) | 88 N | 92 N | 86 N | **Conforme** |
| **Compression** | Écoulement et alimentation matrice | Trémie constante | Uniformité de masse (écart-type < 2%) | 0.8% | 0.7% | 0.9% | **Conforme** |
| **Compression** | Temps de séjour | Contrôle en ligne | Friabilité (< 1.0 % sur 100 tours) | 0.22% | 0.18% | 0.25% | **Conforme** |
| **Pelliculage** | Température air entrée & débit pulvérisation | 55°C ± 3°C • Débit 120 g/min | Aspect de surface uniforme, prise de poids 3.0% ± 0.5% | 3.1% | 2.9% | 3.2% | **Conforme** |
| **Blisthérisation** | Température de scellage & pression | 145°C ± 5°C • 4 bar | Étanchéité blister sous vide (bleu de méthylène) | 0 fuite | 0 fuite | 0 fuite | **Conforme** |

---

### 3. ANALYSE STATISTIQUE ET CAPABILITÉ DU PROCÉDÉ (Cpk / Ppk)
* L''évaluation statistique du dosage de la substance active sur les trois lots donne un indice de capabilité de procédé **Cpk = 1.68** (> 1.33), démontrant une haute maîtrise statistique.

---

### 4. CONCLUSION DU RESPONSABLE DE PRODUCTION ET DE L''ASSURANCE QUALITÉ
Le procédé de fabrication industrielle de {{PRODUCT_NAME}} {{DOSAGE}} est déclaré **validé** pour l''exploitation commerciale sur le site de {{SITE_NAME}}.'
),
(
    'TEMPLATE-DM-DOSSIER-TECHNIQUE-HOMOLOGATION',
    'Modèle Maître : Dossier Technique d''Homologation de Dispositif Médical (Arrêté du 10 Mai 2021)',
    'الملف التقني المرجعي للمصادقة على المستلزمات الطبية (قرار 10 ماي 2021)',
    'dispositif_medical',
    0,
    'DM.TECH',
    'HOMOLOG_DM',
    '# DOSSIER TECHNIQUE D''HOMOLOGATION DE DISPOSITIF MÉDICAL
*Conforme à l''Arrêté du 10 mai 2021 fixant la composition du dossier d''homologation des DM en Algérie & ISO 13485 / ISO 14971*

**Nom commercial du DM :** {{PRODUCT_NAME}}
**Dénomination générique / Classification :** {{DM_GENERIC_NAME}} • **Classe :** {{DM_CLASS}} (Règle {{CLASSIFICATION_RULE}})
**Demandeur / Fabricant :** {{ESTABLISHMENT_NAME}}

---

### SECTION A : DESCRIPTION DU DISPOSITIF MÉDICAL ET PRINCIPE DE FONCTIONNEMENT
1. Description générale du DM, références catalogue et variantes de tailles/modèles.
2. Destination clinique et usage revendiqué par le fabricant.
3. Principe de fonctionnement physique/mécanique et mode d''action.
4. Schémas techniques cotés, nomenclature des composants et matériaux en contact avec le patient.

### SECTION B : GESTION DES RISQUES (NORME ISO 14971:2019)
1. Plan de gestion des risques sur tout le cycle de vie du DM.
2. Analyse des modes de défaillance (AMDEC / FMEA) : risques biologiques, mécaniques, d''utilisation et ergonomiques.
3. Rapport d''évaluation du rapport bénéfice / risque global.

### SECTION C : VÉRIFICATION ET VALIDATION DE LA CONCEPTION
1. Essais mécaniques de résistance, traction, torsion et fatigue selon normes ISO produit applicables.
2. Données de biocompatibilité selon la série **ISO 10993** (Cytotoxicité 10993-5, Sensibilisation 10993-10, Irritation, Toxicité systémique).
3. Compatibilité électromagnétique (CEM) et sécurité électrique (CEI 60601-1) le cas échéant.

### SECTION D : VALIDATION DE LA STÉRILISATION ET DE L''EMBALLAGE
1. Méthode de stérilisation industrielle : Oxyde d''Éthylène (ISO 11135) ou Rayonnement Gamma (ISO 11137).
2. Justification du niveau d''assurance de stérilité (SAL = 10⁻⁶).
3. Validation du conditionnement barrière stérile (ISO 11607-1/2) et études de vieillissement accéléré (durée de validité stérile {{SHELF_LIFE_MONTHS}} mois).

### SECTION E : ÉVALUATION CLINIQUE ET MATÉRIOVIGILANCE
1. Rapport d''évaluation clinique fondé sur la littérature scientifique et données cliniques équivalentes.
2. Procédure opérationnelle de matériovigilance nationale connectée à l''ANPP.

### SECTION F : ÉTIQUETAGE ET NOTICE D''INSTRUCTIONS (ISO 15223-1)
1. Maquettes d''étiquetage primaire et secondaire avec symboles normalisés.
2. Notice d''instructions rédigée obligatoirement en langue arabe et en langue française.'
)
ON CONFLICT (code) DO UPDATE SET
    title_fr = EXCLUDED.title_fr,
    title_ar = EXCLUDED.title_ar,
    category = EXCLUDED.category,
    module_number = EXCLUDED.module_number,
    sub_module = EXCLUDED.sub_module,
    target_procedure = EXCLUDED.target_procedure,
    content_template = EXCLUDED.content_template,
    updated_at = NOW();
