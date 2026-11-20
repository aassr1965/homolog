# Protocole de Recette & Qualification de Validation (PVQ)
## Jalon 2.3 : Paramétrage Institutionnel, Redevances & Versioning Temporel ALCOA+

- **Système :** SANAD (Système d'Accompagnement Numérique des Affaires et Dossiers Pharmaceutiques)
- **Version :** 2.3.0
- **Branche Git :** `feat/jalon-2.3-redevances-tenant-settings`
- **Cadre Réglementaire :** 
  - Décret exécutif n° 20-325 & n° 21-82 (Algérie)
  - Cahier des Charges SANAD (Chapitre 6 - Section 6.4, Chapitre 11)
  - Normes d'intégrité des données : **GAMP 5**, **ALCOA+**, **21 CFR Part 11**
- **Date d'exécution :** 20 Novembre 2026
- **Statut Global :** **CONFORME & VALIDÉ (100% de réussite)**

---

## 1. Objectifs de la Qualification

Le Jalon 2.3 répond à l'exigence critique d'isolation stricte multi-tenant et de conformité aux audits des inspecteurs de l'ANPP et commissaires aux comptes :
1. **Découplage strict des coordonnées bancaires et barèmes :** Aucun RIB, compte du Trésor CAS 302.145 ou montant de redevance n'est codé en dur ("hardcoded"). Chaque tenant (laboratoire pharmaceutique / fabricant) dispose de sa propre table de paramètres `tenant_regulatory_settings`.
2. **Versioning temporel & Immutabilité (ALCOA+) :** Les paramètres utilisent un modèle d'intervalles `valid_from` / `valid_to`. Une modification n'écrase jamais l'historique : la version précédente est clôturée et archivée, garantissant qu'un bordereau de versement `ES-FIN-01` généré il y a un an reproduit fidèlement le barème et les coordonnées bancaires de l'époque.
3. **Gouvernance Maker-Checker & Parapheur PDT :** Toute modification initiée par un administrateur passe au statut `PENDING_APPROVAL_DT` avec motif réglementaire obligatoire. Seul le Pharmacien Directeur Technique (PDT) peut approuver et appliquer le changement après déclaration solennelle (art. 21 CFR Part 11).
4. **Alimentation Dynamique des Rapports Officiels :** Le moteur de génération du bordereau `ES-FIN-01` charge dynamiquement les paramètres actifs du tenant.

---

## 2. Matrice de Traçabilité des Exigences (RTM)

| ID Exigence | Description Fonctionnelle & Réglementaire | Composant Logiciel | Cas de Test Associé | Résultat |
| :--- | :--- | :--- | :--- | :---: |
| **REQ-SET-01** | Initialisation automatique des 17 paramètres légaux par tenant (Trésor, BNA, barèmes enregistrement, variations, DM) | `seed_tenant_default_settings(org_id)` / Migration 021 | `TestTenantRegulatorySettingsAndVersioning/01_Default_Settings_Initialization` | **CONFORME** |
| **REQ-SET-02** | Workflow Maker-Checker : proposition par Admin, passage en `PENDING_APPROVAL_DT` avec justification | `ProposeSettingUpdate` / `PUT /api/v1/admin/settings/regulatory/{key}` | `TestTenantRegulatorySettingsAndVersioning/02_Maker_Checker_Proposal` | **CONFORME** |
| **REQ-SET-03** | Approbation solennelle PDT, archivage version précédente (`valid_to = now`), création nouvelle version active | `ApproveAndApplySetting` / `POST /.../approve` | `TestTenantRegulatorySettingsAndVersioning/03_PDT_Approval_Creates_New_Version` | **CONFORME** |
| **REQ-SET-04** | Cloisonnement étanche multi-tenant : la modification chez Tenant A n'impacte pas Tenant B | `tenant_regulatory_settings.organization_id` & RLS | `TestTenantRegulatorySettingsAndVersioning/04_Multi_Tenant_Isolation` | **CONFORME** |
| **REQ-SET-05** | Application directe d'urgence réservée au Pharmacien Directeur Technique avec engagement immédiat | `DirectApplySetting` | `TestTenantRegulatorySettingsAndVersioning/05_Direct_Apply_By_PDT` | **CONFORME** |
| **REQ-SET-06** | Intégration dynamique dans le bordereau officiel `ES-FIN-01` (PDF / HTML certifié ANPP) | `renderESFIN01HTML` / `loadTenantSettings` | `TestTenantRegulatorySettingsAndVersioning/06_Dynamic_ES_FIN_01_Generation` | **CONFORME** |

---

## 3. Résultats Détaillés des Tests d'Intégration & Recette

Exécution de la suite automatisée Go (`go test -v ./internal/api -run TestTenantRegulatorySettingsAndVersioning`) :

```
=== RUN   TestTenantRegulatorySettingsAndVersioning
=== RUN   TestTenantRegulatorySettingsAndVersioning/01_Default_Settings_Initialization
    --- PASS: TestTenantRegulatorySettingsAndVersioning/01_Default_Settings_Initialization (0.64s)
=== RUN   TestTenantRegulatorySettingsAndVersioning/02_Maker_Checker_Proposal
    --- PASS: TestTenantRegulatorySettingsAndVersioning/02_Maker_Checker_Proposal (0.42s)
=== RUN   TestTenantRegulatorySettingsAndVersioning/03_PDT_Approval_Creates_New_Version
    --- PASS: TestTenantRegulatorySettingsAndVersioning/03_PDT_Approval_Creates_New_Version (0.71s)
=== RUN   TestTenantRegulatorySettingsAndVersioning/04_Multi_Tenant_Isolation
    --- PASS: TestTenantRegulatorySettingsAndVersioning/04_Multi_Tenant_Isolation (0.83s)
=== RUN   TestTenantRegulatorySettingsAndVersioning/05_Direct_Apply_By_PDT
    --- PASS: TestTenantRegulatorySettingsAndVersioning/05_Direct_Apply_By_PDT (0.48s)
=== RUN   TestTenantRegulatorySettingsAndVersioning/06_Dynamic_ES_FIN_01_Generation
    --- PASS: TestTenantRegulatorySettingsAndVersioning/06_Dynamic_ES_FIN_01_Generation (0.68s)
--- PASS: TestTenantRegulatorySettingsAndVersioning (3.76s)
PASS
ok  	sanad/internal/api	4.218s
```

### Vérification de non-régression globale :
- `TestReport` (Bordereaux financiers, récépissés de dépôt, certificats d'analyse) : **PASS**
- `TestParapheur` (Signatures électroniques, conformité 21 CFR Part 11, horodatage RFC 3161) : **PASS**

---

## 4. Livrables Logiciels & Fichiers Clés

1. **Base de Données :**
   - Migration : [021_tenant_regulatory_settings_and_fees.sql](file:///C:/dev/homolog/internal/database/migrations/021_tenant_regulatory_settings_and_fees.sql)
   - Enregistrement : [database.go](file:///C:/dev/homolog/internal/database/database.go)
2. **Backend Go :**
   - Modèles de données : [settings_models.go](file:///C:/dev/homolog/internal/domain/settings_models.go)
   - Dépôt (Repository) : [settings_repository.go](file:///C:/dev/homolog/internal/repository/settings_repository.go)
   - Handlers & API REST : [settings_handler.go](file:///C:/dev/homolog/internal/api/settings_handler.go)
   - Génération de rapports dynamiques : [report_handler.go](file:///C:/dev/homolog/internal/api/report_handler.go)
   - Configuration des routes : [main.go](file:///C:/dev/homolog/cmd/server/main.go)
   - Suite de tests de qualification : [settings_test.go](file:///C:/dev/homolog/internal/api/settings_test.go)
3. **Frontend Web & Cockpit d'Administration :**
   - Interface d'administration : [admin.js](file:///C:/dev/homolog/web/js/admin.js)
   - Arborescence de navigation : [app.js](file:///C:/dev/homolog/web/js/app.js)
   - Vue principale : [index.html](file:///C:/dev/homolog/web/index.html)
4. **Documentation & Roadmap :**
   - Roadmap actualisée : [roadmap.md](file:///C:/dev/homolog/docs/roadmap.md)

---

## 5. Conclusion & Avis de Qualification

Le Jalon 2.3 est qualifié avec succès sans réserve. L'ensemble des exigences de découplage, d'isolation multi-tenant, de traçabilité ALCOA+ et de validation par le Pharmacien Directeur Technique est opérationnel et validé par les tests de régression.

La branche `feat/jalon-2.3-redevances-tenant-settings` est prête à être fusionnée dans la branche principale.
