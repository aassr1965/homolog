# PROTOCOLE DE QUALIFICATION OPÉRATIONNELLE ET DE PERFORMANCE (OQ / PQ)
## CAHIER DE TESTS DE RECETTE UTILISATEUR & DE CONFORMITÉ 21 CFR PART 11 / GAMP 5
### SYSTÈME INFORMATISÉ PHARMACEUTIQUE : SANAD PHARMA (v2.5)

---

| **Document ID** | **VAL-SANAD-OQ-PQ-2026-001** |
| :--- | :--- |
| **Statut** | **Approuvé pour Exécution (Ready for UAT Execution)** |
| **Version** | **2.5.0** |
| **Date d'Émission** | **24 Novembre 2026** |
| **Destinataire** | **Prestataire Externe de Validation Logicielle / Auditeur Tiers** |
| **Référentiels Réglementaires** | **FDA 21 CFR Part 11, EU GMP Annexe 11, GAMP 5 Catégorie 4/5, Décrets ANPP / MIPH (Algérie)** |
| **Principe d'Intégrité des Données**| **ALCOA+ (Attributable, Legible, Contemporaneous, Original, Accurate + Complete, Consistent, Enduring, Available)** |

> [!IMPORTANT]
> **Directive de maintenance documentaire permanente :**
> Conformément au plan d'assurance qualité logicielle (GxP / CSV), ce protocole de validation doit être **obligatoirement et systématiquement mis à jour** lors de chaque ajout fonctionnel, refonte d'écran (UI), modification de workflow ou évolution réglementaire. Aucune version logicielle ne peut être déclarée « Qualifiée pour Usage en Production » sans alignement préalable de ce cahier de recette.

---

## SOMMAIRE EXÉCUTIF

1. [Objet & Périmètre de la Qualification](#1-objet--périmètre-de-la-qualification)
2. [Guide d'Utilisation à l'Attention du Prestataire Externe](#2-guide-dutilisation-à-lattention-du-prestataire-externe)
3. [Environnement de Test, Matrice des Comptes & Rôles BPF](#3-environnement-de-test-matrice-des-comptes--rôles-bpf)
4. [Procédures Détaillées de Tests Opérationnels (Step-by-Step Scripts)](#4-procédures-détaillées-de-tests-opérationnels)
   - **Module 01 :** [Sas de Sécurité (Gate) & Authentification Multi-Rôles](#module-01--sas-de-sécurité-gate--authentification-multi-rôles)
   - **Module 02 :** [Gestion des Sessions & Timeout d'Inactivité](#module-02--gestion-des-sessions--timeout-dinactivité)
   - **Module 03 :** [Dossiers Pharmaceutiques CTD / eCTD (Modules 1 à 5)](#module-03--dossiers-pharmaceutiques-ctd--ectd-modules-1-à-5)
   - **Module 04 :** [Workflow de Double Contrôle Maker-Checker & Non-Auto-Approbation](#module-04--workflow-de-double-contrôle-maker-checker--non-auto-approbation)
   - **Module 05 :** [Lettres d'Observation ANPP (Deficiency Letters) & Alertes Délais](#module-05--lettres-dobservation-anpp-deficiency-letters--alertes-délais)
   - **Module 06 :** [Barème des Redevances & Bordereaux de Paiement ANPP](#module-06--barème-des-redevances--bordereaux-de-paiement-anpp)
   - **Module 07 :** [Homologation & Régularisation des Dispositifs Médicaux (DM)](#module-07--homologation--régularisation-des-dispositifs-médicaux-dm)
   - **Module 08 :** [Fabricants Étrangers, CPP OMS & Secret Industriel (Guest / Black-Box)](#module-08--fabricants-étrangers-cpp-oms--secret-industriel-guest--black-box)
   - **Module 09 :** [Agréments MIPH / ANPP & Télésoumission Tabadol](#module-09--agréments-miph--anpp--télésoumission-tabadol)
   - **Module 10 :** [Audit Trail ALCOA+, Flux Live (3s), Pause/Reprendre, Géolocalisation Haute Précision & Chaînage SHA-256](#module-10--audit-trail-alcoa-flux-live-3s-pausereprendre-géolocalisation-haute-précision--chaînage-sha-256)
   - **Module 11 :** [Signature Électronique 21 CFR Part 11 & Cartouche Visuelle](#module-11--signature-électronique-21-cfr-part-11--cartouche-visuelle)
5. [Gestion des Anomalies, Déviations & Fiches d'Incident](#5-gestion-des-anomalies-déviations--fiches-dincident)
6. [Rapport de Clôture & Procès-Verbal de Qualification GxP](#6-rapport-de-clôture--procès-verbal-de-qualification-gxp)

---

## 1. OBJET & PÉRIMÈTRE DE LA QUALIFICATION

Le présent document formalise le protocole d'essais exhaustif permettant à un auditeur ou prestataire externe indépendant de vérifier et d'attester la conformité de l'application **SANAD PHARMA** aux spécifications fonctionnelles, aux exigences réglementaires de l'ANPP et aux règles d'intégrité des données (**21 CFR Part 11 / GAMP 5**).

Le périmètre couvre l'intégralité du cycle de vie des enregistrements électroniques :
* La protection des accès et la séparation stricte des fonctions pharmaceutiques (SoD / RBAC).
* L'immuabilité et la traçabilité intégrale des opérations (Audit Trail horodaté avec géolocalisation haute précision et chaînage cryptographique).
* La signature électronique ayant valeur juridique probante.
* L'absence d'altération, de falsification ou d'écrasement silencieux des données sensibles de santé.

---

## 2. GUIDE D'UTILISATION À L'ATTENTION DU PRESTATAIRE EXTERNE

Puisque vous découvrez l'interface de **SANAD PHARMA**, chaque fiche de test est rédigée selon une méthodologie rigoureusement univoque :
1. **Identifiant Unique** : Chaque cas possède une référence (ex. `OQ-AUTH-01`).
2. **Localisation Visuelle dans l'UI** : Les éléments graphiques sont décrits avec leur nom exact, leur icône et leur positionnement (ex. *« Barre de navigation latérale gauche > Onglet 'Dossiers' > Bouton bleu '+ Nouveau Dossier' en haut à droite »*).
3. **Données de Test Clé en Main** : Vous ne devez rien inventer. Des valeurs pré-établies sont fournies dans les fiches.
4. **Critère d'Acceptation** : Spécification exacte du résultat attendu à l'écran, du comportement du système et de l'enregistrement dans l'Audit Trail.
5. **Preuve Documentaire Exigée** : Chaque test réussi ou échoué doit être consigné avec une capture d'écran horodatée nommée selon le standard `SCR-[ID_TEST]-[STATUT].png` (ex. `SCR-OQ-AUTH-01-PASS.png`).

---

## 3. ENVIRONNEMENT DE TEST, MATRICE DES COMPTES & RÔLES BPF

### 3.1. Accès Technique à l'Environnement de Qualification
* **URL d'accès UAT :** `http://localhost:8090` (ou URL dédiée fournie par l'administrateur système).
* **Navigateurs supportés :** Google Chrome (v115+), Microsoft Edge (v115+), Mozilla Firefox (v118+).
* **Résolution d'écran recommandée :** 1920 × 1080 (Full HD).

### 3.2. Matrice des 10 Comptes & Rôles BPF Pré-configurés
Tous les comptes de démonstration et de qualification partagent le mot de passe initial : **`Admin123!`**.
Sur la mire de connexion, une boîte interactive intitulée **« ⚡ Comptes de test & Matrice des 10 Rôles BPF »** permet de pré-remplir instantanément les identifiants d'un simple clic.

| Rôle BPF / Fonction | Nom du Profil | Identifiant (Email) | Mot de passe | Rôle Système & Habilitations |
| :--- | :--- | :--- | :--- | :--- |
| **Directeur Technique (DT)** | Dr. Yacine Benkhalid | `pdt@pharma.dz` | `Admin123!` | Signataire légal habilité Art. 1.A.7, approbateur ultime des dossiers et levées d'observations ANPP. |
| **Pharmacien Assistant (Intérim DT)** | Dr. Lynda A. | `assistant.dt@pharma.dz` | `Admin123!` | Suppléant qualifié avec délégation temporaire traçable. |
| **Admin Laboratoire** | Karim M. | `admin.lab@pharma.dz` | `Admin123!` | Gestionnaire des utilisateurs internes et des paramétrages organisationnels. |
| **Manager Affaires Réglementaires** | Sarah B. | `ra.lead@pharma.dz` | `Admin123!` | Supervision des dossiers d'enregistrement, validation préliminaire et suivi des échéances. |
| **Chargé Réglementaire (Maker)** | Sofiane H. | `ra.assoc@pharma.dz` | `Admin123!` | Création, saisie des dossiers et téléversement des pièces justificatives (Maker). |
| **Manager Assurance Qualité (AQ)** | Nadir T. | `qa.lead@pharma.dz` | `Admin123!` | Contrôle de conformité documentaire et respect des BPF / Annexe 11. |
| **Contrôle Qualité & Stabilité** | Leila M. | `qc.lead@pharma.dz` | `Admin123!` | Dépôt et validation des rapports d'analyses physico-chimiques et études de stabilité (Module 3). |
| **Production & Validation Procédé** | Omar B. | `prod.lead@pharma.dz` | `Admin123!` | Validation des protocoles de fabrication et validation de procédé industriel (Module 3). |
| **Consultant Externe (CRO / Auditeur)** | David M. | `consultant.ext@cro.com` | `Admin123!` | Accès restreint « besoin d'en connaître » : occultation des formules secrètes et prix. |
| **Titulaire d'API (Active Substance)** | Dr. Wei Zhang | `api.holder@zhejiang-chem.com`| `Admin123!` | Dépôt exclusif de la partie fermée du dossier matière active (ASMF 3.2.S). |
| **SuperAdmin Éditeur SaaS** | Administrateur Système | `super.admin@homolog.dz` | `Admin123!` | Gestion de l'infrastructure, journaux systèmes et clés de chiffrement. |

---

## 4. PROCÉDURES DÉTAILLÉES DE TESTS OPÉRATIONNELS

```
                                      ARCHITECTURE DES PROCÉDURES DE TEST
                                                       │
         ┌─────────────────────────┬───────────────────┴───────────────────┬─────────────────────────┐
         ▼                         ▼                                       ▼                         ▼
   SÉCURITÉ & ACCÈS        MÉTIER RÉGLEMENTAIRE                   WORKFLOWS & CONTRÔLE       ALCOA+ & PART 11
   • OQ-AUTH (Multi-Rôles) • PQ-CTD (Dossiers eCTD)               • PQ-MC (Maker-Checker)    • OQ-AUD (Audit Live/Geo)
   • OQ-SES (Timeouts)     • PQ-DEF (Observations ANPP)           • PQ-CONF (Secret/CRO)     • OQ-SIG (Signature Élec)
   • OQ-GATE (Sas d'Accès) • PQ-FEE (Redevances & Bordereaux)     • PQ-CORP (Tabadol/MIPH)   • OQ-EXP (Archivage PDF/A)
                           • PQ-DM (Dispositifs Médicaux)                                    • OQ-UI (Ergonomie Responsive)
```

---

### MODULE 01 : SAS DE SÉCURITÉ (GATE) & AUTHENTIFICATION MULTI-RÔLES

#### Test Ref : `OQ-AUTH-01`
* **Titre :** Vérification de l'authentification nominative nominale et affichage du cockpit personnalisé.
* **Exigence Réglementaire :** 21 CFR §11.10(d), §11.200 (Identification unique par opérateur).
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).
* **Prérequis :** Navigateur ouvert sur l'URL d'accueil.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se rendre sur `http://localhost:8090`. | URL d'accueil | La page de connexion s'affiche avec le logo SANAD PHARMA, les champs Email, Mot de passe et le sélecteur rapide de comptes. |
| 2 | Cliquer sur le bouton démo **`💊 Dr. Yacine`** ou saisir manuellement les identifiants. | Email: `pdt@pharma.dz`<br>MDP: `Admin123!` | Les champs Email et Mot de passe sont complétés automatiquement. Le champ mot de passe masque les caractères par des puces. |
| 3 | Cliquer sur le bouton bleu **`Se connecter au cockpit ➔`**. | Clic bouton | 1. Le bouton affiche brièvement « Vérification en cours... ».<br>2. L'écran de connexion disparaît au profit du Cockpit Réglementaire.<br>3. En haut à droite, le bandeau utilisateur affiche : *« Dr. Yacine Benkhalid — Directeur Technique »* avec le badge de rôle vert. |
| 4 | Cliquer sur le bouton de l'Audit Trail (icône de bouclier 🛡️ dans la barre supérieure). | Clic icône 🛡️ | La modale d'Audit Trail s'ouvre. La première ligne du tableau enregistre l'événement :<br>- **Action :** `AUTH_LOGIN`<br>- **Utilisateur :** `pdt@pharma.dz`<br>- **Rôle :** `DIRECTEUR_TECHNIQUE`<br>- **Lieu :** Badge avec drapeau et localisation résolue. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUTH-01.png`

---

#### Test Ref : `OQ-AUTH-02`
* **Titre :** Rejet des tentatives de connexion avec mot de passe erroné et protection anti-intrusion.
* **Exigence Réglementaire :** 21 CFR §11.300 (Détection des accès non autorisés).
* **Rôle Requis :** Tout compte.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Sur l'écran de connexion, saisir une adresse valide mais un mot de passe erroné. | Email: `pdt@pharma.dz`<br>MDP: `FauxMotDePasse2026!` | Données saisies. |
| 2 | Cliquer sur **`Se connecter au cockpit ➔`**. | Clic bouton | 1. L'accès au cockpit est formellement bloqué.<br>2. Un message d'alerte rouge apparaît : *« Identifiants incorrects ou compte verrouillé »*.<br>3. Aucun jeton de session n'est émis. |
| 3 | Se connecter avec un compte administrateur et ouvrir l'Audit Trail. | Compte Admin | Un événement de sécurité est enregistré :<br>- **Action :** `AUTH_LOGIN_FAILED`<br>- **Détails :** tentative infructueuse tracée avec horodatage et adresse IP source. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUTH-02.png`

---

#### Test Ref : `OQ-AUTH-03`
* **Titre :** Vérification du cloisonnement strict des privilèges selon le rôle (RBAC).
* **Exigence Réglementaire :** 21 CFR §11.10(g) (Contrôle d'autorité et des autorisations opérationnelles).

| Étape | Action de l'Auditeur (Navigation UI) | Compte Utilisé | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se connecter avec le profil **Chargé Réglementaire (Maker)**. | `ra.assoc@pharma.dz` | Le profil accède à la création des dossiers. Le bouton de signature légale finale « Approuver & Signer DT » est grisé ou absent. |
| 2 | Se déconnecter, puis se connecter avec le profil **Directeur Technique (DT)**. | `pdt@pharma.dz` | Le bouton de signature « Approuver & Signer DT » est actif et accessible. |
| 3 | Tenter d'accéder directement par l'API à une route interdite au Maker (ex. validation DT). | Appel API avec token Maker | Le serveur répond systématiquement par un code HTTP `403 Forbidden` avec log d'alerte dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUTH-03.png`

---

### MODULE 02 : GESTION DES SESSIONS & TIMEOUT D'INACTIVITÉ

#### Test Ref : `OQ-SES-01`
* **Titre :** Déclenchement automatique de la modale de pré-avertissement après inactivité.
* **Exigence Réglementaire :** 21 CFR §11.10(d), Annexe 11 §12 (Verrouillage automatique de session sans surveillance).
* **Rôle Requis :** Tout profil connecté.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se connecter au cockpit et rester totalement inactif (sans mouvement de souris ni frappe clavier). | Délai configuré (ou forcé à 1 min en test) | À l'approche de la fin de session, une modale d'alerte modale critique surgit au premier plan : **« Session bientôt expirée »**. |
| 2 | Observer les informations présentées dans la modale. | Observation visuelle | 1. Un compte à rebours dynamique s'égrène en secondes (ex. *« Déconnexion dans 60 secondes... »*).<br>2. Deux options sont offertes : bouton bleu **« Prolonger ma session »** et bouton gris **« Se déconnecter »**. |
| 3 | Cliquer sur le bouton bleu **« Prolonger ma session »**. | Clic bouton | 1. La modale se referme instantanément.<br>2. Le jeton d'activité est réinitialisé avec succès.<br>3. L'utilisateur demeure dans son contexte de travail sans perte de saisie. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-SES-01.png`

---

#### Test Ref : `OQ-SES-02`
* **Titre :** Déconnexion forcée automatique à l'issue du compte à rebours sans réponse.
* **Exigence Réglementaire :** 21 CFR §11.10(d).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Laisser le compte à rebours de la modale d'inactivité arriver à zéro (0s). | Inaction complète | 1. Dès que le décompte atteint 0s, la session est révoquée côté client et serveur.<br>2. L'application bascule automatiquement sur l'écran de verrouillage / login.<br>3. Le message rouge d'information indique : *« Votre session a expiré pour cause d'inactivité (protection 21 CFR Part 11) »*.<br>4. L'utilisation du bouton « Précédent » du navigateur ne permet en aucun cas de recharger le cockpit sans ré-authentification. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-SES-02.png`

---

### MODULE 03 : DOSSIERS PHARMACEUTIQUES CTD / eCTD (MODULLES 1 À 5)

#### Test Ref : `PQ-CTD-01`
* **Titre :** Création d'un nouveau dossier d'enregistrement de médicament avec structure CTD complète.
* **Exigence Réglementaire :** Directives ICH CTD M4, Réglementation ANPP Dépôt d'Enregistrement.
* **Rôle Requis :** Chargé Réglementaire (`ra.assoc@pharma.dz`) ou Manager Réglementaire (`ra.lead@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données de Test à Saisir | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le menu de navigation gauche, cliquer sur l'onglet **« Dossiers d'Enregistrement »** (icône 📁). | Clic onglet | La vue des dossiers s'affiche avec la liste des produits en cours d'homologation. |
| 2 | Cliquer sur le bouton vert **`+ Nouveau Dossier d'Enregistrement`** en haut à droite. | Clic bouton | La modale de création de dossier s'ouvre. |
| 3 | Compléter les champs du formulaire :<br>- **Nom de Spécialité :**<br>- **DCI (Molécule) :**<br>- **Dosage :**<br>- **Forme Pharmaceutique :**<br>- **Classe Thérapeutique :**<br>- **Type de Procédure :** | `AMOXIPHARM`<br>`Amoxicilline trihydratée`<br>`500 mg`<br>`Gélule`<br>`Antibiotique (Bêta-lactamine)`<br>`Enregistrement Initial (National)` | Tous les champs sont renseignés conformément. La validation du formulaire s'active. |
| 4 | Cliquer sur le bouton **`Créer et Initialiser la Structure CTD`**. | Clic bouton | 1. Le dossier est créé avec succès avec un numéro de référence unique (ex. `DOS-2026-AMX-001`).<br>2. L'arborescence CTD s'initialise automatiquement avec les 5 modules ICH :<br>   - Module 1 : Informations Administratives & Régionales ANPP<br>   - Module 2 : Résumés et Synthèses du CTD<br>   - Module 3 : Qualité Pharmaceutique (CMC / Substance & Produit)<br>   - Module 4 : Rapports Précliniques<br>   - Module 5 : Rapports d'Études Cliniques & Bioéquivalence |
| 5 | Vérifier dans l'Audit Trail. | Consultation 🛡️ | Événement enregistré : `DOSSIER_CREATE` avec identifiant du dossier, nom de spécialité et horodatage ALCOA+. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-CTD-01.png`

---

#### Test Ref : `PQ-CTD-02`
* **Titre :** Téléversement, contrôle d'intégrité et visualisation sécurisée d'une pièce jointe CTD.
* **Exigence Réglementaire :** 21 CFR §11.10(b), §11.10(c) (Gestion de documents électroniques inaltérables).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le dossier créé, déplier le **Module 1 (Administratif)** > Volet **1.2 Demande d'Enregistrement**. | Navigation arborescence | Le volet documentaire s'ouvre, affichant l'exigence de la lettre officielle de demande. |
| 2 | Cliquer sur le bouton **`Téléverser un document`** et sélectionner un fichier PDF valide (ex. `Lettre_Demande_ANPP.pdf`). | Fichier PDF de test | 1. Le système vérifie l'extension et le type MIME (`application/pdf`).<br>2. Une barre de progression s'affiche lors de l'envoi.<br>3. Le fichier est chiffré et consigné dans le coffre-fort numérique avec son hash SHA-256 calculé instantanément. |
| 3 | Tenter de téléverser un fichier non autorisé (ex. `script.exe` ou `test.bat`). | Fichier exécutable | Le système rejette immédiatement le fichier avec un message d'erreur : *« Format non conforme aux spécifications eCTD (PDF uniquement) »*. |
| 4 | Cliquer sur l'icône de l'œil (Visionneuse de document 👁️) en face du PDF téléversé. | Clic icône œil | 1. La visionneuse sécurisée intégrée (`#modal-doc-viewer`) s'ouvre en superposition.<br>2. Le document est parfaitement lisible.<br>3. Un filigrane de sécurité dynamique affiche la mention : *« COPIE CONTRÔLÉE SANAD PHARMA — CONSULTÉ PAR [Nom Utilisateur] LE [Date/Heure] »* (empêchant la fuite d'informations). |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-CTD-02.png`

---

### MODULE 04 : WORKFLOW DE DOUBLE CONTRÔLE MAKER-CHECKER & NON-AUTO-APPROBATION

#### Test Ref : `PQ-MC-01`
* **Titre :** Soumission d'un jalon documentaire par le rédacteur (Maker) pour revue qualité.
* **Exigence Réglementaire :** 21 CFR §11.10(f) (Séquençage opérationnel obligatoire des étapes de validation).
* **Rôle Requis :** Chargé Réglementaire Maker (`ra.assoc@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Connecté en tant que **Maker** (`ra.assoc@pharma.dz`), ouvrir le dossier `AMOXIPHARM`. | Dossier en statut `Brouillon` | L'état du dossier indique le badge gris `BROUILLON (EN COURS DE RÉDACTION)`. |
| 2 | Cliquer sur le bouton orange **`Soumettre pour Revue Qualité & Approbation DT`**. | Clic bouton | Une boîte de confirmation s'affiche demandant la confirmation du dépôt et un commentaire facultatif. |
| 3 | Saisir le commentaire : `Dossier complet, pièces Module 1 à 3 vérifiées` et valider. | Clic **Confirmer la soumission** | 1. L'état du dossier bascule immédiatement à `EN ATTENTE D'APPROBATION DT` (badge jaune ambré).<br>2. Les champs de saisie du dossier sont désormais verrouillés en écriture pour le Maker.<br>3. Un événement `DOSSIER_SUBMIT_FOR_APPROVAL` est consigné dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-MC-01.png`

---

#### Test Ref : `PQ-MC-02`
* **Titre :** Règle de non-auto-approbation stricte (Ségrégation des tâches - SoD).
* **Exigence Réglementaire :** GAMP 5, 21 CFR §11.10(f) (Interdiction d'approuver ses propres soumissions).
* **Rôle Requis :** Même utilisateur que l'étape précédente (`ra.assoc@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Toujours connecté avec le profil du **Maker** (`ra.assoc@pharma.dz`), visualiser le dossier soumis. | Vue dossier | Le bouton vert **« Approuver et Signer (DT) »** est totalement inactif, grisé ou masqué. |
| 2 | Tenter de forcer l'appel HTTP de signature avec le jeton du Maker. | Requête POST `/api/v1/dossiers/{id}/approve` | Le serveur rejette la requête avec l'erreur `403 Forbidden` : *« Ségrégation des tâches BPF : Le rédacteur du dossier ne peut pas être son approbateur »*. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-MC-02.png`

---

#### Test Ref : `PQ-MC-03`
* **Titre :** Approbation formelle et scellement du dossier par le Directeur Technique (Checker).
* **Exigence Réglementaire :** 21 CFR §11.50, §11.70.
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se déconnecter et se connecter avec le profil **Directeur Technique** (`pdt@pharma.dz`). | `pdt@pharma.dz` | Le cockpit affiche le dossier `AMOXIPHARM` dans la section « Actions Requises / Approbations en Attente ». |
| 2 | Ouvrir le dossier et cliquer sur le bouton vert **`Approuver et Apposer la Signature DT`**. | Clic bouton | Le modal de signature électronique 21 CFR Part 11 surgit (voir Module 11 pour le déroulé de signature). |
| 3 | Après signature valide, observer le statut du dossier. | Vue dossier | 1. Le statut passe à `APPROUVÉ PAR LE DIRECTEUR TECHNIQUE - PRÊT POUR TÉLÉSOUMISSION ANPP` (badge vert vif).<br>2. L'empreinte cryptographique globale du dossier (Master Manifest Hash) est générée et scellée. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-MC-03.png`

---

### MODULE 05 : LETTRES D'OBSERVATION ANPP (DEFICIENCY LETTERS) & ALERTES DÉLAIS

#### Test Ref : `PQ-DEF-01`
* **Titre :** Enregistrement d'une lettre d'observation ANPP avec calcul automatique du compte à rebours légal.
* **Exigence Réglementaire :** Réglementation ANPP sur le traitement des compléments d'information (Délai légal de réponse).
* **Rôle Requis :** Manager Réglementaire (`ra.lead@pharma.dz`) ou Pharmacien DT.

| Étape | Action de l'Auditeur (Navigation UI) | Données de Test à Saisir | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le menu de gauche, cliquer sur l'onglet **« Lettres d'Observation ANPP »** (icône 📑). | Clic onglet | La table de suivi des fiches d'observations et demandes de compléments s'ouvre. |
| 2 | Cliquer sur le bouton **`+ Enregistrer une Notification ANPP`**. | Clic bouton | La modale de création d'observation s'affiche. |
| 3 | Renseigner les champs obligatoires :<br>- **Dossier Concerné :**<br>- **Réf. Courrier ANPP :**<br>- **Date Réception :**<br>- **Délai accordé par l'ANPP :**<br>- **Objet de la demande :** | Sélectionner `AMOXIPHARM`<br>`ANPP/DEDP/N°2026-489`<br>Date du jour (`2026-11-24`)<br>`90 jours`<br>`Complément d'études de stabilité sous climat de Zone IVb (30°C/75% RH)` | Formulaire complété. |
| 4 | Cliquer sur **`Enregistrer et Déclencher le Rétroplanning`**. | Clic bouton | 1. La lettre est enregistrée dans le système.<br>2. La date butoir légale est calculée automatiquement : `Date Réception + 90 jours ouvrés/calendaires`.<br>3. Un badge d'alerte temporel dynamique apparaît : affiche le nombre de jours restants avec un code couleur :<br>   - Vert (> 30 jours)<br>   - Orange (entre 15 et 30 jours)<br>   - Rouge clignotant (< 15 jours). |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-DEF-01.png`

---

### MODULE 06 : BARÈME DES REDEVANCES & BORDEREAUX DE PAIEMENT ANPP

#### Test Ref : `PQ-FEE-01`
* **Titre :** Calcul automatique des droits d'enregistrement et génération du bordereau de versement.
* **Exigence Réglementaire :** Décret interministériel fixant le tarif des redevances dues à l'ANPP.
* **Rôle Requis :** Manager Réglementaire (`ra.lead@pharma.dz`) ou Admin Labo.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le menu de gauche, cliquer sur **« Redevances & Bordereaux ANPP »** (icône 💳). | Clic onglet | Le tableau de bord financier réglementaire s'affiche. |
| 2 | Cliquer sur le bouton **`Générer un Bordereau de Paiement`**. | Clic bouton | La calculatrice réglementaire de redevance s'ouvre. |
| 3 | Sélectionner le type d'acte :<br>- **Nature de la demande :** `Enregistrement Initial Produit Fabriqué Localement`<br>- **Nombre de dosages :** `1`<br>- **Formes associées :** `Forme sèche classique` | Sélection des options | Le système applique automatiquement le barème légal officiel ANPP (montant en Dinars Algériens DZD). Le montant total hors taxes et TTC est calculé sans possibilité de modification manuelle arbitraire. |
| 4 | Cliquer sur **`Éditer le Bordereau Officiel de Versement`**. | Clic bouton | 1. Un bordereau normalisé est généré au format PDF/A comportant le RIB officiel de l'ANPP, le numéro de référence unique de transaction et le code-barres d'identification.<br>2. L'événement `FEE_SLIP_GENERATED` est archivé dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-FEE-01.png`

---

### MODULE 07 : HOMOLOGATION & RÉGULARISATION DES DISPOSITIFS MÉDICAUX (DM)

#### Test Ref : `PQ-DM-01`
* **Titre :** Renseignement et validation du Formulaire Officiel d'Homologation DM en 6 sections normées (`ES-REG-02-DM`).
* **Exigence Réglementaire :** Décret exécutif n° 20-324 & Arrêté du 10 mai 2021 (Composition du dossier d'homologation).
* **Rôle Requis :** Responsable Affaires Réglementaires DM (`ra.lead@pharma.dz`) ou Pharmacien DT.

| Étape | Action de l'Auditeur (Navigation UI) | Données de Test à Saisir | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans la barre d'activités gauche, cliquer sur l'icône Dispositifs Médicaux (icône 💉 ou seringue, `#act-dm`). | Clic icône `#act-dm` | Le cockpit DM s'ouvre avec les 3 onglets thématiques : *Portefeuille & 6 Sections*, *Mesures Transitoires & Régularisation*, *Modifications Post-Homologation*. |
| 2 | Dans l'onglet **« 🩺 Portefeuille & 6 Sections »**, repérer le dispositif *« Cathéter intraveineux stérile »* et cliquer sur le bouton **`Fiche 6 Sections`**. | Clic bouton | La modale `#modal-dm-homologation` surgit, affichant le formulaire officiel `ES-REG-02-DM` structuré en 6 sections normées. |
| 3 | Parcourir les 6 sections techniques :<br>- **Sec. 1 (Administratif) :** Nomenclature GMDN, Classe de risque.<br>- **Sec. 2 (Technique) :** Mode d'action, dimensions, stérilisation ETO.<br>- **Sec. 3 (Fabricant & Sites) :** Sites de fabrication et d'assemblage.<br>- **Sec. 4 (Toxicologie & Sécurité) :** Déclarations Latex Free, sans CMR, sans phtalates DEHP.<br>- **Sec. 5 (Normes ISO) :** ISO 14971, ISO 10993, ISO 14155, ISO 11607.<br>- **Sec. 6 (Échantillons) :** Lots pilotes LNCM. | Modifier une référence (ex. Référence ISO 14971 : `PGR-ISO14971-2026-REV02`) | Les champs sont éditables avec contrôle de validité en temps réel. |
| 4 | Cliquer sur le bouton bleu **`💾 Enregistrer la Fiche 6 Sections`**. | Clic bouton | 1. La fiche est mise à jour avec succès.<br>2. La modale se referme et affiche la notification verte : *« Données d'homologation 6 sections mises à jour »*.<br>3. L'événement est consigné dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-DM-01.png`

---

#### Test Ref : `PQ-DM-02`
* **Titre :** Audit de recevabilité en 13 points (`ES-REG-03-DM`), scoring et scellement du Récépissé Officiel par visa 2FA du Pharmacien DT.
* **Exigence Réglementaire :** Arrêté du 10 mai 2021 (Annexe II), 21 CFR §11.50 & §11.70.
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans l'onglet **« 🩺 Portefeuille & 6 Sections »**, cliquer sur le bouton **`Évaluer Recevabilité`** en face du dossier de DM. | Clic bouton | La modale `#modal-dm-recevabilite` s'affiche, présentant la grille d'évaluation de la recevabilité en 13 points de contrôle éliminatoires de l'ANPP. |
| 2 | Examiner le barème de recevabilité affiché : | Observation visuelle | 1. Les 13 points de contrôle sont listés (lettre de demande, formulaire 6 sections, agrément MIPH, redevance BNA, CLV/FSC, marquage CE, ISO 13485, matières premières, matériovigilance, stérilisation, ISO 14971/10993, clinique ISO 14155, 04 exemplaires électroniques).<br>2. Le badge de score indique **`Score : 13/13 (100%) — COMPLET ET RECEVABLE`** (fond vert). |
| 3 | Cliquer sur le bouton vert **`Sceller et Délivrer le Récépissé Officiel (Visa DT)`**. | Clic bouton scellement | 1. Le système exige la confirmation de signature légale du Pharmacien DT.<br>2. Le Récépissé Officiel `ES-REG-03-DM` est scellé avec génération d'un numéro d'enregistrement DDM (ex. `ANPP/DDM/REC/2026/0148`).<br>3. Le hash cryptographique SHA-256 du récépissé est gravé et un événement `DM_RECEVABILITE_SEAL` est enregistré dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-DM-02.png`

---

#### Test Ref : `PQ-DM-03`
* **Titre :** Suivi des Mesures Transitoires, quotas de distribution hospitalière vs officine et export du canevas dérogatoire ministériel.
* **Exigence Réglementaire :** Dispositions transitoires du Décret 20-324, Canevas ministériel ANPP.
* **Rôle Requis :** Responsable Affaires Réglementaires DM (`ra.lead@pharma.dz`) ou Pharmacien DT.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le cockpit DM, cliquer sur le deuxième onglet : **`⏱️ Mesures Transitoires & Régularisation`**. | Clic onglet | Le tableau de bord de régularisation s'ouvre, affichant les indicateurs clés (Total DM en régularisation, volume trimestriel, alertes forclusion). |
| 2 | Vérifier les données du tableau : | Observation visuelle | 1. La liste affiche les DM en circulation sous dérogation (codes `REG-DM-...`).<br>2. Les colonnes détaillent : Statut dérogatoire, Accusé DDM, Échéance légale avec badge dynamique (ex. `J-22` ou `J-90`), Quotas Hôpitaux PCH (%) vs Officines (%). |
| 3 | Cliquer sur le bouton bleu **`📊 Exporter Canevas Dérogatoire (.xlsx)`** dans la barre d'outils. | Clic bouton export | Le fichier officiel de déclaration trimestrielle `CANEVAS_REGULARISATION_DM_ANPP.csv` / `.xlsx` est téléchargé instantanément, prêt pour télétransmission à la DDM de l'ANPP. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-DM-03.png`

---

#### Test Ref : `PQ-DM-04`
* **Titre :** Gestion des Modifications Post-Homologation (Articles 22 à 25 du Décret 20-324) et matrice d'impact normatif ISO.
* **Exigence Réglementaire :** Décret 20-324 (Articles 22 à 25), Ord. 21-07 Art. 31 (Redevance d'avenant).
* **Rôle Requis :** Responsable Affaires Réglementaires DM (`ra.lead@pharma.dz`) ou Pharmacien DT.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le cockpit DM, cliquer sur le troisième onglet : **`🔄 Modifications Post-Homologation`**. | Clic onglet | Le registre des demandes de modification s'affiche avec la répartition : Modifications Substantielles (avec demande d'avenant) vs Non-Substantielles (déclarations annuelles). |
| 2 | Cliquer sur **`➕ Nouvelle Modification DM`**. | Clic bouton | La modale de saisie de modification post-homologation `#modal-dm-modification-new` s'ouvre. |
| 3 | Renseigner les champs :<br>- **Dispositif Médical :**<br>- **Nature de la modification :**<br>- **Catégorie :**<br>- **Description :**<br>- **Impact Normatif ISO :** | Sélectionner `Cathéter intraveineux`<br>`SUBSTANTIELLE (Avenant Décision DDM)`<br>`DM_SUB_03 - Changement de site de stérilisation`<br>`Transfert du procédé ETO vers nouveau site qualifié ISO 11135`<br>Cocher `ISO 11135 (Stérilisation)` et `ISO 14971 (Risques)` | Les champs sont renseignés et le montant de la redevance légale d'avenant est calculé automatiquement (150 000 DA). |
| 4 | Cliquer sur **`Enregistrer la Déclaration de Modification`**. | Clic bouton | 1. La modification est enregistrée dans le système avec son code unique.<br>2. La matrice d'impact normatif ISO est gravée et l'événement est consigné dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-DM-04.png`

---

### MODULE 08 : FABRICANTS ÉTRANGERS, CPP OMS & SECRET INDUSTRIEL (GUEST / BLACK-BOX)

#### Test Ref : `PQ-CONF-01`
* **Titre :** Vérification de l'occultation des données confidentielles pour le profil Consultant Externe (Black-Box).
* **Exigence Réglementaire :** 21 CFR §11.10(d), Protection du Secret Industriel et Commercial.
* **Rôle Requis :** Consultant Externe CRO (`consultant.ext@cro.com`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se connecter avec le compte invité **Consultant Externe** (`consultant.ext@cro.com`). | `consultant.ext@cro.com` | L'interface s'adapte en mode restreint « Invité / Auditeur ». |
| 2 | Ouvrir la fiche détaillée du médicament `AMOXIPHARM` et naviguer vers le Module 3 (Composition & Formule unitaire). | Navigation écran | 1. Les quantités exactes des principes actifs, les procédés de synthèse confidentiels et les coûts de revient sont remplacés par des bandeaux noirs hachurés marqués : **`[CONFIDENTIEL - SECRET INDUSTRIEL MASQUÉ]`**.<br>2. Les boutons de téléchargement brut des sources ou d'export de masse sont totalement inaccessibles.<br>3. Chaque consultation d'écran est consignée avec le tag `DOC_VIEW_RESTRICTED` dans l'Audit Trail. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-CONF-01.png`

---

### MODULE 09 : AGRÉMENTS MIPH / ANPP & TÉLÉSOUMISSION TABADOL

#### Test Ref : `PQ-CORP-01`
* **Titre :** Contrôle de validité des agréments corporate et préparation de la télésoumission Tabadol.
* **Exigence Réglementaire :** Décrets Ministère de l'Industrie Pharmaceutique (MIPH) / Plateforme Tabadol.
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Cliquer sur l'onglet **« Agréments Corporate & Tabadol »** (icône 🏛️). | Clic onglet | L'espace des agréments institutionnels s'ouvre. |
| 2 | Vérifier l'état de l'Agrément d'Établissement Pharmaceutique de Fabrication. | Consultation visuelle | La fiche affiche le numéro d'agrément ministériel, la date de délivrance et le statut de validité. |
| 3 | Cliquer sur le bouton **`Préparer le Paquet Télésoumission Tabadol (gate.miph.gov.dz)`**. | Clic bouton | Le système compile les métadonnées et pièces requises au format structuré attendu par le guichet national Tabadol, avec bordereau d'export scellé cryptographiquement. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-PQ-CORP-01.png`

---

### MODULE 10 : AUDIT TRAIL ALCOA+, FLUX LIVE (3S), PAUSE/REPRENDRE, GÉOLOCALISATION HAUTE PRÉCISION & CHAÎNAGE SHA-256

#### Test Ref : `OQ-AUD-01`
* **Titre :** Vérification du flux temps réel de l'Audit Trail, contrôle de cadencement (3s), boutons Pause et Reprendre.
* **Exigence Réglementaire :** 21 CFR §11.10(e) (Horodatage contemporain, continu et indépendant des enregistrements).
* **Rôle Requis :** Tout utilisateur connecté (ou Administrateur / DT).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans la barre d'outils supérieure, cliquer sur le bouton de l'Audit Trail (icône de bouclier 🛡️). | Clic icône 🛡️ | La modale de l'Audit Trail (`#modal-audit`) s'ouvre au centre de l'écran. |
| 2 | Observer l'en-tête de la modale d'Audit Trail. | Observation visuelle | 1. Le badge d'état `#audit-live-badge` affiche **`🟢 EN DIRECT (3s)`** avec une pastille verte animée d'une pulsation rythmée (`auditPulseLive`).<br>2. Le bouton `#audit-live-toggle-btn` affiche **`⏸️ Pause`**.<br>3. Le bouton `#audit-refresh-btn` affiche **`🔄 Actualiser`**.<br>4. Le pied de page affiche : *« Synchro : [HH:MM:SS] »* mis à jour toutes les 3 secondes. |
| 3 | Cliquer sur le bouton **`⏸️ Pause`**. | Clic bouton | 1. Le badge bascule instantanément sur **`⏸️ EN PAUSE`** (fond ambré fixe, arrêt de la pulsation).<br>2. Le bouton se transforme en **`▶️ Reprendre`**.<br>3. Le flux automatique en arrière-plan est mis en attente, permettant à l'auditeur d'inspecter une ligne sans défilement intempestif. |
| 4 | Cliquer sur le bouton **`▶️ Reprendre`**. | Clic bouton | Le badge repasse immédiatement à **`🟢 EN DIRECT (3s)`** et le bouton redevient **`⏸️ Pause`**. Le flux en direct redémarre. |
| 5 | Cliquer sur le bouton **`🔄 Actualiser`**. | Clic bouton | L'icône de flèche tourne sur elle-même (`spin-anim`) et la liste des événements s'actualise immédiatement. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUD-01.png`

---

#### Test Ref : `OQ-AUD-02`
* **Titre :** Vérification de la géolocalisation de haute précision (échelle de la commune / daïra) et traçabilité IP.
* **Exigence Réglementaire :** 21 CFR §11.10(e), Principe ALCOA+ (Attributable & Accurate).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le tableau de l'Audit Trail, observer la colonne intitulée **« Adresse IP & Lieu »**. | Observation tableau | Chaque ligne présente l'adresse IP de connexion accompagnée d'un badge géographique avec le drapeau national et la localité résolue (ex. `🇩🇿 El Achour, Alger, Algérie` ou `🇩🇿 Douera, Alger, Algérie`). |
| 2 | Cliquer sur une ligne d'événement `AUTH_LOGIN` pour ouvrir le volet d'inspection latéral ALCOA+. | Clic sur la ligne | 1. Le volet d'inspection s'ouvre à droite.<br>2. La section « Empreinte Réseau & Géolocalisation » détaille :<br>   - **Adresse IP :** IP publique de l'opérateur (ex. `105.101.66.109`).<br>   - **Localisation Scellée :** Commune, Wilaya et Pays avec précision fine.<br>   - **Payload JSON Immuable :** Le champ `"client_location"` est scellé dans les données brutes. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUD-02.png`

---

#### Test Ref : `OQ-AUD-03`
* **Titre :** Preuve d'intégrité cryptographique et chaînage SHA-256 (Anti-altération).
* **Exigence Réglementaire :** 21 CFR §11.10(e) (Registres infalsifiables, interdiction de suppression/écrasement).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Dans le volet d'inspection de l'événement d'Audit Trail, observer le champ **« Empreinte Cryptographique (SHA-256) »**. | Consultation visuelle | Le système affiche le hash SHA-256 de l'événement courant ainsi que le `previous_hash` liant l'événement au bloc précédent. |
| 2 | Tenter une requête directe SQL de modification (`UPDATE audit_logs SET action = 'HACK'`). | Injection manuelle de test | La contrainte de base de données / trigger immuable rejette formellement l'instruction avec l'erreur : *« Rejet ALCOA+ : Les enregistrements d'audit trail sont immuables et ne peuvent être ni altérés ni supprimés »*. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUD-03.png`

---

#### Test Ref : `OQ-AUD-04`
* **Titre :** Modale d'Audit Trail Redimensionnable (Bords & Coins OS), Agrandissement Centré Plein Écran (⛶ / 🗗) et Split-View avec Inspecteur ALCOA+.
* **Exigence Réglementaire :** ALCOA+ (*Available & Legible* / 21 CFR §11.10(b) — Capacité de revue lisible et exhaustive des enregistrements d'audit sans contrainte d'espace ni débordement d'affichage).
* **Rôle Requis :** Pharmacien Directeur Technique ou tout rôle habilité `AUDIT_VIEW`.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Cliquer sur le bouton **Audit Trail** dans la barre supérieure ou la barre de statut. | Clic bouton | La modale `#modal-audit-dialog` s'ouvre au centre de l'écran avec une disposition Split-View : grille des logs à gauche, panneau d'inspection ALCOA+ à droite. |
| 2 | Positionner le curseur sur les bords et les 4 coins de la boîte de dialogue d'Audit Trail. | Survol des bords | Le curseur adopte dynamiquement les curseurs de redimensionnement de type système d'exploitation (`ns-resize`, `ew-resize`, `nwse-resize`, `nesw-resize`). |
| 3 | Cliquer et glisser sur un bord latéral ou un coin pour étirer la fenêtre. | Glisser-déposer | La modale s'ajuste fluidement en temps réel aux dimensions souhaitées (avec limites minimales de sécurité `min-width: 540px; min-height: 420px`). |
| 4 | Cliquer sur le bouton d'agrandissement plein écran **`⛶`** dans l'en-tête de la modale. | Clic sur `.btn-modal-expand` | 1. La modale passe en classe `.modal-fullscreen`.<br>2. La fenêtre est **strictement centrée** sur la surface disponible avec un espacement symétrique de 2vw (gauche/droite) et 2vh (haut/bas), sans décalage hors de l'écran.<br>3. L'icône se transforme en bouton de restauration **`🗗`**.<br>4. Les poignées de redimensionnement sont masquées pour éviter toute interférence. |
| 5 | Cliquer sur le bouton de restauration **`🗗`**. | Clic bouton | La modale quitte le mode plein écran et reprend exactement sa position et ses dimensions précédentes. |
| 6 | Cliquer sur une ligne d'événement dans le tableau de gauche. | Sélection ligne | 1. La ligne sélectionnée est mise en évidence visuelle.<br>2. L'inspecteur ALCOA+ à droite affiche la fiche d'audit détaillée : Identifiant d'événement, Horodatage strict UTC et local, Opérateur et rôle, Empreinte réseau (Adresse IP client & User-Agent), Hash SHA-256 complet avec bouton de copie en un clic, et Charge utile JSON formatée. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUD-04-RESIZE-FULLSCREEN.png`

---

#### Test Ref : `OQ-AUD-05`
* **Titre :** Navigation par Onglets Métier Catégorisés (7 Catégories BPF), Filtrage Multi-Critères (dont Adresse IP) et Export CSV RFC 4180 avec UTF-8 BOM.
* **Exigence Réglementaire :** 21 CFR §11.10(b), §11.10(e) (Filtrage opérationnel des flux et génération de copies exportables conformes pour inspection).
* **Rôle Requis :** Pharmacien Directeur Technique ou tout rôle habilité `AUDIT_VIEW`.

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Observer la barre d'onglets supérieure `#audit-tabs-bar` de la modale. | Observation visuelle | 7 onglets thématiques sont affichés avec leurs badges de comptage dynamique :<br>• `🌐 Tous`<br>• `🔐 Sécurité & Accès`<br>• `✍️ Visas & Signatures DT`<br>• `📋 Dossiers CTD & Gating`<br>• `🏛️ Agréments MIPH & Tabadol`<br>• `🔄 Variations & Compléments`<br>• `🛡️ Coffre-fort CAS & Système` |
| 2 | Cliquer successivement sur chaque onglet thématique (ex. `✍️ Visas & Signatures DT`). | Clic onglet | Le tableau filtre instantanément les événements de la catégorie correspondante et met à jour l'indicateur de volume scellé (`X événement(s) scellé(s)`). |
| 3 | Saisir une adresse IP (ex. `192.168.1.45` ou `127.0.0.1`) dans le champ de recherche textuelle `#audit-filter-input`. | Saisie adresse IP | Le tableau filtre en temps réel et n'affiche que les actions initiées depuis cette adresse IP. La colonne dédiée **« Adresse IP »** met en évidence l'adresse en typographie monospace bleue. |
| 4 | Cliquer sur le bouton **`📥 Exporter la vue (CSV)`**. | Clic bouton export | Un fichier CSV certifié RFC 4180 est téléchargé, encodé en UTF-8 avec BOM (`\uFEFF`) pour compatibilité Excel, contenant les colonnes : `Horodatage_ISO, Horodatage_Local, Code_Action, Intitule_Action, Categorie, Operateur, Email_Operateur, Role, Type_Entite, Cible_Entite, Empreinte_SHA256, Statut_Immuable, Adresse_IP`. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-AUD-05-TABS-IP-CSV.png`

---

### MODULE 11 : SIGNATURE ÉLECTRONIQUE 21 CFR PART 11 & CARTOUCHE VISUELLE

#### Test Ref : `OQ-SIG-01`
* **Titre :** Exécution d'une signature électronique légale avec double authentification et sélection obligatoire du motif.
* **Exigence Réglementaire :** 21 CFR §11.50 (Manifestation de signature), §11.200 (Composants d'identification distincts).
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données de Test à Saisir | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Sur un dossier en attente d'approbation (ex. `AMOXIPHARM`), cliquer sur **`Approuver et Apposer la Signature DT`**. | Clic bouton | La boîte de dialogue formelle de signature électronique 21 CFR Part 11 s'ouvre. |
| 2 | Observer les champs obligatoires du formulaire de signature : | Observation | La modale exige obligatoirement :<br>1. La ré-authentification par mot de passe du signataire.<br>2. La sélection impérative du motif légal dans une liste normalisée.<br>3. Le code 2FA TOTP (pour les actes critiques de libération). |
| 3 | Tenter de signer en laissant le champ du motif vide ou sans saisir le mot de passe. | Clic **Signer** avec champs vides | Le système refuse l'action et met les champs manquants en évidence rouge : *« Le mot de passe et le motif légal sont obligatoires »*. |
| 4 | Renseigner les éléments :<br>- **Mot de passe :**<br>- **Motif légal (Meaning) :** | `Admin123!`<br>Sélectionner : *« J'approuve et engage la responsabilité pharmaceutique en tant que Directeur Technique »* | Champs validés. |
| 5 | Cliquer sur le bouton **`Sceller et Apposer la Signature Électronique`**. | Clic bouton | 1. Le système vérifie les identifiants avec l'algorithme Argon2id.<br>2. La signature est calculée et scellée avec le hash du document.<br>3. La modale se referme et affiche la notification de succès verte : *« Signature 21 CFR Part 11 scellée avec succès »*. |
| 6 | Vérifier dans l'Audit Trail. | Consultation 🛡️ | Événement enregistré : `DOSSIER_PDT_SIGN` avec :<br>- **Signataire :** `pdt@pharma.dz`<br>- **Motif légal :** Libellé exact choisi<br>- **Horodatage UTC/Local :** Timestamp de signature. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-SIG-01.png`

---

#### Test Ref : `OQ-SIG-02`
* **Titre :** Vérification de la cartouche de signature visuelle sur les documents PDF exportés.
* **Exigence Réglementaire :** 21 CFR §11.50(a), §11.50(b) (Visibilité de la signature sur tout format lisible par l'homme).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Cliquer sur le bouton **`Exporter le Dossier Officiel & Attestation (PDF/A)`**. | Clic bouton export | Le fichier PDF certifié est téléchargé. |
| 2 | Ouvrir le document PDF téléchargé et se positionner sur la page finale d'approbation. | Ouverture PDF | Une cartouche officielle de signature électronique conforme 21 CFR Part 11 est visible et inamovible :<br>┌─────────────────────────────────────────────────────────────┐<br>│ **SIGNATURE ÉLECTRONIQUE CONFORME 21 CFR PART 11 / ALCOA+** │<br>│ **Signataire :** Dr. Yacine Benkhalid (Directeur Technique) │<br>│ **Date & Heure :** 2026-11-24 07:42:00 UTC+01:00           │<br>│ **Motif :** Approbation & Responsabilité Pharmaceutique     │<br>│ **Empreinte Doc (SHA-256) :** `a4f8e91c...b29d83`          │<br>│ **Certificat d'Intégrité :** SANAD-CA-2026-TRUSTED          │<br>└─────────────────────────────────────────────────────────────┘ |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-SIG-02.png`

---

### MODULE 12 : QUALIFICATION ERGONOMIQUE MULTI-BREAKPOINTS & INTÉGRITÉ VISUELLE ALCOA+

#### Test Ref : `OQ-UI-01`
* **Titre :** Vérification de l'intégrité visuelle ALCOA+ (Lisibilité / Non-scission des boutons et badges) et adaptabilité multi-résolutions.
* **Exigence Réglementaire :** ALCOA+ (*Legible / Lisible* — Les données réglementaires, statuts et contrôles doivent rester non tronqués, sans scission anarchique de texte, quelle que soit la résolution du poste de travail).
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).
* **Résolutions de Test :** 1920×1080 (Desktop), 1366×768 (Laptop Standard), 1200×800 (Laptop Compact / iPad Pro), 992×768 (Tablette).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Se connecter en Directeur Technique à la résolution 1920×1080 et naviguer sur le module **Dossiers d'Enregistrement (CTD)**. | Viewport 1920×1080 | • Carte de dossier : Le code dossier (`DOS-2026-PARAC-NF`), le délai (`J-48`) et le statut (`EN PRÉPARATION`) sont affichés sur une seule ligne sans scission de texte.<br>• Barre d'outils centrale : Alignée sur une seule ligne fluide.<br>• Frise Smart Planning : Les 6 jalons (`J-240` à `J-0`) forment un ruban chronologique continu.<br>• Volet d'Aperçu : Les onglets (`📄 Pièce`, `📋 Fiche ANPP`, `🛠️ Atelier & Visa`) et le bouton `✕ Alt+P` sont nets et alignés sans chevauchement ni débordement.<br>• Cartes de paramétrage Admin Lab & aperçus GED (`.pdf-sim-page`) : Les textes longs et empreintes SHA-256 restent strictement confinés à l'intérieur de leur card sans déborder sur les cards adjacentes. |
| 2 | Redimensionner le navigateur à la résolution Laptop Standard (**1366×768**). | Viewport 1366×768 | • Les boutons de barres d'outils restent strictement sur 1 rangée horizontale défilante (aucun empilement vertical parasite sur 3 ou 7 lignes).<br>• Aucun badge ou bouton ne subit de retour à la ligne au milieu d'un mot ou d'un chiffre. |
| 3 | Naviguer vers le module **Dispositifs Médicaux (DDM)** et tester la résolution **1200×800**. | Viewport 1200×800 | • Les boutons de filtre de classe (`Classe I`, `Classe IIa`, `Classe IIb`, `Classe III`) sont affichés d'un seul bloc, sans coupure en syllabes (`CLAS` / `SE I`).<br>• Les badges du tableau (`Classe IIa`, `🏭 Fabrication Locale`, `En Enregistrement`) conservent leur intégrité visuelle sans retour à la ligne.<br>• Le volet d'aperçu latéral se convertit en **slide-over drawer** rétractable (`Alt+P`), préservant 100% de la largeur utile du tableau central. |
| 4 | Tester la résolution **992×768** sur le module **Agréments Corporate (Décret 21-82)**. | Viewport 992×768 | • Le sélecteur de rôle démo se masque automatiquement pour éviter l'écrasement de la barre supérieure.<br>• Le pied de page masque les métriques secondaires pour ne conserver que l'heure UTC+1 d'Alger.<br>• Les badges `[S] Standard`, `[E] Éliminatoire Tabadol` et `J-180 / J-90` ne sont jamais coupés. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-UI-01-DESKTOP.png`, `SCR-OQ-UI-01-LAPTOP.png`, `SCR-OQ-UI-01-COMPACT.png`, `SCR-OQ-UI-01-TABLET.png`

---

#### Test Ref : `OQ-UI-02`
* **Titre :** Vérification de l'affichage des intitulés réglementaires officiels complets au survol (Infobulles GxP).
* **Exigence Réglementaire :** ALCOA+ (*Accurate / Exact* & *Legible / Lisible*) & GAMP 5 — Tout texte législatif ou réglementaire cité de façon compacte dans l'interface (badge, bouton, tableau, barre d'outils ou Veille Juridique) doit permettre à l'utilisateur de consulter son intitulé officiel intégral (numéro, dates hégirienne et grégorienne, référence JORADP) au survol sans encombrer l'espace de travail.
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Sur le module **Dossiers d'Enregistrement (CTD)**, survoler avec la souris un badge réglementaire dans l'arborescence ou le tableau (ex. `Décret 20-325`, `Arrêté 10 Mai`, `Décret 20-324`). | Curseur sur badge | • Le curseur prend l'aspect d'aide (`cursor: help`) et un soulignement pointillé discret apparaît.<br>• Une infobulle native (`title`) affiche l'intitulé officiel complet du texte ainsi que sa référence JORADP. |
| 2 | Ouvrir la modale **📜 Veille Juridique**, survoler une référence légale et cliquer sur le bouton **« 📄 PDF (… Ko) »** d'une ligne directrice ANPP (`LD-DVIV04/V01`, `ANPP-BPPV-V2`, etc.). | Clic sur `📜 Veille Juridique` & bouton `📄 PDF` | • Chaque ligne du corpus affiche la référence et son titre ; le survol révèle l'intitulé complet enrichi.<br>• Les **19 Guidelines officielles ANPP** sont indexées avec leur fichier PDF complet stocké en base (`regulatory_watch_items.document_pdf` + SHA-256), ouvrable hors-ligne via `GET /api/v1/regulatory-watch/{itemId}/pdf` (**HTTP 200**, `application/pdf`). |
| 3 | Naviguer vers **Agréments Corporate (Décret 21-82)** et **Variations (LD-DEPP07/V01)** et survoler les mentions réglementaires. | Navigation onglets | • Les mentions dynamiques (`Décret 21-82`, `LD-DEPP07/V01`, `Arrêté 22/09/2025`, `21 CFR Part 11`) reçoivent automatiquement leur intitulé officiel complet au survol via le `MutationObserver`. |
| 4 | Ouvrir la modale **🛡️ Audit Trail**, saisir la poignée droite d'un en-tête de colonne (`Horodatage UTC` ou `Cible / Entité`) et la faire glisser horizontalement, puis double-cliquer sur une poignée. | Clic-glisser sur `.audit-col-resizer` | • Le curseur passe en `col-resize` et la colonne s'élargit ou se rétrécit en temps réel (minimum 60px) avec persistance dans `localStorage`.<br>• Le double-clic sur une poignée réinitialise toutes les largeurs par défaut. |
| 5 | Consulter les checklists dans **Dossiers CTD**, **Agréments Corporate (Décret 21-82)** et la **Grille de Recevabilité DM (13 points)**. | Inspection des exigences | • Chaque pièce éliminatoire (`requirement_level = 'E'` / obligatoire) affiche un badge discret `.badge-eliminatoire-discret` (`Éliminatoire` ou `[E] Éliminatoire`) avec infobulle explicative au survol, sans encombrer la ligne. |
| 6 | Sur les écrans **Dossiers CTD** et **Agréments Corporate (Décret 21-82)** à `1366×768`, vérifier la barre d'outils centrale (`#main-toolbar`), les menus déroulants (`▾`) et les sous-onglets du volet d'aperçu Corporate. | Clic sur menus déroulants | • Tous les boutons de `#main-toolbar` (CTD : `✍️ Visa PDT`, `🚀 E-Services ▾`, `💰 Redevances ▾`, `⚙️ Dossier ▾` ; Corporate : `📁 Vues ▾`, `📤 Pièces ▾`, `🏛️ Tabadol ▾`, `🔄`, `🔐`, `👁️`) tiennent intégralement dans la largeur visible sans débordement.<br>• Les 4 sous-onglets du volet Corporate (`📋 Checklist`, `📝 e-Form`, `🌐 Tabadol`, `🏛️ MIPH`) restent visibles sans troncature.<br>• Chaque menu déroulant s'ouvre proprement au-dessus du contenu et se ferme au clic extérieur ou sur `Échap`. |
| 7 | Tester l'accès à l'**Audit Trail** (`AUDIT_VIEW`) avec un compte **Pharmacien Assistant** (`pharmacien_assistant`) hors période d'intérim, puis durant un mandat d'intérim DT actif (`dt_delegations`). | Compte `pharmacien_assistant` | • **Hors intérim :** L'accès à `GET /api/v1/audit/logs` est bloqué (**HTTP 403**) et le bouton Audit Trail n'est pas actif.<br>• **Durant un intérim DT actif :** La permission `AUDIT_VIEW` est dynamiquement octroyée pour la durée du mandat (`start_date` à `end_date`) et révoquée automatiquement dès la clôture/révocation du mandat. |
| 8 | Tester la signature GED (`/api/v1/ged/documents/{id}/sign-dt`) par le **DT intérimaire** (`pharmacien_assistant` avec mandat actif) sur un document *Qualité / Production* (`02_Assurance_Qualite`, `03_Developpement_Production`, `04_Controle_Qualite`) vs un document *Affaires Réglementaires* (`01_Affaires_Reglementaires`). | Mandat intérim actif | • **Qualité & Production (`02_`, `03_`, `04_`) :** Signature acceptée (**HTTP 200**) avec mention `[Intérim DT — Mandat ...]` dans la déclaration scellée.<br>• **Hors périmètre (`01_Affaires_Reglementaires`) ou hors intérim :** Bloqué (**HTTP 403**). |
| 9 | Hors période d'intérim : dans la modale **Registre des Mandats & Habilitations** du Parapheur DT, créer une **Délégation Permanente Ciblée** (`NOTE-HAB-DT-2026-01`) au profit d'un `pharmacien_assistant` sur `02_Assurance_Qualite` et `03_Developpement_Production`, puis tester la signature GED depuis l'interface GED. | Habilitation permanente active (hors intérim) | • **Création avec `01_Affaires_Reglementaires` :** Refusée (**HTTP 400**).<br>• **Signature GED sur `02_` / `03_` :** Bouton & modale de signature accessibles en GED, scellement réussi (**HTTP 200**) avec mention `[Par délégation permanente DT — Réf. NOTE-HAB-DT-2026-01]`.<br>• **Après révocation ou sur `01_` :** Bloqué (**HTTP 403**). |
| 10 | Déclencher une confirmation, une saisie de motif et une notification de succès/erreur dans les modules (ex. création/révocation de délégation, changement de rôle Admin, soumission/scellement workflow, veille réglementaire, frais ANPP). | Actions UI diverses | • **Zéro boîte native :** Aucune fenêtre native du navigateur (`alert`, `confirm`, `prompt`) n'apparaît.<br>• Toutes les confirmations et saisies passent par la modale unifiée `#modal-universal-dialog` (`showConfirmModal` / `showPromptModal`) ou `#modal-revoke-delegation`, et les retours utilisent `showToast(...)` ou `showNotifyModal(...)`. |
| 11 | Dans les modules **Dispositifs Médicaux (DDM)** et **Dossiers CTD (DEPP)**, cliquer sur **« 📥 ANPP (.docx) »**, **« 💶 BV (.docx) »**, **« 📊 Canevas Officiel (.xlsx) »**, **« 📥 DEPP02 (.docx) »**, **« 📥 Pré-soumission (.docx) »**, **« 🧾 Structure Prix (.docx) »** et **« 📋 Récépissé PP (.docx) »**. | Fiche DM & Dossier CTD | • Téléchargement effectif de fichiers `.docx` et `.xlsx` basés sur les gabarits officiels téléchargés depuis `anpp.dz` (dont le bordereau de versement homologation DM `300 000 DA`, la déclaration journalière DM `.xlsx` et les 9 formulaires DEPP incluant `Pré-soumission`, `Structure de prix` et `Récépissé PP`, embarqués et pré-remplis côté serveur avec les données de l'établissement, du produit et des cases cochées `☑`).<br>• Archivage automatique en GED (`01_06_Correspondances_ANPP`, statut `BROUILLON`) des bordereaux de versement (`bv-*`) et des formulaires DEPP officiels (hors modèle `Structure Prix`) avec notification toast du code GED. |
| 12 | Dans **Dispositifs Médicaux (DDM)**, cliquer sur **« 📦 Lot DCE »**, saisir un lot importé (`LOT-2026-881`), cocher **« 🤝 Reliance Tests Labo (ANPP) »** (`EDQM` / `OMCL-2026-77`), soumettre et tester les boutons **Autres Formulaires & Canevas Officiels DCE (ANPP)** ; puis dans la barre du **Parapheur DT**, cliquer sur **« 📥 Fiche DVTIV PDT/QPPV (.docx) »** et vérifier la GED (`04.01 Pharmacovigilance`) ainsi que la checklist Exploitation (`EXP.PV.01`). | Lot DM importé, Reliance Labo & PDT/QPPV | • Enregistrement `POST /api/v1/lot-releases` (**HTTP 201**) avec persistance `is_reliance_lab` / `reliance_lab_authority` et téléchargement du bordereau officiel DCE (`ANPP-DCE-Reception-DM-Importes_*.docx`) pré-rempli + archivage GED (`02_02_Certificats_BPF_ISO`, `BROUILLON`), ainsi que des certificats/déclarations DCE (`01_06_Correspondances_ANPP`, `BROUILLON`), tandis que `Composition Dossier` et `CA SR Secondaire` restent des modèles en téléchargement direct.<br>• Téléchargement direct de la fiche officielle DVTIV (`ANPP-DVTIV-Base-PDT-QPPV_*.docx`) pré-remplie (nom, adresse/wilaya, mentions de statut inutiles et options PSMF/PSMF national `Présence`/`Absence` barrées via `<w:strike/>`, titres protégés contre les orphelins via `<w:keepNext/>`), **archivage automatique en GED** au statut `BROUILLON` (*À signer*), **pré-rattachement** automatique à l'item `EXP.PV.01` au statut `rattache` (sans validation automatique avant signature/revue qualité), et notification toast de confirmation. |
| 13 | Dans un **Dossier CTD**, cliquer sur **« 🤝 Reliance »**, activer le parcours Confiance Réglementaire (`EMA`, `enregistrement`, `EMEA/H/C/005432`) et enregistrer. | Dossier CTD (`NI` / `NF`) | • Enregistrement `PUT /api/v1/dossiers/{id}/reliance` (**HTTP 200**), journalisation `DOSSIER_RELIANCE_UPDATED` dans l'Audit Trail et affichage du badge **`🤝 Reliance (EMA)`** dans l'en-tête du dossier. |
| 14 | Ouvrir le menu utilisateur puis **« 🚨 Rappels de Lots 24h (ANPP) »**, déclarer un rappel ou exercice à blanc (`CLASSE_I`), cliquer sur **« 📢 Notifier ANPP (24h) »** puis **« 🔒 Clôturer & PV GED »**. | Rappel `Biomox 1g` / `LOT-RCL-01` | • Création (`POST /api/v1/recalls`, **HTTP 201**), horodatage ANPP (`within_24h_sla: true`) et clôture avec **archivage automatique du PV de retrait/destruction dans la GED (`04_03_Rappels_Lots`)** et traçabilité Audit Trail. |
| 15 | Ouvrir le menu utilisateur puis **« 🔍 Inspections & CAPA (LD-DVIV04) »**, enregistrer un écart `CRITIQUE`, soumettre le **Plan CAPA** puis cliquer sur **« 🔒 Clôturer & GED »**. | Observation `INSP-2026-01` (`CRITIQUE`) | • Création (`POST /api/v1/inspections/findings`, **HTTP 201**), soumission CAPA (`capa_soumis`) et clôture avec **archivage automatique de la fiche CAPA dans la GED (`02_04_CAPA_Deviations`)** et traçabilité Audit Trail. |
| 16 | Passer un produit au statut **`suspendu`** (`PUT /api/v1/regulatory-status/dm/{id}`), tenter une libération de lot DCE sans motif, puis contourner en tant que **PDT** avec un motif justifié. | DM `suspendu` + Lot DCE | • Sans motif ou hors rôle PDT : blocage **HTTP 409 Conflict** (`pdt_override_only: true`).<br>• Avec rôle **PDT** et `pdt_override_reason` saisi via `showPromptModal` : création autorisée (**HTTP 201**) avec persistance du motif de dérogation et trace Audit Trail. |
| 17 | Cliquer sur **`✏️ Saisir Structure Prix`** (dans un dossier PP), **`✏️ Compléter QPPV/PSMF`** (dans le Parapheur DT), **`✏️ Compléter SA / Lots`** ou **`✏️ Saisir CA SR Secondaire`** (dans la modale DCE) -> Renseigner les champs, cliquer sur **Enregistrer brouillon**, fermer puis réouvrir le formulaire, modifier une valeur et cliquer sur **Enregistrer & Générer document**. | 4 formulaires brouillons ANPP | • Les données restent en statut **`brouillon`** (toujours modifiables sans verrouillage), persistent entre deux ouvertures (`GET`/`PUT /api/v1/anpp-drafts/{kind}`) et sont injectées automatiquement dans le `.docx`/`.xlsx` officiel téléchargé. |
| 18 | Dans la modale **Veille Réglementaire**, dans le menu utilisateur ou dans le **Cockpit IT** (`super_admin_editeur` / `admin_systeme_local`), cliquer sur **`📄 Gabarits ANPP (.docx/.xlsx)`**, cliquer sur **`👁️ Visualiser`** pour prévisualiser un modèle `.docx`/`.xlsx`, remplacer un gabarit (portée `Établissement` pour RA/PDT ou `🌐 Global` pour SuperAdmin SaaS) avec motif puis cliquer sur **`↩️ Défaut`**. | 28 gabarits ANPP (`anpp_template_overrides`) | • L'aperçu structuré (paragraphes, tableaux Word ou feuilles Excel) s'affiche en modale (`GET /api/v1/anpp-templates/{name}/preview`) ; le nouveau modèle est stocké en base (`SHA-256` + version + auteur) et utilisé immédiatement **sans recompilation ni redémarrage**, avec journalisation Audit Trail et restauration possible vers `go:embed`. |
| 19 | Ouvrir n'importe quelle modale de l'application (ex. *Veille Réglementaire*, *Gabarits ANPP*, *Brouillon ANPP*, *Parapheur DT*, *Rappels de Lots*, *Confirmation*), puis la déplacer par sa barre d'en-tête et la redimensionner par ses bords/angles. | Toutes les modales (`.modal-overlay > .modal`) | • Chaque modale se déplace librement dans la fenêtre via glisser-déposer sur `.modal-header` et se redimensionne sur les 8 directions (`.modal-resize-edge`) sans sortir du viewport. |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-UI-02-TOOLTIPS.png`

---

#### Test Ref : `OQ-CORP-02`
* **Titre :** Vérification de l'alignement réglementaire Décret exécutif n° 21-82 (5ᵉ catégorie Exportation, parcours Fabrication en 2 étapes, baux notariés généralisés et Veille Juridique 9 Piliers).
* **Exigence Réglementaire :** Décret exécutif n° 21-82 (Art. 2, 12, 14, 16, 24), Arrêtés d'application 2021–2025/2026 (Importation 300 m²/90 m²/390 m², Distribution BPDG, Exploitation, Fabrication 2 étapes) & ALCOA+.
* **Rôle Requis :** Pharmacien Directeur Technique (`pdt@pharma.dz`).

| Étape | Action de l'Auditeur (Navigation UI) | Données d'Entrée | Résultat Attendu (Acceptance Criteria) |
| :---: | :--- | :--- | :--- |
| 1 | Ouvrir le module **Agréments Corporate (Décret 21-82)** et vérifier l'arborescence de gauche. | Clic sur `🏛️ Agréments MIPH` | • Les 5 catégories officielles sont présentes : `Fabrication`, `Importation`, `Distribution`, `Exploitation`, et `🌍 Exportation Pharmaceutique (Art. 12)`. |
| 2 | Cliquer sur `➕ Nouvelle Demande` et inspecter les sélecteurs **Catégorie** et **Nature de la Demande**. | Modale Création | • La catégorie `EXPORTATION` (`CNRC 702103/702106`) est sélectionnable.<br>• Les démarches Fabrication en 2 étapes (`Agrément Préalable de Réalisation` vs `Agrément d'Ouverture Définitive`) ainsi que `Décision d'Exercice PDT` et `Assistant` sont présentes. |
| 3 | Créer une demande `EXPORTATION` (`AGREMENT_INITIAL`) et ouvrir sa checklist. | Création dossier Export | • Le formulaire est `eForm-MIPH-EXPORTATION`.<br>• La checklist génère **12 exigences unitaires bloquantes `[E]`** (1 pièce = 1 document : `EXT.ADM.01` Statuts, `EXT.ADM.02` CNRC 702103/702106, `EXT.ADM.03` Extrait de rôle, `EXT.ADM.04` CNAS/CASNOS, `EXT.ADM.06` Titre de propriété/bail notarié, `EXT.PDT.01` à `03` PDT, etc.). |
| 4 | Ouvrir ou créer un dossier `IMPORTATION`, `DISTRIBUTION`, `EXPLOITATION` ou `FABRICATION (PREALABLE_REALISATION)` et vérifier la granularité des pièces. | Dossiers Corporate | • Chaque document (Statuts, CNRC, Extrait de rôle, CNAS/CASNOS, Bail notarié/Propriété, Diplôme PDT, CNOP, Contrat PDT) dispose de sa propre ligne unitaire (`15` pts Import, `13` pts Distribution, `11` pts Exploitation, `10` pts Préalable Fabrication). |
| 5 | Ouvrir la modale **📜 Veille Juridique**. | Clic `📜 Veille Juridique` | • Le corpus affiche sans doublons les textes officiels des **9 Piliers** (Loi 18-11, Décret 21-82, Arrêtés 2021–2026, Décret 22-247 BPF, Décret 21-551, `LD-DEPP07/V01`, etc.). |

* **Résultat de l'Auditeur :** [ ] PASS  |  [ ] FAIL  |  [ ] DÉVIATION
* **Preuve Documentaire :** `SCR-OQ-CORP-02-EXPORT.png`, `SCR-OQ-CORP-02-VEILLE.png`

---

## 5. GESTION DES ANOMALIES, DÉVIATIONS & FICHES D'INCIDENT

Tout écart constaté lors de l'exécution d'un test doit être consigné dans la fiche de déviation ci-dessous :

### Fiche de Déviation UAT (Modèle à reproduire)
* **ID Déviation :** `DEV-[ID_TEST]-[N° ORDRE]` (ex. `DEV-OQ-AUD-01-01`)
* **Criticité GxP :** 
  - [ ] **Critique** (Blocage d'une exigence réglementaire légale 21 CFR Part 11 / ANPP, altération de données, faille d'authentification).
  - [ ] **Majeure** (Dysfonctionnement d'un workflow métier ou filtre, contournable sans risque pour la santé publique).
  - [ ] **Mineure** (Défaut cosmétique d'alignement UI ou coquille textuelle).
* **Description de l'anomalie :** Décrire précisément le comportement observé vs attendu.
* **Étapes de reproduction :** Liste numérotée exacte.
* **Preuve jointe :** Nom du screenshot ou extrait du journal système.
* **Impact sur la libération :** Bloquant / Non bloquant sous condition.

---

## 6. RAPPORT DE CLÔTURE & PROCÈS-VERBAL DE QUALIFICATION GXP

À l'issue de l'exécution complète du présent cahier de tests :

| Statistique de Qualification | Valeur Observée | Seuil d'Acceptabilité GxP |
| :--- | :---: | :---: |
| **Nombre total de tests exécutés :** | `...` | 100% des cas prévus |
| **Nombre de tests CONFORMES (PASS) :** | `...` | ≥ 98 % |
| **Nombre de déviations mineures acceptées :** | `...` | ≤ 2 (avec plan d'action CAPA) |
| **Nombre de déviations critiques :** | `...` | **0 Tolérance Zéro** |

### Visa et Approbation Formelle de la Recette

| Rôle | Nom & Prénom | Organisme / Société | Décision (Favorable / Réserves) | Signature & Date |
| :--- | :--- | :--- | :---: | :--- |
| **Auditeur Externe (Prestataire UAT)** | .............................. | .............................. | [ ] Qualifié  [ ] Non Qualifié | .............................. |
| **Responsable Assurance Qualité** | Nadir Toumi | Laboratoire Titulaire | [ ] Approuvé GxP | .............................. |
| **Pharmacien Directeur Technique** | Dr. Yacine Benkhalid | Laboratoire Titulaire | [ ] Bon pour Production | .............................. |
