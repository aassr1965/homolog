-- ==============================================================================
-- Migration 026 : Moteur de Variations ANPP (Ligne Directrice LD-DEPP07/V01)
-- Référentiel des 34 Rubriques, Escalade Dynamique & Grouping (Arrêté 07/11/2021)
-- ==============================================================================

-- 1. Table des Demandes de Variations / Modifications Post-AMM
CREATE TABLE IF NOT EXISTS anpp_variation_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    medicine_id UUID NOT NULL REFERENCES medicines(id) ON DELETE CASCADE,
    dossier_id UUID REFERENCES dossiers(id) ON DELETE SET NULL, -- Dossier CTD MI/MF optionnel lié
    tracking_number VARCHAR(80) NOT NULL, -- Ex: VAR-2026-AMOX-01
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Procédure & Grouping (Art. 09 Arrêté 07/11/2021)
    is_grouping BOOLEAN NOT NULL DEFAULT FALSE,
    grouping_justification TEXT,
    
    -- Classification globale calculée (Mmin, Mmod, Mmaj)
    overall_classification VARCHAR(16) NOT NULL DEFAULT 'Mmin',
    nominal_classification VARCHAR(16) NOT NULL DEFAULT 'Mmin',
    has_escalation BOOLEAN NOT NULL DEFAULT FALSE,
    escalation_summary TEXT,

    -- Portée du Changement (Section 5.3 LD-DEPP07/V01)
    current_approved_situation TEXT NOT NULL, -- État actuel approuvé
    proposed_situation TEXT NOT NULL,         -- Situation nouvelle proposée
    technical_justification TEXT NOT NULL,    -- Justificatif technico-réglementaire
    
    -- Attestation formelle de non-modification des autres sections
    non_modification_attested BOOLEAN NOT NULL DEFAULT FALSE,
    non_modification_details TEXT,

    -- Redevance Réglementaire (Décret 21-82)
    fee_amount_dzd NUMERIC(14, 2) NOT NULL DEFAULT 0.00,
    fee_slip_id UUID REFERENCES anpp_payment_slips(id) ON DELETE SET NULL,

    -- Statut du cycle de vie BPF
    status VARCHAR(32) NOT NULL DEFAULT 'brouillon', -- 'brouillon', 'en_evaluation', 'reclassifie', 'pret_visa_dt', 'soumis_anpp', 'approuve', 'rejete'

    -- Scellement légal par le Pharmacien DT (21 CFR Part 11)
    dt_verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    dt_verified_at TIMESTAMPTZ,
    dt_signature_seal VARCHAR(128),
    dt_verification_notes TEXT,

    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_variation_tracking UNIQUE (organization_id, tracking_number)
);

CREATE INDEX IF NOT EXISTS idx_variations_org ON anpp_variation_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_variations_medicine ON anpp_variation_requests(medicine_id);
CREATE INDEX IF NOT EXISTS idx_variations_status ON anpp_variation_requests(status);
CREATE INDEX IF NOT EXISTS idx_variations_class ON anpp_variation_requests(overall_classification);

-- 2. Table des Rubriques Sélectionnées dans une Demande (1 à n rubriques par demande)
CREATE TABLE IF NOT EXISTS anpp_variation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variation_request_id UUID NOT NULL REFERENCES anpp_variation_requests(id) ON DELETE CASCADE,
    rubrique_code VARCHAR(32) NOT NULL, -- Ex: '1', '7.d.1', '19.a', '30.b'
    rubrique_title VARCHAR(255) NOT NULL,
    category VARCHAR(64) NOT NULL, -- 'ADMINISTRATIVE', 'CEP_API', 'SUBSTANCE_ACTIVE_32S', 'PRODUIT_FINI_32P'
    
    -- Classification nominale vs effective après évaluation des conditions
    nominal_classification VARCHAR(16) NOT NULL, -- 'Mmin', 'Mmod', 'Mmaj'
    effective_classification VARCHAR(16) NOT NULL, -- 'Mmin', 'Mmod', 'Mmaj'
    is_escalated BOOLEAN NOT NULL DEFAULT FALSE,
    escalation_reason TEXT,

    -- Évaluation des conditions préalables C1..Cn (JSON map: condition_id -> boolean)
    conditions_evaluated JSONB NOT NULL DEFAULT '{}'::jsonb,
    all_conditions_met BOOLEAN NOT NULL DEFAULT TRUE,

    -- Détails spécifiques à cette rubrique
    current_state TEXT,
    proposed_state TEXT,
    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_variation_items_req ON anpp_variation_items(variation_request_id);
CREATE INDEX IF NOT EXISTS idx_variation_items_code ON anpp_variation_items(rubrique_code);

-- 3. Table des 12 Pièces Documentaires Requises (Section 5.3 LD-DEPP07/V01)
CREATE TABLE IF NOT EXISTS anpp_variation_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variation_request_id UUID NOT NULL REFERENCES anpp_variation_requests(id) ON DELETE CASCADE,
    doc_index INT NOT NULL, -- 1 à 12 selon la liste obligatoire de la LD-DEPP07
    doc_code VARCHAR(32) NOT NULL, -- Ex: 'DOC-01-LETTRE-DEMANDE', 'DOC-02-TABLEAU-COMPARATIF'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    is_attached BOOLEAN NOT NULL DEFAULT FALSE,
    ged_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    verified_status VARCHAR(32) NOT NULL DEFAULT 'en_attente', -- 'en_attente', 'conforme', 'non_conforme'
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_variation_doc UNIQUE (variation_request_id, doc_index)
);

CREATE INDEX IF NOT EXISTS idx_variation_docs_req ON anpp_variation_documents(variation_request_id);
