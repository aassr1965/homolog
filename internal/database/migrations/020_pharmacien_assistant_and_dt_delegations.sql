-- =====================================================================================
-- Migration 020 : Rôle Pharmacien Assistant et Délégations / Intérim DT (Décret 21-82)
-- Conforme Décret exécutif n° 21-82 (Articles 20 à 25), BPF Algériennes et 21 CFR Part 11
-- =====================================================================================

-- 1. Ajout du Rôle Réglementaire Pharmacien Assistant / Adjoint DT
INSERT INTO roles (id, name_fr, name_ar, description, is_system) VALUES
    ('pharmacien_assistant', 'Pharmacien Assistant / Adjoint DT', 'الصيدلي المساعد / نائب المدير التقني', 'Pharmacien diplômé inscrit à l''Ordre (Section B) titulaire d''une décision ministérielle nominative MIPH (Décret 21-82). Assiste le Directeur Technique et assure la suppléance légale par intérim.', TRUE)
ON CONFLICT (id) DO UPDATE SET
    name_fr = EXCLUDED.name_fr,
    name_ar = EXCLUDED.name_ar,
    description = EXCLUDED.description;

-- 2. Attribution des permissions de base pour le Pharmacien Assistant
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('pharmacien_assistant', 'REF_VIEW'),
    ('pharmacien_assistant', 'DOSSIER_VIEW'),
    ('pharmacien_assistant', 'DOC_READ'),
    ('pharmacien_assistant', 'DOC_DRAFT_MAKER'),
    ('pharmacien_assistant', 'DOC_CHECK_REVIEW'),
    ('pharmacien_assistant', 'DOC_LOCK_CHECKOUT'),
    ('pharmacien_assistant', 'EXPORT_PACKAGE'),
    ('pharmacien_assistant', 'AUDIT_VIEW')
ON CONFLICT DO NOTHING;

-- 3. Table des Délégations & Mandats d'Intérim du Pharmacien DT (Décret 21-82 Art. 24)
CREATE TABLE IF NOT EXISTS dt_delegations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    delegator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    delegate_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    miph_notification_ref VARCHAR(100) NOT NULL,
    reason TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES users(id),
    revoke_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_delegation_dates CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_dt_delegations_lookup 
ON dt_delegations(organization_id, delegate_id, is_active, start_date, end_date);
