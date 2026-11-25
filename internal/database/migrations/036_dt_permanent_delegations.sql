-- =====================================================================================
-- Migration 036 : Délégations Permanentes Ciblées (Hors Intérim — Qualité & Production)
-- Matrice d'habilitation BPF signée par le DT (21 CFR Part 11 / Annexe 11 / Décret 21-82)
-- =====================================================================================

CREATE TABLE IF NOT EXISTS dt_permanent_delegations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    delegator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    delegate_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    habilitation_ref VARCHAR(120) NOT NULL,
    allowed_taxonomy_prefixes TEXT[] NOT NULL DEFAULT ARRAY['02_', '03_', '04_'],
    notes TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES users(id),
    revoke_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dt_perm_delegations_lookup
ON dt_permanent_delegations(organization_id, delegate_id, is_active);
