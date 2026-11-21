-- ==============================================================================
-- Migration 024 : Volet 1.A.4 - Contrôle des Bordereaux de Versement, Quittances Bancaires & Attestations de Change Banque d'Algérie
-- Conforme Décret Exécutif 20-325, Décret 21-82, Arrêté du 10 mai 2021 & Guide CTD ANPP 2025
-- ==============================================================================

-- 1. Enrichissement de anpp_payment_slips pour le contrôle exhaustif de la quittance et de l'attestation de change
ALTER TABLE anpp_payment_slips
    ADD COLUMN IF NOT EXISTS currency VARCHAR(8) NOT NULL DEFAULT 'DZD',
    ADD COLUMN IF NOT EXISTS currency_amount NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS bank_name VARCHAR(120),
    ADD COLUMN IF NOT EXISTS payment_mode VARCHAR(64) NOT NULL DEFAULT 'virement', -- 'virement', 'cheque_certifie', 'versement_especes'
    ADD COLUMN IF NOT EXISTS bank_receipt_ref VARCHAR(120),
    ADD COLUMN IF NOT EXISTS bank_receipt_date DATE,
    ADD COLUMN IF NOT EXISTS requires_fx_attestation BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS fx_rate NUMERIC(14, 6),
    ADD COLUMN IF NOT EXISTS fx_rate_date DATE,
    ADD COLUMN IF NOT EXISTS fx_attestation_ref VARCHAR(120),
    ADD COLUMN IF NOT EXISTS fx_attestation_doc_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR(32) NOT NULL DEFAULT 'pending', -- 'pending', 'conforme', 'non_conforme'
    ADD COLUMN IF NOT EXISTS verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verification_notes TEXT,
    ADD COLUMN IF NOT EXISTS amount_concordance_ok BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS sod_check_passed BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Index de performance
CREATE INDEX IF NOT EXISTS idx_payment_slips_verification ON anpp_payment_slips(dossier_id, verification_status);
CREATE INDEX IF NOT EXISTS idx_payment_slips_receipt_doc ON anpp_payment_slips(receipt_document_id);
CREATE INDEX IF NOT EXISTS idx_payment_slips_fx_doc ON anpp_payment_slips(fx_attestation_doc_id);
