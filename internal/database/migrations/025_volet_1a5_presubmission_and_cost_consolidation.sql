-- ==============================================================================
-- Migration 025 : Volet 1.A.5 - Pré-soumission ANPP & Consolidation des Coûts
-- Conforme Guide CTD ANPP Juillet 2025, Décret Exécutif 20-325 & Décret 21-82
-- ==============================================================================

-- 1. Table des Décisions & Frais de Pré-soumission ANPP (Section 1.A.5 du CTD)
CREATE TABLE IF NOT EXISTS anpp_pre_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    anpp_reference VARCHAR(120) NOT NULL, -- N° accord/réponse officielle ANPP
    anpp_notification_date DATE NOT NULL, -- Date notification favorable
    decision_type VARCHAR(64) NOT NULL DEFAULT 'favorable', -- 'favorable', 'reserve', 'defavorable'
    fee_amount_dzd NUMERIC(14, 2) NOT NULL DEFAULT 50000.00, -- Montant redevance étude préalable
    slip_reference VARCHAR(120) NOT NULL, -- N° quittance ou avis de règlement
    payment_date DATE NOT NULL, -- Date effective du règlement
    response_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL, -- Scan réponse officielle ANPP
    payment_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL, -- Scan justificatif de paiement
    verification_status VARCHAR(32) NOT NULL DEFAULT 'conforme', -- 'conforme', 'non_conforme', 'pending'
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verification_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_presubmission_dossier UNIQUE (dossier_id)
);

CREATE INDEX IF NOT EXISTS idx_pre_submissions_org ON anpp_pre_submissions(organization_id);
CREATE INDEX IF NOT EXISTS idx_pre_submissions_dossier ON anpp_pre_submissions(dossier_id);
CREATE INDEX IF NOT EXISTS idx_pre_submissions_status ON anpp_pre_submissions(verification_status);
