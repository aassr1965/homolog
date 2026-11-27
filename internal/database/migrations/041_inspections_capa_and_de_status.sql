-- ============================================================================
-- Migration 041 : Inspections Réglementaires (LD-DVIV04/V01), Écarts BPF/CAPA
-- et Statuts Réglementaires DE/DH (Refus, Suspension, Retrait)
-- ============================================================================

CREATE TABLE IF NOT EXISTS inspection_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    inspection_reference VARCHAR(64) NOT NULL,
    inspection_date DATE NOT NULL DEFAULT CURRENT_DATE,
    inspector_authority VARCHAR(128) NOT NULL DEFAULT 'ANPP (DVTIV)',
    severity VARCHAR(16) NOT NULL DEFAULT 'MAJEUR', -- CRITIQUE, MAJEUR, AUTRE
    description TEXT NOT NULL,
    capa_plan TEXT NOT NULL DEFAULT '',
    responsible_name VARCHAR(128) NOT NULL DEFAULT '',
    due_date DATE,
    status VARCHAR(32) NOT NULL DEFAULT 'ouvert', -- ouvert, capa_soumis, cloture
    ged_report_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    closed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inspection_findings_org ON inspection_findings(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_inspection_findings_sev ON inspection_findings(organization_id, severity, status);

-- Statuts réglementaires post-décision DE / DH (actif, suspendu, retire_temporaire, retire_definitif, refuse)
ALTER TABLE medicines
    ADD COLUMN IF NOT EXISTS regulatory_status VARCHAR(32) NOT NULL DEFAULT 'actif',
    ADD COLUMN IF NOT EXISTS status_decision_ref VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status_effective_date DATE,
    ADD COLUMN IF NOT EXISTS status_reason TEXT NOT NULL DEFAULT '';

ALTER TABLE medical_devices
    ADD COLUMN IF NOT EXISTS regulatory_status VARCHAR(32) NOT NULL DEFAULT 'actif',
    ADD COLUMN IF NOT EXISTS status_decision_ref VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status_effective_date DATE,
    ADD COLUMN IF NOT EXISTS status_reason TEXT NOT NULL DEFAULT '';

-- Traçabilité du contournement PDT sur libération de lot d'un produit suspendu/retiré
ALTER TABLE import_lot_releases
    ADD COLUMN IF NOT EXISTS pdt_override_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pdt_override_by UUID REFERENCES users(id) ON DELETE SET NULL;
