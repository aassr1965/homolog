---
document_metadata:
  code_ctd_anpp: "1.P.2"
  title: "Notice Destinée au Patient Bilingue (Français / Arabe) - Modèle Harmonisé ANPP"
  legal_reference: "Guide CTD ANPP (Juillet 2025) Module 1.P.2 ; Décret exécutif n° 20-325, Article 5 ; Arrêté interministériel relatif à l'étiquetage"
  target_procedures: ["NF", "NI", "RF", "RI", "VF", "VI"]
  security_roles:
    maker: "Chargé Affaires Réglementaires / Traduction Médicale"
    checker: "Pharmacien Information Médicale"
    approver: "Pharmacien Directeur Technique (PDT)"
  gating_dependency: "Jalon 2.4 - Approbation Packaging & Bon à Tirer (BAT)"
  alcoa_compliance: "21 CFR Part 11 Compliant - Document Patient Bilingue Certifié"
---

# AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
## NOTICE PATIENT BILINGUE : FRANÇAIS / ARABE
### EN SIX (06) RUBRIQUES LÉGALES OBLIGATOIRES

---

## NOTICE : INFORMATION DE L'UTILISATEUR
## النشرة: معلومات موجهة للمريض

**`{{product.brand_name}} {{product.dosage}}, {{product.pharmaceutical_form}}`**  
**`{{product.dci}}`**

---

### 1. QU'EST-CE QUE `{{product.brand_name}}` ET DANS QUELS CAS EST-IL UTILISÉ ?
### 1. ما هو `{{product.brand_name}}` وفي أي حالات يُستعمل ؟

**[Texte Français] :**  
`{{product.brand_name}}` contient la substance active `{{product.dci}}`. Il appartient à la classe thérapeutique des `{{product.atc_description}}`. Ce médicament est indiqué dans le traitement de :
- `{{patient.indication_1_fr}}`
- `{{patient.indication_2_fr}}`

**[النص العربي] :**  
يحتوي `{{product.brand_name}}` على المادة الفعالة `{{product.dci}}`. وهو ينتمي إلى فئة `{{product.atc_description}}`. يُوصف هذا الدواء لعلاج:
- `{{patient.indication_1_ar}}`
- `{{patient.indication_2_ar}}`

---

### 2. QUELLES SONT LES INFORMATIONS À CONNAÎTRE AVANT DE PRENDRE `{{product.brand_name}}` ?
### 2. ما هي المعلومات الواجب معرفتها قبل تناول `{{product.brand_name}}` ؟

**[Texte Français] :**  
**Ne prenez jamais ce médicament :**  
- Si vous êtes allergique à la substance active `{{product.dci}}` ou à l'un des autres composants contenus dans ce médicament (listés à la rubrique 6).  
- `{{patient.contraindications_fr}}`

**Avertissements et précautions :**  
Adressez-vous à votre médecin ou pharmacien avant de prendre ce médicament si vous souffrez de : `{{patient.precautions_fr}}`.  
*Enfants et adolescents :* `{{patient.pediatric_warning_fr}}`.  
*Grossesse et allaitement :* Si vous êtes enceinte ou que vous allaitez, demandez conseil à votre médecin avant de prendre ce médicament.

**[النص العربي] :**  
**لا تتناول هذا الدواء أبداً في الحالات التالية:**  
- إذا كنت تعاني من حساسية تجاه المادة الفعالة `{{product.dci}}` أو أي من المكونات الأخرى لهذا الدواء (المذكورة في القسم 6).  
- `{{patient.contraindications_ar}}`

**التحذيرات والاحتياطات:**  
تحدث إلى طبيبك أو الصيدلي قبل تناول هذا الدواء إذا كنت تعاني من: `{{patient.precautions_ar}}`.  
*الأطفال والمراهقون:* `{{patient.pediatric_warning_ar}}`.  
*الحمل والرضاعة:* إذا كنت حاملاً أو مرضعة، استشيري طبيبك قبل استخدام هذا الدواء.

---

### 3. COMMENT PRENDRE `{{product.brand_name}}` ?
### 3. كيف يُستعمل `{{product.brand_name}}` ؟

**[Texte Français] :**  
Veillez à toujours prendre ce médicament en suivant exactement les indications de votre médecin ou pharmacien.  
- **Posologie recommandée chez l'adulte :** `{{patient.dosage_adult_fr}}`.  
- **Mode d'administration :** Voie orale. Avalez le comprimé avec un verre d'eau `{{patient.administration_moment_fr}}`.  
- **Si vous avez oublié de prendre une dose :** Ne prenez pas de dose double pour compenser la dose oubliée. Prenez la dose suivante à l'heure habituelle.

**[النص العربي] :**  
احرص دائماً على تناول هذا الدواء تماماً كما أخبرك به طبيبك أو الصيدلي.  
- **الجرعة الموصى بها للبالغين:** `{{patient.dosage_adult_ar}}`.  
- **طريقة وتوقيت الاستعمال:** يُؤخذ عن طريق الفم مع كأس ماء `{{patient.administration_moment_ar}}`.  
- **إذا نسيت تناول الجرعة:** لا تتناول جرعة مضاعفة لتعويض الجرعة المنسية، وتناول الجرعة التالية في وقتها المعتاد.

---

### 4. QUELS SONT LES EFFETS INDÉSIRABLES ÉVENTUELS ?
### 4. ما هي الآثار الجانبية المحتملة ؟

**[Texte Français] :**  
Comme tous les médicaments, ce médicament peut provoquer des effets indésirables, mais ils ne surviennent pas systématiquement chez tout le monde.  
- **Effets indésirables graves (consultez immédiatement un médecin) :** Réaction allergique brutale (gonflement du visage, gêne respiratoire), éruption cutanée sévère.  
- **Effets indésirables fréquents :** `{{patient.side_effects_frequent_fr}}`.  
*Déclaration des effets indésirables :* Si vous ressentez un quelconque effet indésirable, parlez-en à votre médecin ou pharmacien. Vous pouvez également déclarer les effets indésirables au Centre National de Pharmacovigilance et Matériovigilance (CNPM).

**[النص العربي] :**  
مثل جميع الأدوية، يمكن أن يسبب هذا الدواء آثاراً جانبية، على الرغم من أنها لا تحدث لدى الجميع.  
- **أعراض جانبية خطيرة (استشر الطبيب فوراً):** رد فعل تحسسي مفاجئ (انتفاخ الوجه، صعوبة التنفس)، طفح جلدي حاد.  
- **أعراض جانبية شائعة:** `{{patient.side_effects_frequent_ar}}`.  
*التبليغ عن الأعراض الجانبية:* إذا لاحظت أي أثر جانبي، تحدث إلى طبيبك أو الصيدلي. يمكنك أيضاً التبليغ مباشرة إلى المركز الوطني لليقظة الصيدلانية ومراقبة المواد الطبية (CNPM).

---

### 5. COMMENT CONSERVER `{{product.brand_name}}` ?
### 5. كيف يُحفظ `{{product.brand_name}}` ؟

**[Texte Français] :**  
- Tenir ce médicament hors de la vue et de la portée des enfants.  
- N'utilisez pas ce médicament après la date de péremption indiquée sur l'emballage. La date de péremption fait référence au dernier jour de ce mois.  
- **Conditions de conservation (Zone IVb) :** À conserver à une température ne dépassant pas $30^\circ\text{C}$, dans l'emballage d'origine, à l'abri de l'humidité et de la lumière.

**[النص العربي] :**  
- يُحفظ هذا الدواء بعيداً عن أنظار ومتناول الأطفال.  
- لا تستعمل هذا الدواء بعد تاريخ انتهاء الصلاحية المذكور على العلبة. تاريخ انتهاء الصلاحية يشير إلى آخر يوم من ذلك الشهر.  
- **شروط الحفظ (المنطقة المناخية IVb):** يُحفظ في درجة حرارة لا تتجاوز 30 درجة مئوية، في العبوة الأصلية، بعيداً عن الرطوبة والضوء.

---

### 6. CONTENU DE L'EMBALLAGE ET AUTRES INFORMATIONS
### 6. محتوى العبوة ومعلومات أخرى

**[Texte Français] :**  
- **Ce que contient ce médicament :** La substance active est `{{product.dci}}` (`{{product.dosage}}`). Les autres composants sont : `{{product.excipients_short_list}}`.  
- **Aspect du médicament et contenu de l'emballage :** Boîte contenant `{{product.pack_sizes}}` unités sous plaquettes thermoformées avec notice.  
- **Titulaire et Exploitant :** `{{establishment.name}}`, `{{establishment.address}}` - Algérie.  
- **Fabricant :** `{{site_bulk.name}}`, `{{site_bulk.address}}`.

**[النص العربي] :**  
- **محتويات هذا الدواء:** المادة الفعالة هي `{{product.dci}}` بتركيز (`{{product.dosage}}`). المكونات الأخرى هي: `{{product.excipients_short_list}}`.  
- **شكل الدواء ومحتوى العبوة:** علبة تحتوي على `{{product.pack_sizes}}` وحدة في صفائح مغلفة مع النشرة.  
- **المؤسسة الصيدلانية المسؤولة عن التسويق:** `{{establishment.name}}`، `{{establishment.address}}` - الجزائر.  
- **المُصنّع:** `{{site_bulk.name}}`، `{{site_bulk.address}}`.

---

## VALIDATION ET BON À TIRER (BAT)

| Rôle BPF | Nom & Prénom | Titre | Date & Heure UTC | Empreinte Numérique (SHA-256) |
|:---|:---|:---|:---|:---|
| **Maker** | `{{audit.maker_name}}` | Traducteur / Affaires Réglem. | `{{audit.maker_timestamp}}` | `{{audit.maker_sha256}}` |
| **Checker** | `{{audit.checker_name}}` | Pharmacien Information Médicale | `{{audit.checker_timestamp}}` | `{{audit.checker_sha256}}` |
| **Approver** | `Dr. {{pdt.full_name}}` | Pharmacien Directeur Technique | `{{audit.approver_timestamp}}` | `{{audit.approver_sha256}}` |
