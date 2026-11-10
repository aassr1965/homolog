# Cahier des charges fonctionnel & technique détaillé
## Plateforme collaborative de gestion de l'enregistrement des médicaments, de l'homologation des dispositifs médicaux et GED pharmaceutique en Algérie

**Version 0.7 (Validé : Ergonomie 'Pharma Explorer' & Haute Densité Windows)** — 10 novembre 2026.  
*Document de référence intégrant le socle réglementaire ANPP/MIPH, la GED Qualité & Réglementaire BPF, l'architecture Go / PostgreSQL 16+, l'authentification 21 CFR Part 11, le déploiement hybride On-Premise / SaaS souverain, et l'ergonomie visuelle dense inspirée de Windows Explorer.*

---

## 1. Contexte et objectifs

### 1.1 Contexte
En Algérie, la mise sur le marché d'un médicament à usage humain passe par une décision d'enregistrement, et celle d'un dispositif médical (DM) par une décision d'homologation, toutes deux instruites par l'Agence nationale des produits pharmaceutiques (ANPP). Pour un laboratoire, ces démarches impliquent de nombreux documents, plusieurs intervenants internes et externes, des paiements, des échanges avec l'agence et des échéances de renouvellement. 

Parallèlement, la gestion documentaire interne des laboratoires (certificats BPF, dossiers de fabrication, méthodes d'analyse, spécifications, étiquetages) est souvent dispersée sur des partages réseau sans traçabilité, exposant à des risques majeurs d'obsolescence et de non-conformité.

### 1.2 Objectifs
- **Centraliser** le suivi de tous les dossiers d'enregistrement et d'homologation d'un laboratoire.
- **Intégrer une GED pharmaceutique native** (Gestion Électronique de Documents) assurant le cycle de vie, la signature électronique et l'intégrité (ALCOA+) des documents de référence.
- **Lier dynamiquement la GED aux dossiers ANPP** : une pièce validée en GED alimente automatiquement les dossiers d'enregistrement (CTD) ou d'homologation DM.
- **Garantir la traçabilité complète et l'audit trail** (qui a fait quoi, quand, sur quelle version, avec quel motif).
- **Anticiper les échéances critiques** (renouvellements, réponses aux compléments ANPP, redevances, expiration de certificats BPF/ISO).
- **Réduire les erreurs de complétude** des dossiers grâce à des modèles de dossiers configurables sans modifier le code.
- **Offrir une flexibilité totale de déploiement** : utilisable sur site en local (100% hors-ligne) ou dans le Cloud sous forme de SaaS sécurisé avec réversibilité totale.
- **Offrir une ergonomie visuelle haute densité inspirée de Windows Explorer** : maximiser l'information visible sans scroll superflu, navigation par arbre à gauche, vue détaillée centrale et volet de prévisualisation immédiat à droite.

---

## 2. Décisions de conception validées (Matrice de Cadrage)

| Domaine | Arbitrage validé | Justification & Impact |
|---|---|---|
| **Ergonomie & UI** | **Modèle 'Pharma Explorer' à 3 volets (façon Windows Explorer)** | Arbre de navigation à gauche, vue détaillée compacte (lignes de 32px, 25-30 docs visibles) au centre, et volet de prévisualisation PDF à droite (Alt+P) avec barre d'adresse breadcrumb. |
| **Authentification & Sessions** | **Argon2id + Cookies HttpOnly révocables** | Sessions protégées contre le vol XSS (pas de token en `localStorage`), révocation instantanée côté serveur en base PostgreSQL, déconnexion auto sur inactivité (15 min). |
| **Second Facteur (2FA)** | **TOTP RFC 6238 100% hors-ligne** | Clé secrète et QR code local pour Google Authenticator/FreeOTP. Fonctionne sans internet, sans réseau mobile et sans passerelle SMS. 10 codes de secours chiffrés. |
| **Signature Électronique BPF** | **Re-authentification stricte (21 CFR Part 11)** | Re-saisie obligatoire du mot de passe ou 2FA avec déclaration solennelle du rôle (*« En qualité de Directeur Technique... »*) et scellement d'intégrité SHA-256. |
| **Dossiers Médicaments** | **Modèle CTD standardisé (Modules 1 à 5)** adapté à l'ANPP | Module 1 spécifique Algérie (agrément, prix, notices arabisées), Modules 2 à 5 pour qualité, sécurité et efficacité. Standard international et local. |
| **Dossiers Dispositifs Médicaux** | **Questionnaire de qualification dynamique** | La saisie de la classe (I, IIa, IIb, III, DMDIV), de la stérilité et de l'usage génère automatiquement la checklist légale exacte de l'arrêté du 10 mai 2021. |
| **Validation Directeur Technique (DT)** | **Verrouillage légal strict (BPF/ANPP)** | Blocage strict du passage au statut *« Prêt pour dépôt / Validé interne »* tant que 100% des pièces ne sont pas validées et signées électroniquement par le pharmacien DT. |
| **Compléments ANPP** | **Studio de traitement des compléments unitaire** | Découpage de chaque question du courrier officiel en tâche unitaire avec responsable dédié, compte à rebours légal et validation DT avant transmission. |
| **Impact GED $\leftrightarrow$ Variations** | **Matrice d'impact guidée** | Dès qu'un document GED est révisé en version majeure, détection des produits enregistrés liés et suggestion guidée d'ouverture d'un dossier de modification (variation). |
| **Moteur d'échéances & Alertes** | **Système multi-niveaux avec escalade** | Alertes in-app + bandeau rouge d'urgence ($J < 30$j) + e-mails SMTP (J-180, J-90, J-30, J-7) + escalade automatique au DT si une échéance à J-30 est orpheline. |
| **Consultants externes** | **Cloisonnement étanche & Expiration automatique** | Accès restreint exclusivement aux dossiers explicitement partagés (zéro accès à la GED globale), avec date d'expiration de contrat et audit trail renforcé. |
| **Bilinguisme FR / AR** | **Bilinguisme intégral (UI + Données) RTL natif** | Saisie bilingue des dénominations, notices et références, inversion dynamique RTL, police locale `Cairo` embarquée, rendu PDF arabe parfait. |
| **Évolutions réglementaires** | **Paquets signés (.pkg / YAML) hors-ligne** | Import par l'admin sans développeur ; les dossiers en cours gardent leur modèle avec comparaison visuelle (diff) avant migration validée par l'humain. |
| **Sauvegarde & Réversibilité** | **Archive universelle chiffrée `.homolog`** | Dump PostgreSQL + blobs GED dédupliqués + hash SHA-256. Restaurable en 1 clic Cloud $\leftrightarrow$ On-Premise avec contrôle d'intégrité cryptographique. |
| **Priorité Phase 1 (Socle MVP)** | **Parcours 360° complet** | Tableau de bord (alertes J-180/J-90) $\to$ Checklist dossier dynamique (CTD/DM) $\to$ Coffre-fort GED avec prévisualisation PDF et signature électronique DT. |

---

## 3. Architecture technique, Ergonomie & Déploiement

### 3.1 Stack technologique unifiée ("One Codebase")
- **Backend & Serveur applicatif** : **Golang (Go 1.26+)**
  - Performance native, empreinte mémoire minimale (~40 Mo de RAM au repos), démarrage instantané (< 100 ms).
  - Gestion du streaming de fichiers volumineux par chunks de 5 Mo avec reprise automatique.
  - Driver haute performance `jackc/pgx/v5` avec pool de connexions asynchrone.
- **Frontend embarqué (`//go:embed`)** :
  - Développé en TypeScript / Tailwind CSS / Lucide Icons.
  - Assets compilés embarqués directement dans le binaire Go (`homolog`).
  - **Zéro serveur Node.js en production.**
- **Base de données** : **PostgreSQL 16+**
  - Support natif des champs `JSONB` indexés pour les modèles de dossiers configurables.
  - Recherche plein texte intégrée (`tsvector` / `tsquery`).
  - Table du Journal d'Audit immuable strictement *append-only*.
- **Stockage abstrait (`StorageProvider`)** :
  - On-Premise : stockage local direct chiffré sur disque (CAS SHA-256 avec déduplication).
  - SaaS : stockage local ou Object Storage S3.

### 3.2 Double mode de déploiement (Hybride)
```
[DEPLOYMENT_MODE=on-premise]
  ├── Instance locale du laboratoire (100% hors-ligne)
  ├── Stockage local sur disque chiffré (/data/storage)
  └── SMTP local du laboratoire

[DEPLOYMENT_MODE=saas]
  ├── Datacenter souverain multi-tenant
  ├── Sous-domaines isolés : https://<labo>.homolog.dz
  ├── Schéma PostgreSQL dédié par laboratoire (Isolation étanche)
  └── Console Super-Admin Éditeur (provisioning & quotas)
```

### 3.3 Stratégie d'authentification et sécurité (Conforme BPF / 21 CFR Part 11)
- **Identités nominatives et hachage Argon2id** : Comptes individuels stricts (interdiction des comptes génériques). Mots de passe hachés avec Argon2id, politique de complexité paramétrable (min. 12 caractères).
- **Double facteur (2FA) TOTP hors-ligne (RFC 6238)** :
  - Fonctionne entièrement hors-ligne grâce à une clé secrète générée localement et affichée sous forme de QR Code (compatible Google Authenticator, Microsoft Authenticator, FreeOTP).
  - Génération de 10 codes de secours à usage unique chiffrés en base de données.
- **Gestion des sessions par Cookies HttpOnly sécurisés** :
  - Cookies marqués `HttpOnly`, `Secure` et `SameSite=Lax/Strict`, immunisés contre les vols de tokens par attaques XSS.
  - Sessions enregistrées en base PostgreSQL permettant la **révocation instantanée** de session en cas de départ ou de suspension d'un utilisateur.
  - Verrouillage automatique de session après inactivité prolongée (Idle timeout paramétrable, 15 min par défaut pour respect des BPF).
- **Protection Anti-Brute Force** :
  - Verrouillage temporaire du compte (15 minutes) après 5 échecs consécutifs d'authentification.
  - Journalisation de sécurité immédiate (succès, échecs, IP, User-Agent) dans la table `audit_logs`.
- **Re-authentification pour signature électronique légale** :
  - Toute approbation formelle en GED ou validation finale d'un dossier ANPP requiert une re-saisie obligatoire du mot de passe ou code 2FA avec déclaration d'intention légale enregistrée dans l'audit trail.

### 3.4 Ergonomie visuelle : Le modèle 'Pharma Explorer' haute densité
- **Structure en 3 volets inspirée de Windows Explorer Pro** :
  1. **Volet Gauche (Navigation Tree)** : Arborescence dépliable avec chevrons (les 5 modules CTD, les chapitres de l'arrêté DM ou les 5 branches GED), intégrant des compteurs compacts de complétude (ex: `(4/4)` en vert, `(2/5)` en rouge).
  2. **Zone Centrale (Details View)** : Tableau dense compact (hauteur de ligne **32px**, typographie 12-13px) affichant 25 à 30 documents à l'écran sans défilement, avec colonnes triables (*Statut, Nom, Version, Échéance, Rédacteur, Validé DT, Hash SHA-256*).
  3. **Volet Droit (Preview Pane)** : Volet de prévisualisation instantanée (PDF.js) redimensionnable et masquable (`Alt+P`), permettant de lire le document sélectionné et d'apposer la signature électronique sans quitter la liste.
- **Barre d'outils et d'adresse (Breadcrumb & Search)** :
  - Chemin navigable façon Windows (`📁 GED > 01_Qualite > Certificats_BPF > ...`).
  - Barre d'outils compacte (*Nouveau document, Téléverser Drag & Drop, Exporter ZIP*).
  - Champ de filtrage instantané en haut à droite filtrant la liste en temps réel.
- **Barre d'état inférieure (Status Bar)** :
  - Affiche le nombre d'éléments, le nombre de pièces obligatoires manquantes, et l'espace disque consommé par le laboratoire.
- **Adaptation RTL Arabe** :
  - Inversion naturelle du layout : l'arbre de navigation bascule à droite, le tableau au centre, et le volet de prévisualisation à gauche.

---

## 4. Matrice des Rôles et Permissions

1. **Super-Administrateur Éditeur (SaaS)** : Gestion des instances labos, quotas de stockage.
2. **Administrateur Système Local (On-Premise)** : Maintenance serveur, certificats TLS, sauvegardes, sans accès aux dossiers.
3. **Administrateur de Laboratoire** : Paramétrage du laboratoire, gestion des comptes internes.
4. **Responsable Affaires Réglementaires** : Création et gestion des dossiers CTD/DM, coordination des pièces, suivi ANPP.
5. **Directeur Technique (Pharmacien DT)** : Approbation finale bloquante des dossiers, signature électronique légale, consultation globale.
6. **Contributeur Qualité / R&D / Production** : Dépôt de documents techniques en GED, réalisation des tâches de compléments.
7. **Consultant Externe** : Accès limité aux seuls dossiers partagés, date d'expiration automatique.
8. **Lecteur** : Consultation en lecture seule (Direction générale).
9. **Auditeur** : Consultation du Journal d'Audit immuable, aucun droit d'écriture.

---

## 5. Taxonomie de la GED Pharmaceutique (Standard BPF)

1. **`01_Affaires_Reglementaires`**
   - `01_01_Agrements_Etablissement` (Agréments MIPH)
   - `01_02_Decisions_Enregistrement` (Décisions médicaments en vigueur)
   - `01_03_Decisions_Homologation` (Décisions DM en vigueur)
   - `01_04_Dossiers_CTD_Reference` (Modules 1 à 5 maîtres)
   - `01_05_Notices_Articles_Conditionnement` (Notices arabisées, maquettes étuis)
   - `01_06_Correspondances_ANPP` (Courriers officiels et récépissés)
2. **`02_Assurance_Qualite_BPF`**
   - `02_01_Manuel_Qualite_Procedures` (POS / SOP en vigueur)
   - `02_02_Certificats_BPF_ISO` (Certificats GMP, ISO 13485)
   - `02_03_Rapports_Audit_Inspections` (Audits internes et ANPP)
   - `02_04_CAPA_Deviations`
3. **`03_Developpement_Production`**
   - `03_01_Formules_Fabrication_DMF`
   - `03_02_Methodes_Controle_Validation`
   - `03_03_Etudes_Stabilite`
   - `03_04_Specifications_Matieres_PF`
   - `03_05_Dossiers_Techniques_DM`
4. **`04_Vigilance_Post_Market`**
   - `04_01_Pharmacovigilance` (PSUR, rapports périodiques)
   - `04_02_Materiovigilance` (Signalements incidents DM)
   - `04_03_Rappels_Lots`
5. **`05_Contrats_Tiers`**
   - `05_01_Accords_Qualite_QTA`
   - `05_02_Contrats_Faconnage_SousTraitance`
   - `05_03_Certificats_Fournisseurs_APIs`

---

## 6. Prochaines étapes

La phase d'alignement et de cadrage est **100 % finalisée**. L'ensemble des exigences, choix d'architecture, règles métier et flux de validation sont validés et documentés. Le projet est prêt pour le développement de la Phase 1.
