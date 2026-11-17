-- ============================================================================
-- Migration 014 : Profil 10 RBAC (Partenaire Fournisseur / API Holder) & Harmonisation Sanad
-- Conforme Section 7.1 & 7.2 du Cahier des Charges (v1.0 Novembre 2026)
-- Ségrégation ASMF/DMF Substance Active (3.2.S) et Dépôt Chiffré Partenaire
-- ============================================================================

-- 1. Insertion du 10e Profil Utilisateur Strict
INSERT INTO roles (id, name_fr, name_ar, description, is_system) VALUES
    ('partenaire_fournisseur_api', 'Partenaire Fournisseur / API Holder', 'شريك مورد المادة الفعالة / صاحب ملف الدواء الرئيسي', 'Détenteur de DMF/ASMF Substance Active déposant les données confidentielles de synthèse (Module 3.2.S chiffré)', TRUE)
ON CONFLICT (id) DO UPDATE SET
    name_fr = EXCLUDED.name_fr,
    name_ar = EXCLUDED.name_ar,
    description = EXCLUDED.description;

-- 2. Matrice des Permissions Minimales pour le Partenaire Fournisseur / API Holder
-- Le partenaire peut voir le dossier assigné, lire les pièces associées et déposer ses projets (Maker) sur 3.2.S
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('partenaire_fournisseur_api', 'DOSSIER_VIEW'),
    ('partenaire_fournisseur_api', 'DOC_READ'),
    ('partenaire_fournisseur_api', 'DOC_DRAFT_MAKER')
ON CONFLICT (role_id, permission_id) DO NOTHING;
