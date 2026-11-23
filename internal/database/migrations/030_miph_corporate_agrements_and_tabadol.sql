-- =========================================================================
-- Migration 030: Module Sanad Corporate — Agréments MIPH & Tabadol
-- Décret exécutif n° 21-82 du 23 février 2021 & Arrêtés sept/oct 2025
-- =========================================================================

-- 1. Table des Demandes d'Agréments Ministériels MIPH
CREATE TABLE IF NOT EXISTS corporate_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    establishment_id UUID REFERENCES establishments(id) ON DELETE SET NULL,
    category VARCHAR(32) NOT NULL, -- FABRICATION, IMPORTATION, DISTRIBUTION, EXPLOITATION
    nature VARCHAR(32) NOT NULL,   -- PREALABLE_REALISATION, OUVERTURE_INITIALE, SOUS_TRAITANCE, EXERCICE_PDT, EXERCICE_ASSISTANT, RENOUVELLEMENT, MODIF_SUBSTANTIELLE, MODIF_NON_SUBSTANTIELLE
    application_number VARCHAR(64) NOT NULL,
    internal_reference VARCHAR(128),
    title VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'brouillon', -- brouillon, en_revue, vise_pdt, depose_tabadol, depose_bo, agrement_octroye, rejete
    form_type VARCHAR(64) NOT NULL,
    form_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    completeness_percent NUMERIC(5,2) DEFAULT 0,
    pdt_signed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    pdt_signed_at TIMESTAMPTZ,
    pdt_signature_seal VARCHAR(128),
    tabadol_submission_ref VARCHAR(128),
    tabadol_submitted_at TIMESTAMPTZ,
    tabadol_submission_notes TEXT,
    miph_receipt_ref VARCHAR(128),
    miph_receipt_date DATE,
    decision_ref VARCHAR(128),
    decision_date DATE,
    license_expiry_date DATE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_corporate_app_num UNIQUE (organization_id, application_number)
);

-- Index pour requêtes fréquentes
CREATE INDEX IF NOT EXISTS idx_corp_app_org_created ON corporate_applications (organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_corp_app_org_status ON corporate_applications (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_corp_app_category ON corporate_applications (organization_id, category);
CREATE INDEX IF NOT EXISTS idx_corp_app_nature ON corporate_applications (organization_id, nature);
CREATE INDEX IF NOT EXISTS idx_corp_app_establishment ON corporate_applications (establishment_id);

-- 2. Table des Pièces de Checklist pour les Agréments MIPH
CREATE TABLE IF NOT EXISTS corporate_checklist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES corporate_applications(id) ON DELETE CASCADE,
    item_code VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    section VARCHAR(64) NOT NULL, -- ADMINISTRATIF, TECHNIQUE_LOCAUX, ORGANISATION_BPF, PHARMACIEN_DT, CONDITIONS_TECHNIQUES, MODIFICATIONS
    requirement_level VARCHAR(8) NOT NULL DEFAULT 'E', -- E, C, F, S
    is_mandatory BOOLEAN NOT NULL DEFAULT true,
    condition_rule TEXT,
    sort_order INT NOT NULL DEFAULT 10,
    ged_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'manquant', -- manquant, rattache, valide_qualite, rejete
    notes TEXT,
    uploaded_at TIMESTAMPTZ,
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_corp_item_code UNIQUE (application_id, item_code)
);

CREATE INDEX IF NOT EXISTS idx_corp_items_app_sort ON corporate_checklist_items (application_id, sort_order ASC, item_code ASC);
CREATE INDEX IF NOT EXISTS idx_corp_items_status ON corporate_checklist_items (application_id, status);
CREATE INDEX IF NOT EXISTS idx_corp_items_ged ON corporate_checklist_items (ged_document_id);
CREATE INDEX IF NOT EXISTS idx_corp_items_completeness ON corporate_checklist_items (application_id, is_mandatory, requirement_level, status);

-- 3. Registre des Titres Réglementaires & Agréments Quinquennaux (MIPH Licenses)
CREATE TABLE IF NOT EXISTS corporate_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    establishment_id UUID REFERENCES establishments(id) ON DELETE SET NULL,
    application_id UUID REFERENCES corporate_applications(id) ON DELETE SET NULL,
    license_type VARCHAR(32) NOT NULL, -- AGREMENT_REALISATION, AGREMENT_OUVERTURE, DECISION_PDT, DECISION_ASSISTANT, AGREMENT_MODIF
    license_number VARCHAR(128) NOT NULL,
    decision_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    expiry_date DATE,
    pharmacist_name VARCHAR(255),
    pharmacist_order_number VARCHAR(64),
    status VARCHAR(32) NOT NULL DEFAULT 'actif', -- actif, a_renouveler_j180, alerte_j90, critique_j30, expire, modifie, abroge
    document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_corp_license_num UNIQUE (organization_id, license_number)
);

CREATE INDEX IF NOT EXISTS idx_corp_lic_org_exp ON corporate_licenses (organization_id, expiry_date ASC);
CREATE INDEX IF NOT EXISTS idx_corp_lic_status ON corporate_licenses (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_corp_lic_establishment ON corporate_licenses (establishment_id);
CREATE INDEX IF NOT EXISTS idx_corp_lic_document ON corporate_licenses (document_id);
