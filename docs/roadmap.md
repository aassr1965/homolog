# Feuille de route du projet (Roadmap)
## Plateforme collaborative de gestion de l'enregistrement des médicaments, de l'homologation des DM et GED pharmaceutique en Algérie

**Dernière mise à jour :** 15 novembre 2026  
**Document de référence :** [`docs/cahier_des_charges.md`](file:///c:/dev/homolog/docs/cahier_des_charges.md) (v1.0 - Harmonisation Réglementaire ANPP & Chapitre 10 - Spécifications Exhaustives des États de Sortie)  
**Textes réglementaires officiels algériens intégrés :**
- **Guide pour la soumission du Dossier Technique Commun (CTD)** destiné à l'enregistrement des produits pharmaceutiques à usage humain (DEPP/ANPP, Version 01, Date d'effet : 31/07/2025).
- **Ligne directrice relative aux modifications des décisions d’enregistrement** des produits pharmaceutiques à usage de la médecine humaine (Code : `LD-DEPP07/V01`, Version 01, Date : 22/04/2026).
- **Décret exécutif n° 20-325** du 22 novembre 2020 (Modalités d'enregistrement des médicaments).
- **Arrêté du 3 octobre 2021** (Modalités de modification de la DE) & **Arrêté du 7 novembre 2021** (Article 09 : Regroupement de modifications / *Grouping*).
- **Décret exécutif n° 20-324** et **Arrêté du 10 mai 2021** (Modalités d'homologation des dispositifs médicaux).  
- **Décret exécutif n° 21-82** (Barèmes des redevances et droits des prestations de l'ANPP - Formulaires BV1 et BV2).
- **Arrêté BPF 2024 / ICH Q10 / 21 CFR Part 11** (Intégrité des données ALCOA+, Piste d'audit immuable et Signature électronique légale).
**Architecture cible :** Go 1.26+ | PostgreSQL 16+ (`pgx/v5`) | Frontend embarqué TypeScript/Tailwind (RTL FR/AR) | Mode Hybride (On-Premise & SaaS)

---

## Vue d'ensemble des jalons

```mermaid
gantt
    title Planning de déploiement et d'implémentation Homolog
    dateFormat  YYYY-MM-DD
    section Phase 1 : Socle & GED
    Infra, Base PostgreSQL & CAS Dédupliqué (1.1)    :done, p1_1, 2026-11-10, 1d
    Auth 2FA, PSSI, RBAC 14 Perms & SoD (1.2)        :done, p1_2, 2026-11-10, 4d
    Référentiel Produits & Établissements ANPP (1.3) :done, p1_3, 2026-11-11, 1d
    Coffre-Fort GED Pharmaceutique BPF/ALCOA+ (1.4)  :done, p1_4, after p1_3, 4d
    Moteur Dossiers ANPP (CTD 2025 & DM 2021) (1.5)  :done, p1_5, after p1_4, 4d
    UI Haute Densité 'Pharma Explorer' 3 Volets (1.6):done, p1_6, after p1_5, 4d
    Studio Compléments & Assistant Soumission ANPP (1.7) :done, p1_7, after p1_6, 3d
    Packaging Docker & Archive .homolog (1.8)        :done, p1_8, after p1_7, 2d
    section Phase 2 : Complétude & Variations
    Atelier Élaboration & Workflow BPF Maker-Checker (2.1):active, p2_1, after p1_8, 4d
    Redevances, Devises & Justificatifs Bancaires(2.2):p2_2, after p2_1, 3d
    Moteur Variations ANPP (Mmin/Mmod/Mmaj)(2.3)     :p2_3, after p2_2, 4d
    Espace Consultants Externes Cloisonné (2.4)      :p2_4, after p2_3, 2d
    Régularisation DM Décret 20-324 (2.5)            :p2_5, after p2_4, 3d
    Recherche Plein Texte FTS Bilingue FR/AR (2.6)   :p2_6, after p2_5, 3d
    Console Super-Admin Éditeur Cloud SaaS (2.7)     :p2_7, after p2_6, 3d
    Mises à Jour Réglementaires Hors-Ligne .pkg (2.8):p2_8, after p2_7, 3d
    section Phase 3 : Extensions
    Vigilance (Matériovigilance & Pharmacovigilance) :p3_1, after p2_8, 5d
    Intégrations d'Entreprise SSO / LDAP / iCal      :p3_2, after p3_1, 4d
    Passerelle Officielle API ANPP                   :p3_3, after p3_2, 4d
```

---

## Phase 1 : Socle Fondateur, GED Pharmaceutique & Parcours 360° (Priorité M)
> **Objectif stratégique :** Disposer d'une plateforme opérationnelle de bout en bout permettant de créer un produit (Princeps, Générique importé, Générique fabriqué localement), monter un dossier d'enregistrement CTD conforme au guide ANPP 2025 (procédures NI, RI, MI, NF, RF, MF) ou d'homologation DM, déposer et signer électroniquement des pièces en GED, et exporter l'archive de réversibilité.

### Jalon 1.1 : Infrastructure, Base de données & Configuration Hybride
- [x] Initialiser l'arborescence Go (`cmd/server`, `internal/{api,domain,ged,regulatory,storage,database,audit}`)
- [x] Configurer la connexion PostgreSQL via `jackc/pgx/v5` avec pool de connexions asynchrone
- [x] Écrire les migrations initiales SQL (Schémas, tables de base, contraintes d'intégrité)
- [x] Gérer la configuration hybride par variables d'environnement (`DEPLOYMENT_MODE=on-premise` ou `saas`)
- [x] Mettre en place le `StorageProvider` (Driver CAS local sur disque chiffré avec hachage SHA-256 et déduplication physique)
- [x] Valider la suite de tests unitaires du stockage CAS (`storage_test.go`)

### Jalon 1.2 : Authentification Forte, 2FA Hors-Ligne, Rôles & Audit Trail (21 CFR Part 11) — *[Terminé : RBAC 14 Perms, SoD & Console Admin]*
- [x] **Socle d'authentification et gestion des sessions BPF :**
  - [x] Implémenter le hachage sécurisé des mots de passe avec **Argon2id** (mémoire 64 Mo, 3 itérations) et politique de complexité (min. 12 car., minuscules, majuscules, chiffres, caractères spéciaux).
  - [x] Développer la gestion des sessions par **Cookies HttpOnly sécurisés** (SameSite=Strict, HTTPS) avec table de sessions PostgreSQL pour révocation instantanée.
  - [x] Développer le second facteur **TOTP (RFC 6238)** 100% hors-ligne (QR code local, Google/Microsoft Authenticator, 10 codes de secours chiffrés).
  - [x] Mettre en place le verrouillage automatique de session sur inactivité (**Idle timeout 15 min BPF**) et protection anti-brute force (5 essais = 15 min de blocage).
  - [x] Valider les suites de tests unitaires d'authentification (`password_test.go`, `totp_test.go`, `auth_handler_test.go`).
- [x] **Intégrité ALCOA+, Piste d'audit & Signature légale PDT :**
  - [x] Créer la table `audit_logs` strictement *append-only* (triggers SQL PostgreSQL interdisant `UPDATE` et `DELETE`).
  - [x] Logger automatiquement toutes les actions sensibles (connexion, consultation, téléchargement, approbation, signature, modification de rôle).
  - [x] Créer le flux de **re-authentification pour signature électronique légale du PDT** :
    - Re-saisie du mot de passe ou validation TOTP avec la déclaration solennelle obligatoire :  
      *« En qualité de Pharmacien Directeur Technique, j'atteste sur l'honneur l'exactitude, la véracité et la complétude des données du présent dossier soumis à l'ANPP conformément aux lois et réglementations en vigueur. »*
    - Scellement cryptographique par empreinte SHA-256 liant le binaire du document, l'identité du signataire, son rôle officiel, l'horodatage UTC et l'adresse IP.
- [x] **Mise en conformité PSSI, RBAC Dynamique & Console Dédiée (Sections 6.3, 7.1 & 7.2 du CDC) :**
  - [x] Définir les 9 profils utilisateurs stricts dans le domaine métier (`internal/domain/rbac.go`).
  - [x] **Modèle de Données PostgreSQL Dynamique (Migration `005_rbac_and_dynamic_permissions.sql`) :**
    - Tables `roles`, `permissions`, `role_permissions`, `user_permissions`.
    - Pré-remplissage des 14 permissions unitaires et de la matrice par défaut des 9 rôles sans perturber les données existantes.
  - [x] **Moteur RBAC Backend Go & Garde-fous BPF Inviolables :**
    - Service `RBACService` et middleware `RequirePermission(perm)` / `RequireAnyPermission(perms...)`.
    - Garde-fous réglementaires inviolables en code Go : exercice du visa PDT réservé exclusivement au DT titulaire de la décision ministérielle (1.A.7), filtre Zero-Trust bloquant le téléchargement de documents pour les admins techniques (`admin_systeme_local`, `super_admin_editeur`), et non-auto-approbation SoD par pièce (*Maker* $\neq$ *Checker* $\neq$ *Approver*).
    - Système d'invalidation instantanée par versioning de permissions (`permissions_version`) pour répercussion immédiate sur sessions actives.
    - Interception stricte de chaque endpoint API (`HTTP 403 Forbidden`) avec journalisation d'alerte automatique dans `audit_logs` (`AUDIT_SECURITY_403_BLOCKED`).
  - [x] **Console Dédiée d'Administration du Laboratoire (Section 6.3 CDC) :**
    - Interface autonome dédiée pour l'Administrateur de Laboratoire (`web/js/admin.js`, gestion collaborateurs, modification interactive de rôle, matrice dynamique des 14 permissions unitaires, 4 politiques visuelles de garde-fous BPF/PSSI, export CSV certifié).
    - Commutateur de test instantané **Dev Role Switcher** dans la barre supérieure (`POST /api/v1/auth/dev-switch`, verrouillé strictement et désactivé en production `APP_ENV=production`, cookies avec attribut `Secure` dynamique TLS/proxy).
  - [x] **Gardes de Navigation Frontend & Masquage Dynamique de l'UI (Pharma Explorer) :**
    - Route Guards frontend interceptant les accès directs non autorisés par domaine avec popups sécurisées d'interdiction.
    - Moteur de rendu conditionnel masquant les composants non autorisés (`data-perm`, `data-perm-any`, bouton signature PDT, console admin, boutons téléchargement Zero-Trust).
    - Provisionnement de 9 collaborateurs réalistes avec mot de passe universel `Homolog2026!Dev` pour les tests complets.
  - [x] Développer et valider la suite de tests unitaires RBAC / Zero-Trust / Session (`internal/api/rbac_test.go`).

### Jalon 1.3 : Référentiel Établissements & Produits (Conforme Guide ANPP 2025)
- [x] **Modèle et tables PostgreSQL pour les Établissements :**
  - Table `establishments` : Raison sociale, statut juridique, agrément ministériel MIPH (numéro, date d'octroi, échéance), coordonnées du Pharmacien Directeur Technique (PDT - 1.A.7) avec décision ministérielle nominative.
  - Table `manufacturing_sites` (1.H.8) : Liste exhaustive des sites intervenants, rôle (Synthèse SA, Intermédiaire, Formulateur PF, Conditionneur primaire, Conditionneur secondaire, Laboratoire de contrôle analytique), licences de fabrication (ML 1.M.S.1, 1.M.P.1) et certificats BPF/GMP (1.M.S.2, 1.M.P.2) avec dates d'échéance.
- [x] **Modèle et tables pour les Médicaments (Typologie ANPP Juillet 2025) :**
  - Table `medicines` : Dénomination commerciale, DCI, forme galénique, dosage, voie d'administration, classe pharmacothérapeutique ATC (Code OMS).
  - Classification d'origine : `princeps` (innovateur), `generique_importe`, `generique_local`.
  - Données SA / API : Présence de monographie pharmacopée (Ph. Eur.), détention d'un CEP EDQM (1.M.S.3 / 3.2.S.3) avec lettre d'accès nominative, numéro APIMF/DMF (DPSA), forme polymorphique et criticité granulométrique (1.P.4 / 3.2.S.1.3 / 3.2.S.4).
  - Données Produit Fini : Conditions de stockage (°C, humidité), durée de conservation initiale et après reconstitution/ouverture. Dossiers de lots de validation industrielle (3.2.P.3.5 / 1.P.8) pour la fabrication locale.
  - Données économiques réglementaires : Structure de prix PCSU pour fabrication locale (1.H.4) ou Attestation prix FOB visée Chambre de Commerce et dossier de Benchmark international pour l'importation (1.H.4 / 1.H.5).
  - Références administratives : N° de Décision d'Enregistrement (DE), date d'enregistrement initial, date d'échéance quinquennale.
- [x] **Modèle et tables pour les Dispositifs Médicaux :**
  - Table `medical_devices` : Dénomination, référence fabricant, classe (I, IIa, IIb, III, DMDIV), stérilité, usage unique, décision d'homologation selon le décret exécutif n° 20-324 et l'arrêté du 10 mai 2021.
- [x] Support complet des champs bilingues FR/AR (`name_fr`, `name_ar`, `inn_fr`, `inn_ar`, `presentation_fr`, `presentation_ar`).
- [x] API RESTful : CRUD établissements, sites, médicaments et DM avec filtrage et pagination.

### Jalon 1.4 : Coffre-Fort GED Pharmaceutique (Standard BPF / ALCOA+)
- [x] **Arborescence standardisée préconfigurée en 5 branches maîtresses normalisées (Section 8 du CDC) :**
  - `01_Affaires_Reglementaires` :
    - `01_01_Agrements_Etablissement` (Agréments MIPH demandeur, façonniers et sous-traitants)
    - `01_02_Decisions_Enregistrement` (DE initiales, renouvellements et avenants ministériels)
    - `01_03_Decisions_Homologation` (Décisions d'homologation DM/DMDIV)
    - `01_04_Dossiers_CTD_Reference` (Dossiers maîtres archivés et dossiers de variations)
    - `01_05_Notices_Articles_Conditionnement` (Notices bilingues FR/AR, étiquettes, maquettes à l'échelle)
    - `01_06_Correspondances_ANPP` (Courriers de compléments, notifications, accusés de réception)
  - `02_Assurance_Qualite_BPF` :
    - `02_01_Manuel_Qualite_Procedures` (Manuel qualité, procédures opératoires normalisées - POS / SOP)
    - `02_02_Certificats_BPF_ISO` (Certificats GMP fabricants SA et PF, certificats ISO 13485)
    - `02_03_Rapports_Audit_Inspections` (Rapports d'inspection ANPP, audits fournisseurs)
    - `02_04_CAPA_Deviations` (Déviations qualité et actions correctives/préventives CAPA)
  - `03_Developpement_Production` :
    - `03_01_Formules_Fabrication_DMF` (Dossiers de lot maîtres, Drug Master Files / DPSA)
    - `03_02_Methodes_Controle_Validation` (Méthodes d'analyse, transferts et validations analytiques)
    - `03_03_Etudes_Stabilite` (Protocoles et rapports de stabilité Zone IVa/IVb, temps réel et accéléré)
    - `03_04_Specifications_Matieres_PF` (Spécifications SA, excipients, articles de conditionnement, PF)
    - `03_05_Dossiers_Techniques_DM` (Fichiers techniques DM selon l'arrêté du 10 mai 2021)
  - `04_Vigilance_Post_Market` :
    - `04_01_Pharmacovigilance` (Rapports périodiques de sécurité - PSUR, fiches de notification d'effets indésirables)
    - `04_02_Materiovigilance` (Signalements d'incidents DM, fiches d'alerte)
    - `04_03_Rappels_Lots` (Procédures de retrait et dossiers d'investigation de lot)
  - `05_Contrats_Tiers` :
    - `05_01_Accords_Qualite_QTA` (Quality Technical Agreements avec façonniers et fournisseurs)
    - `05_02_Contrats_Faconnage_SousTraitance` (Contrats de fabrication, conditionnement, analyse)
    - `05_03_Certificats_Fournisseurs_APIs` (CEP EDQM, lettres d'accès APIMF, déclarations EST/OGM)
- [x] Gestion du cycle de vie documentaire (*Brouillon* → *En revue* → *En approbation* → *En vigueur* → *Obsolète*).
- [x] Gestion des versions mineures/majeures (`v0.1`, `v1.0`, `v2.0`) avec saisie obligatoire du motif de révision et traçabilité.
- [x] Mécanisme de verrouillage de document (Check-out / Check-in) pour prévenir les éditions concurrentes.
- [x] Circuit d'approbation et **Signature électronique légale du Directeur Technique** (avec re-authentification 2FA, horodatage UTC et scellement SHA-256).
- [x] Téléversement résilient par morceaux (Chunked Upload 5 Mo) et déduplication physique par hash SHA-256 via le driver CAS.
- [x] Interface utilisateur Haute Densité 'Pharma Explorer' connectée en temps réel aux APIs GED du backend.

### Jalon 1.5 : Moteur de Dossiers ANPP & Checklists Dynamiques (Guide CTD Juillet 2025)
- [x] **Générateur dynamique de dossier CTD selon les 6 procédures officielles :**
  - **`NI`** : Enregistrement Initial d’un produit Importé.
  - **`RI`** : Renouvellement de la DE d'un produit Importé.
  - **`MI`** : Modification relative à un produit Importé.
  - **`NF`** : Enregistrement Initial d’un produit Fabriqué localement.
  - **`RF`** : Renouvellement de la DE d'un produit Fabriqué localement.
  - **`MF`** : Modification relative à un produit Fabriqué localement.
- [x] **Implémentation intégrale de la nomenclature officielle du Module 1 Algérie :**
  - **Sous-module 1.A (Applicant) :** Récépissé de dépôt (1.A.1), Courrier de demande (1.A.2), Formulaire officiel paraphé/signé/cacheté PDT (1.A.3), Bordereau de versement + document bancaire / quittance + attestation de change Banque d'Algérie (1.A.4), Réponse pré-soumission + BV (1.A.5), Agrément MIPH (1.A.6), Décision d'exercice du PDT (1.A.7), Contrats exploitation/façonnage (1.A.8), Historique DE et anciens dépôts (1.A.9).
  - **Sous-module 1.H (Holder) :** Courrier (1.H.1), AMM pays d'origine (1.H.2), Certificat CPP modèle OMS (1.H.3), Structure PCSU ou attestation FOB visée Chambre de Commerce (1.H.4), Benchmark prix international (1.H.5 - obligatoire import NI/RI/MI), Approbation pays d'origine (1.H.6), Descriptif mise à jour (1.H.7), Liste des sites intervenants visée (1.H.8), Déclaration de similarité (1.H.9), Déclaration de non-modification (1.H.10 - renouvellements RI/RF), Accord transfert DE (1.H.11), Date d'implémentation pays d'origine (1.H.12).
  - **Sous-module 1.M (Manufacturers) :** Agréments et BPF API (1.M.S.1/2), Lettre d'accès / CEP API (1.M.S.3), Agrément et BPF Produit Fini (1.M.P.1/2), Étude de bioéquivalence ou justification d'exonération Biowaiver (1.M.P.3).
  - **Sous-module 1.P (Product Information) :** Textes pays d'origine RCP/Notice/Étiquette (1.P.1.1 à 1.P.1.3), Modèle Algérie visé PDT avec notice bilingue FR/AR et maquettes couleurs échelle 100% (1.P.2.1 à 1.P.2.4), Échantillons (1.P.3), Bulletins SA avec granulométrie/polymorphisme et recontrôle (1.P.4), Bulletins PF et prémix (1.P.5), Attestation correspondance lots (1.P.6), Certificat libération lot (1.P.7), **Lettre d'engagement de non-commercialisation des lots de validation** (1.P.8 - obligatoire fabrication locale NF/MF), **Lettre d'engagement suivi stabilité PF en temps réel Zone IVa/IVb** (1.P.9 - obligatoire fabrication locale NF/MF).
- [x] **Règles de complétude et Spécificités Régionales des Modules 2 à 5 :**
  - Contrôle strict de présence des Modules 2 à 5 pour les Princeps (`E` sur l'ensemble des modules).
  - Dérogation automatisée pour les Génériques : Module 1 exigé (`E`), Module 2 incorporé, Module 3 exigé (`E`), Module 4 facultatif (`F`), Module 5 centré sur la Bioéquivalence (`P`).
  - **Exigences Module 3.2.S (Substance Active) :** Intégration du CEP EDQM avec lettre d'accès nominative, contrôle obligatoire du polymorphisme cristallin (3.2.S.1.3) et de la granulométrie (3.2.S.4).
  - **Exigences Module 3.2.P (Produit Fini) :** Dépôt obligatoire des dossiers de lots industriels de validation en 3.2.P.3.5 pour la fabrication locale (NF/MF), spécifications à libération et à péremption (3.2.P.5.1).
  - **Exigences Module 3.2.R (Documentation Régionale Algérienne) :**
    - `3.2.R.1.1` : Exécution des documents de production (Master Batch Records exécutés).
    - `3.2.R.1.2` : Documents de production principaux (Formules cadres de fabrication et conditionnement).
    - `3.2.R.2` : Procédures analytiques et informations de validation sur site.
  - **Exigences Module 5 (Données Cliniques et Bioéquivalence Génériques) :**
    - Étude comparative de Biodisponibilité (BA) et Bioéquivalence (BE) in vivo en 5.3.1.2 (intervalle de confiance 90 % compris entre 80,00 % et 125,00 % sur Cmax et AUC).
    - Validation complète des méthodes bioanalytiques humaines en 5.3.1.4 (BPL/GCP).
    - Rattachement obligatoire du dossier de justification d'exonération *Biowaiver* en cas d'absence d'étude in vivo (profils de dissolution comparatifs multipoints à pH 1.2, 4.5, 6.8 et classification biopharmaceutique BCS classe I/III en 1.M.P.3 et 5.3.1.2).
- [x] Questionnaire de qualification dynamique **Dispositifs Médicaux** (génération automatique de checklist selon le décret exécutif n° 20-324 et l'arrêté du 10 mai 2021).
- [x] Liaison dynamique GED ↔ Checklist CTD/DM avec calcul du pourcentage de complétude en temps réel.
- [x] **Verrouillage légal strict :** Interdiction de passage au statut *« Validé interne / Prêt pour dépôt ANPP »* si complétude < 100% ou absence de signature électronique légale du Pharmacien Directeur Technique.

### Jalon 1.6 : Interface Web Haute Densité 'Pharma Explorer' (3 Volets, Responsive & Thème Clair)
- [x] Architecture frontend embarquée via `//go:embed` (zéro serveur Node.js ni CDN externe en production).
- [x] Layout 'Pharma Explorer' à 3 volets façon Windows Explorer Pro avec splitters redimensionnables à la souris et mémorisation :
  - **Volet gauche (Tree Explorer) :** Navigation hiérarchique dépliable (les 5 modules CTD, rubriques 1.A/1.H/1.M/1.P, chapitres DM ou branches GED) avec chevrons `▶ / ▼`, badges de procédure et compteurs de complétude dynamiques (ex: `(9/9)`, `(0/28)`).
  - **Zone centrale (Vue détaillée haute densité) :** Tableau compact (lignes de 32px), groupement par sous-module sous forme d'accordéons pliables/dépliables avec ratios de complétude, champ de recherche instantanée en temps réel et barre de pills de filtrage rapide (`Tous`, `🔴 Manquants`, `🟢 Rattachés`, `⚠️ Obligatoires [E]`).
  - **Volet droit (Volet d'aperçu contextuel - Alt+P) à double onglet intelligent :**
    - Onglet `📄 Aperçu Pièce GED` : cartouche cryptographique d'intégrité CAS SHA-256, statut de validation, horodatage UTC, métadonnées BPF et bouton de téléchargement du fichier original.
    - Onglet `📋 Fiche ANPP` : fiche légale officielle issue du Guide CTD ANPP Juillet 2025, texte bilingue, niveau d'exigence (`[E]`, `[F]`, `[P]`) et traçabilité 21 CFR Part 11.
- [x] Mode Plein Écran « 🎯 Focus » permettant d'escamoter instantanément les volets latéraux pour maximiser l'espace du tableau de checklist.
- [x] Thème Clair Pharmaceutique activé par défaut (fond blanc pur/gris perle, bleu marine réglementaire) avec bascule instantanée Mode Sombre.
- [x] Architecture Responsive à 3 paliers opérationnelle :
  - Desktop (≥ 1200px) : 3 volets avec splitters ajustables.
  - Laptop / Tablette (768px - 1199px) : 2 volets fixes et volet droit d'aperçu en tiroir latéral coulissant (*slide-over* avec backdrop).
  - Mobile (< 768px) : Menu hamburger ☰, volet gauche en *drawer* coulissant, tableau adapté et volet d'aperçu escamotable.
- [x] Interface 100% en Français institutionnel avec préservation des textes juridiques bilingues FR/AR dans les fiches réglementaires ANPP.
- [x] Barre d'adresse fil d'Ariane navigable (`📁 DOSSIERS_ANPP > NI > DOS-2026-AMOX-NI > Amoxicilline 500mg`).
- [x] Barre d'état inférieure et indicateurs KPI de complétude avec jauges colorimétriques ALCOA+.

### Jalon 1.7 : Studio de Compléments ANPP, Assistant de Soumission E-Services & Veille RDV
- [x] **Gestion des Compléments et Notifications ANPP (Migration `010_anpp_deficiency_letters_and_tracking.sql`) :**
  - [x] Enregistrement des courriers officiels ANPP (date émission, date réception décharge bureau d'ordre, analyse du délai de latence postale, numéro de référence officiel, motif).
  - [x] Découpage des questions du courrier en sous-tâches unitaires par module CTD (Module 1, 2, 3, 4, 5, DM) avec assignation et priorité.
  - [x] Moteur de compte à rebours légal multi-niveaux (Décret exécutif 20-325, 90j par défaut, 60j ou date butoir personnalisée) :
    - [x] Bandeau d'urgence visuel et badges dynamiques dans l'application (J ≤ 15 Critique, J ≤ 30 Urgent, J ≤ 60 Alerte, Dans les délais).
    - [x] Modale de supervision des échéances critiques accessible depuis le bandeau supérieur de l'application.
    - [x] Visa légal régalien du Pharmacien Directeur Technique (1.A.7 / 21 CFR Part 11) avec scellement de chaque réponse dans l'audit trail BPF.
  - [x] **Numérisation, Versement & Visualisation In-App des Scans Officiels (GED ALCOA+) :**
    - [x] Archivage cryptographique au coffre-fort numérique GED (CAS SHA-256) sous l'arborescence `01_Affaires_Reglementaires/01_06_Correspondances_ANPP`.
    - [x] Colonne dédiée « Pièce Officielle GED » dans la table des compléments avec boutons d'accès rapide `👁️ Aperçu Scan ANPP` et `📎 Joindre Scan`.
    - [x] Visualiseur universel de documents in-app (`#modal-doc-preview`) avec affichage PDF intégré, bascule plein écran `⛶ Agrandir` et téléchargement `📥`.
    - [x] Habilitations RBAC étendues garantissant le téléversement et le remplacement sans friction des pièces officielles par le Pharmacien DT.
- [x] **Assistant de Soumission E-Services ANPP (E-TASJIL & RDV PP/DM) :**
  - [x] **Calculateur des redevances & Bordereaux officiels BV1 / BV2 (Migration `011_anpp_fee_schedules_and_payment_slips.sql`) :**
    - [x] Barèmes légaux configurables selon les procédures officielles ANPP (NI, RI, MI, NF, RF, MF, DM).
    - [x] Production automatique des bordereaux de versement **BV1** (Enregistrement initial / Homologation) et **BV2** (Modifications / Renouvellements) au format HTML imprimable conforme avec calcul automatique des montants en DZD et devises étrangères.
  - [x] **Veille calendaire et Assistant Prise de Rendez-vous (Migration `012_anpp_appointments_and_regulatory_watch.sql`) :**
    - [x] Alertes proactives avant ouverture des guichets : mardi à 09h50 pour les DM (`rdvdm.anpp.dz`), mercredi à 09h50 pour les Produits Pharmaceutiques (`rdvpp.anpp.dz`).
    - [x] Fast-Fill sécurisé préparant les données du dossier (raison sociale, agrément MIPH, PDT, produit, annexe Alger/Oran/Constantine) pour réservation immédiate du créneau.
    - [x] Corpus juridique et registre de veille réglementaire avec analyse d'impact et plan de mise en conformité par le Pharmacien DT.
  - [x] **Submission Package pour E-TASJIL (`etasjil.anpp.dz`) :**
    - [x] Audit de conformité pré-dépôt bloquant : interdiction de verrouillage si pièces obligatoires manquantes, certificats expirés ou absence de signature électronique du PDT (21 CFR Part 11).
    - [x] Génération des manifestes d'intégrité SHA-256 (JSON et XML) et scellement formel du dossier de soumission.
    - [x] Fiche récapitulative de télé-saisie E-TASJIL pour saisie rapide alignée champ par champ avec les masques du portail officiel.

### Jalon 1.8 : Packaging, Déploiement & Réversibilité (Socle Phase 1 100% Validé)
- [x] **Environnement Docker & Docker-Compose de Production :**
  - [x] `Dockerfile` multi-stage durci avec binaire statique Go CGO_ENABLED=0 (`v0.8.0`), image Alpine minimale, utilisateur non-root `homolog:homolog`, santé `/health`.
  - [x] `docker-compose.yml` orchestrant l'application et la base PostgreSQL 16 Alpine avec healthcheck `pg_isready` et volumes persistants.
  - [x] Fichier `.env.production` avec durcissement PSSI (désactivation stricte du commutateur dev, Cookie Secure forcé, clé JWT 64+ car., timeout 15 min BPF).
- [x] **Assistant de Premier Démarrage (Setup Wizard BPF 1.A.7) :**
  - [x] Détection automatique du statut vierge de l'instance (`GET /api/v1/setup/status`).
  - [x] Provisionnement initial de l'établissement MIPH et du compte Pharmacien DT titulaire (`POST /api/v1/setup/initialize`).
  - [x] Écran guidé d'initialisation in-app (`#screen-setup`) avec hachage Argon2id et journalisation dans la piste d'audit immuable.
  - [x] Garde-fou strict interdisant toute réinitialisation sur une instance déjà active (HTTP 409 Conflict).
- [x] **Archive Universelle Portative .homolog & Réversibilité ALCOA+ :**
  - [x] Spécification du conteneur d'archive universelle `HOMOLOG_PORTABLE_ARCHIVE_V1` (flux tar.gz autonome sans vendor lock-in).
  - [x] Moteur d'export (`GET /api/v1/system/backup/export`) : sérialisation JSON des 17 tables métier (`database/`), collecte des pièces GED physiques dédupliquées (`blobs/<sha256>`), manifeste signé (`manifest.json`) et sceau d'archive global SHA-256 (`X-Homolog-Seal`).
  - [x] Moteur de restauration certifié (`POST /api/v1/system/backup/restore`) : validation d'intégrité du manifeste, contrôle systématique du hash SHA-256 de chaque document CAS avec rejet immédiat en cas d'altération (HTTP 422), réinjection transactionnelle des 17 tables et scellement dans l'audit log.
  - [x] Interface cockpit IT (`web/js/system_admin.js`) : section « Souveraineté & Réversibilité », téléchargement direct et modale de restauration (`#modal-restore-homolog`) avec barre de progression.
- [x] **Validation de la réversibilité complète :**
  - [x] Suite de tests unitaires et d'intégration validée à 100% (`backup_archive_test.go` : 5 tests réussis).
  - [x] Phase 1 du socle fondateur achevée à 100% (8/8 jalons complétés).

---

## Phase 2 : Complétude, Moteur de Variations ANPP & Efficacité (Priorité S)
> **Objectif stratégique :** Intégrer l'atelier d'élaboration documentaire collaborative BPF (Maker-Checker-Approver), les modèles maîtres ANPP/ICH, le moteur expert de modifications post-enregistrement selon la ligne directrice ANPP LD-DEPP07/V01 (Avril 2026), la gestion des redevances, la régularisation des DM, la recherche plein texte et le multi-tenant SaaS.

### Jalon 2.1 : Atelier d'Élaboration Documentaire, Templates Maîtres & Workflow Multi-Rôles BPF (Maker-Checker-Approver)
- [ ] **Cartographie & Attribution Décentralisée des Responsabilités Rédactionnelles :**
  - Matrice d'affectation métier selon l'expertise (Section 7.4 du Cahier des Charges) :
    - *Affaires Réglementaires (RA)* : Module 1 complet, notices patient bilingues FR/AR (1.P.2), résumés des caractéristiques (RCP 1.P.1), maquettes, bordereaux BV1/BV2, lettres d'engagements (1.P.8, 1.P.9, 1.H.10) et dossiers de variations (LD-DEPP07/V01).
    - *Contrôle Qualité (CQ) & Stabilités* : Spécifications analytiques SA/PF (3.2.S.4, 3.2.P.5.1), protocoles et rapports de validation des méthodes d'analyse selon ICH Q2(R1) (3.2.P.5.3), bulletins d'analyse libératoires CoA (1.P.4, 1.P.5), et protocoles/rapports d'études de stabilité sous conditions climatiques Zone IVb (3.2.P.8).
    - *Recherche & Développement (R&D) / Galénique* : Justification de la formulation et développement pharmaceutique (3.2.P.2), profils de dissolution comparatifs ($f_1$, $f_2$) pour *Biowaiver* (1.M.P.3), et Résumé Global de la Qualité (QOS Module 2.3).
    - *Production Industrielle & Ingénierie* : Formules de fabrication unitaire et par lot (3.2.P.3.2), procédés et contrôles en cours IPC (3.2.P.3.3), protocoles et rapports de validation de procédé sur 3 lots industriels (3.2.P.3.5 / 1.P.8).
    - *Assurance Qualité (AQ)* : Rôle de relecture indépendante (*Checker*), procédures opératoires (SOP/POS), revue annuelle de la qualité (PQR), audits BPF et gestion des CAPA.
    - *Partenaires Externes (Tiers)* : Dépôt sécurisé des ASMF/DMF Substance Active (3.2.S) par les fabricants d'API et rapports d'essais cliniques/bioéquivalence (Module 5) par les CROs.
  - Assignation des sections CTD/DM à des contributeurs désignés avec dates butoirs et notifications automatiques de suivi.
- [ ] **Moteur de Workflow Rédactionnel à 5 États (Maker-Checker-Approver BPF) :**
  - Cycle de vie : `À Rédiger` ➔ `En Cours` ➔ `En Revue Qualité/Checker` ➔ `Prêt pour Visa DT` ➔ `Approuvé & Scellé`.
  - Mécanisme de verrouillage collaboratif (**Check-out / Check-in**) évitant les conflits d'édition simultanée.
  - Gestion granulaire du versionnage : révisions de travail mineures incrémentales (`v0.1`, `v0.2`, `v0.3`...) jusqu'à la révision majeure officielle scellée (`v1.0`) par le PDT.
  - Circuit de revue par les pairs / AQ (*Checker*) avec annotations, demandes d'ajustements motivées et journal d'audit complet.
- [ ] **Moteur de Dépendances Chronologiques, Chemin Critique & Déverrouillage Conditionnel (*Gating*) :**
  - Modélisation du graphe orienté acyclique (DAG) des dépendances techniques et réglementaires CTD/DM en 5 phases successives (Matières & Méthodes ➔ Lots pilotes & CoA ➔ Stabilités Zone IVb & Bioéquivalence ➔ QOS & Notice/RCP ➔ Engagements & Visa PDT).
  - Système de verrous bloquants prévenant les erreurs de rédaction prématurée (verrouillage de la notice 1.P.2 et du RCP 1.P.1 tant que les stabilités 3.2.P.8 ne sont pas conclues ; verrouillage du QOS 2.3 tant que le Module 3 n'est pas validé).
  - Horloge de suivi des épreuves temporelles incompressibles : suivi des jalons de stabilité Zone IVb (M0, M3, M6, M9, M12, M24) et alerte automatique à l'atteinte du seuil minimal légal de soumission (M6).
  - Mécanisme de propagation descendante des métadonnées critiques : injection automatique de la formule validée, de la durée de conservation et des conditions de stockage dans les modèles maîtres de Notice bilingue, RCP, étuis et formulaire 1.A.3.
- [ ] **Bibliothèque de Modèles Maîtres (*Master Templates*) conformes ANPP & ICH :**
  - Bibliothèque native de canevas paramétrables : Notice patient bilingue FR/AR (1.P.2), RCP (1.P.1), Spécifications et méthodes de contrôle produit fini (ICH Q6A / 3.2.P.5), Protocoles et rapports de stabilité Zone IVb (ICH Q1A(R2) / 3.2.P.8), Résumé global de la qualité (QOS Module 2.3), Rapport de validation de procédé (3.2.P.3.5), Dossier technique DM (Arrêté du 10 mai 2021).
  - Pré-remplissage dynamique des variables documentaires depuis la base de données (nom commercial, DCI, forme galénique, dosage, titulaire, fabricant, etc.).
- [ ] **Générateur Automatisé de Formulaires & Lettres d'Engagement en 1 Clic :**
  - Moteur de génération automatique instantanée de documents officiels au format PDF/A prêt pour visa PDT :
    - Lettre d'engagement de validation sur 3 premiers lots industriels de fabrication locale (1.P.8).
    - Lettre d'engagement de suivi de stabilité continue post-commercialisation en Zone IVb (1.P.9).
    - Attestation d'évaluation et de conformité par le Pharmacien Directeur Technique (1.H.10).
    - Attestation de non-modification des autres sections pour dossiers de variations (LD-DEPP07/V01).
- [ ] **Moteur Universel de Restitution & des 15 États de Sortie BPF / ANPP (Chapitre 10 du Cahier des Charges) :**
  - Architecture du moteur backend Go (`ReportingService` dans `internal/reporting`) et génération haute fidélité au format **PDF/A-1b et PDF/A-2b** (ISO 19005).
  - Système de **Filigranes Dynamiques BPF** selon l'état de cycle de vie : `BROUILLON (DRAFT)` (translucide), `OFFICIEL / APPROUVÉ` (vert BPF avec visa PDT), `ARCHIVÉ / PÉRIMÉ` (rouge).
  - Dispositif anti-falsification par **QR-Code cryptographique** en pied de page renvoyant vers l'URL de vérification instantanée (`https://homolog.labo.dz/verify/{sha256}`) pour contrôle mobile immédiat par l'ANPP ou auditeurs.
  - Paginage légal indélébile **`Page X sur Y`** et cartouche d'identification normalisé (raison sociale, agrément MIPH, DCI, version, horodatage UTC+1 certifié).
  - **Couverture intégrale des 5 familles d'états de sortie (15 états réglementaires officiels) :**
    - *Famille 1 (Dépôt & Procédures CTD) :* ES-REG-01 (Bordereau de transmission 5 modules), ES-REG-02 (Formulaire 1.A.3), ES-REG-03 (Index et table des matières CTD avec SHA-256), ES-REG-04 (Fiche de variation LD-DEPP07/V01), ES-REG-05 (Bordereau d'envoi échantillons LNCPP).
    - *Famille 2 (Redevances & Financier Décret 21-82) :* ES-FIN-01 (Bordereau BV1 initial/DM), ES-FIN-02 (Bordereau BV2 variations/renouvellements avec décompte Grouping), ES-FIN-03 (Fiche d'imputation et rapprochement quittance bancaire).
    - *Famille 3 (Vigilance & Compléments) :* ES-VIG-01 (Bordereau de réponse aux compléments avec compte à rebours 60j/90j), ES-VIG-02 (Fiche de préparation RDV ANPP), ES-VIG-03 (Rapport périodique de veille et échéancier AMM J-180/90/30).
    - *Famille 4 (Qualité & Audit Trail BPF) :* ES-QUA-01 (Registre d'Audit Trail scellé ALCOA+), ES-QUA-02 (Manifeste de signature électronique Maker-Checker), ES-QUA-03 (Fiche de contrôle documentaire pré-dépôt).
    - *Famille 5 (Pilotage & Réversibilité) :* ES-DIR-01 (Dashboard exécutif du portefeuille AMM/DM), ES-EXP-01 (Procès-verbal de réversibilité archive universelle `.homolog`).
  - Exports tabulaires certifiés **Excel (XLSX) et CSV UTF-8** avec métadonnées d'audit trail (pour comptabilité, finance et analyses de portefeuille).
  - Routes REST API sécurisées (`GET /api/v1/dossiers/{id}/reports/{code}`, `POST .../seal`, `GET /api/v1/public/verify/{sha256}`).
- [ ] **Cockpit de Supervision & Parapheur Électronique pour le Pharmacien Directeur Technique :**
  - Tableau de bord de pilotage temps réel : vue matricielle d'avancement par module CTD et par département rédacteur, alertes sur retards et goulots d'étranglement.
  - Parapheur électronique consolidé pour les pièces au statut *Prêt pour Visa DT*, avec comparaison visuelle des versions (*diff*) et rapport du *Checker*.
  - Signature électronique par lot (*Batch Signing*) avec validation TOTP unique, audit trail individuel 21 CFR Part 11 et scellement cryptographique SHA-256.

### Jalon 2.2 : Suivi des Redevances Réglementaires & Devises
- [ ] Barèmes configurables des redevances ANPP selon la procédure (**NI, RI, MI, NF, RF, MF**) et la classe de produit.
- [ ] Enregistrement et contrôle des bordereaux de versement (1.A.4), quittances bancaires et attestations de taux de change Banque d'Algérie.
- [ ] Suivi des frais de pré-soumission (1.A.5) et tableau de bord consolidé des coûts d'enregistrement par produit et par an.

### Jalon 2.3 : Moteur de Variations ANPP & Reclassement Automatique (Ligne Directrice LD-DEPP07/V01)
- [ ] **Implémentation du référentiel officiel algérien des variations :**
  - **Mmin (Modification Mineure) :** Impact minime ou nul.
  - **Mmod (Modification Modérée) :** Impact potentiel sur qualité, sécurité, efficacité.
  - **Mmaj (Modification Majeure) :** Répercussions significatives.
  *(Remplacement définitif de la nomenclature européenne Type IA/IB/II).*
- [ ] **Catalogue interactif des 34 rubriques officielles de modifications :**
  - **Rubriques administratives (1 à 5) :** Dénomination commerciale PF (1), Dénomination SA/excipient/conditionnement (2), Code ATC OMS (3), Nom/adresse titulaire AMM/DPSA/fabricant/conditionneur/contrôleur (4), Suppression de sites (5).
  - **Rubriques CEP et préqualification API (6.a et 6.b) :** Fabricant approuvé ou nouveau fabricant (Mmin ou Mmod).
  - **Rubriques qualitatives Substance Active 3.2.S (7 à 16) :** Nouveau site/fabricant API (7 - Mmin/Mmod/Mmaj), Nouveau bloc/unité (8.a - Mmin), Procédé API (9 - mineur Mmod, important Mmaj), Taille de lot API (10 - Mmin/Mmod), Contrôles en cours API (11), Spécifications API (12.a-f), Procédures d'essai API (13.a-e), Conditionnement primaire SA (14), Durée de vie/recontrôle API (15), Conditions de stockage API (16).
  - **Rubriques qualitatives Produit Fini 3.2.P (17 à 34) :** Système d'aromatisation/coloration (17), Masse enrobage/capsules (18.a), Composition formes solides (19 - Mmod/Mmaj), Site de fabrication PF (20.a conditionnement secondaire Mmin, 20.b conditionnement primaire Mmin, 20.c fabrication conventionnelle Mmod), Site de contrôle PF (21.a - Mmin), Taille de lot PF (22 - x10 Mmin, >x10 Mmod), Procédé de fabrication PF (23 - Mmod), Contrôles en cours PF (24), Remplacement excipient risque EST par végétal/synthétique (25 - Mmin), Spécifications/méthodes excipient (26), Monographie pharmacopée PF (27 - Mmin), Spécifications PF (28), Procédures analytiques PF (29), Type de contenant primaire (30 - non stérile Mmod, stérile Mmaj), Composition emballage primaire (31), Durée de conservation PF (32 - Mmod), Durée après première ouverture/reconstitution (33 - Mmod), Conditions de stockage PF (34 - Mmod).
- [ ] **Moteur d'évaluation des "Conditions à remplir" et Reclassement Automatique (Escalade) :**
  - Questionnaire guidé pour chaque modification sélectionnée vérifiant les critères techniques.
  - **Algorithme d'escalade :** Dès qu'une condition n'est pas satisfaite, le système reclasse immédiatement la modification (Mmin → Mmod ou Mmaj, et Mmod → Mmaj) et recalcule dynamiquement la liste des pièces justificatives obligatoires.
- [ ] **Générateur du Dossier de Modification ANPP (Section 5.3 LD-DEPP07/V01) :**
  - Production automatique du **Tableau comparatif standardisé** (*État actuel approuvé vs Situation proposée*).
  - Production automatique du modèle officiel d'**Attestation formelle de non-modification des autres sections**.
  - Contrôle et agrégation des **12 pièces obligatoires exigées pour toute variation (Section 5.3 du CDC)** :
    1. Lettre de demande officielle visée par le PDT.
    2. Tableau comparatif standardisé approuvé vs proposé.
    3. Justificatif technico-réglementaire motivant le changement.
    4. Bordereau de versement de la redevance ANPP (+ quittance bancaire).
    5. Sections modifiées du dossier CTD (remplacement page à page).
    6. Liste actualisée et visée de tous les intervenants (SA, PF, conditionnement, contrôle).
    7. Attestation formelle de non-modification des autres sections.
    8. Copie de la décision d’enregistrement (DE) en vigueur.
    9. Accusés de dépôt des modifications et renouvellements en cours d'instruction.
    10. Approbation de la modification dans le pays d'origine (produits importés).
    11. Autorisations des sites et certificats BPF/BPL en vigueur.
    12. Informations "produit" révisées (RCP, notice bilingue FR/AR, étiquetage et maquettes couleurs visées PDT).
- [ ] **Module de Regroupement de Modifications (*Grouping* - Article 09 de l'arrêté du 07/11/2021) :**
  - Création d'une soumission groupée pour un même produit combinant plusieurs modifications interdépendantes.
  - Calcul du niveau de criticité global du groupe (alignement sur la modification la plus sévère du groupe).

### Jalon 2.4 : Espace Sécurisé pour Consultants Externes
- [ ] Cloisonnement étanche : accès restreint exclusivement aux seuls dossiers CTD/DM explicitement assignés.
- [ ] Zéro accès à la GED transverse ou aux autres produits du laboratoire.
- [ ] Date d'expiration de contrat automatique révoquant les accès à l'échéance.
- [ ] Journalisation et marquage d'audit renforcé sur chaque action de consultation ou dépôt par un consultant.

### Jalon 2.5 : Procédure de Régularisation des Dispositifs Médicaux (Décret 20-324)
- [ ] Inventaire des DM déjà mis sur le marché sans décision d'homologation.
- [ ] Checklist spécifique de régularisation selon l'arrêté du 10 mai 2021.
- [ ] Tableau de suivi des mesures transitoires et des échéances de dépôt de régularisation.

### Jalon 2.6 : Moteur de Recherche Plein Texte Bilingue (PostgreSQL FTS)
- [ ] Indexation textuelle automatique des documents PDF, Word (.docx) et Excel (.xlsx) déposés en GED.
- [ ] Moteur de recherche plein texte PostgreSQL (`tsvector`/`tsquery`) supportant les requêtes en français et en arabe.
- [ ] Surlignage dynamique des termes trouvés directement dans la vue tableau et le volet de prévisualisation.

### Jalon 2.7 : Console Super-Admin Éditeur (Mode Cloud SaaS)
- [ ] Interface d'administration multi-laboratoires et suivi du parc d'instances.
- [ ] Provisioning d'un laboratoire en un clic avec schéma PostgreSQL dédié et sous-domaine isolé (`<labo>.homolog.dz`).
- [ ] Gestion des quotas de stockage par laboratoire et suivi de consommation.

### Jalon 2.8 : Mises à Jour Réglementaires Hors-Ligne (.pkg Signés)
- [ ] Format de paquets de mise à jour réglementaire signés cryptographiquement (`.pkg`).
- [ ] Déploiement sans redémarrage des nouveaux modèles de checklists CTD ou fiches de variations émises par l'ANPP.
- [ ] Vue comparative (Diff) entre le modèle de dossier actuel et le nouveau modèle avant application validée par l'administrateur.

---

## Phase 3 : Extensions & Écosystème (Priorité C)
> **Objectif :** Étendre la plateforme à la surveillance post-marché, aux interconnexions d'entreprise et à l'interopérabilité directe avec l'ANPP.

### Jalon 3.1 : Registre de Vigilance Post-Marché
- [ ] Module de Pharmacovigilance (suivi des PSUR, déclarations d'effets indésirables, fiches de notification officielles).
- [ ] Module de Matériovigilance des DM (déclaration d'incidents, signalements de défaillances de dispositifs).
- [ ] Suivi des enquêtes de lot et gestion des rappels/retraits ordonnés par l'ANPP ou initiés par le laboratoire.

### Jalon 3.2 : Intégrations d'Entreprise
- [ ] Authentification unique SSO d'entreprise (SAML 2.0 / OpenID Connect) et synchronisation d'annuaire LDAP / Active Directory.
- [ ] Flux de calendrier réglementaire iCal (synchronisation Outlook/Google Calendar des dates d'échéance ANPP et des renouvellements).

### Jalon 3.3 : Passerelle Officielle API ANPP (Télétransmission Directe M2M)
- [ ] Connecteur d'interopérabilité API (`internal/anpp`) dès publication des spécifications techniques de télé-procédure ou webhook par l'ANPP.
- [ ] Télétransmission automatisée sécurisée des paquets de soumission CTD scellés (Submission Packages) avec accusé de réception cryptographique.
- [ ] Synchronisation bidirectionnelle du statut d'instruction en temps réel entre E-TASJIL et Homolog.

---

## Guide d'avancement et métriques de suivi

| Indicateur clé | Objectif Phase 1 | État actuel | Conformité réglementaire |
|---|---|---|---|
| **Spécifications fonctionnelles validées** | 100 % | 100 % ([`cahier_des_charges.md`](file:///c:/dev/homolog/docs/cahier_des_charges.md) v1.0 - 11 Domaines) | Aligné Guide CTD ANPP 2025, LD Modifications 2026 & Chapitre 10 États de Sortie ✅ |
| **Jalons Phase 1 complétés** | 8 / 8 jalons | 8 / 8 jalons complétés (100 %) (Jalons 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7 & 1.8 ✅) | Phase 1 validée à 100% : socle souverain, GED BPF, compléments ANPP, BV1/BV2 et réversibilité .homolog ✅ |
| **Couverture des tests automatisés** | > 80 % sur le domaine métier & audit | Tests unitaires stockage CAS, Auth, Produits & Réversibilité validés (100%) | Argon2id, TOTP, Sessions, CAS, RBAC & Backup conformes BPF ✅ |
| **Conformité BPF / ALCOA+** | Validée sur audit trail, GED et signatures | Implémentée (Triggers immuabilité + CAS SHA-256 + Sceau d'archive) | Scellement cryptographique et rôles validés ✅ |
| **Nomenclature des Variations** | Conforme LD-DEPP07/V01 | Modélisée (34 rubriques Mmin/Mmod/Mmaj) | Reclassement dynamique & Grouping intégrés ✅ |
| **États de Sortie & Reporting BPF** | Conforme Chapitre 10 CDC & Décret 21-82 | 15 états spécifiés (5 familles métier, QR-Code d'intégrité, PDF/A, XLSX/CSV) | Valeur probante, anti-falsification et paginage « Page X sur Y » conformes BPF ✅ |
