-- ============================================================================
-- Migration 042 : Brouillons éditables des formulaires complémentaires ANPP
-- (Structure de prix, CA SR Secondaire, QPPV/PSMF DVTIV, Complément SA DCE)
-- ============================================================================

CREATE TABLE IF NOT EXISTS anpp_form_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    form_kind VARCHAR(64) NOT NULL,
    entity_id UUID NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'brouillon',
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_anpp_form_draft UNIQUE (organization_id, form_kind, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_anpp_form_drafts_org ON anpp_form_drafts(organization_id, form_kind);
