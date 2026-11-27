-- ============================================================================
-- Migration 040 : Registre des Rappels de Lots & Retraits PP/DM (Délai 24h)
-- Guides ANPP Rappel de lots & Retrait PP/DM non conformes + Exercice à blanc
-- ============================================================================

CREATE TABLE IF NOT EXISTS lot_recalls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    recall_reference VARCHAR(64) NOT NULL,
    product_kind VARCHAR(8) NOT NULL DEFAULT 'pp', -- pp, dm
    product_name VARCHAR(255) NOT NULL,
    lot_numbers TEXT NOT NULL,
    severity_class VARCHAR(16) NOT NULL DEFAULT 'CLASSE_I', -- CLASSE_I, CLASSE_II, CLASSE_III
    is_mock_drill BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT NOT NULL DEFAULT '',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notified_anpp_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'ouvert', -- ouvert, notifie_anpp, cloture
    ged_pv_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_by_pdt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lot_recalls_org ON lot_recalls(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lot_recalls_status ON lot_recalls(organization_id, status);
