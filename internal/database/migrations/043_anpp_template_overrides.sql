-- Migration 043 : Surcharge dynamique en base des gabarits officiels ANPP (.docx / .xlsx)
-- Permet de remplacer un template sans recompiler le binaire tout en conservant le go:embed comme repli.

CREATE TABLE IF NOT EXISTS anpp_template_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    template_name VARCHAR(128) NOT NULL,
    content_bytes BYTEA NOT NULL,
    sha256_hash VARCHAR(64) NOT NULL,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    uploaded_by_name VARCHAR(255) NOT NULL DEFAULT '',
    version INT NOT NULL DEFAULT 1,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_anpp_template_org
    ON anpp_template_overrides (organization_id, template_name)
    WHERE organization_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_anpp_template_global
    ON anpp_template_overrides (template_name)
    WHERE organization_id IS NULL;
