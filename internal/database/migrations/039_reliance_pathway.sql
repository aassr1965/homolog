-- ============================================================================
-- Migration 039 : Parcours Confiance Réglementaire (Reliance & Reconnaissance)
-- Lignes directrices ANPP : Enregistrement/Modification, Tests Labo & BPF
-- ============================================================================

ALTER TABLE dossiers
    ADD COLUMN IF NOT EXISTS is_reliance BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS reliance_authority VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reliance_type VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reliance_reference VARCHAR(128) NOT NULL DEFAULT '';

ALTER TABLE import_lot_releases
    ADD COLUMN IF NOT EXISTS is_reliance_lab BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS reliance_lab_authority VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reliance_lab_reference VARCHAR(128) NOT NULL DEFAULT '';

-- Index d'optimisation pour la veille réglementaire
CREATE INDEX IF NOT EXISTS idx_reg_watch_org_docfile ON regulatory_watch_items(organization_id, document_filename);
