-- ============================================================================
-- Migration 012 : Gestion des Rendez-vous Dépôt ANPP & Veille Réglementaire
-- Jalon 1.7 Phase 4 : Planification des séances ANPP et corpus juridique JORADP
-- ============================================================================

-- 1. Table des Rendez-vous et Séances de Dépôt ANPP
CREATE TABLE IF NOT EXISTS anpp_deposit_appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID REFERENCES dossiers(id) ON DELETE SET NULL,
    appointment_type VARCHAR(64) NOT NULL DEFAULT 'depot_initial', -- depot_initial, seance_commission, expertise_lncm, recours_audition, depot_complements
    scheduled_at TIMESTAMPTZ NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT 'ANPP - Direction de l''Enregistrement, Annexe Dely Ibrahim, Alger',
    slot_reference VARCHAR(100) NOT NULL, -- ex: RDV-ANPP-2026-N°0412
    contact_officer VARCHAR(255), -- ex: Bureau de Réception et de Conformité Documentaire
    assigned_delegate_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'planifie', -- planifie, confirme_anpp, effectue, reporte, annule
    notes TEXT,
    confirmation_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_anpp_appointments_org ON anpp_deposit_appointments(organization_id, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_anpp_appointments_dossier ON anpp_deposit_appointments(dossier_id);
CREATE INDEX IF NOT EXISTS idx_anpp_appointments_status ON anpp_deposit_appointments(status);

-- 2. Table de la Veille Réglementaire et Corpus Juridique Pharmaceutique (JORADP / ANPP)
CREATE TABLE IF NOT EXISTS regulatory_watch_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title_fr VARCHAR(255) NOT NULL,
    title_ar VARCHAR(255),
    legal_reference VARCHAR(255) NOT NULL, -- ex: Décret Exécutif N° 20-325, Arrêté du 10 mai 2021
    joradp_number VARCHAR(100), -- ex: JORADP N° 78 du 28/11/2020
    published_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    category VARCHAR(64) NOT NULL DEFAULT 'enregistrement_amm', -- enregistrement_amm, dispositifs_medicaux, redevances_tarifs, bpf_inspection, pharmacovigilance, essais_cliniques_be
    impact_level VARCHAR(32) NOT NULL DEFAULT 'majeur', -- critique, majeur, mineur
    summary_fr TEXT NOT NULL,
    impact_analysis TEXT,
    compliance_status VARCHAR(32) NOT NULL DEFAULT 'conforme', -- en_vigueur, en_cours_analyse, plan_action_requis, conforme
    source_url VARCHAR(255),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_regulatory_watch_org ON regulatory_watch_items(organization_id, category);
CREATE INDEX IF NOT EXISTS idx_regulatory_watch_impact ON regulatory_watch_items(impact_level);
CREATE INDEX IF NOT EXISTS idx_regulatory_watch_effective ON regulatory_watch_items(effective_date DESC);
