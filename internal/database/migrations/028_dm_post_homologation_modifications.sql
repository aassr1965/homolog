-- ============================================================================
-- Migration 028 : Modifications Post-Homologation des Dispositifs Médicaux
-- Conforme au Décret exécutif n° 20-324 du 22 novembre 2020 (Articles 22 à 25)
-- Modifications Substantielles vs Non-Substantielles, Avenants & Normes ISO
-- ============================================================================

CREATE TABLE IF NOT EXISTS dm_modifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    medical_device_id UUID NOT NULL REFERENCES medical_devices(id) ON DELETE CASCADE,
    dossier_id UUID REFERENCES dossiers(id) ON DELETE SET NULL,

    -- Identification de la modification
    modification_code VARCHAR(64) NOT NULL,
    modification_type VARCHAR(32) NOT NULL, -- 'SUBSTANTIELLE', 'NON_SUBSTANTIELLE'
    category_code VARCHAR(32) NOT NULL,    -- 'DM_SUB_01'..'DM_SUB_07', 'DM_MIN_01'..'DM_MIN_04'
    category_label VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    justification TEXT,

    -- Situation Avant / Après
    current_state TEXT,
    proposed_state TEXT,

    -- Matrice d'Impact Normatif ISO (BPF / Dispositifs Médicaux)
    impact_risk_iso14971 BOOLEAN DEFAULT FALSE,
    impact_biocompatibility_iso10993 BOOLEAN DEFAULT FALSE,
    impact_sterilization_iso11135 BOOLEAN DEFAULT FALSE,
    impact_packaging_iso11607 BOOLEAN DEFAULT FALSE,
    impact_clinical_iso14155 BOOLEAN DEFAULT FALSE,
    impact_quality_iso13485 BOOLEAN DEFAULT FALSE,

    -- Aspects Financiers & Redevances (Ord. 21-07 Art. 31)
    fee_amount NUMERIC(12,2) NOT NULL DEFAULT 45000.00,
    fee_receipt_ref VARCHAR(100),
    fee_payment_date DATE,

    -- Statut du Workflow Réglementaire
    status VARCHAR(32) NOT NULL DEFAULT 'brouillon',
    -- 'brouillon', 'soumis_ddm', 'en_instruction', 'approuve_avenant', 'rejete', 'nouveau_dossier_requis'

    -- Décision & Avenant DDM / ANPP
    anpp_submission_date DATE,
    anpp_decision_date DATE,
    avenant_number VARCHAR(100),
    avenant_date DATE,
    decision_notes TEXT,

    -- Scellement & Visa Pharmacien DT (21 CFR Part 11 / Mandat 1.A.7)
    pdt_signed_at TIMESTAMPTZ,
    pdt_signed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    pdt_signature_seal VARCHAR(64),
    pdt_signature_notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_dm_mod_org_code UNIQUE (organization_id, modification_code)
);

CREATE INDEX IF NOT EXISTS idx_dm_mod_org ON dm_modifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_dm_mod_device ON dm_modifications(medical_device_id);
CREATE INDEX IF NOT EXISTS idx_dm_mod_type ON dm_modifications(modification_type);
CREATE INDEX IF NOT EXISTS idx_dm_mod_status ON dm_modifications(status);
