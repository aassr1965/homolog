# Directives Projet & Règles Permanentes pour les Agents — SANAD PHARMA

## 1. Règle Permanente : Maintien et Mise à Jour du Protocole de Qualification (OQ / PQ)
> [!IMPORTANT]
> **Règle absolue de maintien documentaire GxP / CSV :**
> Le document de référence `docs/validation/PROTOCOLE_QUALIFICATION_OQ_PQ_SANAD.md` constitue le **Cahier de Recette Utilisateur et Protocole de Qualification Opérationnelle & de Performance (OQ/PQ)** officiel de l'application, destiné aux prestataires externes de validation et aux auditeurs réglementaires (ANPP / BPF / 21 CFR Part 11).
>
> **Consigne obligatoire pour tout développement futur :**
> À chaque nouveau développement, évolution d'interface graphique (UI), ajout de rôle, modification de formulaire, nouveau connecteur réglementaire ou mise à jour de workflow (Maker-Checker, signature électronique, gestion documentaire, etc.) :
> 1. Vous devez **systématiquement mettre à jour** le fichier `docs/validation/PROTOCOLE_QUALIFICATION_OQ_PQ_SANAD.md`.
> 2. Toute nouvelle fonctionnalité doit comporter sa fiche de test pas-à-pas détaillée (« click-by-click ») avec ses données d'entrée, ses captures d'écran attendues et ses critères de validation ALCOA+.
> 3. Ne jamais livrer une fonctionnalité sans avoir mis à jour ce document de validation.

## 2. Architecture & Environnement Technique
- **Backend** : Go 1.24+ avec `pgx/v5` sur PostgreSQL (port 5434 en local).
- **Frontend** : Vanilla HTML5, CSS3 modulaire et JavaScript ES6+, embarqué au binaire via `//go:embed` dans `web/embed.go`.
- **Règles de compilation** : Toute modification de fichiers sous `web/` nécessite une recompilation de `server.exe` (`go build -o server.exe ./cmd/server`) avant redémarrage du service.
- **Conformité Réglementaire** : 21 CFR Part 11, EU GMP Annexe 11, GAMP 5, exigences ANPP (Algérie). Traçabilité immuable, signature électronique à double authentification, Audit Trail chaîné en SHA-256 avec géolocalisation haute précision (commune/daïra).
