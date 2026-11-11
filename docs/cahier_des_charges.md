# Cahier des charges fonctionnel & technique détaillé
## Plateforme collaborative de gestion de l'enregistrement des médicaments, de l'homologation des dispositifs médicaux et GED pharmaceutique en Algérie

**Version :** 1.0 (Harmonisation Réglementaire ANPP & Lignes Directrices Officielles) — Novembre 2026.  
**Documents sources et référentiels officiels de l'ANPP intégrés :**
1. **Guide pour la soumission du Dossier Technique Commun (CTD)** destiné à l'enregistrement des produits pharmaceutiques à usage humain (DEPP/ANPP, Version 01, Date d'effet : 31/07/2025).
2. **Ligne directrice relative aux modifications des décisions d’enregistrement** des produits pharmaceutiques à usage de la médecine humaine (Code : `LD-DEPP07/V01`, Version 01, Date : 22/04/2026).
3. **Décret exécutif n° 20-325** du 22 novembre 2020 relatif aux modalités d'enregistrement des produits pharmaceutiques.
4. **Arrêté du 26 Safar 1443 (3 octobre 2021)** fixant les modalités de modification de la décision d'enregistrement des produits pharmaceutiques à usage humain.
5. **Arrêté du 07 novembre 2021** (Notamment article 09 régissant le regroupement des modifications - *Grouping*).
6. **Décret exécutif n° 20-324** et **Arrêté du 10 mai 2021** fixant les modalités d'homologation des dispositifs médicaux (DM).

---

## 1. Contexte, Cadre Juridique et Objectifs Stratégiques

### 1.1 Contexte Réglementaire Algérien
En Algérie, la mise sur le marché d'un médicament à usage de la médecine humaine est soumise à l'obtention d'une **Décision d'Enregistrement (DE)**, délivrée après évaluation technique, pharmaceutique, biologique, préclinique et clinique par la **Direction de l'Enregistrement des Produits Pharmaceutiques de l'Agence Nationale des Produits Pharmaceutiques (ANPP/DEPP)**. De même, les dispositifs médicaux (DM) et dispositifs médicaux de diagnostic in vitro (DMDIV) sont assujettis à une **Décision d'Homologation**.

Deux textes fondamentaux publiés par l'ANPP structurent l'ensemble du cycle de vie des médicaments :
1. **Le Guide CTD ANPP (Juillet 2025) :** Instaure l'exigence du format CTD (Common Technical Document - Modules 1 à 5) pour toute soumission d'enregistrement initial, de renouvellement quinquennal et de variation, en formalisant les exigences spécifiques pour les produits princeps (innovateurs), génériques importés et génériques fabriqués localement.
2. **La Ligne Directrice des Modifications LD-DEPP07/V01 (Avril 2026) :** Établit la nomenclature algérienne des modifications post-enregistrement en trois classes rigoureuses : **Modification Mineure (Mmin)**, **Modification Modérée (Mmod)** et **Modification Majeure (Mmaj)**, rompant définitivement avec les terminologies étrangères obsolètes (Types IA/IB/II européens). Elle formalise également le principe de reclassement automatique en cas de non-respect d'une condition et la possibilité de regroupement de modifications (*Grouping*, article 09 de l'arrêté du 07/11/2021).

### 1.2 Objectifs de la Plateforme "Homolog"
- **Centraliser et piloter le cycle de vie complet 360°** de tous les dossiers d'enregistrement de médicaments et d'homologation de DM au sein d'une instance unique.
- **Assurer la conformité native au format CTD ANPP 2025** : instancier automatiquement les checklists et arborescences selon le type de produit (Princeps, Générique importé, Générique local) et le type de procédure (**NI**, **RI**, **MI**, **NF**, **RF**, **MF**).
- **Intégrer un Coffre-Fort GED Pharmaceutique (Standard BPF / ALCOA+)** garantissant l'intégrité, la gestion des versions (mineures/majeures), les circuits d'approbation et la signature électronique légale du Directeur Technique.
- **Intégrer un Moteur Intelligent de Variations (LD-DEPP07/V01)** : détecter l'impact d'une modification documentaire, qualifier le type de modification (Mmin, Mmod, Mmaj) parmi les 34 rubriques réglementaires, vérifier les conditions, exécuter le reclassement dynamique et générer automatiquement le dossier complet de modification (avec tableau comparatif *État actuel vs Proposé* et attestation de non-modification des autres sections).
- **Anticiper les échéances critiques et péremptions** : renouvellements quinquennaux à J-180/J-90, réponses aux compléments ANPP avec compte à rebours légal, expiration des certificats BPF, CEP et licences d'exploitation.
- **Garantir la souveraineté numérique et la flexibilité hybride** : déployable en mode On-Premise 100 % hors-ligne (en intranet sécurisé de laboratoire) ou en mode Cloud SaaS souverain multi-tenant avec réversibilité absolue via l'archive chiffrée `.homolog`.
- **Offrir une ergonomie "Pharma Explorer" haute densité** : navigation 3 volets inspirée de Windows Explorer Pro, affichage compact (32px par ligne), volet de prévisualisation PDF instantané et bilinguisme intégral FR/AR avec support RTL natif.

---

## 2. Typologie des Procédures et Matrice CTD ANPP (Guide Juillet 2025)

### 2.1 Les 6 Parcours Officiels de Soumission
La plateforme distingue nativement les six codes de procédure définis par l'ANPP :

| Code ANPP | Intitulé officiel de la procédure | Portée | Périmètre documentaire principal |
|---|---|---|---|
| **NI** | **Enregistrement Initial d’un produit Importé** | Médicaments fabriqués à l'étranger importés en Algérie | Dossier CTD complet + CPP + Prix FOB Chambre Commerce + Benchmark |
| **RI** | **Renouvellement de la DE d'un produit Importé** | Renouvellement quinquennal d'un produit importé | Module 1 actualisé + Déclaration de non-modification + BPF/CPP à jour |
| **MI** | **Modification d’un produit Importé** | Variations administratives ou qualitatives | Dossier de modification ciblé + Approbation pays d'origine + Sections CTD modifiées |
| **NF** | **Enregistrement Initial d’un produit Fabriqué localement** | Médicaments produits sur un site industriel national | CTD complet + Structure de prix PCSU + Dossiers lots validation + Non-commercialisation |
| **RF** | **Renouvellement de la DE d’un produit Fabriqué localement** | Renouvellement quinquennal production locale | Module 1 + Déclaration de non-modification + BPF national + Revue qualité produit |
| **MF** | **Modification d’un produit Fabriqué localement** | Variations administratives ou techniques locales | Dossier de variation ciblé + Données comparatives de lots/stabilité + Sections CTD |

### 2.2 Structure des 5 Modules CTD selon le Type d'Origine
Conformément au tableau récapitulatif officiel de l'ANPP :

```
                                 ┌──────────────────────────────────────────────────────────┐
                                 │                STRUCTURE GLOBALE CTD ANPP                │
                                 └────────────────────────────┬─────────────────────────────┘
                                                              │
                     ┌────────────────────────────────────────┴────────────────────────────────────────┐
                     ▼                                                                                 ▼
     ┌───────────────────────────────┐                                                 ┌───────────────────────────────┐
     │       PRODUIT PRINCEPS        │                                                 │       PRODUIT GÉNÉRIQUE       │
     │      (Médicament Innovant)    │                                                 │    (Importé ou Localisé)      │
     ├───────────────────────────────┤                                                 ├───────────────────────────────┤
     │ Module 1 : Administratif [E]  │                                                 │ Module 1 : Administratif [E]  │
     │ Module 2 : Résumés [E]        │                                                 │ Module 2 : Incorporé (3, 4, 5)│
     │ Module 3 : Qualité [E]        │                                                 │ Module 3 : Qualité [E]        │
     │ Module 4 : Non Clinique [E]   │                                                 │ Module 4 : Facultatif [F]     │
     │ Module 5 : Clinique [E]       │                                                 │ Module 5 : Bioéquivalence [P] │
     └───────────────────────────────┘                                                 └───────────────────────────────┘
     [E] = Exigé obligatoirement      [F] = Facultatif (ANPP peut exiger à tout moment)   [P] = Partiellement requis (BE/BA)
```

- **Règle Princeps :** Modules 1, 2, 3, 4 et 5 obligatoirement exigés (`E`).
- **Règle Générique :**
  - **Module 1 :** Exigé (`E`) dans son intégralité selon les exigences algériennes.
  - **Module 2 :** Non requis sous forme de classeurs isolés 2.1 à 2.7 ; les résumés et synthèses sont directement incorporés dans les modules 3, 4 et 5.
  - **Module 3 :** Exigé (`E`) à l'identique d'un princeps (Substance Active 3.2.S, Produit Fini 3.2.P, Documentation régionale 3.2.R).
  - **Module 4 :** Facultatif (`F`) : non requis lors de la soumission initiale, mais l'ANPP se réserve formellement le droit d'exiger des données de toxicité ou pharmacologie si des impuretés nouvelles ou des excipients nouveaux apparaissent.
  - **Module 5 :** Partiellement requis (`P`) : focalisé impérativement sur la démonstration de la bioéquivalence (5.3.1.2) ou son exonération légitime (*Biowaiver* justifié en 1.M.P.3), les méthodes bioanalytiques humaines (5.3.1.4), et la tolérance clinique.

---

## 3. Spécification Exhaustive du Module 1 (Spécificités ANPP Algérie)

Le Module 1 est découpé en quatre sous-sections normalisées (`1.A`, `1.H`, `1.M`, `1.P`). Chaque pièce possède un identifiant strict, un titre légal et des conditions d'exigibilité selon le vecteur de soumission (NI, RI, MI, NF, RF, MF) :

### 3.1 Sous-Module 1.A : APPLICANT (Demandeur & Établissement)

| Réf. ANPP | Désignation de l'exigence documentaire | NI | RI | MI | NF | RF | MF | Règles métier de validation |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|---|
| **1.A.1** | Récépissé de dépôt du dossier | E | E | E | E | E | E | Document scanné avec numéro d'enregistrement ANPP et horodatage |
| **1.A.2** | Courrier de demande officiel | E | E | E | E | E | E | À en-tête du laboratoire, signé par le Pharmacien Directeur Technique |
| **1.A.3** | Formulaire de demande parafé, daté, signé et cacheté par le PDT | E | E | E | E | E | E | Formulaire officiel ANPP, paraphes sur chaque page, signature et cachet DT |
| **1.A.4** | Bordereau de versement + document bancaire / quittance ou autres + Attestation de taux de change si applicable | E | E | E | E | E | E | Preuve de paiement des redevances réglementaires, attestation de change Banque d'Algérie si devises |
| **1.A.5** | Copie de la réponse à la demande de pré-soumission (+ copie justificatif de paiement + BV) | E | NA | E | E | NA | E | Obligatoire pour NI, MI, NF, MF. Accord formel de pré-soumission de l'ANPP |
| **1.A.6** | Agrément d'exploitation / Fabrication / Sous-traitance | E | E | E | E | E | E | Agrément en cours de validité délivré par le Ministère de l'Industrie Pharmaceutique |
| **1.A.7** | Décision d'exercice du Pharmacien Directeur Technique (PDT) | E | E | E | E | E | E | Décision ministérielle nominative nommant le DT responsable |
| **1.A.8** | Copie des Contrats (exploitation, façonnage, sous-traitance et tout autre accord qualité) | E | E | E | E | E | E | Contrats liant le demandeur aux différents sites et détenteurs |
| **1.A.9** | Copie(s) de la DE, des approbations et des accusés des anciens dépôts (récépissés, notifications, paiements antérieurs) | NA | E | E | NA | E | E | Historique complet indispensable pour les renouvellements et variations |

### 3.2 Sous-Module 1.H : HOLDER (Détenteur de la Décision d'Enregistrement)

| Réf. ANPP | Désignation de l'exigence documentaire | NI | RI | MI | NF | RF | MF | Règles métier de validation |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|---|
| **1.H.1** | Courrier de demande du détenteur | E | E | E | E | E | E | Courrier officiel autorisant le demandeur local à déposer |
| **1.H.2** | Autorisation de Mise sur le Marché (original ou copie certifiée conforme) | E | E | E | E | E | E | AMM délivrée dans le pays d'origine par l'autorité compétente |
| **1.H.3** | Certificat de Produit Pharmaceutique (CPP - Modèle OMS) | E | E | E | E | E | E | CPP légalisé/apostillé conforme au modèle de certification OMS |
| **1.H.4** | Structure de prix (PCSU Fabrication) / Attestation de prix FOB Algérie visée Chambre de Commerce | E | E | E | E | E | E | Structure PCSU (Prix de Revient/Sortie Usine) pour le local (NF/RF/MF) ; Attestation FOB visée pour l'import (NI/RI/MI) |
| **1.H.5** | Dossier BENCHMARK de prix international | E | E | E | NA | NA | NA | Comparatif de prix dans les pays de référence (obligatoire pour importations) |
| **1.H.6** | Approbation de la variation par le pays d'origine | NA | NA | E | NA | NA | E | Preuve que la modification est approuvée dans le pays d'origine (ou justification) |
| **1.H.7** | Descriptif de la mise à jour signé | NA | NA | E | NA | NA | E | Document de synthèse expliquant le rationnel technico-réglementaire |
| **1.H.8** | Liste des sites intervenants dans la fabrication des substances actives et du produit fini (signée et visée) | E | E | E | E | E | E | Cartographie exhaustive des sites (SA, intermédiaires, PF, conditionnement, contrôle) |
| **1.H.9** | Déclaration de similarité | E | NA | NA | E | NA | NA | Attestation certifiant la similarité de composition et procédé |
| **1.H.10** | Déclaration de non-modification | NA | E | NA | NA | E | NA | Attestation formelle certifiant qu'aucune modification non autorisée n'a été opérée |
| **1.H.11** | Accord du détenteur actuel sur le transfert de la DE | NA | NA | E | NA | NA | E | Requis si la variation implique un transfert de titulaire de la DE |
| **1.H.12** | Attestation de déclaration de la date d’implémentation de la variation dans le pays d’origine | NA | NA | E | NA | NA | NA | Engagement sur la date de bascule industrielle dans le pays d'origine |

### 3.3 Sous-Module 1.M : MANUFACTURERS (Fabricants SA et Produit Fini)

| Réf. ANPP | Désignation de l'exigence documentaire | Cible | NI | RI | MI | NF | RF | MF | Exigences techniques clés |
|---|---|---|:---:|:---:|:---:|:---:|:---:|:---:|---|
| **1.M.S.1** | Agrément(s) du/des site(s) de fabrication SA (Manufacturing License) | API / SA | E | E | E | E | E | E | Licence de fabrication officielle de l'usine d'API |
| **1.M.S.2** | Certificat(s) BPF/GMP de conformité aux Bonnes Pratiques de Fabrication de la SA | API / SA | E | E | E | E | E | E | Certificat GMP valide émis par autorité compétente |
| **1.M.S.3** | Lettre d’accès / Certificate of Suitability (CEP) | API / SA | E | E | E | E | E | E | CEP EDQM valide avec déclaration d'accès nominative au nom du fabricant du FPP |
| **1.M.P.1** | Agrément d’ouverture d'établissement pharmaceutique (ML) | FPP (PF) | E | E | E | E | E | E | Licence d'établissement pour le fabricant du produit fini |
| **1.M.P.2** | Certificat(s) de conformité BPF/GMP du produit fini | FPP (PF) | E | E | E | E | E | E | Certificat BPF valide pour toutes les étapes de fabrication et conditionnement |
| **1.M.P.3** | Étude de bioéquivalence OU justificatif d’exonération de l'étude (Biowaiver) | FPP (PF) | E | NA | NA | E | NA | NA | Rapport clinique BE in vivo ou justification scientifique robuste (BCS classe I/III) |

### 3.4 Sous-Module 1.P : PRODUCT INFORMATION (Informations sur le Produit & Contrôle)

| Réf. ANPP | Désignation de l'exigence documentaire | NI | RI | MI | NF | RF | MF | Règles métier et contraintes |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|---|
| **1.P.1** | **Textes officiels pays d'origine :**<br>• 1.P.1.1 RCP d'origine<br>• 1.P.1.2 Notice d'origine<br>• 1.P.1.3 Étiquette d'origine | E | E | E | E | E | E | Documents approuvés dans le pays exportateur ou de référence |
| **1.P.2** | **Modèle Algérie (visé obligatoirement par le PDT) :**<br>• 1.P.2.1 RCP Modèle Algérie<br>• 1.P.2.2 Notice Modèle Algérie (Bilingue FR/AR)<br>• 1.P.2.3 Étiquette Modèle Algérie<br>• 1.P.2.4 Maquette Modèle Algérie (Conditionnement primaire & secondaire) | E | E | E | E | E | E | Textes conformes à la réglementation algérienne (notamment arabe obligatoire), maquettes en couleurs à échelle 100%, visées par le DT |
| **1.P.3** | Échantillons du produit fini | E | NA | E | E | NA | E | Enregistrement de la remise physique au laboratoire de contrôle de l'ANPP |
| **1.P.4** | Bulletin d'analyse de la substance active (avec granulométrie et polymorphisme) et son recontrôle | E | E | E | E | E | E | CoA du fabricant SA + bulletin de recontrôle complet par le fabricant du produit fini |
| **1.P.5** | Bulletin d’analyse du produit fini (et bulletin du prémix le cas échéant) | E | E | E | E | E | E | Certificat d'analyse libératoire complet conforme aux spécifications |
| **1.P.6** | Attestation de correspondance entre les lots | E | E | E | E | E | E | Traçabilité entre le lot soumis, le lot de bioéquivalence et les lots de stabilité |
| **1.P.7** | Certificat de libération de lot (précisant la libération pays d'origine / pays tiers) | E | E | E | E | E | E | Certificat officiel de libération par la personne qualifiée (QP/DT) |
| **1.P.8** | Lettre d’engagement de non-commercialisation des lots de validation | NA | NA | NA | E | NA | E | **Spécifique Fabrication Locale (NF/MF) :** Engagement formel signé du DT à ne pas commercialiser les lots industriels de validation avant l'obtention de la DE |
| **1.P.9** | Lettre d’engagement du demandeur à compléter les résultats des études de stabilité du PF | NA | NA | NA | E | NA | E | **Spécifique Fabrication Locale (NF/MF) :** Engagement à poursuivre les stabilités en temps réel (Zone IVa/IVb) et transmettre les résultats à l'ANPP |

---

## 4. Modules Techniques CTD 2, 3, 4 et 5 (Exigences et Spécificités Régionales)

### 4.1 Module 3 : Qualité Pharmaceutique, Chimique et Biologique
Le Module 3 est obligatoire (`E`) pour toutes les catégories (Princeps, Génériques importés, Génériques locaux). Le système intègre des contrôles stricts sur les éléments critiques imposés par l'ANPP :
1. **Partie 3.2.S (Substance Active) :**
   - Acceptabilité formelle du **CEP (Certificat de conformité à la Pharmacopée Européenne)** : la présence d'un CEP valide avec sa déclaration d'accès et ses annexes allège la documentation technique requise tout en garantissant la conformité analytique.
   - Contrôle obligatoire du **polymorphisme cristallin** et de la **granulométrie** (taille des particules) en 3.2.S.1.3 et 3.2.S.4.
2. **Partie 3.2.P (Produit Fini) :**
   - **Exigence spécifique fabrication locale (3.2.P.3.5) :** Dépôt obligatoire des dossiers complets de fabrication et de conditionnement des **lots industriels de validation** exécutés sur le site national.
   - Spécifications complètes à libération et à péremption (3.2.P.5.1).
3. **Partie 3.2.R (Informations Régionales Algériennes) :**
   - `3.2.R.1.1` : Exécution des documents de production (Master Batch Records exécutés).
   - `3.2.R.1.2` : Documents de production principaux (Formules cadres de fabrication et conditionnement).
   - `3.2.R.2` : Procédures analytiques et informations de validation sur site.

### 4.2 Module 5 : Données Cliniques et Bioéquivalence
Pour les médicaments génériques (importés et locaux), le moteur de checklist conditionne la recevabilité du dossier aux exigences suivantes :
- **5.3.1.2 Étude comparative de Biodisponibilité (BA) et Bioéquivalence (BE) :** Rapport complet de l'étude clinique in vivo démontrant la bioéquivalence du produit candidat vis-à-vis du produit princeps de référence (intervalle de confiance 90 % compris entre 80,00 % et 125,00 % sur Cmax et AUC).
- **5.3.1.4 Méthodes bioanalytiques et analytiques :** Validation complète selon les lignes directrices internationales (BPL/GCP).
- **Justification d'exonération (Biowaiver) :** En cas d'absence d'étude in vivo, rattachement obligatoire d'un dossier justificatif complet en 1.M.P.3 et 5.3.1.2 (profils de dissolution comparatifs multipoints à pH 1.2, 4.5, 6.8 et classification biopharmaceutique BCS).

---

## 5. Moteur Intelligent des Modifications (Ligne Directrice ANPP LD-DEPP07/V01)

### 5.1 Cadre Conceptuel et Règle d'Escalade Dynamique
L'ANPP classe toute modification post-approbation de la décision d'enregistrement en trois catégories exclusives :
1. **Modification Mineure (Mmin) :** Répercussions sur la qualité, la sécurité ou l'efficacité minimes ou nulles.
2. **Modification Modérée (Mmod) :** Impact potentiel sur la qualité, la sécurité ou l'efficacité.
3. **Modification Majeure (Mmaj) :** Répercussions significatives sur la qualité, la sécurité ou l'efficacité.

```
       ┌────────────────────────────────────────────────────────────────────────┐
       │             ALGORITHME DE QUALIFICATION & RECLASSEMENT ANPP            │
       └───────────────────────────────────┬────────────────────────────────────┘
                                           │
                                 Saisie de la variation
                             (Sélection parmi rubriques 1 à 34)
                                           │
                                           ▼
                       Évaluation des "Conditions à remplir"
                                           │
                      ┌────────────────────┴────────────────────┐
                      ▼                                         ▼
            Toutes conditions remplies             Au moins une condition non remplie
                      │                                         │
                      ▼                                         ▼
           Classification nominale               RECLASSEMENT AUTOMATIQUE OBLIGATOIRE
             (ex: Mmin ou Mmod)                                 │
                                                ┌───────────────┴───────────────┐
                                                ▼                               ▼
                                         Mmin  ──>  Mmod ou Mmaj         Mmod  ──>  Mmaj
```

> [!IMPORTANT]
> **Règle réglementaire de reclassement (Section 5.1 & 5.2 de la Ligne Directrice) :**
> Dès lors qu'une ou plusieurs conditions prédéfinies pour une modification mineure (`Mmin`) ou modérée (`Mmod`) ne sont pas strictement satisfaites, la plateforme réévalue et **reclasse automatiquement** la demande au niveau supérieur (`Mmin` reclassée en `Mmod` ou `Mmaj`, et `Mmod` reclassée en `Mmaj`), en ajustant instantanément la liste des documents justificatifs requis.

### 5.2 Regroupement de Modifications (*Grouping* - Arrêté du 07/11/2021)
En vertu de l'article 09 de l'arrêté du 07 novembre 2021, la plateforme permet de soumettre plusieurs modifications au sein d'une même procédure groupée (*Grouping*) lorsque :
- Elles concernent une même décision d'enregistrement et découlent d'un même changement technique (ex: changement de site de fabrication de l'API entraînant une modification de taille de lot et de méthode analytique).
- La plateforme calcule le niveau global de risque du groupe : si un groupe contient une modification `Mmaj`, l'ensemble du dossier groupé est instruit selon le niveau de sévérité maximal `Mmaj`.

### 5.3 Documentation Générale Requise pour TOUTE Demande de Modification
Selon la section 5.3 de la Ligne Directrice, toute soumission de variation (Mmin, Mmod, Mmaj) génère obligatoirement le bordereau de pièces maîtresses suivant :
1. **Lettre de demande officielle** de modification de la DE, visée par le PDT.
2. **Tableau comparatif standardisé :** État actuel approuvé vs Situation proposée, définissant clairement la portée du changement.
3. **Justificatif technico-réglementaire** motivant la modification.
4. **Bordereau de versement** de la redevance ANPP applicable (+ justificatif bancaire).
5. **Version modifiée des sections concernées du dossier CTD** (remplacement page à page).
6. **Liste actualisée et visée** de tous les intervenants (fabricants SA, PF, sites de conditionnement et de contrôle).
7. **Attestation formelle de non-modification des autres sections** en dehors de celles déclarées.
8. **Copie de la décision d’enregistrement (DE) en vigueur**.
9. **Accusés de dépôt** des demandes de modifications et/ou renouvellements actuellement en cours d'instruction par l'ANPP avec leurs bordereaux.
10. **Approbation de la modification par les autorités du pays d'origine** (pour les médicaments importés).
11. **Autorisations des sites et preuves BPF/BPL** valides pour les opérations concernées.
12. **Informations "produit" révisées** (RCP, notice bilingue, étiquetage, maquettes visées PDT).

### 5.4 Référentiel des 34 Rubriques de Modifications Réglementaires

#### I. Modifications Administratives (LD-DEPP07 Section 5.4.1)
- **1.** Changement dans la dénomination commerciale du produit fini.
- **2.** Changement dans la dénomination de la substance active, d'un excipient ou d'un article de conditionnement (avec maintien de l'entité chimique/spécification).
- **3.** Changement ou attribution du code ATC par l'OMS.
- **4.** Changement du nom et/ou de l'adresse du titulaire de l'AMM, du titulaire du dossier permanent de la substance active (DPSA/DMF), du site de stockage de la banque de cellules, du site de fabrication de l'API, du produit intermédiaire ou du produit fini, du conditionneur primaire/secondaire, ou du site de contrôle.
- **5.** Suppression d'un site de fabrication, de contrôle, de conditionnement ou de stockage (avec maintien d'au moins un site approuvé opérationnel).

#### II. Modifications relatives au CEP ou Préqualification API (LD-DEPP07 Section 5.4.1.2)
- **6.a.** Soumission d'un CEP nouveau ou mis à jour provenant d'un fabricant déjà approuvé (`6.a.1`  →  Mmin, `6.a.2`  →  Mmin, `6.a.3`  →  Mmod).
- **6.b.** Soumission d'un CEP provenant d'un nouveau fabricant (`6.b.1`  →  Mmin, `6.b.2`  →  Mmod).

#### III. Modifications Qualitatives Substance Active 3.2.S (LD-DEPP07 Section 5.4.2.1)
- **7.** Remplacement ou ajout d'un site de fabrication ou d'un nouveau fabricant d'API :
  - *7.a :* Uniquement le contrôle de la SA (`7.a.1`  →  Mmin, `7.a.2`  →  Mmod).
  - *7.b :* Production du matériau de départ (`7.b.1`  →  Mmin, `7.b.2`  →  Mmaj).
  - *7.c :* Production de l'intermédiaire (`7.c.1`  →  Mmod, `7.c.2`  →  Mmin).
  - *7.d :* Production complète de l'API (`7.d.1`  →  Mmod, `7.d.2`  →  Mmaj).
- **8.a.** Ajout ou changement d'un bloc de fabrication ou d'une unité de production sur un site API approuvé  →  Mmin.
- **9.** Changement dans le procédé de fabrication de l'API :
  - *9.a.1 :* Changement mineur sans altération du profil d'impuretés ni de la voie de synthèse  →  Mmod.
  - *9.a.2 :* Changement important susceptible d'avoir des répercussions significatives  →  Mmaj.
- **10.** Modification de la taille de lot de l'API ou d'un intermédiaire :
  - *10.a.1 :* Augmentation jusqu'à 10 fois la taille acceptée  →  Mmin.
  - *10.b :* Réduction de la taille de lot  →  Mmin.
  - *10.c.1 :* Augmentation supérieure à 10 fois la taille acceptée  →  Mmod.
- **11.** Essais en cours de fabrication ou limites durant la synthèse de l'API :
  - *11.a :* Resserrement des limites  →  Mmin.
  - *11.b :* Changement mineur d'une procédure analytique  →  Mmin.
  - *11.c :* Ajout d'un nouveau paramètre de spécification  →  Mmin.
  - *11.d :* Suppression d'un paramètre redondant  →  Mmin.
  - *11.e :* Ajout suite à un problème de qualité/sécurité  →  Mmod.
  - *11.f :* Élargissement des limites acceptées  →  Mmaj.
- **12.** Changement des spécifications et critères d'acceptation de l'API :
  - *12.a :* Mise à jour suite à révision de monographie de Pharmacopée officielle  →  Mmin.
  - *12.c :* Ajout de paramètre (`12.c.1`  →  Mmin, `12.c.2`  →  Mmod, `12.c.3`  →  Mmaj).
  - *12.b :* Suppression de paramètre (`12.b.1`  →  Mmin, `12.b.2`  →  Mmod, `12.b.3`  →  Mmaj).
  - *12.d :* Remplacement de paramètre (`12.d.1`  →  Mmin, `12.d.2`  →  Mmod, `12.d.3`  →  Mmaj).
  - *12.e :* Resserrement d'un critère d'acceptation  →  Mmin.
  - *12.f :* Élargissement d'un critère (`12.f.1`  →  Mmin, `12.f.2`  →  Mmod, `12.f.3`  →  Mmaj).
- **13.** Procédures d'essai de la SA ou de ses intermédiaires :
  - *13.a :* Mise à jour suite à révision pharmacopée officielle  →  Mmin.
  - *13.b :* Passage d'une méthode interne à une pharmacopée reconnue  →  Mmin.
  - *13.c :* Ajout d'une procédure (`13.c.1`  →  Mmin, `13.c.2`  →  Mmod, `13.c.3`  →  Mmaj).
  - *13.d :* Modification/remplacement de procédure (`13.d.1`  →  Mmin, `13.d.2`  →  Mmod, `13.d.3`  →  Mmaj).
  - *13.e :* Suppression d'une procédure alternative (`13.e.1/2`  →  Mmin, `13.e.3`  →  Mmaj).
- **14.** Changement de conditionnement primaire de la SA :
  - *14.a :* Composition qualitative/quantitative  →  Mmin.
  - *14.b :* SA liquides non stériles  →  Mmod.
- **15.** Période de recontrôle ou durée de vie de l'API :
  - *15.a :* Réduction  →  Mmin.
  - *15.b :* Extension  →  Mmod.
- **16.** Conditions de stockage de l'API sur l'étiquette :
  - *16.a :* Conditions plus restrictives  →  Mmin.
  - *16.b :* Tout autre changement  →  Mmod.

#### IV. Modifications Qualitatives Produit Fini 3.2.P (LD-DEPP07 Section 5.4.2.2)
- **17.** Système d'aromatisation ou de coloration du produit fini :
  - *17.a :* Augmentation ou réduction de teneur  →  Mmin.
  - *17.b :* Ajout, suppression ou remplacement d'arôme/colorant  →  Mmin.
- **18.a.** Masse de l'enrobage ou des capsules vides pour formes orales solides à libération immédiate  →  Mmin.
- **19.** Composition en excipients d'une forme orale solide à libération immédiate :
  - *19.a :* Remplacement d'un excipient par un excipient comparable à concentration similaire  →  Mmod (ou Mmaj si conditions non satisfaites).
  - *19.b :* Changement quantitatif des excipients  →  Mmaj.
- **20.** Site de fabrication du produit fini :
  - *20.a :* Site de conditionnement secondaire  →  Mmin.
  - *20.b :* Site de conditionnement primaire (solides, crèmes, sirops)  →  Mmin.
  - *20.c :* Site chargé de la fabrication (procédé conventionnel non complexe)  →  Mmod.
- **21.a.** Remplacement ou ajout d'un site de contrôle analytique  →  Mmin.
- **22.** Taille de lot du produit fini (formes orales solides ou solutions non stériles) :
  - *22.a :* Augmentation jusqu'à 10 fois la taille approuvée  →  Mmin.
  - *22.b :* Réduction jusqu'à 10 fois  →  Mmin.
  - *22.c :* Plus de 10 fois supérieure ou inférieure  →  Mmod.
- **23.** Changement dans le procédé de fabrication du produit fini (sans changement du principe fondamental)  →  Mmod.
- **24.** Contrôles en cours de fabrication du produit intermédiaire ou fini :
  - *24.a :* Resserrement des limites  →  Mmin.
  - *24.b :* Suppression d'un contrôle redondant  →  Mmin.
  - *24.c / 24.d :* Ajout, révision ou remplacement d'un contrôle  →  Mmin.
- **25.** Remplacement d'un excipient à risque EST par un excipient d'origine végétale ou synthétique  →  Mmin.
- **26.** Spécifications ou procédures analytiques d'un excipient :
  - *26.a :* Suppression d'un paramètre interne non critique  →  Mmin.
  - *26.b :* Ajout d'un nouveau paramètre  →  Mmin.
  - *26.c :* Resserrement des limites  →  Mmin.
  - *26.d :* Changement ou remplacement d'une méthode  →  Mmod.
- **27.** Mise à jour des spécifications du produit fini suite à révision d'une monographie de Pharmacopée officielle  →  Mmin.
- **28.** Spécifications et critères d'acceptation du produit fini :
  - *28.a :* Suppression d'un paramètre non critique redondant  →  Mmin.
  - *28.b :* Ajout d'un paramètre de contrôle  →  Mmin.
  - *28.c :* Resserrement d'un critère  →  Mmin.
  - *28.d :* Remplacement d'un paramètre  →  Mmin.
- **29.** Procédures analytiques du produit fini :
  - *29.a :* Suppression d'une procédure alternative  →  Mmin.
  - *29.b :* Ajout d'une procédure analytique  →  Mmin.
  - *29.c :* Modification ou remplacement d'une procédure  →  Mmod.
  - *29.d :* Mise à jour selon monographie pharmacopée  →  Mmin.
  - *29.e :* Transfert vers une pharmacopée reconnue  →  Mmin.
- **30.** Conditionnement primaire du produit fini (changement ou ajout de type de contenant) :
  - *30.a :* Formes pharmaceutiques non stériles  →  Mmod.
  - *30.b :* Médicaments stériles  →  Mmaj.
  - *30.c :* Suppression d'un contenant sans supprimer de dosage/forme  →  Mmod.
- **31.** Composition qualitative et quantitative du conditionnement primaire :
  - *31.a :* Formes solides (ex: blister vers autre blister équivalent)  →  Mmin.
  - *31.b :* Formes semi-solides et liquides non stériles  →  Mmod.
- **32.** Durée de conservation du produit fini pour la vente :
  - *32.a :* Réduction  →  Mmod.
  - *32.b :* Extension (sur base de données de stabilité temps réel)  →  Mmod.
- **33.** Durée de conservation après première ouverture, dilution ou reconstitution :
  - *33.a :* Réduction  →  Mmod.
  - *33.b :* Extension  →  Mmod.
- **34.** Changement des conditions de stockage du produit fini ou en cours d'utilisation  →  Mmod.

---

## 6. Architecture Technique et Ergonomie "Pharma Explorer"

### 6.1 Stack Technologique Native et Souveraine
- **Backend & Serveur Applicatif :** **Golang (Go 1.26+)**
  - Performance native sans runtime lourd, binaire unique auto-suffisant (`homolog.exe`).
  - Streaming de documents par morceaux (Chunked upload 5 Mo) résilient aux micro-coupures réseau.
  - Driver PostgreSQL natif `jackc/pgx/v5` avec pool de connexions optimisé.
- **Stockage Abstrait Sécurisé (`StorageProvider`) :**
  - Architecture CAS (Content-Addressed Storage) basée sur l'empreinte cryptographique **SHA-256**.
  - Déduplication physique des fichiers à la source et chiffrement AES-GCM sur disque.
- **Base de Données Relationnelle :** **PostgreSQL 16+**
  - Schémas configurables via colonnes `JSONB` indexées (checklists dynamiques, modèles de dossiers).
  - Recherche plein texte bilingue intégrée (`tsvector` / `tsquery`).
  - Table `audit_logs` strictement *append-only* (triggers SQL interdisant formellement `UPDATE` et `DELETE` selon 21 CFR Part 11).
- **Frontend Embarqué (`//go:embed`) :**
  - TypeScript strict, Tailwind CSS, Lucide Icons, PDF.js local sans appel à aucun CDN public externe.
  - Zéro dépendance Node.js requise sur le serveur en production.

### 6.2 Modèle Ergonomique "Pharma Explorer" à 3 Volets
Inspiré des outils d'exploration professionnelle haute densité, l'écran de travail optimise chaque pixel disponible :

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ [📁 Navigation]  C:\Homolog > Dossiers > ENR-2026-0042 (Amoxicilline 500mg) > CTD_Module_1 > 1.P     │
├────────────────────┬─────────────────────────────────────────────────┬─────────────────────────────────┤
│ ARBRE DE NAVIGATION│ VUE TABLEAU COMPACT (Hauteur ligne : 32px)      │ VOLET DE PRÉVISUALISATION       │
│ (Volet Gauche)     │ (Zone Centrale)                                 │ (Volet Droit - Alt+P)           │
├────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────┤
│ ▼ ENR-2026-0042    │ Réf  | Statut | Document        | Ver | Validé  │ [Visionneuse PDF.js Intégrée]   │
│   ► Mod 1 : Admin  │ 1.P.1| Valide | RCP_Origine.pdf | 1.0 | ✅ DT    │                                 │
│     ▼ 1.P Info     │ 1.P.2| Valide | Notice_FR_AR.pdf| 2.0 | ✅ DT    │ Titre : Notice_FR_AR.pdf        │
│       • 1.P.1 (3/3)│ 1.P.3| Manque | Échantillons    | -   | ❌       │ Version : 2.0 (En vigueur)      │
│       • 1.P.2 (4/4)│ 1.P.4| Valide | CoA_API_Lot.pdf | 1.0 | ✅ DT    │ SHA-256 : e3b0c44298fc1c...     │
│       • 1.P.3 (0/1)│ 1.P.5| Valide | CoA_PF_Lot.pdf  | 1.0 | ✅ DT    │ Horodatage : 2026-11-10 14:32Z  │
│       • 1.P.8 (1/1)│ 1.P.8| Valide | Engag_NonCom.pdf| 1.0 | ✅ DT    │                                 │
│   ► Mod 2 : Résumé │ 1.P.9| Valide | Engag_Stab.pdf  | 1.0 | ✅ DT    │ [ Signer Électroniquement (DT) ]│
│   ► Mod 3 : Qualité│                                                 │ [ Télécharger avec Certificat ] │
│   ► Mod 5 : BE/BA  │                                                 │                                 │
├────────────────────┴─────────────────────────────────────────────────┴─────────────────────────────────┤
│ Éléments : 7 | Complétude : 85 % (1 obligatoire manquante) | Stockage : 1.42 Go / 50 Go                │
└────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

- **Adaptation RTL Arabe Intégrale :**
  - Bascule instantanée LTR  ↔  RTL via `dir="rtl"`.
  - Inversion symétrique naturelle de l'interface : arbre de navigation ancré à droite, tableau de données au centre, volet de prévisualisation à gauche.
  - Polices locales sans appel réseau : `Cairo` (typographie arabe fluide) et `Inter` (typographie latine).

---

## 7. Sécurité, Rôles et Conformité 21 CFR Part 11 / BPF

### 7.1 Matrice des Rôles Stricts (9 Profils)
1. **Super-Administrateur Éditeur (Mode Cloud SaaS) :** Provisioning d'instances laboratoires, isolation des schémas PostgreSQL, quotas de stockage. Aucun accès aux données métier.
2. **Administrateur Système Local (Mode On-Premise) :** Sauvegardes, certificats TLS, surveillance système. Aucun accès aux documents pharmaceutiques.
3. **Administrateur de Laboratoire :** Gestion des comptes utilisateurs internes, attribution des profils.
4. **Responsable Affaires Réglementaires :** Création et pilotage des dossiers CTD/DM, coordination des pièces justificatives, suivi des échanges ANPP.
5. **Directeur Technique (Pharmacien DT) :** Seul titulaire du pouvoir légal d'approbation et de scellement. Droit de signature électronique légale bloquante pour validation finale des dossiers.
6. **Contributeur Qualité / BPF :** Gestion documentaire interne, téléversement de SOP, certificats BPF, rapports d'audit.
7. **Contributeur R&D / Production :** Dépôt des formules cadres, dossiers de lots de validation (3.2.P.3.5), données de stabilité et méthodes d'analyse.
8. **Consultant Externe :** Accès strictement cloisonné aux seuls dossiers auxquels il est invité (aucune visibilité sur le reste du laboratoire), avec date d'expiration contractuelle automatique.
9. **Auditeur / Inspecteur :** Consultation en lecture seule avec accès au journal d'audit immuable (`audit_logs`).

### 7.2 Signature Électronique Légale et Intégrité ALCOA+
- **Double facteur (2FA) TOTP 100 % hors-ligne (RFC 6238) :** Compatible Google Authenticator, Microsoft Authenticator, FreeOTP, avec 10 codes de secours à usage unique chiffrés.
- **Re-authentification à la signature :** Toute signature de pièce critique ou validation de dossier impose la re-saisie du mot de passe ou validation TOTP accompagnée de la mention solennelle légale obligatoire :
  > *"En qualité de Pharmacien Directeur Technique, j'atteste sur l'honneur l'exactitude, la véracité et la complétude des données du présent dossier soumis à l'ANPP conformément aux lois et réglementations en vigueur."*
- **Scellement cryptographique :** Génération d'une empreinte SHA-256 scellant le binaire du document, l'identifiant nominatif du signataire, le rôle officiel, l'horodatage UTC et l'adresse IP.

---

## 8. Taxonomie du Coffre-Fort GED Pharmaceutique (Standard BPF)

Les documents de référence du laboratoire sont organisés en cinq branches maîtresses interconnectées avec les checklists des dossiers d'enregistrement :

1. **`01_Affaires_Reglementaires`**
   - `01_01_Agrements_Etablissement` (Agréments MIPH demandeur et fabricants)
   - `01_02_Decisions_Enregistrement` (DE initiales, renouvellements et avenants)
   - `01_03_Decisions_Homologation` (Décisions d'homologation DM/DMDIV)
   - `01_04_Dossiers_CTD_Reference` (Dossiers maîtres archivés et fiches de variations)
   - `01_05_Notices_Articles_Conditionnement` (Notices bilingues FR/AR, étiquettes, maquettes à l'échelle)
   - `01_06_Correspondances_ANPP` (Courriers de compléments, notifications, accusés de réception)
2. **`02_Assurance_Qualite_BPF`**
   - `02_01_Manuel_Qualite_Procedures` (Procédures opératoires normalisées - POS / SOP)
   - `02_02_Certificats_BPF_ISO` (Certificats GMP fabricants SA et PF, certificats ISO 13485)
   - `02_03_Rapports_Audit_Inspections` (Rapports d'inspection ANPP, audits fournisseurs)
   - `02_04_CAPA_Deviations` (Déviations qualité et actions correctives/préventives)
3. **`03_Developpement_Production`**
   - `03_01_Formules_Fabrication_DMF` (Dossiers de lot maîtres, Drug Master Files / DPSA)
   - `03_02_Methodes_Controle_Validation` (Méthodes d'analyse, transferts et validations analytiques)
   - `03_03_Etudes_Stabilite` (Protocoles et rapports de stabilité Zone IVa/IVb, temps réel et accéléré)
   - `03_04_Specifications_Matieres_PF` (Spécifications SA, excipients, articles de conditionnement, PF)
   - `03_05_Dossiers_Techniques_DM` (Fichiers techniques DM selon arrêté du 10 mai 2021)
4. **`04_Vigilance_Post_Market`**
   - `04_01_Pharmacovigilance` (Rapports périodiques de sécurité - PSUR, fiches de notification d'effets indésirables)
   - `04_02_Materiovigilance` (Incidents DM, fiches d'alerte)
   - `04_03_Rappels_Lots` (Procédures de retrait et dossiers d'investigation)
5. **`05_Contrats_Tiers`**
   - `05_01_Accords_Qualite_QTA` (Quality Technical Agreements avec façonniers et fournisseurs)
   - `05_02_Contrats_Faconnage_SousTraitance` (Contrats de fabrication, conditionnement, analyse)
   - `05_03_Certificats_Fournisseurs_APIs` (CEP EDQM, lettres d'accès APIMF, déclarations EST/OGM)

---

## 9. Conclusion et Validation

Le présent cahier des charges constitue le socle fonctionnel et technique exhaustif de la plateforme **Homolog**. Il intègre l'état de l'art du droit pharmaceutique algérien, assure une traçabilité rigoureuse tout au long du cycle de vie du médicament (enregistrement, renouvellement, variations 1 à 34), et garantit l'alignement strict entre la gestion documentaire BPF et les impératifs réglementaires de l'ANPP.
